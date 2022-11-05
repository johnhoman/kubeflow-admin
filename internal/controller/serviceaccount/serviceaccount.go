package serviceaccount

import (
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/johnhoman/kubeflow-admin/internal/controller/serviceaccount/imagepullsecrets"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Setup reconcilers for profile service accounts
func Setup(mgr ctrl.Manager, o controller.Options) error {
	funcs := []func(mgr ctrl.Manager, options controller.Options) error{
		imagepullsecrets.Setup,
	}

	for _, f := range funcs {
		if err := f(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
