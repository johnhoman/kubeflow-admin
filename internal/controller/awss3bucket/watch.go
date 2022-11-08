package awss3bucket

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	"github.com/johnhoman/kubeflow-admin/internal/types/profile"
)

func NewEnqueueRequestsForNamespaces(reader client.Reader) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(o client.Object) []ctrl.Request {
		switch obj := o.(type) {
		case *corev1.ServiceAccount:
			owner := metav1.GetControllerOf(obj)
			if owner == nil {
				return nil
			}
			gv, err := schema.ParseGroupVersion(owner.APIVersion)
			if err != nil {
				return nil
			}
			if gv.WithKind(owner.Kind).GroupKind() == profile.GroupKind {
				return []ctrl.Request{{NamespacedName: client.ObjectKey{Name: obj.Name}}}
			}
		case *corev1.Secret:
			// ClusterSecret reflects a secret into target namespaces. If the namespace
			// isn't selected by the cluster secret, don't queue. If the secret type is not
			// .dockerconfigjson, then don't queue
			owner := metav1.GetControllerOf(obj)
			if owner == nil {
				return nil
			}
			gv, err := schema.ParseGroupVersion(owner.APIVersion)
			if err != nil {
				return nil
			}
			if gv.WithKind(owner.Kind).GroupKind() == profile.GroupKind {
				switch obj.Type {
				case corev1.DockerConfigJsonKey, corev1.DockerConfigKey:
					return []ctrl.Request{{NamespacedName: client.ObjectKey{Name: obj.Name}}}
				}
			}
		}
		return nil
	})
}
