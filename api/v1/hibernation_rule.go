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

// HibernationRuleList contains a list of HibernationRule
type HibernationRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HibernationRule `json:"items"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// HibernationRule is the Schema for the hibernationrules API
type HibernationRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HibernationRuleSpec   `json:"spec,omitempty"`
	Status HibernationRuleStatus `json:"status,omitempty"`
}

// HibernationRuleSpec defines the desired state of HibernationRule
type HibernationRuleSpec struct {
	//+kubebuilder:default:=true
	Enabled bool `json:"enabled,omitempty"`

	//+kubebuilder:default:= false
	DryRun bool `json:"dryRun,omitempty"`

	//+kubebuilder:validation:MinItems:=1
	Matchers []Matchers `json:"matchers"`

	//+kubebuilder:validation:MinItems:=1
	Schedules []ReentrantSchedule `json:"schedules"`
}

// HibernationRuleStatus defines the observed state of HibernationRule
type HibernationRuleStatus struct {
	LastModified metav1.Time `json:"lastModified,omitempty"`
}

func init() {
	SchemeBuilder.Register(&HibernationRule{}, &HibernationRuleList{})
}
