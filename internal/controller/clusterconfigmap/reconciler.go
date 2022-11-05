package clusterconfigmap

import (
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/johnhoman/kubeflow-admin/apis/v1alpha1"
	"github.com/johnhoman/kubeflow-admin/internal/types/profile"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	errListNamespaces       = "failed to list cluster namespaces"
	errReadReferencedSecret = "failed to read referenced secret"
	errReadNamespaceOwner   = "failed to read namespace owner"
	errApplySecret          = "failed to apply secret"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := "kubeflow-ext/service-account"

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.ClusterSecret{}).
		Owns(&corev1.Secret{}).
		Watches(
			&source.Kind{Type: &corev1.Namespace{}},
			NewEnqueueRequestsForClusterConfigMaps(mgr.GetClient()),
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

	clusterSecret := &v1alpha1.ClusterSecret{}
	if err := r.client.Get(ctx, req.NamespacedName, clusterSecret); err != nil {
		return ctrl.Result{}, errors.Wrap(client.IgnoreNotFound(err), "could not read namespace")
	}

	nsSelector := labels.Everything()
	if clusterSecret.Spec.Selector.Namespace != nil {
		// If the spec doesn't specify a selector, then choose every
		// namespace by default
		var err error
		nsSelector, err = metav1.LabelSelectorAsSelector(clusterSecret.Spec.Selector.Namespace)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	namespaceList := &corev1.NamespaceList{}
	if err := r.client.List(ctx, namespaceList, client.MatchingLabelsSelector{Selector: nsSelector}); err != nil {
		return ctrl.Result{}, errors.Wrap(err, errListNamespaces)
	}

	namespaces := sets.NewString()
	for _, item := range namespaceList.Items {
		if item.Labels == nil || item.Labels["app.kubernetes.io/part-of"] != "kubeflow-profile" {
			continue
		}
		selector := clusterSecret.Spec.Selector.Subject
		if selector == nil {
			namespaces.Insert(item.Name)
		} else {
			// check if the profile is eligible for this cluster secret
			u := profile.NewUnstructured()
			owner := metav1.GetControllerOf(&item)
			if owner == nil {
				continue
			}
			u.SetName(owner.Name)
			key := client.ObjectKeyFromObject(u)
			if err := r.client.Get(ctx, key, u); err != nil {
				r.logger.Debug(errReadNamespaceOwner, "error", err.Error())
				continue
			}
			pr, err := profile.NewFromUnstructured(u)
			if err != nil {
				// TODO: record error
				r.logger.Debug(err.Error())
				continue
			}
			sub, err := pr.GetOwner()
			if err != nil {
				r.logger.Debug(err.Error())
				continue
			}
			if selector.Matches(sub) {
				namespaces.Insert(item.Name)
			}
		}
	}

	ref := &corev1.Secret{}
	ref.SetName(clusterSecret.Spec.SecretRef.Name)
	ref.SetNamespace(clusterSecret.Spec.SecretRef.Namespace)
	if err := r.client.Get(ctx, client.ObjectKeyFromObject(ref), ref); err != nil {
		return ctrl.Result{}, errors.Wrap(err, errReadReferencedSecret)
	}

	controllerRef := metav1.NewControllerRef(clusterSecret,
		v1alpha1.SchemaGroupVersion.WithKind(v1alpha1.ClusterSecretKind),
	)

	// Create all secrets that should exist in selected namespaces
	for namespace := range namespaces {
		secret := &corev1.Secret{}
		secret.SetName(clusterSecret.Name)
		secret.SetNamespace(namespace)

		res, err := controllerutil.CreateOrPatch(ctx, r.client, secret, func() error {
			secret.SetOwnerReferences([]metav1.OwnerReference{*controllerRef})
			secret.Type = ref.Type
			secret.Data = ref.Data
			secret.StringData = ref.StringData
			// Skip immutable, that should be enforced at the reference secret level

			if secret.Labels == nil {
				secret.Labels = make(map[string]string)
			}
			secret.Labels["app.kubernetes.io/managed-by"] = controllerRef.Name
			secret.Labels["admin.kubeflow.org/claim-namespace"] = namespace
			return nil
		})
		if err != nil {
			r.logger.Debug(errors.Wrap(err, errApplySecret).Error())
		}
		switch res {
		case controllerutil.OperationResultCreated, controllerutil.OperationResultUpdated:
			r.logger.Debug("finished applying secret", "namespace", namespace)
		}
	}

	labelSelector := &metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{{
			Key:      "app.kubernetes.io/managed-by",
			Operator: metav1.LabelSelectorOpIn,
			Values:   []string{controllerRef.Name},
		}},
	}
	if namespaces.Len() > 0 {
		labelSelector.MatchExpressions = append(labelSelector.MatchExpressions, metav1.LabelSelectorRequirement{
			Key:      "admin.kubeflow.org/claim-namespace",
			Operator: metav1.LabelSelectorOpNotIn,
			Values:   namespaces.UnsortedList(),
		})
	}

	selector, err := metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Delete all secrets that shouldn't exist in namespaces
	secretList := &corev1.SecretList{}
	if err := r.client.List(ctx, secretList, client.MatchingLabelsSelector{Selector: selector}); err != nil {
		return ctrl.Result{}, errors.New(errListNamespaces)
	}
	for _, secret := range secretList.Items {
		r.logger.Debug("removing orphaned secret", "namespace", secret.Namespace)
		if err := r.client.Delete(ctx, &secret); client.IgnoreNotFound(err) != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}
