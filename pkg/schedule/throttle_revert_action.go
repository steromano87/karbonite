package schedule

import (
	"context"
	"errors"
	"github.com/go-logr/logr"
	karbonitev1 "github.com/steromano87/karbonite/api/v1"
	v1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type ThrottleRevertAction struct {
	Log                logr.Logger
	Timeout            time.Duration
	SourceThrottleRule *karbonitev1.ThrottlingRule
	AffectedResources  []karbonitev1.AffectedResource
}

func (a ThrottleRevertAction) Run(kubeClient client.Client) error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), a.Timeout)
	defer cancelFunc()

	a.Log.Info("Started reverting throttles", "sourceThrottlingRule", a.SourceThrottleRule.GetName())

	for _, affectedResource := range a.AffectedResources {
		a.Log.Info(
			"Reverting throttle for affected resource",
			"affectedResource", affectedResource,
			"originalReplicas", affectedResource.ResourceScalingSpec.OriginalReplicas,
			"currentReplicas", affectedResource.ResourceScalingSpec.OriginalReplicas,
			"timeout", a.Timeout,
		)

		var err error
		switch affectedResource.Kind {
		case deploymentKind:
			err = a.revertDeploymentThrottling(ctx, kubeClient, affectedResource)
			if err != nil {
				a.Log.Error(err, "Error while reverting throttle for Deployment", "affectedResource", affectedResource.String())
				continue
			}

		case statefulSetKind:
			err = a.revertStatefulSetThrottling(ctx, kubeClient, affectedResource)
			if err != nil {
				a.Log.Error(err, "Error while reverting throttle for StatefulSet", "affectedResource", affectedResource.String())
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
			"Successfully reverted throttle for affected resource", "affectedResource", affectedResource.String())
	}

	a.Log.Info("Throttles have been successfully reverted", "sourceThrottlingRule", a.SourceThrottleRule.GetName())

	// Update throttling rule status by removing the active throttle revert entry
	a.SourceThrottleRule.Status.ActiveReentrantThrottle = nil
	err := kubeClient.Status().Update(ctx, a.SourceThrottleRule)
	if err != nil {
		a.Log.Error(err, "Error updating source throttling rule", "sourceThrottlingRule", a.SourceThrottleRule.GetName())
	}

	return nil
}

func (a ThrottleRevertAction) revertDeploymentThrottling(ctx context.Context, kubeClient client.Client, affectedResource karbonitev1.AffectedResource) error {
	targetResource := &v1.Deployment{}
	err := kubeClient.Get(ctx, affectedResource.NamespacedName(), targetResource)
	if err != nil {
		a.Log.Error(err, "Cannot find resource", "affectedResource", affectedResource)
		return err
	}

	originalReplicas := int32(affectedResource.ResourceScalingSpec.OriginalReplicas)
	targetResource.Spec.Replicas = &originalReplicas

	err = kubeClient.Update(ctx, targetResource)
	if err != nil {
		a.Log.Error(err, "Cannot modify resource replicas", "affectedResource", affectedResource)
		return err
	}

	return nil
}

func (a ThrottleRevertAction) revertStatefulSetThrottling(ctx context.Context, kubeClient client.Client, affectedResource karbonitev1.AffectedResource) error {
	targetResource := &v1.StatefulSet{}
	err := kubeClient.Get(ctx, affectedResource.NamespacedName(), targetResource)
	if err != nil {
		a.Log.Error(err, "Cannot find resource", "affectedResource", affectedResource)
		return err
	}

	originalReplicas := int32(affectedResource.ResourceScalingSpec.OriginalReplicas)
	targetResource.Spec.Replicas = &originalReplicas

	err = kubeClient.Update(ctx, targetResource)
	if err != nil {
		a.Log.Error(err, "Cannot modify resource replicas", "affectedResource", affectedResource)
		return err
	}

	return nil
}
