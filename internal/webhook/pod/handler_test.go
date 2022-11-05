package pod

import (
    "context"
    "encoding/json"
    "testing"

    qt "github.com/frankban/quicktest"
    "gomodules.xyz/jsonpatch/v2"
    v1 "k8s.io/api/admission/v1"
    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/client-go/kubernetes/scheme"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func TestHandler_Handle(t *testing.T) {
    cases := map[string]struct{
        pod  *corev1.Pod
        mutateFunc MutateFunc
        want []jsonpatch.JsonPatchOperation
    } {
        "ShouldPatchAPod": {
            pod: &corev1.Pod{},
            mutateFunc: func(ctx context.Context, _ client.Reader, pod *corev1.Pod) error {
                if pod.Annotations == nil {
                    pod.Annotations = make(map[string]string)
                }
                pod.Annotations["foo"] = "bar"
                return nil
            },
            want: []jsonpatch.JsonPatchOperation{{
                Operation: "add",
                Path: "/metadata/annotations",
                Value: map[string]interface{}{"foo": "bar"},
            }},
        },
    }

    for name, subtest := range cases {
        t.Run(name, func(t *testing.T) {

            raw, err := json.Marshal(subtest.pod)
            qt.Assert(t, err, qt.IsNil)

            h := NewHandler(WithMutateFunc(subtest.mutateFunc))
            decoder, err := admission.NewDecoder(scheme.Scheme)
            qt.Assert(t, err, qt.IsNil)
            qt.Assert(t, h.InjectDecoder(decoder), qt.IsNil)
            resp := h.Handle(context.Background(), admission.Request{AdmissionRequest: v1.AdmissionRequest{
                Object: runtime.RawExtension{Raw: raw},
            }})
            qt.Assert(t, resp.Patches, qt.DeepEquals, subtest.want)
        })
    }
}
