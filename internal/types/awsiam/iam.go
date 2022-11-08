package awsiam

import (
	"time"

	"github.com/johnhoman/kubeflow-admin/internal/types/awsiam/ack"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var (
	_ Role = &ack.Role{}
)

type Unstructured interface {
	UnstructuredContent() map[string]any
	ToUnstructured() *unstructured.Unstructured
}

type Role interface {
	Unstructured
	GetAssumedRolePolicyDocument() (string, error)
	SetAssumedRolePolicyDocument(doc string) error
	GetDescription() (string, error)
	SetDescription(desc string) error
	GetMaxDurationSeconds() (time.Duration, error)
	SetMaxDurationSeconds(duration time.Duration) error
	GetName() (string, error)
	SetName(name string) error
	GetPath() (string, error)
	SetPath(path string) error
	GetPermissionBoundary() (string, error)
	SetPermissionBoundary(boundary string) error
	GetPolicies() ([]string, error)
	SetPolicies(policies []string) error
	GetTags() (map[string]string, error)
	SetTags(tags map[string]string) error
	Arn() (string, error)
	Id() (string, error)
}
