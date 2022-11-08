package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BucketConfigName struct {
	Prefix   *string `json:"prefix,omitempty"`
	Template *string `json:"template,omitempty"`
}

type BucketConfigPolicy struct {
	Base string `json:"base"`
}

type BucketConfigSpec struct {
	Name   BucketConfigName    `json:"name"`
	Region string              `json:"region"`
	Policy *BucketConfigPolicy `json:"policy,omitempty"`
}

// +kubebuilder:object:root=true

type BucketConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec BucketConfigSpec `json:"spec"`
}

// +kubebuilder:object:root=true

type BucketConfigList struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Items []BucketConfig `json:"items,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

type ClusterBucketConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

// +kubebuilder:object:root=true

type ClusterClusterConfigList struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Items []RoleConfig `json:"items,omitempty"`
}
