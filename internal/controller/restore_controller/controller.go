package restorecontroller

import (
	"context"
	"errors"
	"fmt"

	storageresourcek8sv1alpha1 "github.com/alka-infracloud/storage-customized-controller/api/v1alpha1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciler reconciles a UserRestore instance.
type Reconciler struct {
	client.Client
	Logger logr.Logger
	Scheme *runtime.Scheme
}

// In this function, we fetch the UserRestore instance, check if it exists,
// and handle the creation of a PersistentVolumeClaim (PVC) based on the
// UserRestore specification. If the UserRestore instance is deleted, it logs
// the delete event and exits gracefully. If any error occurs during the process,
// it logs the error and requeues the request for further processing.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	r.Logger.Info("Reconcile request", "name", req.Name, "namespace", req.Namespace)

	userRestoreObj := storageresourcek8sv1alpha1.UserRestore{}
	// Get UserRestore Instance.
	err := r.Client.Get(context.Background(), req.NamespacedName, &userRestoreObj)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			r.Logger.Info("Got delete event: PVC is removed.")
			return reconcile.Result{}, nil
		}

		r.Logger.Error(err, "Failed to get UserBackup instance")
		return reconcile.Result{}, err
	}

	if err := r.handleCreateEvent(ctx, userRestoreObj); err != nil {
		r.Logger.Error(err, "Failed to handle create PVC for UserRestore instance")

		// Update status of userRestoreObj instance with failure.
		userRestoreObj.Status.Phase = "unknown"
		userRestoreObj.Status.Conditions = []storageresourcek8sv1alpha1.UserRestoreCondition{
			{
				Type:               storageresourcek8sv1alpha1.UserRestoreConditionReady,
				Status:             storageresourcek8sv1alpha1.RestoreConditionFalse,
				LastTransitionTime: metav1.Now(),
				Reason:             pvcAlreadyExistErr,
				Message:            "PVC is not ready",
			},
		}

		statusErr := r.Status().Update(ctx, &userRestoreObj)
		fmt.Println("DBG ALKA ", statusErr, userRestoreObj)

		if statusErr != nil {
			r.Logger.Error(statusErr, "Failed to update userRestore status with PVC creation failure")
		}

		return reconcile.Result{}, err
	}

	return ctrl.Result{}, nil
}

// handleCreateEvent handles the creation of a PersistentVolumeClaim (PVC) for the given UserRestore object.
// It first verifies the existence of the corresponding UserBackup instance.
func (r *Reconciler) handleCreateEvent(ctx context.Context, restoreObj storageresourcek8sv1alpha1.UserRestore) error {
	// Verify UserBackup instance exists or not.
	userBackupObj := storageresourcek8sv1alpha1.UserBackup{}
	err := r.Get(context.Background(), client.ObjectKey{Name: restoreObj.Spec.UserBackUpName, Namespace: restoreObj.GetNamespace()}, &userBackupObj)
	if err != nil {
		r.Logger.Error(err, "Failed to get UserBackup instance")
		return err
	}

	// Create PVC.
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      restoreObj.GetName(),
			Namespace: restoreObj.GetNamespace(),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.PersistentVolumeAccessMode(userBackupObj.Status.PvcAccessMode)},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(userBackupObj.Status.RestoreSize),
				},
			},
			StorageClassName: &userBackupObj.Status.PvcStorageClassName,
			DataSource: &corev1.TypedLocalObjectReference{
				APIGroup: &[]string{"snapshot.storage.k8s.io"}[0],
				Kind:     []string{"VolumeSnapshot"}[0],
				Name:     userBackupObj.Name,
			},
		},
	}

	if restoreObj.Spec.AccessMode != "" {
		pvc.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{corev1.PersistentVolumeAccessMode(restoreObj.Spec.AccessMode)}
	}

	err = r.Get(ctx, client.ObjectKey{Name: pvc.GetName(), Namespace: pvc.GetNamespace()}, pvc)
	if err == nil {
		if checkPVCOwnedByUserRestore(pvc) {
			return nil
		}

		r.Logger.Error(err, pvcAlreadyExistErr)
		return errors.New(pvcAlreadyExistErr)
	}

	// Set the owner reference for the PVC.
	// This ensures that the PVC will be cleaned up when the UserRestore custom resource is deleted.
	err = controllerutil.SetControllerReference(&restoreObj, pvc, r.Scheme)
	if err != nil {
		r.Logger.Error(err, "unable to set owner reference on pvc")
		return err
	}

	// Set blockOwnerDeletion to false to allow child deletion even when parent exists.
	for i := range pvc.OwnerReferences {
		pvc.OwnerReferences[i].BlockOwnerDeletion = &[]bool{false}[0]
	}

	err = r.Create(ctx, pvc)
	if err != nil {
		r.Logger.Error(err, "Failed to create restored pvc")
		return err
	}

	r.Logger.Info("Successfully created restored pvc", pvc.Name, pvc.Namespace)

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&storageresourcek8sv1alpha1.UserRestore{}).
		Named("UserRestoreController").
		Complete(r)
}
