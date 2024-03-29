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
//+kubebuilder:resource:shortName=dlr
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Enabled",type="boolean",JSONPath=".spec.enabled",description="Whether the DeletionRule is enforced or not"
//+kubebuilder:printcolumn:name="Dry-run",type="boolean",JSONPath=".spec.dryRun",description="Whether the DeletionRule runs in dry-run mode (i.e. only logging affected resources)"
//+kubebuilder:printcolumn:name="Schedules",type="string",priority=1,JSONPath=".spec.schedules[*]",description="The active schedules"
//+kubebuilder:printcolumn:name="Last run",type="string",format="date",JSONPath=".status.lastRun.timestamp",description="Last run date"
//+kubebuilder:printcolumn:name="Run count",type="integer",JSONPath=".status.runCount",description="Total runs of the rule"
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

	Selector Selector `json:"selector"`

	//+kubebuilder:validation:MinItems:=1
	Schedules []string `json:"schedules"`
}

// DeletionRuleStatus defines the observed state of DeletionRule
type DeletionRuleStatus struct {
	NextRun metav1.Time `json:"nextRun,omitempty"`

	//+kubebuilder:default:=0
	RunCount int `json:"runCount"`

	//+kubebuilder:validation:Optional
	LastRun *LastRunInfo `json:"lastRun,omitempty"`
}

func init() {
	SchemeBuilder.Register(&DeletionRule{}, &DeletionRuleList{})
}
