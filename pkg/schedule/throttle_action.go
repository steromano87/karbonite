package schedule

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-co-op/gocron"
	"github.com/go-logr/logr"
	karbonitev1 "github.com/steromano87/karbonite/api/v1"
	"k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type ThrottleAction struct {
	Log               logr.Logger
	Selector          karbonitev1.Selector
	DesiredReplicas   int
	DryRun            bool
	ReentrantSchedule string
	Timeout           time.Duration

	ReferenceThrottlingRule *karbonitev1.ThrottlingRule

	CronScheduler *gocron.Scheduler

	affectedResources []karbonitev1.AffectedResource
}

func (a *ThrottleAction) Run(kubeClient client.Client, job gocron.Job) error {
	var startMessage string
	if a.DryRun {
		startMessage = "Dry-running throttling rule. This is a drill!"
	} else {
		startMessage = "Running throttling rule. This is not a drill!!!"
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), a.Timeout)
	defer cancelFunc()
	a.Log.Info(startMessage, "timeout", a.Timeout)

	a.affectedResources = make([]karbonitev1.AffectedResource, 0)
	allMatchingResources, err := a.Selector.FindMatchingResources(ctx, kubeClient)
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
		a.Log.Info("Resource throttling started")
		err := a.throttleResources(ctx, kubeClient, allMatchingResources)
		if err != nil {
			return err
		}
		a.Log.Info("Resource throttling completed")
	}

	// Add throttle revert action, if the End field is defined
	err = a.scheduleThrottleRevertAction(kubeClient)
	if err != nil {
		a.Log.Error(err, "Error scheduling throttle revert")
	}

	// Add 1 to the run count, because it is increased only when the job has completed the function execution
	lastRun := metav1.NewTime(time.Now())
	a.ReferenceThrottlingRule.Status.RunCount = job.RunCount() + 1
	a.ReferenceThrottlingRule.Status.LastRun = &karbonitev1.LastRunInfo{
		Timestamp:         &lastRun,
		AffectedResources: a.affectedResources,
	}

	err = kubeClient.Status().Update(ctx, a.ReferenceThrottlingRule)
	if err != nil {
		a.Log.Error(err, "Error updating reference throttling rule", "throttlingRule", a.ReferenceThrottlingRule.GetName())
	}

	return nil
}

func (a *ThrottleAction) throttleResources(ctx context.Context, kubeClient client.Client, targetResources []unstructured.Unstructured) error {
	var err error

	for _, rawResource := range targetResources {
		namespacedName := types.NamespacedName{Namespace: rawResource.GetNamespace(), Name: rawResource.GetName()}

		switch rawResource.GetKind() {
		case statefulSetKind:
			err = a.throttleStatefulSet(ctx, kubeClient, namespacedName)
		case deploymentKind:
			err = a.throttleDeployment(ctx, kubeClient, namespacedName)
		default:
			err = errors.New(fmt.Sprintf("unmanaged type %s for resource %s", rawResource.GetKind(), namespacedName))
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (a *ThrottleAction) throttleStatefulSet(ctx context.Context, kubeClient client.Client, namespacedName types.NamespacedName) error {
	statefulSet := &v1.StatefulSet{}
	err := kubeClient.Get(ctx, namespacedName, statefulSet)
	if err != nil {
		a.Log.Error(err, "Error retrieving StatefulSet", "name", namespacedName.String())
		return err
	}

	newReplicas := int32(a.DesiredReplicas)
	oldReplicas := *statefulSet.Spec.Replicas
	statefulSet.Spec.Replicas = &newReplicas
	err = kubeClient.Update(ctx, statefulSet)
	if err != nil {
		a.Log.Error(err, "Error scaling StatefulSet", "name", namespacedName.String(), "desiredReplicas", a.DesiredReplicas)
		return err
	}

	a.Log.Info("Successfully scaled StatefulSet",
		"name", namespacedName.String(),
		"oldReplicas", oldReplicas,
		"newReplicas", newReplicas,
	)

	// Add affected resource only if there was a real change
	if newReplicas != oldReplicas {
		a.affectedResources = append(a.affectedResources, karbonitev1.AffectedResource{
			Namespace: namespacedName.Namespace,
			Resource:  namespacedName.Name,
			Kind:      statefulSetKind,
			ResourceScalingSpec: &karbonitev1.ResourceScalingSpec{
				OriginalReplicas: int(oldReplicas),
				CurrentReplicas:  a.DesiredReplicas,
			},
		})
	}

	return nil
}

func (a *ThrottleAction) throttleDeployment(ctx context.Context, kubeClient client.Client, namespacedName types.NamespacedName) error {
	deployment := &v1.Deployment{}
	err := kubeClient.Get(ctx, namespacedName, deployment)
	if err != nil {
		a.Log.Error(err, "Error retrieving Deployment", "name", namespacedName.String())
		return err
	}

	newReplicas := int32(a.DesiredReplicas)
	oldReplicas := *deployment.Spec.Replicas
	deployment.Spec.Replicas = &newReplicas
	err = kubeClient.Update(ctx, deployment)
	if err != nil {
		a.Log.Error(err, "Error scaling Deployment", "name", namespacedName.String(), "desiredReplicas", a.DesiredReplicas)
		return err
	}

	a.Log.Info("Successfully scaled Deployment",
		"name", namespacedName.String(),
		"oldReplicas", oldReplicas,
		"newReplicas", newReplicas,
	)

	// Add affected resource only if there was a real change
	if newReplicas != oldReplicas {
		a.affectedResources = append(a.affectedResources, karbonitev1.AffectedResource{
			Namespace: namespacedName.Namespace,
			Resource:  namespacedName.Name,
			Kind:      deploymentKind,
			ResourceScalingSpec: &karbonitev1.ResourceScalingSpec{
				OriginalReplicas: int(oldReplicas),
				CurrentReplicas:  a.DesiredReplicas,
			},
		})
	}

	return nil
}

func (a *ThrottleAction) scheduleThrottleRevertAction(kubeClient client.Client) error {
	if len(a.affectedResources) == 0 {
		a.Log.Info("No resources were affected by the throttling rule, skipping throttle revert scheduling")
		return nil
	}

	if a.ReentrantSchedule != "" {
		throttleRevertAction := ThrottleRevertAction{
			Log:                a.Log.WithName("revert"),
			SourceThrottleRule: a.ReferenceThrottlingRule,
			Timeout:            a.Timeout,
			AffectedResources:  a.affectedResources,
		}

		revertJobTags := []string{
			fmt.Sprintf("%s/%s", a.ReferenceThrottlingRule.GetNamespace(), a.ReferenceThrottlingRule.GetName()),
			karbonitev1.ThrottleRevertTag,
		}

		// Add the revert job, run it only once and then remove it from the scheduler
		revertJob, err := a.CronScheduler.Cron(a.ReentrantSchedule).Tag(revertJobTags...).LimitRunsTo(1).Do(throttleRevertAction.Run, kubeClient)
		revertJob.SingletonMode()
		if err != nil {
			a.Log.Error(err, "Error scheduling throttle revert")
			return err
		}

		a.ReferenceThrottlingRule.Status.ActiveReentrantThrottle = &karbonitev1.ActiveReentrantThrottle{
			AffectedResources: a.affectedResources,
			ReentrantOn:       metav1.NewTime(revertJob.NextRun()),
		}

		a.Log.Info("Scheduled throttle revert")
	} else {
		a.ReferenceThrottlingRule.Status.ActiveReentrantThrottle = &karbonitev1.ActiveReentrantThrottle{}
		a.Log.Info("No re-entrant schedule has been set, skipping throttle revert scheduling")
	}

	return nil
}
