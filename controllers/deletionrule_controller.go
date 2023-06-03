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
	"fmt"
	"github.com/go-co-op/gocron"
	"github.com/go-logr/logr"
	"github.com/robfig/cron/v3"
	karbonitev1 "github.com/steromano87/karbonite/api/v1"
	"github.com/steromano87/karbonite/pkg/schedule"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

const deletionRuleFinalizer = "karbonite.io/deletionrule-finalizer"

// DeletionRuleReconciler reconciles a DeletionRule object
type DeletionRuleReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	CronScheduler        *gocron.Scheduler
	CronValidator        cron.Parser
	ScheduledJobsTimeout time.Duration
}

//+kubebuilder:rbac:groups=karbonite.io,resources=deletionrules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=karbonite.io,resources=deletionrules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=karbonite.io,resources=deletionrules/finalizers,verbs=update
//+kubebuilder:rbac:groups=*,resources=*,verbs=get;list;watch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *DeletionRuleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reconcileLog := r.Log.WithName("DeletionRuleReconciler").WithValues("rule", req.NamespacedName)

	reconcileLog.Info("Running reconcile loop")

	deletionRule := &karbonitev1.DeletionRule{}
	err := r.Client.Get(ctx, req.NamespacedName, deletionRule)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			reconcileLog.Info("The requested resource was not found (deletion in progress?), skipping")
		} else {
			reconcileLog.Error(err, "Error while retrieving deletion rule spec")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Handle finalizer
	if deletionRule.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so the finalizer is registered
		if !controllerutil.ContainsFinalizer(deletionRule, deletionRuleFinalizer) {
			reconcileLog.Info("Registering finalizer", "finalizer", deletionRuleFinalizer)
			controllerutil.AddFinalizer(deletionRule, deletionRuleFinalizer)
			if err := r.Client.Update(ctx, deletionRule); err != nil {
				r.Log.Error(err, "Error while registering finalizer")
				return ctrl.Result{}, err
			}
			reconcileLog.Info("Finalizer registered", "finalizer", deletionRuleFinalizer)
		}
	} else {
		reconcileLog.Info("Throttling rule is being deleted, starting termination reconciliation")
		return r.runFinalizerCleanupActivities(logr.NewContext(ctx, reconcileLog), deletionRule, req)
	}

	reconcileLog.Info("Parsing deletion rule",
		"schedulesCount", len(deletionRule.Spec.Schedules))

	// Delete any already saved schedule for te given tag (i.e. the namespaced name of the DeletionRule)
	err = r.removeExistingSchedules(logr.NewContext(ctx, reconcileLog), req.NamespacedName.String())
	if err != nil {
		return ctrl.Result{}, err
	}

	// Add the deletion schedules
	if deletionRule.Spec.Enabled {
		reconcileLog.Info("Deletion rule is enabled, adding deletion cronjobs")
		err = r.scheduleDeletionActions(logr.NewContext(ctx, reconcileLog), req, deletionRule)
		if err != nil {
			return ctrl.Result{}, err
		}
	} else {
		reconcileLog.Info("Deletion rule is disabled, skipping")
	}

	// Update rule by setting lastModified field
	deletionRule.Status.RunCount = 0
	err = r.Client.Status().Update(ctx, deletionRule)
	if err != nil {
		reconcileLog.Error(err, "Error updating analyzed rule")
	}

	return ctrl.Result{}, err
}

func (r *DeletionRuleReconciler) runFinalizerCleanupActivities(ctx context.Context, deletionRule *karbonitev1.DeletionRule, req ctrl.Request) (ctrl.Result, error) {
	reconcileLog, _ := logr.FromContext(ctx)

	// The objects is being deleted, so perform the cleanup...
	if controllerutil.ContainsFinalizer(deletionRule, deletionRuleFinalizer) {
		reconcileLog.Info("Finalizer registered, running cleanup activities before deletion", "finalizer", deletionRuleFinalizer)

		// Delete any already saved schedule for te given tag (i.e. the namespaced name of the DeletionRule)
		err := r.removeExistingSchedules(logr.NewContext(ctx, reconcileLog), req.NamespacedName.String())
		if err != nil {
			// Do not log errors, because they are already handled in the inner function
			return ctrl.Result{}, err
		}

		// ... and remove the finalizer
		reconcileLog.Info("Cleanup activities completed, de-registering finalizer", "finalizer", deletionRuleFinalizer)
		controllerutil.RemoveFinalizer(deletionRule, deletionRuleFinalizer)
		if err := r.Client.Update(ctx, deletionRule); err != nil {
			r.Log.Error(err, "Error while removing finalizer")
			return ctrl.Result{}, err
		}
		reconcileLog.Info("Finalizer deregistered")
	}

	// Stop reconciliation, because the rule is being deleted
	reconcileLog.Info("Termination reconciliation completed")
	return ctrl.Result{}, nil
}

func (r *DeletionRuleReconciler) removeExistingSchedules(ctx context.Context, namespacedRuleName string) error {
	reconcileLog, _ := logr.FromContext(ctx)
	reconcileLog.Info("Checking for previously saved schedules to delete")
	previouslySavedSchedules, err := r.CronScheduler.FindJobsByTag(namespacedRuleName, karbonitev1.DeletionTag)

	if err != nil {
		if err == gocron.ErrJobNotFoundWithTag {
			reconcileLog.Info("No previously saved schedules have been found")
			return nil
		}

		reconcileLog.Error(err, "Error while retrieving previously saved schedules")
		return err
	}

	reconcileLog.Info("Found previously saved schedules, deleting...", "affectedItems", len(previouslySavedSchedules))
	err = r.CronScheduler.RemoveByTags(namespacedRuleName, karbonitev1.DeletionTag)
	if err != nil {
		reconcileLog.Error(err, "Error while removing previously saved schedules")
		return err
	}

	reconcileLog.Info("Successfully deleted previously saved schedules")
	return nil
}

func (r *DeletionRuleReconciler) scheduleDeletionActions(ctx context.Context, req ctrl.Request, deletionRule *karbonitev1.DeletionRule) error {
	reconcileLog, _ := logr.FromContext(ctx)

	// If no namespace matcher is explicitly given, set it to the origin namespace
	if len(deletionRule.Spec.Selector.MatchNamespaces) == 0 {
		reconcileLog.Info("No explicit namespace matcher has been set, defaulting to DeletionRule namespace exact matcher",
			"namespace", req.Namespace)
		deletionRule.Spec.Selector.MatchNamespaces = []string{fmt.Sprintf("^%s$", req.Namespace)}
	}

	for _, ruleSchedule := range deletionRule.Spec.Schedules {
		action := schedule.DeletionAction{
			Log:                   r.Log.WithName("DeleteAction"),
			Selector:              deletionRule.Spec.Selector,
			DryRun:                deletionRule.Spec.DryRun,
			Timeout:               r.ScheduledJobsTimeout,
			ReferenceDeletionRule: deletionRule,
		}

		// Validate the cron expression before accepting it!
		_, err := r.CronValidator.Parse(ruleSchedule)
		if err != nil {
			reconcileLog.Error(err, "Error parsing cron expression")
			return err
		}

		// Schedule the deletion action
		scheduledAction, err := r.CronScheduler.Cron(ruleSchedule).Tag(
			req.NamespacedName.String(), karbonitev1.DeletionTag,
		).DoWithJobDetails(action.Run, r.Client)

		if err != nil {
			reconcileLog.Error(err, "Error scheduling deletion job")
			return err
		}

		scheduledAction.SingletonMode()
		reconcileLog.Info("Added deletion cronjob", "schedule", ruleSchedule, "isDryRun", action.DryRun)
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DeletionRuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Log.Info(
		"Starting deletion actions scheduler",
		"referenceTimezone", r.CronScheduler.Location(),
		"scheduledJobsTimeout", r.ScheduledJobsTimeout.String(),
	)
	r.CronScheduler.StartAsync()

	return ctrl.NewControllerManagedBy(mgr).
		For(&karbonitev1.DeletionRule{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 1,
		}).
		WithEventFilter(deletionRuleIgnorePredicate()).
		Complete(r)
}
