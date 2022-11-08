package watch

import (
	"github.com/johnhoman/kubeflow-admin/internal/types/profile"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	errRequeueServiceAccountFromNamespaceEvent = "failed to requeue services accounts from namespace event"
	errParseGVKFromControllerRef               = "failed to parse GVK from controller reference"
)

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

var IsOwnedByProfilePredicate = predicate.NewPredicateFuncs(func(obj client.Object) bool {
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
