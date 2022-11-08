package ack

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	errPolicyKind        = "cannot convert to Policy, invalid kind"
	errSetName           = "Could not set attribute Name value on kind Policy"
	errGetName           = "Could not get attribute Name value from kind Policy"
	errSetPath           = "Could not set attribute Path value on kind Policy"
	errGetPath           = "Could not get attribute Path value from kind Policy"
	errSetDescription    = "Could not set attribute Description value on kind Policy"
	errGetDescription    = "Could not get attribute Description value from kind Policy"
	errSetPolicyDocument = "Could not set attribute PolicyDocument value on kind Policy"
	errGetPolicyDocument = "Could not get attribute PolicyDocument value from kind Policy"
	errSetArn            = "Could not set attribute Arn value on kind Policy"
	errGetArn            = "Could not get attribute Arn value from kind Policy"
)

type Policy struct {
	obj map[string]any
}

func (w *Policy) ToUnstructured() *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: w.obj}
}

func (w *Policy) UnstructuredContent() map[string]any {
	return w.obj
}

func (w *Policy) GetName() (string, error) {
	const def = ""
	name, ok, err := unstructured.NestedString(w.obj, "spec", "name")

	if err != nil {
		return def, errors.Wrap(err, errGetName)
	}
	if !ok {
		return def, nil
	}
	return name, nil
}

func (w *Policy) SetName(name string) error {
	err := unstructured.SetNestedField(w.obj, name, "spec", "name")
	if err != nil {
		return errors.Wrap(err, errSetName)
	}
	return nil
}

func (w *Policy) GetPath() (string, error) {
	const def = ""
	path, ok, err := unstructured.NestedString(w.obj, "spec", "path")

	if err != nil {
		return def, errors.Wrap(err, errGetPath)
	}
	if !ok {
		return def, nil
	}
	return path, nil
}

func (w *Policy) SetPath(path string) error {
	err := unstructured.SetNestedField(w.obj, path, "spec", "path")
	if err != nil {
		return errors.Wrap(err, errSetPath)
	}
	return nil
}

func (w *Policy) GetDescription() (string, error) {
	const def = ""
	description, ok, err := unstructured.NestedString(w.obj, "spec", "description")

	if err != nil {
		return def, errors.Wrap(err, errGetDescription)
	}
	if !ok {
		return def, nil
	}
	return description, nil
}

func (w *Policy) SetDescription(description string) error {
	err := unstructured.SetNestedField(w.obj, description, "spec", "description")
	if err != nil {
		return errors.Wrap(err, errSetDescription)
	}
	return nil
}

func (w *Policy) GetPolicyDocument() (string, error) {
	const def = ""
	policyDocument, ok, err := unstructured.NestedString(w.obj, "spec", "policyDocument")

	if err != nil {
		return def, errors.Wrap(err, errGetPolicyDocument)
	}
	if !ok {
		return def, nil
	}
	return policyDocument, nil
}

func (w *Policy) SetPolicyDocument(policyDocument string) error {
	err := unstructured.SetNestedField(w.obj, policyDocument, "spec", "policyDocument")
	if err != nil {
		return errors.Wrap(err, errSetPolicyDocument)
	}
	return nil
}

func (w *Policy) GetArn() (string, error) {
	const def = ""
	arn, ok, err := unstructured.NestedString(w.obj, "status", "ackResourceMetadata", "arn")

	if err != nil {
		return def, errors.Wrap(err, errGetArn)
	}
	if !ok {
		return def, nil
	}
	return arn, nil
}

func (w *Policy) SetArn(arn string) error {
	err := unstructured.SetNestedField(w.obj, arn, "status", "ackResourceMetadata", "arn")
	if err != nil {
		return errors.Wrap(err, errSetArn)
	}
	return nil
}

func NewPolicyFromUnstructured(u *unstructured.Unstructured) (*Policy, error) {
	if u.GroupVersionKind() != GroupVersion.WithKind(PolicyKind) {
		return nil, errors.New(errPolicyKind)
	}
	return &Policy{obj: u.Object}, nil
}

func NewUnstructuredPolicy() *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": GroupVersion.String(),
		"kind":       PolicyKind,
	}}
}

var (
	PolicyKind      = "Policy"
	PolicyGroupKind = GroupVersion.WithKind(PolicyKind).GroupKind()
)
