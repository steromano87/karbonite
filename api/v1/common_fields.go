package v1

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type LastRunInfo struct {
	//+kubebuilder:validation:Optional
	Timestamp *metav1.Time `json:"timestamp,omitempty"`

	//+kubebuilder:validation:Type:=array
	//+kubebuilder:validation:Optional
	AffectedResources []AffectedResource `json:"affectedResources,omitempty"`
}

type AffectedResource struct {
	Namespace           string               `json:"namespace"`
	Resource            string               `json:"resource"`
	Kind                string               `json:"kind"`
	ResourceScalingSpec *ResourceScalingSpec `json:"scalingSpec,omitempty"`
}

func (in *AffectedResource) NamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: in.Namespace,
		Name:      in.Resource,
	}
}

func (in *AffectedResource) String() string {
	return fmt.Sprintf("%s/%s [%s]", in.Namespace, in.Resource, in.Kind)
}

type ResourceScalingSpec struct {
	//+kubebuilder:validation:Minimum:=0
	OriginalReplicas int `json:"originalReplicas"`
	//+kubebuilder:validation:Minimum:=0
	CurrentReplicas int `json:"currentReplicas"`
}
