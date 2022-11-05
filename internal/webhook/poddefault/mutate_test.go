package poddefault

import (
	"context"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/johnhoman/kubeflow-admin/apis/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_Mutate(t *testing.T) {

	cases := map[string]struct {
		original *corev1.Pod
		objects  []client.Object
		want     *corev1.Pod
	}{
		"CanAddAServiceAccount": {
			original: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "foo",
						Image: "python:3.9",
					}},
				},
			},
			objects: []client.Object{
				&v1alpha1.ClusterPodDefault{
					ObjectMeta: metav1.ObjectMeta{
						Name: "foo",
					},
					Spec: v1alpha1.ClusterPodDefaultSpec{
						Selector: &metav1.LabelSelector{},
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								ServiceAccountName: "foo-user",
							},
						},
					},
				},
			},
			want: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "foo-user",
					Containers: []corev1.Container{{
						Name:  "foo",
						Image: "python:3.9",
					}},
				},
			},
		},
		"AppliesInPriorityOrder": {
			original: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "foo",
						Image: "python:3.9",
					}},
				},
			},
			objects: []client.Object{
				&v1alpha1.ClusterPodDefault{
					ObjectMeta: metav1.ObjectMeta{
						Name: "foo",
					},
					Spec: v1alpha1.ClusterPodDefaultSpec{
						Selector: &metav1.LabelSelector{},
						Priority: pointer.Int(100),
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								ServiceAccountName: "foo-user",
							},
						},
					},
				},
				&v1alpha1.ClusterPodDefault{
					ObjectMeta: metav1.ObjectMeta{
						Name: "bar",
					},
					Spec: v1alpha1.ClusterPodDefaultSpec{
						Selector: &metav1.LabelSelector{},
						Priority: pointer.Int(10),
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								ServiceAccountName: "bar-user",
							},
						},
					},
				},
			},
			want: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "foo-user",
					Containers: []corev1.Container{{
						Name:  "foo",
						Image: "python:3.9",
					}},
				},
			},
		},
		"AppliesInAlphabeticalOrder": {
			original: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "foo",
						Image: "python:3.9",
					}},
				},
			},
			objects: []client.Object{
				&v1alpha1.ClusterPodDefault{
					ObjectMeta: metav1.ObjectMeta{
						Name: "foo",
					},
					Spec: v1alpha1.ClusterPodDefaultSpec{
						Selector: &metav1.LabelSelector{},
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								ServiceAccountName: "foo-user",
							},
						},
					},
				},
				&v1alpha1.ClusterPodDefault{
					ObjectMeta: metav1.ObjectMeta{
						Name: "bar",
					},
					Spec: v1alpha1.ClusterPodDefaultSpec{
						Selector: &metav1.LabelSelector{},
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								ServiceAccountName: "bar-user",
							},
						},
					},
				},
			},
			want: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "bar-user",
					Containers: []corev1.Container{{
						Name:  "foo",
						Image: "python:3.9",
					}},
				},
			},
		},
		"CanPatchContainers": {
			original: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "foo",
						Image: "python:3.9",
					}},
				},
			},
			objects: []client.Object{
				&v1alpha1.ClusterPodDefault{
					ObjectMeta: metav1.ObjectMeta{
						Name: "foo",
					},
					Spec: v1alpha1.ClusterPodDefaultSpec{
						Selector: &metav1.LabelSelector{},
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{
									Name: "foo",
									Env: []corev1.EnvVar{{
										Name:  "foobar",
										Value: "baz",
									}},
								}},
							},
						},
					},
				},
			},
			want: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "foo",
						Image: "python:3.9",
						Env: []corev1.EnvVar{{
							Name:  "foobar",
							Value: "baz",
						}},
					}},
				},
			},
		},
	}

	qt.Assert(t, v1alpha1.AddToScheme(scheme.Scheme), qt.IsNil)

	for name, subtest := range cases {
		t.Run(name, func(t *testing.T) {
			k8s := fake.NewClientBuilder().
				WithScheme(scheme.Scheme).
				WithObjects(subtest.objects...).
				Build()

			got := subtest.original
			qt.Assert(t, Mutate(context.Background(), k8s, got), qt.IsNil)
			qt.Assert(t, got, qt.DeepEquals, subtest.want)
		})
	}
}
