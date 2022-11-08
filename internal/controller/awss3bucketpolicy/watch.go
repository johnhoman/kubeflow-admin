package awss3bucketpolicy

import (
	"context"

	"github.com/johnhoman/kubeflow-admin/apis/aws/v1alpha1"
	"github.com/johnhoman/kubeflow-admin/internal/types/awss3"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func EnqueueRequestsForServiceAccounts(reader client.Reader) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(o client.Object) []ctrl.Request {

		switch o.GetObjectKind().GroupVersionKind().GroupKind() {
		case v1alpha1.BucketConfigGroupKind, awss3.BucketGroupKind:
			serviceAccountList := &corev1.ServiceAccountList{}
			if err := reader.List(context.Background(), serviceAccountList,
				client.InNamespace(o.GetNamespace()),
			); err != nil {
				return nil
			}
			reqs := make([]ctrl.Request, 0)
			for _, item := range serviceAccountList.Items {
				if objAssociatedRoleArn(o) && objAssociatedRoleId(o) {
					reqs = append(reqs, ctrl.Request{
						NamespacedName: client.ObjectKeyFromObject(&item),
					})
				}
			}
			return reqs
		}

		return nil
	})
}

const (
	annotationIamRoleId  = "eks.amazonaws.com/role-id"
	annotationIamRoleArn = "eks.amazonaws.com/role-arn"
)

func objAssociatedRoleId(obj client.Object) bool {
	annotations := obj.GetAnnotations()
	if annotations != nil {
		_, ok := annotations[annotationIamRoleId]
		return ok
	}
	return false
}

func objAssociatedRoleArn(obj client.Object) bool {
	annotations := obj.GetAnnotations()
	if annotations != nil {
		_, ok := annotations[annotationIamRoleArn]
		return ok
	}
	return false
}

var hasAssociatedRoleId = predicate.NewPredicateFuncs(objAssociatedRoleId)

var hasAssociatedRoleArn = predicate.NewPredicateFuncs(objAssociatedRoleArn)

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
