package awss3

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	errBucketKind               = "cannot convert to Bucket, invalid kind"
	errSetName                  = "Could not set attribute Name value on kind Bucket"
	errGetName                  = "Could not get attribute Name value from kind Bucket"
	errSetRegion                = "Could not set attribute Region value on kind Bucket"
	errGetRegion                = "Could not get attribute Region value from kind Bucket"
	errSetPolicy                = "Could not set attribute Policy value on kind Bucket"
	errGetPolicy                = "Could not get attribute Policy value from kind Bucket"
	errSetBlockPublicACLs       = "Could not set attribute BlockPublicACLs value on kind Bucket"
	errGetBlockPublicACLs       = "Could not get attribute BlockPublicACLs value from kind Bucket"
	errSetBlockPublicPolicy     = "Could not set attribute BlockPublicPolicy value on kind Bucket"
	errGetBlockPublicPolicy     = "Could not get attribute BlockPublicPolicy value from kind Bucket"
	errSetIgnorePublicACLs      = "Could not set attribute IgnorePublicACLs value on kind Bucket"
	errGetIgnorePublicACLs      = "Could not get attribute IgnorePublicACLs value from kind Bucket"
	errSetRestrictPublicBuckets = "Could not set attribute RestrictPublicBuckets value on kind Bucket"
	errGetRestrictPublicBuckets = "Could not get attribute RestrictPublicBuckets value from kind Bucket"
	errSetArn                   = "Could not set attribute Arn value on kind Bucket"
	errGetArn                   = "Could not get attribute Arn value from kind Bucket"
)

type Bucket struct {
	obj map[string]any
}

func (w *Bucket) ToUnstructured() *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: w.obj}
}

func (w *Bucket) UnstructuredContent() map[string]any {
	return w.obj
}

func (w *Bucket) GetName() (string, error) {
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

func (w *Bucket) SetName(name string) error {
	err := unstructured.SetNestedField(w.obj, name, "spec", "name")
	if err != nil {
		return errors.Wrap(err, errSetName)
	}
	return nil
}

func (w *Bucket) GetRegion() (string, error) {
	const def = ""
	locationConstraint, ok, err := unstructured.NestedString(w.obj, "spec", "createBucketConfiguration", "locationConstraint")

	if err != nil {
		return def, errors.Wrap(err, errGetRegion)
	}
	if !ok {
		return def, nil
	}
	return locationConstraint, nil
}

func (w *Bucket) SetRegion(locationConstraint string) error {
	err := unstructured.SetNestedField(w.obj, locationConstraint, "spec", "createBucketConfiguration", "locationConstraint")
	if err != nil {
		return errors.Wrap(err, errSetRegion)
	}
	return nil
}

func (w *Bucket) GetPolicy() (string, error) {
	const def = ""
	policy, ok, err := unstructured.NestedString(w.obj, "spec", "policy")

	if err != nil {
		return def, errors.Wrap(err, errGetPolicy)
	}
	if !ok {
		return def, nil
	}
	return policy, nil
}

func (w *Bucket) SetPolicy(policy string) error {
	err := unstructured.SetNestedField(w.obj, policy, "spec", "policy")
	if err != nil {
		return errors.Wrap(err, errSetPolicy)
	}
	return nil
}

func (w *Bucket) GetBlockPublicACLs() (bool, error) {
	const def = false
	blockPublicACLs, ok, err := unstructured.NestedBool(w.obj, "spec", "publicAccessBlock", "blockPublicACLs")

	if err != nil {
		return def, errors.Wrap(err, errGetBlockPublicACLs)
	}
	if !ok {
		return def, nil
	}
	return blockPublicACLs, nil
}

func (w *Bucket) SetBlockPublicACLs(blockPublicACLs bool) error {
	err := unstructured.SetNestedField(w.obj, blockPublicACLs, "spec", "publicAccessBlock", "blockPublicACLs")
	if err != nil {
		return errors.Wrap(err, errSetBlockPublicACLs)
	}
	return nil
}

func (w *Bucket) GetBlockPublicPolicy() (bool, error) {
	const def = false
	blockPublicPolicy, ok, err := unstructured.NestedBool(w.obj, "spec", "publicAccessBlock", "blockPublicPolicy")

	if err != nil {
		return def, errors.Wrap(err, errGetBlockPublicPolicy)
	}
	if !ok {
		return def, nil
	}
	return blockPublicPolicy, nil
}

func (w *Bucket) SetBlockPublicPolicy(blockPublicPolicy bool) error {
	err := unstructured.SetNestedField(w.obj, blockPublicPolicy, "spec", "publicAccessBlock", "blockPublicPolicy")
	if err != nil {
		return errors.Wrap(err, errSetBlockPublicPolicy)
	}
	return nil
}

func (w *Bucket) GetIgnorePublicACLs() (bool, error) {
	const def = false
	ignorePublicACLs, ok, err := unstructured.NestedBool(w.obj, "spec", "publicAccessBlock", "ignorePublicACLs")

	if err != nil {
		return def, errors.Wrap(err, errGetIgnorePublicACLs)
	}
	if !ok {
		return def, nil
	}
	return ignorePublicACLs, nil
}

func (w *Bucket) SetIgnorePublicACLs(ignorePublicACLs bool) error {
	err := unstructured.SetNestedField(w.obj, ignorePublicACLs, "spec", "publicAccessBlock", "ignorePublicACLs")
	if err != nil {
		return errors.Wrap(err, errSetIgnorePublicACLs)
	}
	return nil
}

func (w *Bucket) GetRestrictPublicBuckets() (bool, error) {
	const def = false
	restrictPublicBuckets, ok, err := unstructured.NestedBool(w.obj, "spec", "publicAccessBlock", "restrictPublicBuckets")

	if err != nil {
		return def, errors.Wrap(err, errGetRestrictPublicBuckets)
	}
	if !ok {
		return def, nil
	}
	return restrictPublicBuckets, nil
}

func (w *Bucket) SetRestrictPublicBuckets(restrictPublicBuckets bool) error {
	err := unstructured.SetNestedField(w.obj, restrictPublicBuckets, "spec", "publicAccessBlock", "restrictPublicBuckets")
	if err != nil {
		return errors.Wrap(err, errSetRestrictPublicBuckets)
	}
	return nil
}

func (w *Bucket) GetArn() (string, error) {
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

func (w *Bucket) SetArn(arn string) error {
	err := unstructured.SetNestedField(w.obj, arn, "status", "ackResourceMetadata", "arn")
	if err != nil {
		return errors.Wrap(err, errSetArn)
	}
	return nil
}

func NewBucketFromUnstructured(u *unstructured.Unstructured) (*Bucket, error) {
	if u.GroupVersionKind() != GroupVersion.WithKind(BucketKind) {
		return nil, errors.New(errBucketKind)
	}
	return &Bucket{obj: u.Object}, nil
}

func NewUnstructuredBucket() *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": GroupVersion.String(),
		"kind":       BucketKind,
	}}
}

var (
	BucketKind      = "Bucket"
	BucketGroupKind = GroupVersion.WithKind(BucketKind).GroupKind()
)
