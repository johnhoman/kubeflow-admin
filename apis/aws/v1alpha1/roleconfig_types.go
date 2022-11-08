package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RoleConfigIssuer struct {
	ARN string `json:"arn"`
}

type RoleConfigSpec struct {
	MaxSessionDuration string           `json:"maxSessionDuration,omitempty"`
	Issuer             RoleConfigIssuer `json:"issuer"`
}

// +kubebuilder:object:root=true

type RoleConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec RoleConfigSpec `json:"spec"`
}

// +kubebuilder:object:root=true

type RoleConfigList struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Items []RoleConfig `json:"items,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

type ClusterRoleConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

// +kubebuilder:object:root=true

type ClusterRoleConfigList struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Items []RoleConfig `json:"items,omitempty"`
}
