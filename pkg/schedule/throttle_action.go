package schedule

import (
	"github.com/go-co-op/gocron"
	"github.com/go-logr/logr"
	v1 "github.com/steromano87/karbonite/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ThrottleAction struct {
	Log               logr.Logger
	Selector          v1.Selector
	DesiredReplicas   int
	DryRun            bool
	ReentrantSchedule string

	ReferenceThrottlingRule *v1.ThrottlingRule
}

func (a ThrottleAction) Run(kubeClient client.Client, job gocron.Job) error {
	if a.DryRun {
		a.Log.Info("Dry-running deletion rule. This is a drill!")
	} else {
		a.Log.Info("Running deletion rule. This is not a drill!!!")
	}

	return nil
}
