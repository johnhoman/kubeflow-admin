package imagepullsecrets

import (
	"context"
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/johnhoman/kubeflow-admin/apis/v1alpha1"
	"github.com/johnhoman/kubeflow-admin/internal/types/profile"
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
		namespace       *corev1.Namespace
		serviceAccounts []client.Object
		secrets         []client.Object
		want            []*corev1.ServiceAccount
	}{
		"ShouldDoNothingWhenThereAreNoSecrets": {
			namespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "foo",
					Labels:      map[string]string{"app.kubernetes.io/part-of": "kubeflow-profile"},
					Annotations: map[string]string{"owner": "foo@example.com"},
				},
			},
			serviceAccounts: []client.Object{
				&corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "default",
						Namespace: "foo",
						OwnerReferences: []metav1.OwnerReference{{
							Controller: pointer.Bool(true),
							Name:       "foo",
							Kind:       "Profile",
							APIVersion: profile.GroupVersion.String(),
						}},
					},
				},
			},
			want: []*corev1.ServiceAccount{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "default",
					Namespace: "foo",
					OwnerReferences: []metav1.OwnerReference{{
						Controller: pointer.Bool(true),
						Name:       "foo",
						Kind:       "Profile",
						APIVersion: profile.GroupVersion.String(),
					}},
				},
			}},
		},
		"ShouldDoNothingWhenThereAreNoSecretsOwnedByClusterSecrets": {
			namespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "foo",
					Labels:      map[string]string{"app.kubernetes.io/part-of": "kubeflow-profile"},
					Annotations: map[string]string{"owner": "foo@example.com"},
				},
			},
			serviceAccounts: []client.Object{
				&corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "default",
						Namespace: "foo",
						OwnerReferences: []metav1.OwnerReference{{
							Controller: pointer.Bool(true),
							Name:       "foo",
							Kind:       "Profile",
							APIVersion: profile.GroupVersion.String(),
						}},
					},
				},
			},
			secrets: []client.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ghcr.io",
						Namespace: "foo",
					},
					Type: corev1.DockerConfigJsonKey,
					Data: map[string][]byte{
						corev1.DockerConfigJsonKey: []byte("{}"),
					},
				},
			},
			want: []*corev1.ServiceAccount{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "default",
					Namespace: "foo",
					OwnerReferences: []metav1.OwnerReference{{
						Controller: pointer.Bool(true),
						Name:       "foo",
						Kind:       "Profile",
						APIVersion: profile.GroupVersion.String(),
					}},
				},
			}},
		},
		"ShouldAddImagePullSecretsOwnedByClusterSecret": {
			namespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "foo",
					Labels:      map[string]string{"app.kubernetes.io/part-of": "kubeflow-profile"},
					Annotations: map[string]string{"owner": "foo@example.com"},
				},
			},
			serviceAccounts: []client.Object{
				&corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "default",
						Namespace: "foo",
						OwnerReferences: []metav1.OwnerReference{{
							Controller: pointer.Bool(true),
							Name:       "foo",
							Kind:       "Profile",
							APIVersion: "kubeflow.org/v1",
						}},
					},
				},
			},
			secrets: []client.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ghcr.io",
						Namespace: "foo",
						OwnerReferences: []metav1.OwnerReference{{
							Controller: pointer.Bool(true),
							APIVersion: "admin.kubeflow.org/v1alpha1",
							Kind:       "ClusterSecret",
						}},
					},
					Type: corev1.DockerConfigJsonKey,
					Data: map[string][]byte{
						corev1.DockerConfigJsonKey: []byte("{}"),
					},
				},
			},
			want: []*corev1.ServiceAccount{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "default",
					Namespace: "foo",
					OwnerReferences: []metav1.OwnerReference{{
						Controller: pointer.Bool(true),
						Name:       "foo",
						Kind:       "Profile",
						APIVersion: profile.GroupVersion.String(),
					}},
				},
				ImagePullSecrets: []corev1.LocalObjectReference{{Name: "ghcr.io"}},
			}},
		},
		"ShouldNotOverwriteExistingImagePullSecrets": {
			namespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "foo",
					Labels:      map[string]string{"app.kubernetes.io/part-of": "kubeflow-profile"},
					Annotations: map[string]string{"owner": "foo@example.com"},
				},
			},
			serviceAccounts: []client.Object{
				&corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "default",
						Namespace: "foo",
						OwnerReferences: []metav1.OwnerReference{{
							Controller: pointer.Bool(true),
							Name:       "foo",
							Kind:       "Profile",
							APIVersion: "kubeflow.org/v1",
						}},
					},
					ImagePullSecrets: []corev1.LocalObjectReference{
						{Name: "foo.com"},
						{Name: "bar.com"},
						{Name: "zab.com"},
					},
				},
			},
			secrets: []client.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ghcr.io",
						Namespace: "foo",
						OwnerReferences: []metav1.OwnerReference{{
							Controller: pointer.Bool(true),
							APIVersion: "admin.kubeflow.org/v1alpha1",
							Kind:       "ClusterSecret",
						}},
					},
					Type: corev1.DockerConfigJsonKey,
					Data: map[string][]byte{
						corev1.DockerConfigJsonKey: []byte("{}"),
					},
				},
			},
			want: []*corev1.ServiceAccount{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "default",
					Namespace: "foo",
					OwnerReferences: []metav1.OwnerReference{{
						Controller: pointer.Bool(true),
						Name:       "foo",
						Kind:       "Profile",
						APIVersion: profile.GroupVersion.String(),
					}},
				},
				ImagePullSecrets: []corev1.LocalObjectReference{
					{Name: "bar.com"},
					{Name: "foo.com"},
					{Name: "ghcr.io"},
					{Name: "zab.com"},
				},
			}},
		},
	}

	qt.Assert(t, v1alpha1.AddToScheme(scheme.Scheme), qt.IsNil)

	for name, subtest := range cases {
		t.Run(name, func(t *testing.T) {

			k8s := fake.NewClientBuilder().
				WithScheme(scheme.Scheme).
				WithObjects(subtest.namespace).
				WithObjects(subtest.serviceAccounts...).
				WithObjects(subtest.secrets...).
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

			for _, want := range subtest.want {
				got := &corev1.ServiceAccount{}
				qt.Assert(t, k8s.Get(ctx, client.ObjectKeyFromObject(want), got), qt.IsNil)
				qt.Assert(t, got, qt.CmpEquals(
					cmpopts.IgnoreUnexported(corev1.ServiceAccount{}),
					cmpopts.IgnoreFields(corev1.ServiceAccount{}, "ResourceVersion", "TypeMeta"),
				), want)
			}
		})
	}
}
