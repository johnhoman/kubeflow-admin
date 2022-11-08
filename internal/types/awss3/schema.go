package awss3

import "k8s.io/apimachinery/pkg/runtime/schema"

const (
	Group   = "s3.services.k8s.aws"
	Version = "v1alpha1"
	Kind    = "Bucket"
)

var (
	GroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	GroupVersionKind = GroupVersion.WithKind(Kind)
	GroupKind        = GroupVersionKind.GroupKind()
)
