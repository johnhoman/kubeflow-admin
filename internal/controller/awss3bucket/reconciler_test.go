package awss3bucket

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/johnhoman/kubeflow-admin/apis/aws/v1alpha1"
	"github.com/johnhoman/kubeflow-admin/internal/types/awss3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestReconciler(t *testing.T) {
	ctx := context.Background()

	cases := map[string]struct {
		namespace *corev1.Namespace
		objects   []client.Object
		want      map[string]any
	}{
		"ShouldCreateABucket": {
			namespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
			},
			objects: []client.Object{
				&corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "edit",
						Namespace: "foo",
						Annotations: map[string]string{
							"eks.amazonaws.com/role-id":  "AROA1234567890EXAMPLE",
							"eks.amazonaws.com/role-arn": "arn:aws:iam::111122223333:role/edit-foo",
						},
					},
				},
				&corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "view",
						Namespace: "foo",
						Annotations: map[string]string{
							"eks.amazonaws.com/role-id":  "AROA1234567891EXAMPLE",
							"eks.amazonaws.com/role-arn": "arn:aws:iam::111122223333:role/view-foo",
						},
					},
				},
				&v1alpha1.BucketConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "default",
						Namespace: "foo",
					},
					Spec: v1alpha1.BucketConfigSpec{
						Name: v1alpha1.BucketConfigName{
							Prefix: pointer.String("xxxx"),
						},
						Region: "us-east-1",
					},
				},
			},
			want: map[string]any{
				"apiVersion": "s3.services.k8s.aws/v1alpha1",
				"kind":       "Bucket",
				"metadata": map[string]any{
					"name":      "default",
					"namespace": "foo",
				},
				"spec": map[string]any{
					"name": "xxxx-foo-default",
					"createBucketConfiguration": map[string]any{
						"locationConstraint": "us-east-1",
					},
					"policy": mustJsonMarshal(map[string]any{
						"Version": "2012-10-17",
						"Statement": []any{
							map[string]any{
								"Effect":    "Deny",
								"Principal": "*",
								"Action":    "s3:*",
								"Resource": []any{
									"arn:aws:s3:::xxxx-foo-default",
									"arn:aws:s3:::xxxx-foo-default/*",
								},
								"Condition": map[string]any{
									"StringNotLike": map[string]any{
										"aws:userId": []any{
											"AROA1234567890EXAMPLE:*",
											"AROA1234567891EXAMPLE:*",
										},
									},
								},
							},
						},
					}),
				},
			},
		},
	}

	qt.Assert(t, v1alpha1.AddToScheme(scheme.Scheme), qt.IsNil)

	for name, subtest := range cases {
		t.Run(name, func(t *testing.T) {

			k8s := fake.NewClientBuilder().
				WithScheme(scheme.Scheme).
				WithObjects(subtest.namespace).
				WithObjects(subtest.objects...).
				Build()

			zl := zap.New(zap.UseDevMode(true))

			reconciler := &Reconciler{
				client: k8s,
				logger: logging.NewLogrLogger(zl),
				record: event.NewNopRecorder(),
			}
			req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(subtest.namespace)}
			_, err := reconciler.Reconcile(ctx, req)
			qt.Assert(t, err, qt.IsNil)

			want := awss3.NewUnstructuredBucket()
			want.SetUnstructuredContent(subtest.want)

			got := awss3.NewUnstructuredBucket()
			qt.Assert(t, k8s.Get(ctx, client.ObjectKeyFromObject(want), got), qt.IsNil)
			qt.Assert(t, got, qt.CmpEquals(
				cmpopts.IgnoreMapEntries(func(k, v any) bool { return k == "resourceVersion" }),
				cmp.FilterPath(
					func(path cmp.Path) bool {
						return strings.Contains(path.Index(-2).String(), "policy")
					},
					cmp.Transformer("", func(in string) map[string]any {
						m := make(map[string]any)
						qt.Assert(t, json.Unmarshal([]byte(in), &m), qt.IsNil)
						return m
					}),
				),
			), want)
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
