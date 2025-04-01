/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// UserBackupSpec defines the desired state of UserBackup.
type UserBackupSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Immutable
	PvcName string `json:"pvcName,omitempty"` // Immutable field
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Immutable
	SnapshotClassName string `json:"snapshotClassName,omitempty"` // Immutable field
}

// UserBackupStatus defines the observed state of UserBackup.
type UserBackupStatus struct {
	PvcAccessMode       string                `json:"pvcAccessMode,omitempty"`
	PvcStorageClassName string                `json:"pvcStorageClassName,omitempty"`
	RestoreSize         string                `json:"restoreSize,omitempty"`
	Conditions          []UserBackupCondition `json:"conditions,omitempty"`
}

// UserBackupCondition describes the state of a userbackup object at a certain point.
type UserBackupCondition struct {
	Type UserBackupConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=UserBackupConditionType"`
	// Status of the condition, one of True, False, Unknown.
	Status ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=ConditionStatus"`
	// The last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,3,opt,name=lastTransitionTime"`
	// The reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,4,opt,name=reason"`
	// A human readable message indicating details about the transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,5,opt,name=message"`
}

type UserBackupConditionType string

const (
	UserBackupConditionReady UserBackupConditionType = "Ready"
)

type ConditionStatus string

const (
	ConditionTrue  ConditionStatus = "True"
	ConditionFalse ConditionStatus = "False"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// UserBackup is the Schema for the userbackups API.
type UserBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserBackupSpec   `json:"spec,omitempty"`
	Status UserBackupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// UserBackupList contains a list of UserBackup.
type UserBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UserBackup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&UserBackup{}, &UserBackupList{})
}
