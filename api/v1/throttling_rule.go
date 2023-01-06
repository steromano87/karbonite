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
//+kubebuilder:printcolumn:name="Active",type="boolean",JSONPath=".status.active",description="Whether a re-entrant schedule is defined"
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

	//+kubebuilder:validation:MinItems:=1
	Matchers []Selector `json:"matchers"`

	//+kubebuilder:validation:MinItems:=1
	Schedules []ThrottlingSchedule `json:"schedules"`
}

type ThrottlingSchedule struct {
	ReentrantSchedule `json:",inline"`

	//+kubebuilder:validation:minValue:=0
	DesiredReplicas int `json:"desiredReplicas"`
}

// ThrottlingRuleStatus defines the observed state of ThrottlingRule
type ThrottlingRuleStatus struct {
	//+kubebuilder:default:=false
	Active bool `json:"active,omitempty"`
}

func init() {
	SchemeBuilder.Register(&ThrottlingRule{}, &ThrottlingRuleList{})
}
