package ack

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	errFmtInvalidRoleKind = "expected %s, but got %s"
)

type Role struct {
	obj map[string]any
}

func (r *Role) UnstructuredContent() map[string]any {
	return r.obj
}

func (r *Role) ToUnstructured() *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: r.obj}
}

func (r *Role) GetAssumedRolePolicyDocument() (string, error) {
	doc, ok, err := unstructured.NestedString(r.obj, "spec", "assumeRolePolicyDocument")
	if err != nil {
		return "", err
	}
	if !ok {
		// maybe return error?
		return "", nil
	}
	return doc, nil
}

func (r *Role) SetAssumedRolePolicyDocument(doc string) error {
	err := unstructured.SetNestedField(r.obj, doc, "spec", "assumeRolePolicyDocument")
	if err != nil {
		return err
	}
	return nil
}

func (r *Role) GetDescription() (string, error) {
	desc, ok, err := unstructured.NestedString(r.obj, "spec", "description")
	if err != nil {
		return "", err
	}
	if !ok {
		return "", nil
	}

	return desc, nil
}

func (r *Role) SetDescription(desc string) error {
	err := unstructured.SetNestedField(r.obj, desc, "spec", "description")
	if err != nil {
		return err
	}
	return nil
}

func (r *Role) GetMaxDurationSeconds() (time.Duration, error) {
	dur, ok, err := unstructured.NestedInt64(r.obj, "spec", "maxSessionDuration")
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, nil
	}

	return time.Second * time.Duration(dur), nil
}

func (r *Role) SetMaxDurationSeconds(duration time.Duration) error {
	err := unstructured.SetNestedField(r.obj, duration.Seconds(), "spec", "maxSessionDuration")
	if err != nil {
		return err
	}
	return nil
}

func (r *Role) GetName() (string, error) {
	name, ok, err := unstructured.NestedString(r.obj, "spec", "name")
	if err != nil {
		return "", err
	}
	if !ok {
		return "", nil
	}

	return name, nil
}

func (r *Role) SetName(name string) error {
	err := unstructured.SetNestedField(r.obj, name, "spec", "name")
	if err != nil {
		return err
	}
	return nil
}

func (r *Role) GetPath() (string, error) {
	path, ok, err := unstructured.NestedString(r.obj, "spec", "path")
	if err != nil {
		return "", err
	}
	if !ok {
		return "", nil
	}

	return path, nil
}

func (r *Role) SetPath(path string) error {
	err := unstructured.SetNestedField(r.obj, path, "spec", "path")
	if err != nil {
		return err
	}
	return nil
}

func (r *Role) GetPermissionBoundary() (string, error) {
	boundary, ok, err := unstructured.NestedString(r.obj, "spec", "boundary")
	if err != nil {
		return "", err
	}
	if !ok {
		return "", nil
	}

	return boundary, nil
}

func (r *Role) SetPermissionBoundary(boundary string) error {
	err := unstructured.SetNestedField(r.obj, boundary, "spec", "boundary")
	if err != nil {
		return err
	}
	return nil
}

func (r *Role) GetPolicies() ([]string, error) {
	policies, ok, err := unstructured.NestedStringSlice(r.obj, "spec", "policies")
	if err != nil {
		return nil, err
	}
	if !ok {
		return make([]string, 0), nil
	}

	return policies, nil
}

func (r *Role) SetPolicies(policies []string) error {
	err := unstructured.SetNestedStringSlice(r.obj, policies, "spec", "policies")
	if err != nil {
		return err
	}
	return nil
}

func (r *Role) GetTags() (map[string]string, error) {
	rv := make(map[string]string)
	tags, ok, err := unstructured.NestedSlice(r.obj, "spec", "tags")
	if err != nil {
		return nil, err
	}
	if !ok {
		return make(map[string]string), nil
	}
	for _, item := range tags {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		tagKey, ok := m["key"]
		if !ok {
			continue
		}
		tagVal, ok := m["value"]
		if !ok {
			continue
		}
		key, ok := tagKey.(string)
		if !ok {
			continue
		}
		val, ok := tagVal.(string)
		if !ok {
			continue
		}
		rv[key] = val
	}
	return rv, nil
}

func (r *Role) SetTags(in map[string]string) error {

	tags := make([]any, 0, len(in))
	for key, value := range in {
		tags = append(tags, map[string]any{
			"key":   key,
			"value": value,
		})
	}

	err := unstructured.SetNestedSlice(r.obj, tags, "spec", "tags")
	if err != nil {
		return err
	}
	return nil
}

func (r *Role) Arn() (string, error) {
	arn, ok, err := unstructured.NestedString(r.obj, "status", "ackResourceMetadata", "arn")
	if err != nil {
		return "", err
	}
	if !ok {
		return "", nil
	}

	return arn, nil
}

func (r *Role) Id() (string, error) {
	arn, ok, err := unstructured.NestedString(r.obj, "status", "roleID")
	if err != nil {
		return "", err
	}
	if !ok {
		return "", nil
	}

	return arn, nil
}

func NewRoleFromUnstructured(u *unstructured.Unstructured) (*Role, error) {
	if u.GroupVersionKind() != GroupVersion.WithKind(RoleKind) {
		return nil, errors.New(fmt.Sprintf(errFmtInvalidRoleKind,
			GroupVersion.WithKind(RoleKind),
			u.GroupVersionKind(),
		))
	}
	return &Role{obj: u.Object}, nil
}

func NewUnstructuredRole() *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(GroupVersion.WithKind("Role"))
	return u
}
