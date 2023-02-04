/*
Copyright 2022.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:object:root=true

// ThrottlingRuleList contains a list of ThrottlingRule
type ThrottlingRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ThrottlingRule `json:"items"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Enabled",type="boolean",JSONPath=".spec.enabled",description="Whether the ThrottlingRule is enforced or not"
//+kubebuilder:printcolumn:name="Dry-run",type="boolean",JSONPath=".spec.dryRun",description="Whether the DeletionRule runs in dry-run mode (i.e. only logging affected resources)"
//+kubebuilder:printcolumn:name="Schedules",type="string",priority=1,JSONPath=".spec.schedules[*]",description="The active schedules"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// ThrottlingRule is the Schema for the throttlingrules API
type ThrottlingRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ThrottlingRuleSpec   `json:"spec,omitempty"`
	Status ThrottlingRuleStatus `json:"status,omitempty"`
}

// ThrottlingRuleSpec defines the desired state of ThrottlingRule
type ThrottlingRuleSpec struct {
	//+kubebuilder:default:=true
	Enabled bool `json:"enabled,omitempty"`

	//+kubebuilder:default:= false
	DryRun bool `json:"dryRun,omitempty"`

	//+kubebuilder:default:={matchKinds:{Deployment,StatefulSet}}
	Selector Selector `json:"selector"`

	//+kubebuilder:validation:MinItems:=1
	Schedules []ThrottlingSchedule `json:"schedules"`
}

type ThrottlingSchedule struct {
	ReentrantSchedule `json:",inline"`

	//+kubebuilder:validation:Minimum:=0
	DesiredReplicas int `json:"desiredReplicas"`
}

// ThrottlingRuleStatus defines the observed state of ThrottlingRule
type ThrottlingRuleStatus struct {
	NextRun metav1.Time `json:"nextRun,omitempty"`

	//+kubebuilder:default:=0
	RunCount int `json:"runCount"`

	//+kubebuilder:validation:Optional
	LastRun *LastRunInfo `json:"lastRun,omitempty"`

	//+kubebuilder:validation:Optional
	ActiveReentrantThrottle *ActiveReentrantThrottle `json:"activeReentrantThrottle,omitempty"`
}

type ActiveReentrantThrottle struct {
	//+kubebuilder:validation:Type:=array
	AffectedResources []AffectedResource `json:"affectedResources,omitempty"`

	//+kubebuilder:validation:Optional
	ReentrantOn metav1.Time `json:"reentrantOn,omitempty"`
}

func init() {
	SchemeBuilder.Register(&ThrottlingRule{}, &ThrottlingRuleList{})
}
