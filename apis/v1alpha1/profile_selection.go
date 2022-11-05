package v1alpha1

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SubjectKind string

const (
	SubjectKindUser  SubjectKind = rbacv1.UserKind
	SubjectKindGroup SubjectKind = rbacv1.GroupKind
)

// Subject references a specific subject kind
type Subject struct {
	// Kind is the subject kind. Can either be User or Group
	// +kubebuilder:validation:Enum=User;Group
	Kind SubjectKind `json:"kind"`

	// Name is the name of a subject
	// +optional
	Name string `json:"name,omitempty"`
}

func (in *Subject) Matches(sub *rbacv1.Subject) bool {
	if in.Name != "" && in.Name != sub.Name {
		return false
	}
	if in.Kind != "" && string(in.Kind) != sub.Kind {
		return false
	}
	return true
}

type Selector struct {
	// Only apply to a specific subject
	// +optional
	Subject *Subject `json:"subject,omitempty"`

	// Optionally limit the namespaces that this secret is reflected into. If both selector
	// and Subject are specified, they result will be ANDed.
	// +optional
	Namespace *metav1.LabelSelector `json:"namespace,omitempty"`
}
