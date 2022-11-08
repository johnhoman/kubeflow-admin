package awss3bucketpolicy

import (
	"context"
	"crypto/sha256"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/johnhoman/kubeflow-admin/internal/types/awsiam/ack"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/johnhoman/kubeflow-admin/apis/aws/v1alpha1"
	"github.com/johnhoman/kubeflow-admin/internal/features"
	"github.com/johnhoman/kubeflow-admin/internal/types/awss3"
	"github.com/johnhoman/kubeflow-admin/internal/watch"
)

const (
	errBucketNotReady = "bucket is missing ARN"
	errPolicyNotReady = "policy is missing ARN"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {
	if !o.Features.Enabled(features.AWSS3Bucket) {
		return nil
	}

	name := fmt.Sprintf("%s/awss3bucketpolicy", v1alpha1.Group)

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ServiceAccount{}, builder.WithPredicates(
			watch.IsOwnedByProfilePredicate,
			hasAssociatedRoleArn,
		)).
		Owns(ack.NewUnstructuredPolicy()).
		Watches(&source.Kind{Type: awss3.NewUnstructuredBucket()},
			EnqueueRequestsForServiceAccounts(mgr.GetClient()),
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

type manager interface {
	GetClient() client.Client
}

func NewReconciler(mgr manager, opts ...ReconcilerOption) *Reconciler {
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

	serviceAccount := &corev1.ServiceAccount{}
	if err := r.client.Get(ctx, req.NamespacedName, serviceAccount); err != nil {
		return ctrl.Result{}, errors.Wrap(client.IgnoreNotFound(err), "could not read service account")
	}

	bucketList := &unstructured.UnstructuredList{}
	bucketList.SetGroupVersionKind(awss3.GroupVersionKind)
	// TODO: List only ones created by this controller?
	if err := r.client.List(ctx, bucketList, client.InNamespace(serviceAccount.Namespace)); err != nil {
		return ctrl.Result{}, err
	}

	errs := make([]error, 0)
	for _, item := range bucketList.Items {

		bucket, err := awss3.NewBucketFromUnstructured(&item)
		if err != nil {
			errs = append(errs, err)
		}

		bucketArn, err := bucket.GetArn()
		if err != nil {
			return ctrl.Result{}, err
		}

		if bucketArn == "" {
			errs = append(errs, errors.New(errBucketNotReady))
			continue
		}

		roleArn, ok := getRoleArn(serviceAccount)
		if !ok {
			continue
		}
		doc := map[string]any{
			"Version": "2012-10-17",
			"Statement": []any{
				map[string]any{
					"Effect": "Allow",
					"Principal": map[string]any{
						"AWS": roleArn,
					},
					"Action": []any{
						"s3:ListBucket",
					},
					"Resource": []any{bucketArn},
				},
				map[string]any{
					"Effect": "Allow",
					"Principal": map[string]any{
						"AWS": roleArn,
					},
					"Action": []any{
						"s3:GetObject",
						"s3:PutObject",
						"s3:DeleteObject",
					},
					"Resource": []any{bucketArn + "/*"},
				},
			},
		}

		document, err := json.Marshal(doc)
		if err != nil {
			return ctrl.Result{}, err
		}

		sum := sha256.Sum256([]byte(roleArn))
		encoded := base32.StdEncoding.EncodeToString(sum[:])

		up := ack.NewUnstructuredPolicy()
		up.SetName(item.GetName() + "-" + strings.ToLower(encoded[:6]))
		up.SetNamespace(serviceAccount.Namespace)
		res, err := controllerutil.CreateOrPatch(ctx, r.client, up, func() error {
			policy, err := ack.NewPolicyFromUnstructured(up)
			if err != nil {
				return err
			}
			bucketName, err := bucket.GetName()
			if err != nil {
				return err
			}
			desc := fmt.Sprintf("Access to bucket %s for iam role %s", bucketName, roleArn)
			if err := policy.SetName(up.GetName()); err != nil {
				return err
			}
			if err := policy.SetDescription(desc); err != nil {
				return err
			}
			if err := policy.SetPolicyDocument(string(document)); err != nil {
				return err
			}

			up.SetUnstructuredContent(policy.UnstructuredContent())
			return nil
		})
		if err != nil {
			return ctrl.Result{}, err
		}
		switch res {
		case controllerutil.OperationResultUpdated:
		case controllerutil.OperationResultCreated:
		}

		policy, err := ack.NewPolicyFromUnstructured(up)
		if err != nil {
			return ctrl.Result{}, err
		}
		arn, err := policy.GetArn()
		if err != nil {
			return ctrl.Result{}, err
		}
		if arn == "" {
			return ctrl.Result{}, errors.New(errPolicyNotReady)
		}

		sa := item.DeepCopy()
		if !hasPolicyAnnotation(sa, arn) {
			patch := client.MergeFrom(sa.DeepCopy())
			addPolicyAnnotation(sa, arn)
			if err := r.client.Patch(ctx, sa, patch); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
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
