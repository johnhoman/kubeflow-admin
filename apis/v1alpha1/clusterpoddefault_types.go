package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterPodDefaultSpec contains the selector and pod spec
// with the patch to apply to a pod
type ClusterPodDefaultSpec struct {

	// Selector selects which pods the PodDefault applies to
	// +optional
	Selector *metav1.LabelSelector  `json:"selector,omitempty"`

	// NamespaceSelector selects which pods the PodDefault applies to
	// +optional
	NamespaceSelector *metav1.LabelSelector  `json:"namespaceSelector,omitempty"`

	// Priority is the order in which the pod defaults will be applied. Higher priority
	// means it will be applied last
	Priority *int `json:"priority,omitempty"`

	// Template is a PodTemplateSpec that will be merged with the Pod. The merge uses a
	// StrategicMergePatch strategy
	Template corev1.PodTemplateSpec `json:"template,omitempty"`
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
