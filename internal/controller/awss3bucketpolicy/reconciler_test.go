package awss3bucketpolicy_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/johnhoman/kubeflow-admin/apis/v1alpha1"
	"github.com/johnhoman/kubeflow-admin/internal/controller/awss3bucketpolicy"
	"github.com/johnhoman/kubeflow-admin/internal/types/awsiam/ack"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)



type manager struct { client client.Client }
func (m *manager) GetClient() client.Client { return m.client }
func newManager(cli client.Client) *manager { return &manager{client: cli} }


func TestReconciler(t *testing.T) {
	ctx := context.Background()

	cases := map[string]struct {
		serviceAccount      *corev1.ServiceAccount
		objects             []client.Object
		want                map[client.Object]client.Object
	}{
		"CreatesAPolicy": {
			serviceAccount: &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "edit",
					Namespace: "foo",
					Annotations: map[string]string{
						"eks.amazonaws.com/role-id":  "AROA1234567890EXAMPLE",
						"eks.amazonaws.com/role-arn": "arn:aws:iam::111122223333:role/edit-foo",
					},
				},
			},
			objects: []client.Object{
				&unstructured.Unstructured{Object: map[string]any{
					"apiVersion": "s3.services.k8s.aws/v1alpha1",
					"kind":       "Bucket",
					"metadata": map[string]any{
						"name":      "xxxx-foo-default",
						"namespace": "foo",
					},
					"spec": map[string]any{
						"name": "xxxx-foo-default",
					},
					"status": map[string]any{
						"ackResourceMetadata": map[string]any{
							"arn": "arn:aws:s3:::xxxx-foo-default",
						},
					},
				}},
			},
			want: map[client.Object]client.Object{
				ack.NewUnstructuredPolicy(): &unstructured.Unstructured{Object: map[string]any{
					"apiVersion": "iam.services.k8s.aws/v1alpha1",
					"kind": "Policy",
					"metadata": map[string]any{
						"name": "xxxx-foo-default-gnaxza",
						"namespace": "foo",
					},
					"spec": map[string]any{
						"name": "xxxx-foo-default-gnaxza",
						"policyDocument": mustJsonMarshal(map[string]any{
							"Version": "2012-10-17",
							"Statement": []any{
								map[string]any{
									"Effect": "Allow",
									"Principal": map[string]any{
										"AWS": "arn:aws:iam::111122223333:role/edit-foo",
									},
									"Action": []any{
										"s3:ListBucket",
									},
									"Resource": []any{"arn:aws:s3:::xxxx-foo-default"},
								},
								map[string]any{
									"Effect": "Allow",
									"Principal": map[string]any{
										"AWS": "arn:aws:iam::111122223333:role/edit-foo",
									},
									"Action": []any{
										"s3:GetObject",
										"s3:PutObject",
										"s3:DeleteObject",
									},
									"Resource": []any{"arn:aws:s3:::xxxx-foo-default/*"},
								},
							},
						}),
						"description": "Access to bucket xxxx-foo-default for iam role arn:aws:iam::111122223333:role/edit-foo",
					},
				}},
			},
		},
	}

	qt.Assert(t, v1alpha1.AddToScheme(scheme.Scheme), qt.IsNil)

	for name, subtest := range cases {
		t.Run(name, func(t *testing.T) {

			k8s := fake.NewClientBuilder().
				WithScheme(scheme.Scheme).
				WithObjects(subtest.serviceAccount).
				WithObjects(subtest.objects...).
				Build()

			zl := zap.New(zap.UseDevMode(true))

			reconciler := awss3bucketpolicy.NewReconciler(
				newManager(k8s),
				awss3bucketpolicy.WithLogger(logging.NewLogrLogger(zl)),
			)
			req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(subtest.serviceAccount)}
			_, err := reconciler.Reconcile(ctx, req)
			qt.Assert(t, err, qt.ErrorMatches, "policy is missing ARN")

			for got, want := range subtest.want {
				qt.Assert(t, k8s.Get(ctx, client.ObjectKeyFromObject(want), got), qt.IsNil)
				qt.Assert(t, got, qt.CmpEquals(
					cmpopts.IgnoreMapEntries(func(k, v any) bool { return k == "resourceVersion" }),
					cmp.FilterPath(
						func(path cmp.Path) bool {
							return strings.Contains(path.Index(-2).String(), "policyDocument")
						},
						cmp.Transformer("", func(in string) map[string]any {
							m := make(map[string]any)
							qt.Assert(t, json.Unmarshal([]byte(in), &m), qt.IsNil)
							return m
						}),
					),
				), want)
			}
		})
	}
}

func mustJsonMarshal(m map[string]any) string {
	raw, err := json.Marshal(m)
	if err != nil {
		panic(err.(any))
	}
	return string(raw)
}
