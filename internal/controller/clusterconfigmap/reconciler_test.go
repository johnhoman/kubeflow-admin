package clusterconfigmap

import (
	"context"
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/johnhoman/kubeflow-admin/apis/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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
		clusterSecret *v1alpha1.ClusterSecret
		objects       []client.Object
		want          []*corev1.Secret
		dontWant      []*corev1.Secret
	}{
		"CreatesOwnedSecret": {
			clusterSecret: &v1alpha1.ClusterSecret{
				TypeMeta: metav1.TypeMeta{
					APIVersion: v1alpha1.SchemaGroupVersion.String(),
					Kind:       v1alpha1.ClusterSecretKind,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo-secret",
					UID:  types.UID("af452288-2cb8-4c0e-8f82-d4f1bdff9a55"),
				},
				Spec: v1alpha1.ClusterSecretSpec{
					SecretRef: v1alpha1.SecretRef{
						Name:      "ref-secret",
						Namespace: "kubeflow-0",
					},
				},
			},
			objects: []client.Object{
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "bar-namespace",
						Labels: map[string]string{
							"app.kubernetes.io/part-of": "kubeflow-profile",
						},
					},
				},
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "baz-namespace",
						Labels: map[string]string{
							"app.kubernetes.io/part-of": "kubeflow-profile",
						},
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ref-secret",
						Namespace: "kubeflow-0",
					},
					Type: corev1.DockerConfigJsonKey,
				},
			},
			want: []*corev1.Secret{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo-secret",
					Namespace: "bar-namespace",
					Labels: map[string]string{
						"admin.kubeflow.org/claim-namespace": "bar-namespace",
						"app.kubernetes.io/managed-by":       "foo-secret",
					},
					OwnerReferences: []metav1.OwnerReference{{
						BlockOwnerDeletion: pointer.Bool(true),
						Controller:         pointer.Bool(true),
						Name:               "foo-secret",
						UID:                types.UID("af452288-2cb8-4c0e-8f82-d4f1bdff9a55"),
						APIVersion:         "admin.kubeflow.org/v1alpha1",
						Kind:               "ClusterSecret",
					}},
				},
				Type: corev1.DockerConfigJsonKey,
			}, {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo-secret",
					Namespace: "baz-namespace",
					Labels: map[string]string{
						"admin.kubeflow.org/claim-namespace": "baz-namespace",
						"app.kubernetes.io/managed-by":       "foo-secret",
					},
					OwnerReferences: []metav1.OwnerReference{{
						BlockOwnerDeletion: pointer.Bool(true),
						Controller:         pointer.Bool(true),
						Name:               "foo-secret",
						UID:                types.UID("af452288-2cb8-4c0e-8f82-d4f1bdff9a55"),
						APIVersion:         "admin.kubeflow.org/v1alpha1",
						Kind:               "ClusterSecret",
					}},
				},
				Type: corev1.DockerConfigJsonKey,
			}},
		},
		"IgnoresNamespacesNotOwnedByProfile": {
			clusterSecret: &v1alpha1.ClusterSecret{
				TypeMeta: metav1.TypeMeta{
					APIVersion: v1alpha1.SchemaGroupVersion.String(),
					Kind:       v1alpha1.ClusterSecretKind,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo-secret",
					UID:  types.UID("af452288-2cb8-4c0e-8f82-d4f1bdff9a55"),
				},
				Spec: v1alpha1.ClusterSecretSpec{
					SecretRef: v1alpha1.SecretRef{
						Name:      "ref-secret",
						Namespace: "kubeflow-0",
					},
				},
			},
			objects: []client.Object{
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "bar-namespace",
						Labels: map[string]string{
							"app.kubernetes.io/part-of": "kubeflow-profile",
						},
					},
				},
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "baz-namespace",
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ref-secret",
						Namespace: "kubeflow-0",
					},
					Type: corev1.DockerConfigJsonKey,
				},
			},
			want: []*corev1.Secret{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo-secret",
					Namespace: "bar-namespace",
					Labels: map[string]string{
						"admin.kubeflow.org/claim-namespace": "bar-namespace",
						"app.kubernetes.io/managed-by":       "foo-secret",
					},
					OwnerReferences: []metav1.OwnerReference{{
						BlockOwnerDeletion: pointer.Bool(true),
						Controller:         pointer.Bool(true),
						Name:               "foo-secret",
						UID:                types.UID("af452288-2cb8-4c0e-8f82-d4f1bdff9a55"),
						APIVersion:         "admin.kubeflow.org/v1alpha1",
						Kind:               "ClusterSecret",
					}},
				},
				Type: corev1.DockerConfigJsonKey,
			}},
			dontWant: []*corev1.Secret{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo-secret",
					Namespace: "baz-namespace",
				},
			}},
		},
	}

	qt.Assert(t, v1alpha1.AddToScheme(scheme.Scheme), qt.IsNil)

	for name, subtest := range cases {
		t.Run(name, func(t *testing.T) {

			k8s := fake.NewClientBuilder().
				WithScheme(scheme.Scheme).
				WithObjects(subtest.clusterSecret).
				WithObjects(subtest.objects...).
				Build()

			zl := zap.New(zap.UseDevMode(true))

			reconciler := &Reconciler{
				client: k8s,
				logger: logging.NewLogrLogger(zl),
				record: event.NewNopRecorder(),
			}
			req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(subtest.clusterSecret)}
			_, err := reconciler.Reconcile(ctx, req)
			qt.Assert(t, err, qt.IsNil)

			for _, want := range subtest.want {
				got := &corev1.Secret{}
				qt.Assert(t, k8s.Get(ctx, client.ObjectKeyFromObject(want), got), qt.IsNil)
				qt.Assert(t, got, qt.CmpEquals(
					cmpopts.IgnoreUnexported(corev1.Secret{}),
					cmpopts.IgnoreFields(corev1.Secret{}, "ResourceVersion", "TypeMeta"),
				), want)
			}
			for _, want := range subtest.dontWant {
				key := client.ObjectKeyFromObject(want)
				qt.Assert(t, apierrors.IsNotFound(k8s.Get(ctx, key, &corev1.Secret{})), qt.IsTrue,
					qt.Commentf("expected secret not to exist in namespace: %s", want.Namespace),
				)
			}
		})
	}
}
