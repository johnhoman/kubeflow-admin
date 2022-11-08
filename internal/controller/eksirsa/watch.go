package eksirsa

import (
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/johnhoman/kubeflow-admin/apis/aws/v1alpha1"
	"github.com/johnhoman/kubeflow-admin/internal/types/profile"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	errRequeueServiceAccountFromNamespaceEvent = "failed to requeue services accounts from namespace event"
	errParseGVKFromControllerRef               = "failed to parse GVK from controller reference"
)

func EnqueueRequestsForServiceAccounts(reader client.Reader, logger logging.Logger) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(o client.Object) []ctrl.Request {
		switch obj := o.(type) {
		case *v1alpha1.RoleConfig:
			return requeueServiceAccountsInNamespace(obj.Namespace, reader, logger)
		}
		return nil
	})
}

func gvkForController(obj client.Object) (schema.GroupVersionKind, error) {
	owner := metav1.GetControllerOf(obj)
	if owner == nil {
		return schema.GroupVersionKind{}, errors.New(errParseGVKFromControllerRef)
	}
	gv, err := schema.ParseGroupVersion(owner.APIVersion)
	if err != nil {
		return schema.GroupVersionKind{}, errors.Wrap(err, errParseGVKFromControllerRef)
	}
	return gv.WithKind(owner.Kind), nil
}

func requeueServiceAccountsInNamespace(namespace string, reader client.Reader, logger logging.Logger) []ctrl.Request {
	serviceAccountList := &corev1.ServiceAccountList{}
	if err := reader.List(context.Background(), serviceAccountList, client.InNamespace(namespace)); err != nil {
		logger.Debug(errors.Wrap(err, errRequeueServiceAccountFromNamespaceEvent).Error())
		return nil
	}
	reqs := make([]ctrl.Request, 0)
	for _, item := range serviceAccountList.Items {

		gvk, err := gvkForController(&item)
		if err != nil {
			logger.Debug(err.Error())
			continue
		}
		if gvk.GroupKind() == profile.GroupKind {
			reqs = append(reqs, ctrl.Request{
				NamespacedName: client.ObjectKeyFromObject(&item),
			})
		}
	}
	return reqs
}

var isOwnedByProfile = predicate.NewPredicateFuncs(func(obj client.Object) bool {
	owner := metav1.GetControllerOf(obj)
	if owner == nil {
		return false
	}
	gvk, err := gvkForController(obj)
	if err != nil {
		return false
	}
	return gvk.GroupKind() == profile.GroupKind
})
