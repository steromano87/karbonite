package schedule

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-co-op/gocron"
	"github.com/go-logr/logr"
	karbonitev1 "github.com/steromano87/karbonite/api/v1"
	"k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ThrottleAction struct {
	Log               logr.Logger
	Selector          karbonitev1.Selector
	DesiredReplicas   int
	DryRun            bool
	ReentrantSchedule string

	ReferenceThrottlingRule *karbonitev1.ThrottlingRule
}

func (a ThrottleAction) Run(kubeClient client.Client, job gocron.Job) error {
	if a.DryRun {
		a.Log.Info("Dry-running deletion rule. This is a drill!")
	} else {
		a.Log.Info("Running deletion rule. This is not a drill!!!")
	}

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

		a.Log.Info("The following resources would be throttled",
			"resourceCount", len(allMatchingResources),
			"resourceNames", resourceNames,
			"desiredReplicas", a.DesiredReplicas,
		)
	} else {
		a.Log.Info("No matching resource has been found")
	}

	if !a.DryRun {
		a.Log.Info("Resource deletion started")
		err := a.throttleResources(kubeClient, allMatchingResources)
		if err != nil {
			return err
		}
		a.Log.Info("Resource deletion completed")
	}

	// Update deletion rule's status
	a.ReferenceThrottlingRule.Status.RunCount = job.RunCount()
	err = kubeClient.Status().Update(context.Background(), a.ReferenceThrottlingRule)
	if err != nil {
		a.Log.Error(err, "Error updating reference throttling rule", "throttlingRule", a.ReferenceThrottlingRule.GetName())
	}

	return nil
}

func (a ThrottleAction) throttleResources(kubeClient client.Client, targetResources []unstructured.Unstructured) error {
	var err error

	for _, rawResource := range targetResources {
		namespacedName := types.NamespacedName{Namespace: rawResource.GetNamespace(), Name: rawResource.GetName()}

		switch rawResource.GetKind() {
		case "StatefulSet":
			err = a.throttleStatefulSet(kubeClient, namespacedName)
		case "Deployment":
			err = a.throttleDeployment(kubeClient, namespacedName)
		default:
			err = errors.New(fmt.Sprintf("unmanaged type %s for resource %s", rawResource.GetKind(), namespacedName))
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (a ThrottleAction) throttleStatefulSet(kubeClient client.Client, namespacedName types.NamespacedName) error {
	statefulSet := &v1.StatefulSet{}
	err := kubeClient.Get(context.Background(), namespacedName, statefulSet)
	if err != nil {
		a.Log.Error(err, "Error retrieving StatefulSet", "name", namespacedName.String())
		return err
	}

	newReplicas := int32(a.DesiredReplicas)
	oldReplicas := *statefulSet.Spec.Replicas
	statefulSet.Spec.Replicas = &newReplicas
	err = kubeClient.Update(context.Background(), statefulSet)
	if err != nil {
		a.Log.Error(err, "Error scaling StatefulSet", "name", namespacedName.String(), "desiredReplicas", a.DesiredReplicas)
		return err
	}

	a.Log.Info("Successfully scaled StatefulSet",
		"name", namespacedName.String(),
		"oldReplicas", oldReplicas,
		"newReplicas", newReplicas,
	)

	return nil
}

func (a ThrottleAction) throttleDeployment(kubeClient client.Client, namespacedName types.NamespacedName) error {
	deployment := &v1.Deployment{}
	err := kubeClient.Get(context.Background(), namespacedName, deployment)
	if err != nil {
		a.Log.Error(err, "Error retrieving Deployment", "name", namespacedName.String())
		return err
	}

	newReplicas := int32(a.DesiredReplicas)
	oldReplicas := *deployment.Spec.Replicas
	deployment.Spec.Replicas = &newReplicas
	err = kubeClient.Update(context.Background(), deployment)
	if err != nil {
		a.Log.Error(err, "Error scaling Deployment", "name", namespacedName.String(), "desiredReplicas", a.DesiredReplicas)
		return err
	}

	a.Log.Info("Successfully scaled Deployment",
		"name", namespacedName.String(),
		"oldReplicas", oldReplicas,
		"newReplicas", newReplicas,
	)

	return nil
}
