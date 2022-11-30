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
	rulesv1 "github.com/steromano87/karbonite/api/v1"
	"github.com/steromano87/karbonite/pkg/schedule"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"time"
)

// DeletionRuleReconciler reconciles a DeletionRule object
type DeletionRuleReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	CronScheduler *gocron.Scheduler
	CronValidator cron.Parser
}

//+kubebuilder:rbac:groups=rules.karbonite.io,resources=deletionrules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rules.karbonite.io,resources=deletionrules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=rules.karbonite.io,resources=deletionrules/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DeletionRule object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *DeletionRuleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reconcileLog := r.Log.WithValues("rule", req.NamespacedName)

	reconcileLog.Info("Running reconcile loop")

	deletionRule := rulesv1.DeletionRule{}
	err := r.Client.Get(context.Background(), req.NamespacedName, &deletionRule)
	if err != nil {
		reconcileLog.Error(err, "Error while retrieving deletion rule spec")
		return ctrl.Result{}, err
	}

	reconcileLog.Info("Parsing deletion rule",
		"matchersCount", len(deletionRule.Spec.Matchers),
		"schedulesCount", len(deletionRule.Spec.Schedules))

	// Delete any already saved schedule for te given tag (i.e. the namespaced name of the DeletionRule)
	err = r.removeExistingSchedules(logr.NewContext(ctx, reconcileLog), req.NamespacedName.String())
	if err != nil {
		return ctrl.Result{}, err
	}

	// Add the deletion schedules
	err = r.scheduleDeletionActions(logr.NewContext(ctx, reconcileLog), req, deletionRule)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Update rule by setting lastModified field
	deletionRule.Status.LastModified = metav1.Time{Time: time.Now()}
	err = r.Client.Update(context.Background(), &deletionRule)
	if err != nil {
		reconcileLog.Error(err, "Error updating analyzed rule")
	}

	return ctrl.Result{}, err
}

func (r *DeletionRuleReconciler) removeExistingSchedules(ctx context.Context, jobTag string) error {
	reconcileLog, _ := logr.FromContext(ctx)
	reconcileLog.Info("Checking for previously saved schedules to delete")
	previouslySavedSchedules, err := r.CronScheduler.FindJobsByTag(jobTag)

	if err != nil {
		if err == gocron.ErrJobNotFoundWithTag {
			reconcileLog.Info("No previously saved schedules have been found")
			return nil
		}

		reconcileLog.Error(err, "Error while retrieving previously saved schedules")
		return err
	}

	reconcileLog.Info("Found previously saved schedules, deleting...", "affectedItems", len(previouslySavedSchedules))
	err = r.CronScheduler.RemoveByTag(jobTag)
	if err != nil {
		reconcileLog.Error(err, "Error while removing previously saved schedules")
		return err
	}

	reconcileLog.Info("Successfully deleted previously saved schedules")
	return nil
}

func (r *DeletionRuleReconciler) scheduleDeletionActions(ctx context.Context, req ctrl.Request, deletionRule rulesv1.DeletionRule) error {
	reconcileLog, _ := logr.FromContext(ctx)
	for _, ruleSchedule := range deletionRule.Spec.Schedules {
		action := schedule.DeletionAction{
			Origin:   req.NamespacedName,
			Matchers: deletionRule.Spec.Matchers,
			DryRun:   deletionRule.Spec.DryRun,
		}

		// Validate the cron expression before accepting it!
		_, err := r.CronValidator.Parse(ruleSchedule)
		if err != nil {
			reconcileLog.Error(err, "Error parsing cron expression")
			return err
		}

		// Schedule the deletion action
		scheduledAction, err := r.CronScheduler.Cron(ruleSchedule).Tag(req.NamespacedName.String()).Do(action.Run, r.Client)
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
	r.Log.Info("Starting actions scheduler", "referenceTimezone", r.CronScheduler.Location())
	r.CronScheduler.StartAsync()

	return ctrl.NewControllerManagedBy(mgr).
		For(&rulesv1.DeletionRule{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 1,
		}).
		WithEventFilter(deletionRuleIgnorePredicate()).
		Complete(r)
}
