package eksirsa

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/johnhoman/kubeflow-admin/internal/types/awsiam/ack"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/johnhoman/kubeflow-admin/apis/aws/v1alpha1"
	"github.com/johnhoman/kubeflow-admin/internal/features"
)

const (
	errReadServiceAccount       = "failed to read service account"
	errReadRoleConfigClaim      = "failed to read role config claim"
	errParseSessionDuration     = "failed to parse session duration from RoleConfig spec"
	errSetSessionDuration       = "failed to set session duration on Role spec"
	errParseIssuerURL           = "invalid issuer arn, could not parse url"
	errReconcileIAMRole         = "failed to reconcile IAM role for service account"
	errWaitForServiceAccountARN = "waiting for valid iam role arn"

	annotationIamRoleClaim = v1alpha1.Group + "/iam-role-claim"
)

// +kubebuilder:rbac:groups=iam.services.k8s.aws,resources=roles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=iam.services.k8s.aws,resources=roles/status,verbs=get;update;patch

func Setup(mgr ctrl.Manager, o controller.Options) error {
	if !o.Features.Enabled(features.EKSIRSA) {
		logging.NewLogrLogger(mgr.GetLogger()).Debug(
			"IAM roles for service accounts not enabled",
		)
		return nil
	}
	name := fmt.Sprintf("%s/service-account/role", v1alpha1.Group)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&corev1.ServiceAccount{}, builder.WithPredicates(isOwnedByProfile)).
		Owns(ack.NewUnstructuredRole()).
		Watches(&source.Kind{Type: &v1alpha1.RoleConfig{}},
			EnqueueRequestsForServiceAccounts(mgr.GetClient(), o.Logger)).
		WithOptions(o.ForControllerRuntime()).
		Complete(NewReconciler(mgr,
			WithLogger(o.Logger.WithValues("controller", name)),
			WithEventRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		))
}

type ReconcilerOption func(r *Reconciler)

func WithLogger(l logging.Logger) ReconcilerOption {
	return func(r *Reconciler) {
		r.logger = l
	}
}

func WithEventRecorder(er event.Recorder) ReconcilerOption {
	return func(r *Reconciler) {
		r.record = er
	}
}

type manager interface {
	GetClient() client.Client
}

func NewReconciler(mgr manager, opts ...ReconcilerOption) *Reconciler {
	r := &Reconciler{
		client: mgr.GetClient(),
		logger: logging.NewNopLogger(),
		record: event.NewNopRecorder(),
	}
	for _, f := range opts {
		f(r)
	}
	return r
}

type Reconciler struct {
	client client.Client
	logger logging.Logger
	record event.Recorder
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	serviceAccount := &corev1.ServiceAccount{}
	if err := r.client.Get(ctx, req.NamespacedName, serviceAccount); err != nil {
		return ctrl.Result{}, errors.Wrap(client.IgnoreNotFound(err), errReadServiceAccount)
	}

	configName := "default"
	if serviceAccount.Annotations != nil {
		if name := serviceAccount.Annotations[annotationIamRoleClaim]; name != "" {
			configName = name
		}
	}

	rc := &v1alpha1.RoleConfig{}
	rc.SetName(configName)
	rc.SetNamespace(req.Namespace)
	if err := r.client.Get(ctx, client.ObjectKeyFromObject(rc), rc); err != nil {
		return ctrl.Result{}, errors.Wrap(client.IgnoreNotFound(err), errReadRoleConfigClaim)
	}

	policyArns := sets.NewString()
	for key, value := range serviceAccount.Annotations {
		if strings.HasPrefix(key, fmt.Sprintf("policy.%s/", v1alpha1.Group)) {
			policyArns.Insert(value)
		}
	}

	iamRole := ack.NewUnstructuredRole()
	iamRole.SetName(serviceAccount.Name)
	iamRole.SetNamespace(serviceAccount.Namespace)
	res, err := controllerutil.CreateOrPatch(ctx, r.client, iamRole, func() error {
		ownerRef := metav1.NewControllerRef(serviceAccount, corev1.SchemeGroupVersion.WithKind("ServiceAccount"))
		iamRole.SetOwnerReferences([]metav1.OwnerReference{*ownerRef})
		role, err := ack.NewRoleFromUnstructured(iamRole)
		if err != nil {
			return err
		}

		if rc.Spec.MaxSessionDuration != "" {
			dur, err := time.ParseDuration(rc.Spec.MaxSessionDuration)
			if err != nil {
				return errors.Wrap(err, errParseSessionDuration)
			}
			if err := role.SetMaxDurationSeconds(dur); err != nil {
				return errors.Wrap(err, errSetSessionDuration)
			}
		}

		saName := fmt.Sprintf("system:serviceaccount:%s:%s",
			serviceAccount.Namespace,
			serviceAccount.Name,
		)

		if err := role.SetName(strings.ReplaceAll(saName, ":", "-")); err != nil {
			return err
		}

		_, issuerURL, ok := strings.Cut(rc.Spec.Issuer.ARN, "/")
		if !ok {
			return errors.New(errParseIssuerURL)
		}

		statement := NewStatement(
			WithEffectAllow(),
			WithIssuerArn(rc.Spec.Issuer.ARN),
			ForServiceAccount(saName, issuerURL),
			WithAction("sts:AssumeRoleWithWebIdentity"),
		)
		document := NewDocument(WithStatement(statement))
		doc, err := json.Marshal(document)
		if err != nil {
			return err
		}
		if err := role.SetAssumedRolePolicyDocument(string(doc)); err != nil {
			return err
		}
		if err := role.SetPolicies(policyArns.List()); err != nil {
			return err
		}
		desc := fmt.Sprintf("iam role for service account %s", saName)
		if err := role.SetDescription(desc); err != nil {
			return err
		}

		iamRole.SetUnstructuredContent(role.UnstructuredContent())
		return nil
	})
	if err != nil {
		r.logger.Debug(errReconcileIAMRole, "error", err.Error())
		return ctrl.Result{}, errors.Wrap(err, errReconcileIAMRole)
	}
	switch res {
	case controllerutil.OperationResultUpdated, controllerutil.OperationResultCreated:
		r.logger.Debug(fmt.Sprintf("%s iam role %s", res, iamRole.GetName()))
	}

	// This probably won't be populated right away
	role, err := ack.NewRoleFromUnstructured(iamRole)
	if err != nil {
		panic(err.(any))
	}
	roleArn, err := role.Arn()
	if err != nil {
		return ctrl.Result{}, err
	}
	roleId, err := role.Id()
	if err != nil {
		return ctrl.Result{}, err
	}
	if roleArn == "" || roleId == "" {
		return ctrl.Result{}, errors.New(errWaitForServiceAccountARN)
	}

	if !hasRoleAnnotation(serviceAccount, roleArn) || !hasRoleIdAnnotation(serviceAccount, roleId) {
		patch := client.MergeFrom(serviceAccount.DeepCopy())
		addRoleAnnotation(serviceAccount, roleArn)
		addRoleIdAnnotation(serviceAccount, roleId)
		if err := r.client.Patch(ctx, serviceAccount, patch); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

const annotationIamRole = "eks.amazonaws.com/role-arn"
const annotationIamRoleId = "eks.amazonaws.com/role-id"

func hasRoleAnnotation(obj client.Object, arn string) bool {
	if obj.GetAnnotations() != nil {
		return obj.GetAnnotations()[annotationIamRole] == arn
	}
	return false
}

func hasRoleIdAnnotation(obj client.Object, id string) bool {
	if obj.GetAnnotations() != nil {
		return obj.GetAnnotations()[annotationIamRoleId] == id
	}
	return false
}

func addRoleAnnotation(obj client.Object, arn string) {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[annotationIamRole] = arn
	obj.SetAnnotations(annotations)
}

func addRoleIdAnnotation(obj client.Object, id string) {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[annotationIamRoleId] = id
	obj.SetAnnotations(annotations)
}
