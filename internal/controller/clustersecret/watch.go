package clustersecret

import (
	"context"

	"github.com/johnhoman/kubeflow-admin/apis/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)


func NewEnqueueRequestsForClusterSecrets(reader client.Reader) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(o client.Object) []ctrl.Request {
		ns, ok := o.(*corev1.Namespace)
		if !ok {
			return nil
		}
		secretList := &v1alpha1.ClusterSecretList{}
		if err := reader.List(context.Background(), secretList); err != nil {
			return nil
		}

		reqs := make([]ctrl.Request, 0)
		for _, item := range secretList.Items {
			selector := labels.Everything()
			if item.Spec.Selector.Namespace != nil {
				var err error
				selector, err = metav1.LabelSelectorAsSelector(item.Spec.Selector.Namespace)
				if err != nil {
					continue
				}
			}
			if selector.Matches(labels.Set(ns.Labels)) {
				reqs = append(reqs, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&item)})
			}
		}
		return reqs
	})
}