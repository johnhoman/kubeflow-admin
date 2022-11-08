package awss3bucket

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/johnhoman/kubeflow-admin/internal/types/awsiam/ack"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/johnhoman/kubeflow-admin/apis/aws/v1alpha1"
	"github.com/johnhoman/kubeflow-admin/internal/features"
	"github.com/johnhoman/kubeflow-admin/internal/types/awss3"
	"github.com/johnhoman/kubeflow-admin/internal/watch"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {
	if !o.Features.Enabled(features.AWSS3Bucket) {
		return nil
	}

	name := fmt.Sprintf("%s/awss3bucket", v1alpha1.Group)

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		Owns(awss3.NewUnstructuredBucket()).
		Owns(ack.NewUnstructuredPolicy()).
		Watches(
			&source.Kind{Type: &corev1.ServiceAccount{}},
			watch.EnqueueNamespaceForObject(),
			builder.WithPredicates(
				watch.IsOwnedByProfilePredicate,
				hasAssociatedRoleId,
			),
		).
		Watches(
			&source.Kind{Type: &v1alpha1.BucketConfig{}},
			watch.EnqueueNamespaceForObject(),
		).
		Complete(NewReconciler(mgr,
			WithLogger(o.Logger.WithValues("controller", name)),
			WithEventRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		))
}

type ReconcilerOption func(r *Reconciler)

func WithLogger(l logging.Logger) ReconcilerOption {
	return func(r *Reconciler) {
		r.logger = l
	}
}

func WithEventRecorder(er event.Recorder) ReconcilerOption {
	return func(r *Reconciler) {
		r.record = er
	}
}

func NewReconciler(mgr ctrl.Manager, opts ...ReconcilerOption) *Reconciler {
	r := &Reconciler{
		client: mgr.GetClient(),
		logger: logging.NewNopLogger(),
		record: event.NewNopRecorder(),
	}
	for _, f := range opts {
		f(r)
	}
	return r
}

type Reconciler struct {
	client client.Client
	logger logging.Logger
	record event.Recorder
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	namespace := &corev1.Namespace{}
	if err := r.client.Get(ctx, req.NamespacedName, namespace); err != nil {
		return ctrl.Result{}, errors.Wrap(client.IgnoreNotFound(err), "could not read namespace")
	}

	bc := &v1alpha1.BucketConfig{}
	bc.SetName("default")
	bc.SetNamespace(namespace.Name)
	if err := r.client.Get(ctx, client.ObjectKeyFromObject(bc), bc); err != nil {
		return ctrl.Result{}, errors.Wrap(client.IgnoreNotFound(err), "could not read bucket config")
	}

	bucketName := fmt.Sprintf("%s-%s-%s",
		strings.TrimPrefix(*bc.Spec.Name.Prefix, "-"),
		namespace.Name,
		bc.Name,
	)

	serviceAccountList := &corev1.ServiceAccountList{}
	if err := r.client.List(ctx, serviceAccountList, client.InNamespace(namespace.Name)); err != nil {
		return ctrl.Result{}, err
	}
	sort.Slice(serviceAccountList.Items, func(i, j int) bool {
		return serviceAccountList.Items[i].Name < serviceAccountList.Items[j].Name
	})

	serviceAccounts := make([]*corev1.ServiceAccount, 0)
	ids := sets.NewString()
	for _, item := range serviceAccountList.Items {
		it := item.DeepCopy()
		if roleId, ok := getRoleId(&item); ok {
			ids.Insert(roleId + ":*")
			serviceAccounts = append(serviceAccounts, it)
		}
	}
	doc := map[string]any{
		"Version": "2012-10-17",
		"Statement": []any{
			map[string]any{
				"Effect":    "Deny",
				"Principal": "*",
				"Action":    "s3:*",
				"Resource": []any{
					fmt.Sprintf("arn:aws:s3:::%s", bucketName),
					fmt.Sprintf("arn:aws:s3:::%s/*", bucketName),
				},
				"Condition": map[string]any{
					"StringNotLike": map[string]any{
						"aws:userId": ids.List(),
					},
				},
			},
		},
	}

	pd, err := json.Marshal(doc)
	if err != nil {
		return ctrl.Result{}, err
	}

	ub := awss3.NewUnstructuredBucket()
	ub.SetName(bc.Name)
	ub.SetNamespace(namespace.Name)
	res, err := controllerutil.CreateOrPatch(ctx, r.client, ub, func() error {
		bucket, err := awss3.NewBucketFromUnstructured(ub)
		if err != nil {
			return err
		}
		// names need to be global unique
		if err := bucket.SetName(bucketName); err != nil {
			return err
		}

		if err := bucket.SetRegion(bc.Spec.Region); err != nil {
			return err
		}

		if err := bucket.SetPolicy(string(pd)); err != nil {
			return err
		}

		ub.SetUnstructuredContent(bucket.UnstructuredContent())
		return nil
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	switch res {
	case controllerutil.OperationResultUpdated:
	case controllerutil.OperationResultCreated:
	}

	return ctrl.Result{}, nil
}

const (
	annotationIamRoleId  = "eks.amazonaws.com/role-id"
	annotationIamRoleArn = "eks.amazonaws.com/role-arn"
)

var hasAssociatedRoleId = predicate.NewPredicateFuncs(func(obj client.Object) bool {
	annotations := obj.GetAnnotations()
	if annotations != nil {
		_, ok := annotations[annotationIamRoleId]
		return ok
	}
	return false
})

var hasAssociatedRoleArn = predicate.NewPredicateFuncs(func(obj client.Object) bool {
	annotations := obj.GetAnnotations()
	if annotations != nil {
		_, ok := annotations[annotationIamRoleArn]
		return ok
	}
	return false
})

func getRoleId(obj client.Object) (string, bool) {
	annotations := obj.GetAnnotations()
	if annotations != nil {
		arn, ok := annotations[annotationIamRoleId]
		return arn, ok
	}
	return "", false

}

func getRoleArn(obj client.Object) (string, bool) {
	annotations := obj.GetAnnotations()
	if annotations != nil {
		arn, ok := annotations[annotationIamRoleArn]
		return arn, ok
	}
	return "", false

}

func addPolicyAnnotation(obj client.Object, arn string) {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	key := fmt.Sprintf("policy.aws.admin.kubeflow.org/%s", obj.GetName())
	annotations[key] = arn
}

func hasPolicyAnnotation(obj client.Object, arn string) bool {
	annotations := obj.GetAnnotations()
	if annotations != nil {
		key := fmt.Sprintf("policy.aws.admin.kubeflow.org/%s", obj.GetName())
		return annotations[key] == arn
	}
	return false
}
