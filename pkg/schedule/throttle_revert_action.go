package schedule

import (
	"context"
	"errors"
	"github.com/go-co-op/gocron"
	"github.com/go-logr/logr"
	karbonitev1 "github.com/steromano87/karbonite/api/v1"
	v1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ThrottleRevertAction struct {
	Log                logr.Logger
	SourceThrottleRule *karbonitev1.ThrottlingRule
	AffectedResources  []karbonitev1.AffectedResource

	CronScheduler *gocron.Scheduler
}

func (a ThrottleRevertAction) Run(kubeClient client.Client, job gocron.Job) error {
	a.Log.Info("Started reverting throttles", "sourceThrottlingRule", a.SourceThrottleRule.GetName())

	for _, affectedResource := range a.AffectedResources {
		a.Log.Info(
			"Reverting throttle for affected resource",
			"affectedResource", affectedResource,
			"originalReplicas", affectedResource.OriginalReplicas,
			"currentReplicas", affectedResource.CurrentReplicas,
		)

		var err error
		switch affectedResource.Kind {
		case deploymentKind:
			err = a.revertDeploymentThrottling(kubeClient, affectedResource)
			if err != nil {
				a.Log.Error(err, "Error while reverting throttle for Deployment", "affectedResource", affectedResource)
				continue
			}

		case statefulSetKind:
			err = a.revertStatefulSetThrottling(kubeClient, affectedResource)
			if err != nil {
				a.Log.Error(err, "Error while reverting throttle for StatefulSet", "affectedResource", affectedResource)
				continue
			}

		default:
			a.Log.Error(
				errors.New("unthrottable kind: "+affectedResource.Kind),
				"Error while reverting throttle for unknown type",
				"affectedResource", affectedResource,
			)
			continue
		}

		a.Log.Info(
			"Successfully reverted throttle for affected resource", "affectedResource", affectedResource)
	}

	a.Log.Info("Throttles have been successfully reverted", "sourceThrottlingRule", a.SourceThrottleRule.GetName())

	return nil
}

func (a ThrottleRevertAction) revertDeploymentThrottling(kubeClient client.Client, affectedResource karbonitev1.AffectedResource) error {
	targetResource := &v1.Deployment{}
	err := kubeClient.Get(context.Background(), affectedResource.NamespacedName(), targetResource)
	if err != nil {
		a.Log.Error(err, "Cannot find resource", "affectedResource", affectedResource)
		return err
	}

	originalReplicas := int32(affectedResource.OriginalReplicas)
	targetResource.Spec.Replicas = &originalReplicas

	err = kubeClient.Update(context.Background(), targetResource)
	if err != nil {
		a.Log.Error(err, "Cannot modify resource replicas", "affectedResource", affectedResource)
		return err
	}

	return nil
}

func (a ThrottleRevertAction) revertStatefulSetThrottling(kubeClient client.Client, affectedResource karbonitev1.AffectedResource) error {
	targetResource := &v1.StatefulSet{}
	err := kubeClient.Get(context.Background(), affectedResource.NamespacedName(), targetResource)
	if err != nil {
		a.Log.Error(err, "Cannot find resource", "affectedResource", affectedResource)
		return err
	}

	originalReplicas := int32(affectedResource.OriginalReplicas)
	targetResource.Spec.Replicas = &originalReplicas

	err = kubeClient.Update(context.Background(), targetResource)
	if err != nil {
		a.Log.Error(err, "Cannot modify resource replicas", "affectedResource", affectedResource)
		return err
	}

	return nil
}
