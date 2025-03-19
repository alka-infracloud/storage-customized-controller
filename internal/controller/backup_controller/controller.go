package backupcontroller

import (
	"context"
	"errors"

	storageresourcek8sv1alpha1 "github.com/alka-infracloud/storage-customized-controller/api/v1alpha1"
	"github.com/go-logr/logr"
	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v8/apis/volumesnapshot/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciler reconciles UserBackup instance.
type Reconciler struct {
	client.Client
	Logger logr.Logger
	Scheme *runtime.Scheme
}

// In this function, we handle the reconciliation logic for the UserBackup custom resource.
// It fetches the UserBackup instance specified in the request, and if it exists,
// it proceeds to handle the creation of a VolumeSnapshot associated with the UserBackup instance.
// If the UserBackup instance is not found, it logs the delete event and returns without error.
// If any error occurs during fetching or handling the UserBackup instance, it logs the error and returns it.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Logger.Info("Reconcile request", "name", req.Name, "namespace", req.Namespace)

	userBackupObj := storageresourcek8sv1alpha1.UserBackup{}
	// Get userBackup Instance.
	err := r.Get(context.Background(), req.NamespacedName, &userBackupObj)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			r.Logger.Info("Got delete event: VolumeSnapshot is removed.")
			return reconcile.Result{}, nil
		}

		r.Logger.Error(err, "Failed to get userBackup instance")
		return reconcile.Result{}, err
	}

	if err := r.handleCreateEvent(ctx, userBackupObj); err != nil {
		r.Logger.Error(err, "Failed to create VolumeSnapshot for userBackup instance")

		// Update status of UserBackup instance with failure.
		userBackupObj.Status.Conditions = getUpdatedConditions(
			userBackupObj.Status.Conditions,
			storageresourcek8sv1alpha1.UserBackupCondition{
				Type:               storageresourcek8sv1alpha1.UserBackupConditionReady,
				Status:             storageresourcek8sv1alpha1.ConditionFalse,
				LastTransitionTime: metav1.Now(),
				Reason:             volumeSnapshotAlreadyExistErr,
				Message:            "VolumeSnapshot is not ready",
			},
		)

		statusErr := r.Status().Update(ctx, &userBackupObj)
		if statusErr != nil {
			r.Logger.Error(statusErr, "Failed to update UserBackup status with VolumeSnapshot creation failure")
		}

		return reconcile.Result{}, err
	}

	return ctrl.Result{}, nil
}

// handleCreateEvent handles the creation of a VolumeSnapshot for a given UserBackup instance.
// It performs the following steps:
// 1. Checks if a VolumeSnapshot with the given name already exists.
//   - If it exists and is owned by the UserBackup instance, it returns nil.
//   - If it exists but is not owned by the UserBackup instance, it returns an error indicating that the VolumeSnapshot already exists and is not owned by the UserBackup.
//
// 2. If the VolumeSnapshot does not exist, it creates a new VolumeSnapshot with the specified SnapshotClassName and PvcName from the UserBackup spec.
// 3. Sets the owner reference for the VolumeSnapshot to ensure it is cleaned up when the UserBackup custom resource is deleted.
// 4. Sets blockOwnerDeletion to false to allow child deletion even when the parent exists.
// 5. Creates the VolumeSnapshot and logs the success or failure of the operation.
func (r *Reconciler) handleCreateEvent(ctx context.Context, backupObj storageresourcek8sv1alpha1.UserBackup) error {
	volumeSnapshot := &snapshotv1.VolumeSnapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      backupObj.GetName(),
			Namespace: backupObj.GetNamespace(),
		},
	}

	err := r.Get(ctx, client.ObjectKey{Name: volumeSnapshot.GetName(), Namespace: volumeSnapshot.GetNamespace()}, volumeSnapshot)
	if err == nil {
		if checkVolumeSnapshotOwnedByUserBackup(volumeSnapshot) {
			return nil
		}

		r.Logger.Error(err, volumeSnapshotAlreadyExistErr)
		return errors.New(volumeSnapshotAlreadyExistErr)
	}

	if k8serrors.IsNotFound(err) {
		// Create VolumeSnapshotCR.
		snapshotClassName := backupObj.Spec.SnapshotClassName
		pvcName := backupObj.Spec.PvcName

		volumeSnapshot.Spec = snapshotv1.VolumeSnapshotSpec{
			Source: snapshotv1.VolumeSnapshotSource{
				PersistentVolumeClaimName: &pvcName,
				// TODO: Will add support of following optional params later. +optional, VolumeSnapshotContentName
			},
			VolumeSnapshotClassName: &snapshotClassName,
		}

		// Set the owner reference for the VolumeSnapshot.
		err := controllerutil.SetControllerReference(&backupObj, volumeSnapshot, r.Scheme)
		if err != nil {
			r.Logger.Error(err, "unable to set owner reference on VolumeSnapshot")
			return err
		}

		// Set blockOwnerDeletion to false.
		for i := range volumeSnapshot.OwnerReferences {
			volumeSnapshot.OwnerReferences[i].BlockOwnerDeletion = &[]bool{false}[0]
		}

		err = r.Create(ctx, volumeSnapshot)
		if err != nil {
			r.Logger.Error(err, "Failed to create VolumeSnapshot")
			return err
		}

		r.Logger.Info("Successfully created VolumeSnapshot", volumeSnapshot.Name, volumeSnapshot.Namespace)
	} else if err != nil {
		r.Logger.Error(err, "Failed to get VolumeSnapshot")
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
// Watching on UserBackup instance.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&storageresourcek8sv1alpha1.UserBackup{}).
		Named("UserBackupController").
		Complete(r)
}
