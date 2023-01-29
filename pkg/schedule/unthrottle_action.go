package schedule

import (
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
)

type UnThrottleAction struct {
	Log               logr.Logger
	AffectedResources []types.NamespacedName
	OriginalReplicas  int
}
