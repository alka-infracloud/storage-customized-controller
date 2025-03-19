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

// UserRestoreSpec defines the desired state of UserRestore.
type UserRestoreSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Immutable
	UserBackUpName string `json:"userBackUpName"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Immutable
	AccessMode string `json:"accessMode"`
}

// UserRestoreStatus defines the observed state of UserRestore.
type UserRestoreStatus struct {
	Phase      string                 `json:"phase"`
	Conditions []UserRestoreCondition `json:"conditions,omitempty"`
}

// UserRestoreCondition describes the state of a userbackup object at a certain point.
type UserRestoreCondition struct {
	Type UserRestoreConditionType `json:"type" protobuf:"bytes,1,opt,name=type"`
	// Status of the condition, one of True, False, Unknown.
	Status RestoreConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status"`
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

type UserRestoreConditionType string

const (
	UserRestoreConditionReady UserRestoreConditionType = "Ready"
)

type RestoreConditionStatus string

const (
	RestoreConditionTrue  RestoreConditionStatus = "True"
	RestoreConditionFalse RestoreConditionStatus = "False"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// UserRestore is the Schema for the userrestores API.
type UserRestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserRestoreSpec   `json:"spec,omitempty"`
	Status UserRestoreStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// UserRestoreList contains a list of UserRestore.
type UserRestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UserRestore `json:"items"`
}

func init() {
	SchemeBuilder.Register(&UserRestore{}, &UserRestoreList{})
}
