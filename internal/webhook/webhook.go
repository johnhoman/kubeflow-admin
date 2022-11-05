package webhook

import (
    "github.com/crossplane/crossplane-runtime/pkg/event"
    "github.com/crossplane/crossplane-runtime/pkg/logging"
    "github.com/johnhoman/kubeflow-admin/internal/webhook/pod"
    "github.com/johnhoman/kubeflow-admin/internal/webhook/poddefault"
    corev1 "k8s.io/api/core/v1"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func Setup(mgr ctrl.Manager) error {
    mgr.GetWebhookServer().Register("", &admission.Webhook{
        Handler: pod.NewHandler(
            pod.WithLogger(logging.NewLogrLogger(mgr.GetLogger().WithValues("webhook", "PodDefault"))),
            pod.WithEventRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor("PodDefaultWebhook"))),
            pod.WithReader(mgr.GetClient()),
            pod.WithMutateFunc(poddefault.Mutate),
            pod.WithPredicate(func(p *corev1.Pod) bool {
                return true
            }),
        ),
    })
    return nil
}
