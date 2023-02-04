package schedule

import (
	"context"
	"fmt"
	"github.com/go-co-op/gocron"
	"github.com/go-logr/logr"
	"github.com/steromano87/karbonite/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type DeletionAction struct {
	Log      logr.Logger
	Selector v1.Selector
	DryRun   bool

	ReferenceDeletionRule *v1.DeletionRule

	affectedResources []v1.AffectedResource
}

func (a DeletionAction) Run(kubeClient client.Client, job gocron.Job) error {
	if a.DryRun {
		a.Log.Info("Dry-running deletion rule. This is a drill!")
	} else {
		a.Log.Info("Running deletion rule. This is not a drill!!!")
	}

	a.affectedResources = make([]v1.AffectedResource, 0)
	allMatchingResources, err := a.Selector.FindMatchingResources(kubeClient)
	if err != nil {
		a.Log.Error(err, "Error retrieving matching resources")
		return err
	}

	if len(allMatchingResources) > 0 {
		resourceNames := make([]string, len(allMatchingResources))
		for index, resource := range allMatchingResources {
			resourceNames[index] = fmt.Sprintf("%s/%s [%s]",
				resource.GetNamespace(), resource.GetName(), resource.GetKind())
		}

		a.Log.Info("The following resources would be deleted",
			"resourceCount", len(allMatchingResources),
			"resourceNames", resourceNames)
	} else {
		a.Log.Info("No matching resource has been found")
	}

	if !a.DryRun {
		a.Log.Info("Resource deletion started")
		err := a.deleteResources(kubeClient, allMatchingResources)
		if err != nil {
			return err
		}
		a.Log.Info("Resource deletion completed")
	}

	// Add 1 to the run count, because it is increased only when the job has complete the function execution
	lastRun := metav1.NewTime(time.Now())
	a.ReferenceDeletionRule.Status.RunCount = job.RunCount() + 1
	a.ReferenceDeletionRule.Status.LastRun = &v1.LastRunInfo{
		Timestamp:         &lastRun,
		AffectedResources: a.affectedResources,
	}

	err = kubeClient.Status().Update(context.Background(), a.ReferenceDeletionRule)
	if err != nil {
		a.Log.Error(err, "Error updating reference deletion rule", "deletionRule", a.ReferenceDeletionRule.GetName())
	}

	return nil
}

func (a DeletionAction) deleteResources(kubeClient client.Client, targetResources []unstructured.Unstructured) error {
	for _, targetResource := range targetResources {
		resourceDescriptor := fmt.Sprintf("%s/%s [%s]",
			targetResource.GetNamespace(), targetResource.GetName(), targetResource.GetKind())

		deleteOptions := []client.DeleteOption{
			client.GracePeriodSeconds(30),
		}

		err := kubeClient.Delete(context.Background(), &targetResource, deleteOptions...)
		if err != nil {
			a.Log.Error(err, "Error while deleting resource", "resource", resourceDescriptor)
			return err
		}
		a.Log.Info("Deleted resource", "resource", resourceDescriptor)

		a.affectedResources = append(a.affectedResources, v1.AffectedResource{
			Namespace: targetResource.GetNamespace(),
			Resource:  targetResource.GetName(),
			Kind:      targetResource.GetKind(),
		})
	}

	return nil
}
