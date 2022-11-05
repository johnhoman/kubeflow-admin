// Package v1alpha1 contains API Schema definitions for the kubeflow-ext v1alpha1 API group
//+kubebuilder:object:generate=true
//+groupName=contrib.kubeflow.org

package v1alpha1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

const (
	Group   = "admin.kubeflow.org"
	Version = "v1alpha1"
)

var (
	// SchemaGroupVersion is the group version
	SchemaGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemaGroupVersion}

	// AddToScheme adds the types in this group-version to a given scheme
	AddToScheme = SchemeBuilder.AddToScheme

	// ClusterSecretKind is the string representation of the ClusterSecret TypeMeta Kind field
	ClusterSecretKind = reflect.TypeOf(&ClusterSecret{}).Elem().Name()
)

func init() {
	SchemeBuilder.Register(
		&ClusterSecret{},
		&ClusterSecretList{},
		&ClusterPodDefault{},
		&ClusterPodDefaultList{},
	)
}
