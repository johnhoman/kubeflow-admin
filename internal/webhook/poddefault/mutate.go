package poddefault

import (
    "context"
    "reflect"
    "sort"

    "github.com/johnhoman/kubeflow-admin/apis/v1alpha1"
    "github.com/pkg/errors"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/labels"
    "k8s.io/apimachinery/pkg/runtime"
    utilerrors "k8s.io/apimachinery/pkg/util/errors"
    "k8s.io/apimachinery/pkg/util/strategicpatch"
    "k8s.io/utils/pointer"
    "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
    errPodDefaultList = "could not list pod defaults"
    errConvertFromPod = "failed to convert pod to unstructured"
    errConvertToPod   = "failed to convert pod from unstructured"
)

func Mutate(ctx context.Context, reader client.Reader, pod *corev1.Pod) error {
    podDefaultList := &v1alpha1.ClusterPodDefaultList{}
    if err := reader.List(ctx, podDefaultList); err != nil {
        return errors.Wrap(err, errPodDefaultList)
    }

    errs := make([]error, 0)
    defaults := make([]*v1alpha1.ClusterPodDefault, 0)
    for _, item := range podDefaultList.Items {
        selector, err := metav1.LabelSelectorAsSelector(item.Spec.Selector)
        if err != nil {
            errs = append(errs, err)
            continue
        }
        if selector.Matches(labels.Set(pod.Labels)) {
            defaults = append(defaults, item.DeepCopy())
        }
    }
    if len(errs) > 0 {
        return utilerrors.NewAggregate(errs)
    }

    sort.Slice(defaults, func(i, j int) bool {
        // Sort in reverse alphabetical order, so that the pod defaults
        // sorted lexicographically will take priority
        return defaults[i].Name > defaults[j].Name
    })
    sort.Slice(defaults, func(i, j int) bool {
        pr1 := defaults[i].Spec.Priority
        pr2 := defaults[j].Spec.Priority
        if pr1 == nil {
            pr1 = pointer.Int(-1)
        }
        if pr2 == nil {
            pr2 = pointer.Int(-1)
        }
        return *pr1 < *pr2
    })

    podMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(pod)
    if err != nil {
        return errors.Wrap(err, errConvertToPod)
    }
    errs = make([]error, 0)
    for _, def := range defaults {

        from, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&def.Spec.Template)
        if err != nil {
            errs = append(errs, err)
            continue
        }
        removeNil(from)

        schema, err := strategicpatch.NewPatchMetaFromStruct(&corev1.PodTemplateSpec{})
        if err != nil {
            errs = append(errs, err)
            continue
        }

        podMap, err = strategicpatch.StrategicMergeMapPatchUsingLookupPatchMeta(podMap, from, schema)
        if err != nil {
            errs = append(errs, err)
            continue
        }
    }
    if len(errs) > 0 {
        return utilerrors.NewAggregate(errs)
    }
    if err := runtime.DefaultUnstructuredConverter.FromUnstructured(podMap, pod); err != nil {
        return errors.Wrap(err, errConvertFromPod)
    }
    return nil
}

// removeNil removes all nil fields from a map. If the nil fields
// are left, they can cause unwanted deletions in a pod spec when merging
func removeNil(m map[string]any) {
    rv := reflect.ValueOf(m)
    for _, key := range rv.MapKeys() {
        value := rv.MapIndex(key)
        if value.IsNil() {
            delete(m, key.String())
            continue
        }
        switch t := value.Interface().(type) {
        case map[string]any:
            removeNil(t)
        case []any:
            for _, lm := range t {
                m, ok := lm.(map[string]any)
                if ok {
                    removeNil(m)
                }
            }
        }
    }
}