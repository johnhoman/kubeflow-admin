package profile

import "k8s.io/apimachinery/pkg/runtime/schema"

var (
    Kind         = "Profile"
    GroupVersion = schema.GroupVersion{Group: "kubeflow.org", Version: "v1"}

    GroupVersionKind = GroupVersion.WithKind(Kind)
    GroupKind = GroupVersionKind.GroupKind()
)
