package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterPodDefaultSpec contains the selector and pod spec
// with the patch to apply to a pod
type ClusterPodDefaultSpec struct {
	Selector          *metav1.LabelSelector  `json:"selector,omitempty"`
	NamespaceSelector *metav1.LabelSelector  `json:"namespaceSelector,omitempty"`
	Priority          *int                   `json:"priority,omitempty"`
	Template          corev1.PodTemplateSpec `json:"template,omitempty"`
}

// ClusterPodDefault configures an admission webhook with defaults to apply
// to selected pods. This ClusterPodDefault accomplishes a similar goal
// to the kubeflow core ClusterPodDefault, but includes a full pod spec

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

type ClusterPodDefault struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ClusterPodDefaultSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

type ClusterPodDefaultList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ClusterPodDefault `json:"items,omitempty"`
}
