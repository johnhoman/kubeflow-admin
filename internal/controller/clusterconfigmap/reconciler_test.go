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
		clusterConfigMap *v1alpha1.ClusterConfigMap
		objects          []client.Object
		want             []*corev1.ConfigMap
		dontWant         []*corev1.ConfigMap
	}{
		"CreatesOwnedSecret": {
			clusterConfigMap: &v1alpha1.ClusterConfigMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: v1alpha1.SchemaGroupVersion.String(),
					Kind:       v1alpha1.ClusterConfigMapKind,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo-cm",
					UID:  types.UID("af452288-2cb8-4c0e-8f82-d4f1bdff9a55"),
				},
				Spec: v1alpha1.ClusterConfigMapSpec{
					ConfigMapRef: v1alpha1.ConfigMapRef{
						Name:      "ref-config-map",
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
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ref-config-map",
						Namespace: "kubeflow-0",
					},
					Data: map[string]string{"foo": "bar"},
				},
			},
			want: []*corev1.ConfigMap{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo-cm",
					Namespace: "bar-namespace",
					Labels: map[string]string{
						"admin.kubeflow.org/claim-namespace": "bar-namespace",
						"app.kubernetes.io/managed-by":       "foo-cm",
					},
					OwnerReferences: []metav1.OwnerReference{{
						BlockOwnerDeletion: pointer.Bool(true),
						Controller:         pointer.Bool(true),
						Name:               "foo-cm",
						UID:                types.UID("af452288-2cb8-4c0e-8f82-d4f1bdff9a55"),
						APIVersion:         "admin.kubeflow.org/v1alpha1",
						Kind:               "ClusterConfigMap",
					}},
				},
				Data: map[string]string{"foo": "bar"},
			}},
		},
	}

	qt.Assert(t, v1alpha1.AddToScheme(scheme.Scheme), qt.IsNil)

	for name, subtest := range cases {
		t.Run(name, func(t *testing.T) {

			k8s := fake.NewClientBuilder().
				WithScheme(scheme.Scheme).
				WithObjects(subtest.clusterConfigMap).
				WithObjects(subtest.objects...).
				Build()

			zl := zap.New(zap.UseDevMode(true))

			reconciler := &Reconciler{
				client: k8s,
				logger: logging.NewLogrLogger(zl),
				record: event.NewNopRecorder(),
			}
			req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(subtest.clusterConfigMap)}
			_, err := reconciler.Reconcile(ctx, req)
			qt.Assert(t, err, qt.IsNil)

			for _, want := range subtest.want {
				got := &corev1.ConfigMap{}
				qt.Assert(t, k8s.Get(ctx, client.ObjectKeyFromObject(want), got), qt.IsNil)
				qt.Assert(t, got, qt.CmpEquals(
					cmpopts.IgnoreUnexported(corev1.ConfigMap{}),
					cmpopts.IgnoreFields(corev1.ConfigMap{}, "ResourceVersion", "TypeMeta"),
				), want)
			}
			for _, want := range subtest.dontWant {
				key := client.ObjectKeyFromObject(want)
				qt.Assert(t, apierrors.IsNotFound(k8s.Get(ctx, key, &corev1.ConfigMap{})), qt.IsTrue,
					qt.Commentf("expected secret not to exist in namespace: %s", want.Namespace),
				)
			}
		})
	}
}
