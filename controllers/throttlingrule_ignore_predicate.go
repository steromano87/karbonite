package controllers

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// throttlingRuleIgnorePredicate skips the reconcile loop for Throttling Rules if:
//   - the object generation is not modified (e.g. a sub-resource only has been modified, like the status)
//   - the number of finalizers has been modified (to avoid the second reconcile loop when adding the finalizer on fresh new rules)
func throttlingRuleIgnorePredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(updateEvent event.UpdateEvent) bool {
			return updateEvent.ObjectOld.GetGeneration() != updateEvent.ObjectNew.GetGeneration() &&
				len(updateEvent.ObjectOld.GetFinalizers()) == len(updateEvent.ObjectNew.GetFinalizers())
		},
	}
}
