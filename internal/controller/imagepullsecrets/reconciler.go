package imagepullsecrets

import (
	"context"
	"fmt"
	"sort"

	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/johnhoman/kubeflow-admin/apis/v1alpha1"
	"github.com/johnhoman/kubeflow-admin/internal/types/profile"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	errListClusterSecret = "failed to list cluster secrets"
	errImagePullSecret   = "failed to update service account with image pull secrets"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := fmt.Sprintf("%s/service-account/image-pull-secrets", v1alpha1.Group)

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		Watches(
			&source.Kind{Type: &corev1.ServiceAccount{}},
			NewEnqueueRequestsForNamespaces(mgr.GetClient()),
		).
		Watches(
			&source.Kind{Type: &corev1.Secret{}},
			NewEnqueueRequestsForNamespaces(mgr.GetClient()),
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

	// check if namespace matches the selectors on the cluster secret
	secretList := &corev1.SecretList{}
	// find all secrets with type imagePullSecret
	if err := r.client.List(ctx, secretList, client.InNamespace(namespace.Name)); err != nil {
		return ctrl.Result{}, errors.Wrap(err, errListClusterSecret)
	}

	desired := sets.NewString()
	for _, secret := range secretList.Items {
		owner := metav1.GetControllerOf(&secret)
		if owner != nil {
			if owner.Kind == v1alpha1.ClusterSecretKind && owner.APIVersion == v1alpha1.SchemaGroupVersion.String() {
				// owned by a ClusterSecretType
				switch secret.Type {
				case corev1.DockerConfigJsonKey, corev1.DockerConfigKey:
					desired.Insert(secret.Name)
				}
			}
		}
	}

	serviceAccountList := &corev1.ServiceAccountList{}
	if err := r.client.List(ctx, serviceAccountList, client.InNamespace(namespace.Name)); err != nil {
		return ctrl.Result{}, err
	}

	for _, sa := range serviceAccountList.Items {
		sa := sa.DeepCopy()
		owner := metav1.GetControllerOf(sa)
		if owner == nil {
			continue
		}
		gv, err := schema.ParseGroupVersion(owner.APIVersion)
		if err != nil {
			continue
		}
		if gv.WithKind(owner.Kind).GroupKind() == profile.GroupKind {
			observed := sets.NewString()
			for _, item := range sa.ImagePullSecrets {
				observed.Insert(item.Name)
			}
			ips := imagePullSecretList{}
			for _, name := range observed.Union(desired).List() {
				ips.Append(name)
			}
			patch := client.MergeFrom(sa.DeepCopy())
			sa.ImagePullSecrets = ips.List()
			if err := r.client.Patch(ctx, sa, patch); err != nil {
				return ctrl.Result{}, errors.Wrap(err, errImagePullSecret)
			}
		}
	}

	return ctrl.Result{}, nil
}

type imagePullSecretList []corev1.LocalObjectReference

func (ips *imagePullSecretList) Sort() {
	sort.Slice(*ips, func(i, j int) bool { return (*ips)[i].Name < (*ips)[j].Name })
}

func (ips *imagePullSecretList) List() []corev1.LocalObjectReference {
	ips.Sort()
	return *ips
}

func (ips *imagePullSecretList) Append(name string) {
	*ips = append(*ips, corev1.LocalObjectReference{Name: name})
}
