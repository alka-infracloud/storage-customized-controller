package restorecontroller

import corev1 "k8s.io/api/core/v1"

const (
	storageApiVersion  = "storageresource.k8s.infracloud/v1alpha1"
	userRestoreKind    = "UserRestore"
	pvcAlreadyExistErr = "Restored pvc with provided name already exists and not ownded by UserRestore instance"
)

// checkPVCOwnedByUserRestore checks if the given PVC is owned by a UserRestore instance.
func checkPVCOwnedByUserRestore(pvc *corev1.PersistentVolumeClaim) bool {
	if pvc.GetOwnerReferences() == nil {
		return false
	}

	for _, ownerRef := range pvc.GetOwnerReferences() {
		if ownerRef.APIVersion == storageApiVersion && ownerRef.Kind == userRestoreKind {
			return true
		}
	}
	return false
}
