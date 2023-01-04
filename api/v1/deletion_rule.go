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

// DeletionRuleList contains a list of DeletionRule
type DeletionRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DeletionRule `json:"items"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Enabled",type="boolean",JSONPath=".spec.enabled",description="Whether the DeletionRule is enforced or not"
//+kubebuilder:printcolumn:name="Schedules",type="string",priority=1,JSONPath=".spec.schedules[*]",description="The active schedules"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// DeletionRule is the Schema for the deletionrules API
type DeletionRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeletionRuleSpec   `json:"spec,omitempty"`
	Status DeletionRuleStatus `json:"status,omitempty"`
}

// DeletionRuleSpec defines the desired state of DeletionRule
type DeletionRuleSpec struct {
	//+kubebuilder:default:=true
	Enabled bool `json:"enabled,omitempty"`

	//+kubebuilder:default:=false
	DryRun bool `json:"dryRun,omitempty"`

	//+kubebuilder:validation:MinItems:=1
	Matchers []Matchers `json:"matchers"`

	//+kubebuilder:validation:MinItems:=1
	Schedules []string `json:"schedules"`
}

// DeletionRuleStatus defines the observed state of DeletionRule
type DeletionRuleStatus struct {
	LastModified metav1.Time `json:"lastModified,omitempty"`
}

func init() {
	SchemeBuilder.Register(&DeletionRule{}, &DeletionRuleList{})
}
