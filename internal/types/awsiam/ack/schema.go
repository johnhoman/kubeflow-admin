package ack

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	Group   = "iam.services.k8s.aws"
	Version = "v1alpha1"
)

var (
	GroupVersion  = schema.GroupVersion{Group: Group, Version: Version}
	RoleKind      = "Role"
	RoleGroupKind = GroupVersion.WithKind(RoleKind).GroupKind()
)
