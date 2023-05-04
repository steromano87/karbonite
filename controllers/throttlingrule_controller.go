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

package controllers

import (
	"context"
	"github.com/go-co-op/gocron"
	"github.com/go-logr/logr"
	"github.com/robfig/cron/v3"
	karbonitev1 "github.com/steromano87/karbonite/api/v1"
	"github.com/steromano87/karbonite/pkg/schedule"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// ThrottlingRuleReconciler reconciles a ThrottlingRule object
type ThrottlingRuleReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	CronScheduler *gocron.Scheduler
	CronValidator cron.Parser
}

//+kubebuilder:rbac:groups=karbonite.io,resources=throttlingrules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=karbonite.io,resources=throttlingrules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=karbonite.io,resources=throttlingrules/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=deployments;statefulsets,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *ThrottlingRuleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reconcileLog := r.Log.WithName("ThrottlingRuleReconciler").WithValues("rule", req.NamespacedName)

	reconcileLog.Info("Running reconcile loop")

	throttlingRule := &karbonitev1.ThrottlingRule{}
	err := r.Client.Get(ctx, req.NamespacedName, throttlingRule)
	if err != nil {
		reconcileLog.Error(err, "Error while retrieving throttling rule spec")
		return ctrl.Result{}, err
	}

	reconcileLog.Info("Parsing throttling rule",
		"schedulesCount", len(throttlingRule.Spec.Schedules))

	// Forcefully re-enter all the active schedules to restore the original resource state
	err = r.undoActiveThrottlingRules(logr.NewContext(ctx, reconcileLog), req.NamespacedName.String())
	if err != nil {
		return ctrl.Result{}, err
	}

	// Delete any already saved schedule for te given tag (i.e. the namespaced name of the DeletionRule)
	err = r.removeExistingSchedules(logr.NewContext(ctx, reconcileLog), req.NamespacedName.String())
	if err != nil {
		// Do not log errors, because they are already handled in the inner function
		return ctrl.Result{}, err
	}

	// Add the deletion schedules
	if throttlingRule.Spec.Enabled {
		reconcileLog.Info("Throttling rule is enabled, adding throttling cronjobs")
		err = r.scheduleThrottlingActions(logr.NewContext(ctx, reconcileLog), req, throttlingRule)
		if err != nil {
			return ctrl.Result{}, err
		}
	} else {
		reconcileLog.Info("Throttling rule is disabled, skipping")
		return ctrl.Result{}, nil
	}

	// Update rule status
	throttlingRule.Status.RunCount = 0
	err = r.Client.Status().Update(ctx, throttlingRule)
	if err != nil {
		reconcileLog.Error(err, "Error updating analyzed rule")
	}

	return ctrl.Result{}, err
}

func (r *ThrottlingRuleReconciler) undoActiveThrottlingRules(ctx context.Context, namespacedRuleName string) error {
	reconcileLog, _ := logr.FromContext(ctx)
	reconcileLog.Info("Checking for previously saved throttling revert schedules to delete")

	err := r.CronScheduler.RemoveByTags(namespacedRuleName, karbonitev1.ThrottleRevertTag)
	if err != nil {
		if err == gocron.ErrJobNotFoundWithTag {
			reconcileLog.Info("No previously saved throttling revert schedules have been found")
			return nil
		}

		reconcileLog.Error(err, "Error while removing previously saved throttling revert schedules")
		return err
	}

	reconcileLog.Info("Successfully deleted previously saved throttling revert schedules")
	return nil
}

func (r *ThrottlingRuleReconciler) removeExistingSchedules(ctx context.Context, namespacedRuleName string) error {
	reconcileLog, _ := logr.FromContext(ctx)
	reconcileLog.Info("Checking for previously saved throttling schedules to delete")

	err := r.CronScheduler.RemoveByTags(namespacedRuleName, karbonitev1.ThrottleTag)
	if err != nil {
		if err == gocron.ErrJobNotFoundWithTag {
			reconcileLog.Info("No previously saved throttling schedules have been found")
			return nil
		}

		reconcileLog.Error(err, "Error while removing previously saved throttling schedules")
		return err
	}

	reconcileLog.Info("Successfully deleted previously saved throttling schedules")
	return nil
}

func (r *ThrottlingRuleReconciler) scheduleThrottlingActions(ctx context.Context, req ctrl.Request, throttlingRule *karbonitev1.ThrottlingRule) error {
	reconcileLog, _ := logr.FromContext(ctx)

	// If no namespace matcher is explicitly given, set it to the origin namespace
	if len(throttlingRule.Spec.Selector.MatchNamespaces) == 0 {
		reconcileLog.Info("No explicit namespace matcher has been set, defaulting to ThrottlingRule namespace",
			"namespace", req.Namespace)
		throttlingRule.Spec.Selector.MatchNamespaces = []string{req.Namespace}
	}

	// If no specific resource kinds are explicitly given, use both Deployments and StatefulSets
	if len(throttlingRule.Spec.Selector.MatchKinds) == 0 {
		reconcileLog.Info("No explicit resource kind matchers have been set, defaulting to Deployment and StatefulSet")
		throttlingRule.Spec.Selector.MatchKinds = []string{"Deployment", "StatefulSet"}
	}

	for _, ruleSchedule := range throttlingRule.Spec.Schedules {
		action := schedule.ThrottleAction{
			Log:                     r.Log.WithName("ThrottleAction - " + req.NamespacedName.String()),
			Selector:                throttlingRule.Spec.Selector,
			DesiredReplicas:         ruleSchedule.DesiredReplicas,
			DryRun:                  throttlingRule.Spec.DryRun,
			ReentrantSchedule:       ruleSchedule.End,
			ReferenceThrottlingRule: throttlingRule,
			CronScheduler:           r.CronScheduler,
		}

		// Validate the cron expression before accepting it!
		_, err := r.CronValidator.Parse(ruleSchedule.Start)
		if err != nil {
			reconcileLog.Error(err, "Error parsing cron expression")
			return err
		}

		// Schedule the deletion action
		scheduledAction, err := r.CronScheduler.Cron(ruleSchedule.Start).Tag(
			req.NamespacedName.String(), karbonitev1.ThrottleTag,
		).DoWithJobDetails(action.Run, r.Client)
		if err != nil {
			reconcileLog.Error(err, "Error scheduling deletion job")
			return err
		}

		scheduledAction.SingletonMode()
		reconcileLog.Info("Added throttling cronjob", "schedule", ruleSchedule, "isDryRun", action.DryRun)
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ThrottlingRuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Log.Info("Starting throttling actions scheduler", "referenceTimezone", r.CronScheduler.Location())
	r.CronScheduler.StartAsync()

	return ctrl.NewControllerManagedBy(mgr).
		For(&karbonitev1.ThrottlingRule{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 1,
		}).
		WithEventFilter(throttlingRuleIgnorePredicate()).
		Complete(r)
}
