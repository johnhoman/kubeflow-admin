package profile

import (
	"github.com/pkg/errors"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	errSubject = "no subject defined on profile"
	errKind    = "cannot convert to profile, invalid kind"
)

type Profile struct {
	obj map[string]interface{}
}

func (p *Profile) GetOwner() (*rbacv1.Subject, error) {
	sub, ok, err := unstructured.NestedMap(p.obj, "spec", "owner")
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New(errSubject)
	}
	subject := &rbacv1.Subject{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(sub, subject); err != nil {
		return nil, err
	}
	return subject, nil
}

func (p *Profile) IsOwnedByUser() (bool, error) {
	sub, err := p.GetOwner()
	if err != nil {
		return false, err
	}
	return sub.Kind == rbacv1.UserKind, nil
}

func (p *Profile) IsOwnedByGroup() (bool, error) {
	sub, err := p.GetOwner()
	if err != nil {
		return false, err
	}
	return sub.Kind == rbacv1.GroupKind, nil
}

func (p *Profile) ToUnstructured() *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: p.obj}
}

func NewFromUnstructured(u *unstructured.Unstructured) (*Profile, error) {
	if u.GroupVersionKind() != GroupVersion.WithKind(Kind) {
		return nil, errors.New(errKind)
	}
	return &Profile{obj: u.Object}, nil
}

func NewUnstructured() *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": GroupVersion.String(),
		"kind": Kind,
	}}
}