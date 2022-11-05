package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SecretRef is the namespace and name of a secret to reflect into all user
// profiles
type SecretRef struct {
	// Name is the name of the secret to propagate to other namespaces
	Name string `json:"name"`
	// Namespace is the namespace the secret lives in. If empty, will default
	// to the namespace the controller is running in
	// +kubebuilder:default=kubeflow
	Namespace string `json:"namespace,omitempty"`
}

// ClusterSecretSpec is the spec for configuring secret reflection into tenant namespaces
type ClusterSecretSpec struct {
	// SecretRef is a reference to the secret to reflect to user
	// namespaces
	SecretRef SecretRef `json:"secretRef"`

	// Select a namespace or profile kind and/or name to apply secrets to
	// +optional
	Selector Selector `json:"selector,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

type ClusterSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ClusterSecretSpec `json:"spec"`
}

// +kubebuilder:object:root=true

type ClusterSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ClusterSecret `json:"items,omitempty"`
}
