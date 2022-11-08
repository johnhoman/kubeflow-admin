package eksirsa_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/johnhoman/kubeflow-admin/internal/controller/eksirsa"
	"github.com/johnhoman/kubeflow-admin/internal/types/awsiam/ack"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/johnhoman/kubeflow-admin/apis/aws/v1alpha1"
)

type manager struct {
	client client.Client
}

func (m *manager) GetClient() client.Client { return m.client }
func newManager(cli client.Client) *manager { return &manager{client: cli} }

func TestReconciler(t *testing.T) {
	ctx := context.Background()

	cases := map[string]struct {
		serviceAccount *corev1.ServiceAccount
		objects        []client.Object
		want           map[string]any
	}{
		"ShouldCreateAnIAMRole": {
			serviceAccount: &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "edit",
					Namespace: "foo",
					UID:       "280aee46-2594-48fa-a51b-f508f76d3530",
					OwnerReferences: []metav1.OwnerReference{{
						Controller: pointer.Bool(true),
						Name:       "foo",
						Kind:       "Profile",
						APIVersion: "kubeflow.org/v1",
					}},
				},
			},
			objects: []client.Object{
				&v1alpha1.RoleConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "default",
						Namespace: "foo",
					},
					Spec: v1alpha1.RoleConfigSpec{
						MaxSessionDuration: "1h",
						Issuer: v1alpha1.RoleConfigIssuer{
							ARN: "arn:aws:iam::111122223333:oidc-provider/oidc.eks.region-code.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE",
						},
					},
				},
			},
			want: map[string]any{
				"apiVersion": "iam.services.k8s.aws/v1alpha1",
				"kind":       "Role",
				"metadata": map[string]any{
					"name":      "edit",
					"namespace": "foo",
					"ownerReferences": []any{
						map[string]any{
							"blockOwnerDeletion": true,
							"controller":         true,
							"name":               "edit",
							"uid":                "280aee46-2594-48fa-a51b-f508f76d3530",
							"kind":               "ServiceAccount",
							"apiVersion":         "v1",
						},
					},
				},
				"spec": map[string]any{
					"assumeRolePolicyDocument": jsonMarshal(t, map[string]any{
						"Version": "2012-10-17",
						"Statement": []any{
							map[string]any{
								"Effect": "Allow",
								"Principal": map[string]any{
									"Federated": "arn:aws:iam::111122223333:oidc-provider/oidc.eks.region-code.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE",
								},
								"Action": "sts:AssumeRoleWithWebIdentity",
								"Condition": map[string]any{
									"StringEquals": map[string]any{
										"oidc.eks.region-code.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE:sub": []string{"system:serviceaccount:foo:edit"},
										"oidc.eks.region-code.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE:aud": []string{"sts.amazonaws.com"},
									},
								},
							},
						},
					}),
					"description":        "iam role for service account system:serviceaccount:foo:edit",
					"maxSessionDuration": int64(3600),
					"name":               "system-serviceaccount-foo-edit",
					"policies":           make([]any, 0),
				},
			},
		},
		"ShouldAddPolicies": {
			serviceAccount: &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "view",
					Namespace: "foo",
					Annotations: map[string]string{
						"policy.aws.admin.kubeflow.org/0": "arn:aws:iam::012345678912:policy/p0",
						"policy.aws.admin.kubeflow.org/1": "arn:aws:iam::012345678912:policy/p1",
					},
					UID: "280aee46-2594-48fa-a51b-f508f76d3530",
					OwnerReferences: []metav1.OwnerReference{{
						Controller: pointer.Bool(true),
						Name:       "foo",
						Kind:       "Profile",
						APIVersion: "kubeflow.org/v1",
					}},
				},
			},
			objects: []client.Object{
				&v1alpha1.RoleConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "default",
						Namespace: "foo",
					},
					Spec: v1alpha1.RoleConfigSpec{
						MaxSessionDuration: "1h",
						Issuer: v1alpha1.RoleConfigIssuer{
							ARN: "arn:aws:iam::012345678912:oidc-provider/oidc.eks.region-code.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE",
						},
					},
				},
			},
			want: map[string]any{
				"apiVersion": "iam.services.k8s.aws/v1alpha1",
				"kind":       "Role",
				"metadata": map[string]any{
					"name":      "view",
					"namespace": "foo",
					"ownerReferences": []any{
						map[string]any{
							"blockOwnerDeletion": true,
							"controller":         true,
							"name":               "view",
							"uid":                "280aee46-2594-48fa-a51b-f508f76d3530",
							"kind":               "ServiceAccount",
							"apiVersion":         "v1",
						},
					},
				},
				"spec": map[string]any{
					"assumeRolePolicyDocument": jsonMarshal(t, map[string]any{
						"Version": "2012-10-17",
						"Statement": []any{
							map[string]any{
								"Effect": "Allow",
								"Principal": map[string]any{
									"Federated": "arn:aws:iam::012345678912:oidc-provider/oidc.eks.region-code.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE",
								},
								"Action": "sts:AssumeRoleWithWebIdentity",
								"Condition": map[string]any{
									"StringEquals": map[string]any{
										"oidc.eks.region-code.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE:sub": []string{"system:serviceaccount:foo:view"},
										"oidc.eks.region-code.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE:aud": []string{"sts.amazonaws.com"},
									},
								},
							},
						},
					}),
					"description":        "iam role for service account system:serviceaccount:foo:view",
					"maxSessionDuration": int64(3600),
					"name":               "system-serviceaccount-foo-view",
					"policies": []any{
						"arn:aws:iam::012345678912:policy/p0",
						"arn:aws:iam::012345678912:policy/p1",
					},
				},
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

			r := eksirsa.NewReconciler(newManager(k8s), eksirsa.WithLogger(logging.NewLogrLogger(zl)))

			req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(subtest.serviceAccount)}
			_, err := r.Reconcile(ctx, req)
			qt.Assert(t, err, qt.ErrorMatches, "waiting for valid iam role arn")

			want := ack.NewUnstructuredRole()
			want.Object = subtest.want

			got := ack.NewUnstructuredRole()
			qt.Assert(t, k8s.Get(ctx, client.ObjectKeyFromObject(want), got), qt.IsNil)
			qt.Assert(t, got, qt.CmpEquals(
				cmpopts.IgnoreMapEntries(func(k, v any) bool { return k == "resourceVersion" }),
				cmp.FilterPath(
					func(path cmp.Path) bool {
						return strings.Contains(path.Index(-2).String(), "assumeRolePolicyDocument")
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

func jsonMarshal(t *testing.T, m map[string]any) string {
	raw, err := json.Marshal(m)
	qt.Assert(t, err, qt.IsNil)
	return string(raw)
}
