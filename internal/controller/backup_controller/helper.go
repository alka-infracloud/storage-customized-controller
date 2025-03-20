package backupcontroller

import (
	storageresourcek8sv1alpha1 "github.com/alka-infracloud/storage-customized-controller/api/v1alpha1"
	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v8/apis/volumesnapshot/v1"
)

const (
	storageApiVersion             = "storageresource.k8s.infracloud/v1alpha1"
	userBackupKind                = "UserBackup"
	volumeSnapshotAlreadyExistErr = "VolumeSnapshot with provided name already exists and not ownded by UserBackup instance"
)

// checkVolumeSnapshotOwnedByUserBackup checks if the given VolumeSnapshot is owned by a UserBackup instance.
func checkVolumeSnapshotOwnedByUserBackup(snapshot *snapshotv1.VolumeSnapshot) bool {
	if snapshot.GetOwnerReferences() == nil {
		return false
	}

	for _, ownerRef := range snapshot.GetOwnerReferences() {
		if ownerRef.APIVersion == storageApiVersion && ownerRef.Kind == userBackupKind {
			return true
		}
	}
	return false
}

// getUpdatedConditions will overwrite the exiting conditions with the update/Add set of conditions.
func getUpdatedConditions(conds []storageresourcek8sv1alpha1.UserBackupCondition,
	newCond storageresourcek8sv1alpha1.UserBackupCondition) []storageresourcek8sv1alpha1.UserBackupCondition {
	var newConds []storageresourcek8sv1alpha1.UserBackupCondition

	isNewCondition := true
	for _, cond := range conds {
		if cond.Type == newCond.Type {
			newConds = append(newConds, newCond)
			isNewCondition = false
		} else {
			newConds = append(newConds, cond)
		}
	}

	if isNewCondition {
		newConds = append(newConds, newCond)
	}

	return newConds
}
