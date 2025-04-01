package backupcontroller

import (
	"context"

	storageresourcek8sv1alpha1 "github.com/alka-infracloud/storage-customized-controller/api/v1alpha1"
	"github.com/go-logr/logr"
	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v8/apis/volumesnapshot/v1"
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

// StatusReconciler reconciles a UserBackup instance status.
type StatusReconciler struct {
	client.Client
	Logger logr.Logger
	Scheme *runtime.Scheme
}

// In this function, it fetches the UserBackup instance specified in the request,
// checks the status of the associated VolumeSnapshot, and updates the UserBackup
// status accordingly. And populates other status field as well when VolumeSnapshot is ready to use.
// If the snapshotis not ready, it updates the status with the appropriate error message.
// Finally, it updates the status of the UserBackup resource in the Kubernetes API.
func (r *StatusReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Logger.Info("Reconcile request", "name", req.Name, "namespace", req.Namespace)

	userBackupObj := storageresourcek8sv1alpha1.UserBackup{}
	// Get UserBackup instance.
	err := r.Get(context.Background(), req.NamespacedName, &userBackupObj)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}

		r.Logger.Error(err, "Failed to get UserBackup instance")
		return reconcile.Result{}, err
	}

	volumeSnapshot := &snapshotv1.VolumeSnapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
	}

	// Update conditions of UserBackup based on the status of VolumeSnapshot.
	err = r.Get(ctx, client.ObjectKey{Name: volumeSnapshot.GetName(), Namespace: volumeSnapshot.GetNamespace()}, volumeSnapshot)
	if errors.IsNotFound(err) {
		return ctrl.Result{}, nil
	} else if err != nil {
		r.Logger.Error(err, "Failed to get VolumeSnapshot")
	}

	if volumeSnapshot.Status != nil && volumeSnapshot.Status.ReadyToUse != nil {
		if *volumeSnapshot.Status.ReadyToUse {
			userBackupObj.Status.Conditions = getUpdatedConditions(
				userBackupObj.Status.Conditions,
				storageresourcek8sv1alpha1.UserBackupCondition{
					Type:               storageresourcek8sv1alpha1.UserBackupConditionReady,
					Status:             storageresourcek8sv1alpha1.ConditionTrue,
					LastTransitionTime: metav1.Now(),
					Message:            "VolumeSnapshot is ready",
				},
			)

			// Get underlying PVC details to populated in userBackupObj status.
			pvc := corev1.PersistentVolumeClaim{}
			err = r.Get(ctx, client.ObjectKey{Name: *volumeSnapshot.Spec.Source.PersistentVolumeClaimName, Namespace: volumeSnapshot.GetNamespace()}, &pvc)
			if err != nil {
				r.Logger.Error(err, "Failed to get PersistentVolumeClaim")
				return reconcile.Result{}, err
			}

			if len(pvc.Spec.AccessModes) > 0 {
				userBackupObj.Status.PvcAccessMode = string(pvc.Spec.AccessModes[0])
			}

			if pvc.Spec.StorageClassName != nil {
				userBackupObj.Status.PvcStorageClassName = *pvc.Spec.StorageClassName
			}

			quantity := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
			userBackupObj.Status.RestoreSize = quantity.String()

		} else {
			var message string
			if volumeSnapshot.Status.Error != nil && volumeSnapshot.Status.Error.Message != nil {
				message = *volumeSnapshot.Status.Error.Message
			}

			userBackupObj.Status.Conditions = getUpdatedConditions(
				userBackupObj.Status.Conditions,
				storageresourcek8sv1alpha1.UserBackupCondition{
					Type:               storageresourcek8sv1alpha1.UserBackupConditionReady,
					Status:             storageresourcek8sv1alpha1.ConditionFalse,
					LastTransitionTime: metav1.Now(),
					Reason:             message,
					Message:            "VolumeSnapshot is not ready",
				},
			)
		}
	} else {
		return ctrl.Result{}, nil
	}

	err = r.Status().Update(ctx, &userBackupObj)
	if err != nil {
		r.Logger.Error(err, "Failed to update UserBackup status")
		return reconcile.Result{}, err
	}

	r.Logger.Info("Successfully updated UserBackup status", volumeSnapshot.Name, volumeSnapshot.Namespace)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StatusReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return checkVolumeSnapshotOwnedByUserBackup(e.ObjectNew.(*snapshotv1.VolumeSnapshot))
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return checkVolumeSnapshotOwnedByUserBackup(e.Object.(*snapshotv1.VolumeSnapshot))
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&snapshotv1.VolumeSnapshot{}).
		WithEventFilter(pred).
		Named("UserBackupStatusController").
		Complete(r)
}
