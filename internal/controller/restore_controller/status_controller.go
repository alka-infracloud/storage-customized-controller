package restorecontroller

import (
	"context"

	storageresourcek8sv1alpha1 "github.com/alka-infracloud/storage-customized-controller/api/v1alpha1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// StatusReconciler reconciles a userRestore object status.
type StatusReconciler struct {
	client.Client
	Logger logr.Logger
	Scheme *runtime.Scheme
}

func (r *StatusReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Logger.Info("Reconcile request", "name", req.Name, "namespace", req.Namespace)

	userRestoreObj := storageresourcek8sv1alpha1.UserRestore{}
	// Get userRestoreInstance.
	err := r.Get(context.Background(), req.NamespacedName, &userRestoreObj)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}

		r.Logger.Error(err, "Failed to get userRestore resource")
		return reconcile.Result{}, err
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
	}

	// Update conditions of userRestore based on the status of Pvc.
	err = r.Get(ctx, client.ObjectKey{Name: pvc.GetName(), Namespace: pvc.GetNamespace()}, pvc)
	if errors.IsNotFound(err) {
		return ctrl.Result{}, nil
	} else if err != nil {
		r.Logger.Error(err, "Failed to get pvc")
		return reconcile.Result{}, err
	}

	userRestoreObj.Status.Phase = string(pvc.Status.Phase)
	userRestoreObj.Status.Conditions = getUpdatedConditions(userRestoreObj.Status.Conditions, pvc.Status.Conditions)

	err = r.Status().Update(ctx, &userRestoreObj)
	if err != nil {
		r.Logger.Error(err, "Failed to update userRestore status")
		return reconcile.Result{}, err
	}

	r.Logger.Info("Successfully updated userRestore status", userRestoreObj.Name, userRestoreObj.Namespace)

	return ctrl.Result{}, nil
}

func getUpdatedConditions(conditions []storageresourcek8sv1alpha1.UserRestoreCondition,
	newConditons []corev1.PersistentVolumeClaimCondition) []storageresourcek8sv1alpha1.UserRestoreCondition {

	conditionsMap := make(map[string]storageresourcek8sv1alpha1.UserRestoreCondition)
	for _, cond := range conditions {
		conditionsMap[string(cond.Type)] = cond
	}

	newConditionsMap := make(map[string]corev1.PersistentVolumeClaimCondition)
	for _, cond := range newConditons {
		newConditionsMap[string(cond.Type)] = cond
	}

	for conditionType, condition := range newConditionsMap {
		restoreCondition := storageresourcek8sv1alpha1.UserRestoreCondition{
			Type:               storageresourcek8sv1alpha1.UserRestoreConditionType(conditionType),
			Status:             storageresourcek8sv1alpha1.RestoreConditionStatus(condition.Status),
			LastTransitionTime: condition.LastTransitionTime,
			Reason:             condition.Reason,
			Message:            condition.Message,
		}

		conditionsMap[conditionType] = restoreCondition
	}

	newConds := make([]storageresourcek8sv1alpha1.UserRestoreCondition, 0, len(conditionsMap))

	// Convert updated conditionsMap to list.
	for _, condition := range conditionsMap {
		newConds = append(newConds, condition)
	}

	return newConds
}

// SetupWithManager sets up the controller with the Manager.
func (r *StatusReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return checkPVCOwnedByUserRestore(e.ObjectNew.(*corev1.PersistentVolumeClaim))
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return checkPVCOwnedByUserRestore(e.Object.(*corev1.PersistentVolumeClaim))
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.PersistentVolumeClaim{}).
		WithEventFilter(pred).
		Named("UserRestoreStatusController").
		Complete(r)
}
