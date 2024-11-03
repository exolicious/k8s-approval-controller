package controller

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"

	acpagchv1 "acp.ag.ch/approval-k8s-controller/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ApprovalReconciler reconciles a Approval object
type ApprovalReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=acp.ag.ch,resources=approvals,verbs=get;update;patch
// +kubebuilder:rbac:groups=acp.ag.ch,resources=approvals/status,verbs=get;update;patch
func (r *ApprovalReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("RECONCILING")
	approval := &acpagchv1.Approval{}
	if err := r.Get(ctx, req.NamespacedName, approval); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Initialize status if it's a new resource
	if approval.Status.State == "" {
		approval.Status.State = "Pending"
		if err := r.Status().Update(ctx, approval); err != nil {
			logger.Error(err, "unable to update Approval status")
			return ctrl.Result{}, err
		}
	}

	if approval.Status.State == "Approved" && approval.Status.ApprovedResource == nil {
		logger.Info("Status is true and no active resource yet")
		now := metav1.NewTime(time.Now())
		approval.Status.DecisionTime = &now

		obj := &unstructured.Unstructured{}
		deserializer := json.NewSerializerWithOptions(json.DefaultMetaFactory, r.Scheme, r.Scheme, json.SerializerOptions{Yaml: false, Pretty: false, Strict: false})
		_, _, err := deserializer.Decode(approval.Spec.ResourceSpec.Raw, nil, obj)
		if err != nil {
			logger.Error(err, "unable to decode RawExtension to unstructured object")
			return ctrl.Result{}, err
		}
		if obj.GetNamespace() == "" {
			obj.SetNamespace(req.Namespace)
		}

		if err := r.Create(ctx, obj); err != nil {
			logger.Error(err, "unable to create resource from RawExtension")
			return ctrl.Result{}, err
		}

		approval.Status.ApprovedResource = &corev1.ObjectReference{
			APIVersion: obj.GetAPIVersion(),
			Kind:       obj.GetKind(),
			Name:       obj.GetName(),
			Namespace:  obj.GetNamespace(),
		}

		if err := r.Status().Update(ctx, approval); err != nil {
			logger.Error(err, "unable to update approval status with active resource reference")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *ApprovalReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&acpagchv1.Approval{}).
		//WithEventFilter(onlyChangeOnStatusApprovedChange()).
		Complete(r)
}
