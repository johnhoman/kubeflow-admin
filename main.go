package main

import (
	"github.com/alecthomas/kong"
	"github.com/johnhoman/kubeflow-admin/apis/v1alpha1"
	"github.com/johnhoman/kubeflow-admin/internal/webhook"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var cli struct {
	LeaderElection         bool   `help:"enabled leader election on the controller" default:"true"`
	HealthProbeBindAddress string `help:"the address the controller should bind the health probe to" default:"8080"`
	MetricsBindAddress     string `help:"the address the controller should bind the metrics probe to" default:":8081"`
	WebhookPort            int    `help:"the address to bind the webhook server port to" default:"9443"`
}

func main() {
	ctx := kong.Parse(&cli)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		NewCache:               newCache,
		LeaderElection:         cli.LeaderElection,
		HealthProbeBindAddress: cli.HealthProbeBindAddress,
		MetricsBindAddress:     cli.MetricsBindAddress,
		Port:                   cli.WebhookPort,
	})
	ctx.FatalIfErrorf(err, "unable to create controller manager")

	ctx.FatalIfErrorf(err, webhook.Setup(mgr), "failed to setup webhook")
	ctx.FatalIfErrorf(mgr.Start(ctrl.SetupSignalHandler()), "failed to start controller manager")
}

var newCache = cache.BuilderWithOptions(cache.Options{
	SelectorsByObject: map[client.Object]cache.ObjectSelector{
		&corev1.Namespace{}: {
			Label: labels.SelectorFromSet(map[string]string{
				"app.kubernetes.io/part-of": "kubeflow-profile",
			}),
		},
		&v1alpha1.ClusterPodDefault{}: {
			Label: labels.Everything(),
			Field: fields.Everything(),
		},
	},
})
