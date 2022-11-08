package controller

import (
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/johnhoman/kubeflow-admin/internal/controller/awss3bucket"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/johnhoman/kubeflow-admin/internal/controller/clusterconfigmap"
	"github.com/johnhoman/kubeflow-admin/internal/controller/clustersecret"
	"github.com/johnhoman/kubeflow-admin/internal/controller/eksirsa"
	"github.com/johnhoman/kubeflow-admin/internal/controller/imagepullsecrets"
)

// Setup reconcilers for profile service accounts
func Setup(mgr ctrl.Manager, o controller.Options) error {
	funcs := []func(mgr ctrl.Manager, options controller.Options) error{
		awss3bucket.Setup,
		clusterconfigmap.Setup,
		clustersecret.Setup,
		eksirsa.Setup,
		imagepullsecrets.Setup,
	}

	for _, f := range funcs {
		if err := f(mgr, o); err != nil {
			return err
		}
	}

	return nil
}
