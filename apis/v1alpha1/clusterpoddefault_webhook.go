package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func (in *ClusterPodDefault) ValidateCreate() error                   { return nil }
func (in *ClusterPodDefault) ValidateUpdate(old runtime.Object) error { return nil }
func (in *ClusterPodDefault) ValidateDelete() error                   { return nil }

var _ admission.Validator = &ClusterPodDefault{}
