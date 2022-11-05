package v1alpha1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConfigMapRef is the namespace and name of a ConfigMap to reflect into all user
// profiles
type ConfigMapRef struct {
    // Name is the name of the secret to propagate to other namespaces
    Name string `json:"name"`
    // Namespace is the namespace the secret lives in. If empty, will default
    // to the namespace the controller is running in
    // +kubebuilder:default=kubeflow
    Namespace string `json:"namespace,omitempty"`
}

// ClusterConfigMapSpec is the spec for configuring secret reflection into tenant namespaces
type ClusterConfigMapSpec struct {
    // ConfigMapRef is a reference to the ConfigMap to reflect to tenant
    // namespaces
    ConfigMapRef ConfigMapRef `json:"configMapRef"`

    // Only apply to a specific subject
    // +optional
    SubjectSelector *Subject `json:"subjectSelector,omitempty"`

    // Optionally limit the namespaces that this secret is reflected into. If both selector
    // and Subject are specified, they result will be ANDed.
    // +optional
    Selector *metav1.LabelSelector `json:"namespaceSelector,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

type ClusterConfigMap struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec ClusterConfigMapSpec `json:"spec"`
}

// +kubebuilder:object:root=true

type ClusterConfigMapList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`

    Items []ClusterConfigMap `json:"items,omitempty"`
}
