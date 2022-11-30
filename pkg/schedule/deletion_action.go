package schedule

import (
	"fmt"
	"github.com/steromano87/karbonite/api/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DeletionAction struct {
	Origin   types.NamespacedName
	Matchers []v1.Matchers
	DryRun   bool
}

func (a DeletionAction) Run(kubeClient client.Client) error {
	if a.DryRun {
		ctrl.Log.Info("Dry-running deletion rule. This is a drill!")
	} else {
		ctrl.Log.Info("Running deletion rule. This is not a drill!!!")
	}

	allMatchingResources := make([]unstructured.Unstructured, 0)

	for _, matcher := range a.Matchers {
		matchingResources, err := matcher.FindMatchingResources(kubeClient, a.Origin.Namespace)
		if err != nil {
			ctrl.Log.Error(err, "Error retrieving matching resources")
		}

		allMatchingResources = append(allMatchingResources, matchingResources...)
	}

	resourceNames := make([]string, len(allMatchingResources))
	for index, resource := range allMatchingResources {
		resourceNames[index] = fmt.Sprintf("%s/%s [%s]",
			resource.GetNamespace(), resource.GetName(), resource.GetKind())
	}

	ctrl.Log.Info("The following resources would be deleted",
		"resource count", len(allMatchingResources),
		"resource names", resourceNames)
	return nil
}
