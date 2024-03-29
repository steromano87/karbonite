//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ActiveReentrantThrottle) DeepCopyInto(out *ActiveReentrantThrottle) {
	*out = *in
	if in.AffectedResources != nil {
		in, out := &in.AffectedResources, &out.AffectedResources
		*out = make([]AffectedResource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.ReentrantOn.DeepCopyInto(&out.ReentrantOn)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ActiveReentrantThrottle.
func (in *ActiveReentrantThrottle) DeepCopy() *ActiveReentrantThrottle {
	if in == nil {
		return nil
	}
	out := new(ActiveReentrantThrottle)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AffectedResource) DeepCopyInto(out *AffectedResource) {
	*out = *in
	if in.ResourceScalingSpec != nil {
		in, out := &in.ResourceScalingSpec, &out.ResourceScalingSpec
		*out = new(ResourceScalingSpec)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AffectedResource.
func (in *AffectedResource) DeepCopy() *AffectedResource {
	if in == nil {
		return nil
	}
	out := new(AffectedResource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DeletionRule) DeepCopyInto(out *DeletionRule) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DeletionRule.
func (in *DeletionRule) DeepCopy() *DeletionRule {
	if in == nil {
		return nil
	}
	out := new(DeletionRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DeletionRule) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DeletionRuleList) DeepCopyInto(out *DeletionRuleList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]DeletionRule, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DeletionRuleList.
func (in *DeletionRuleList) DeepCopy() *DeletionRuleList {
	if in == nil {
		return nil
	}
	out := new(DeletionRuleList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DeletionRuleList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DeletionRuleSpec) DeepCopyInto(out *DeletionRuleSpec) {
	*out = *in
	in.Selector.DeepCopyInto(&out.Selector)
	if in.Schedules != nil {
		in, out := &in.Schedules, &out.Schedules
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DeletionRuleSpec.
func (in *DeletionRuleSpec) DeepCopy() *DeletionRuleSpec {
	if in == nil {
		return nil
	}
	out := new(DeletionRuleSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DeletionRuleStatus) DeepCopyInto(out *DeletionRuleStatus) {
	*out = *in
	in.NextRun.DeepCopyInto(&out.NextRun)
	if in.LastRun != nil {
		in, out := &in.LastRun, &out.LastRun
		*out = new(LastRunInfo)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DeletionRuleStatus.
func (in *DeletionRuleStatus) DeepCopy() *DeletionRuleStatus {
	if in == nil {
		return nil
	}
	out := new(DeletionRuleStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LastRunInfo) DeepCopyInto(out *LastRunInfo) {
	*out = *in
	if in.Timestamp != nil {
		in, out := &in.Timestamp, &out.Timestamp
		*out = (*in).DeepCopy()
	}
	if in.AffectedResources != nil {
		in, out := &in.AffectedResources, &out.AffectedResources
		*out = make([]AffectedResource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LastRunInfo.
func (in *LastRunInfo) DeepCopy() *LastRunInfo {
	if in == nil {
		return nil
	}
	out := new(LastRunInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReentrantSchedule) DeepCopyInto(out *ReentrantSchedule) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReentrantSchedule.
func (in *ReentrantSchedule) DeepCopy() *ReentrantSchedule {
	if in == nil {
		return nil
	}
	out := new(ReentrantSchedule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceScalingSpec) DeepCopyInto(out *ResourceScalingSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceScalingSpec.
func (in *ResourceScalingSpec) DeepCopy() *ResourceScalingSpec {
	if in == nil {
		return nil
	}
	out := new(ResourceScalingSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Selector) DeepCopyInto(out *Selector) {
	*out = *in
	if in.MatchNamespaces != nil {
		in, out := &in.MatchNamespaces, &out.MatchNamespaces
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.MatchKinds != nil {
		in, out := &in.MatchKinds, &out.MatchKinds
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.MatchNames != nil {
		in, out := &in.MatchNames, &out.MatchNames
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.MatchLabels != nil {
		in, out := &in.MatchLabels, &out.MatchLabels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Selector.
func (in *Selector) DeepCopy() *Selector {
	if in == nil {
		return nil
	}
	out := new(Selector)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ThrottlingRule) DeepCopyInto(out *ThrottlingRule) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ThrottlingRule.
func (in *ThrottlingRule) DeepCopy() *ThrottlingRule {
	if in == nil {
		return nil
	}
	out := new(ThrottlingRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ThrottlingRule) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ThrottlingRuleList) DeepCopyInto(out *ThrottlingRuleList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ThrottlingRule, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ThrottlingRuleList.
func (in *ThrottlingRuleList) DeepCopy() *ThrottlingRuleList {
	if in == nil {
		return nil
	}
	out := new(ThrottlingRuleList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ThrottlingRuleList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ThrottlingRuleSpec) DeepCopyInto(out *ThrottlingRuleSpec) {
	*out = *in
	in.Selector.DeepCopyInto(&out.Selector)
	if in.Schedules != nil {
		in, out := &in.Schedules, &out.Schedules
		*out = make([]ThrottlingSchedule, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ThrottlingRuleSpec.
func (in *ThrottlingRuleSpec) DeepCopy() *ThrottlingRuleSpec {
	if in == nil {
		return nil
	}
	out := new(ThrottlingRuleSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ThrottlingRuleStatus) DeepCopyInto(out *ThrottlingRuleStatus) {
	*out = *in
	in.NextRun.DeepCopyInto(&out.NextRun)
	if in.LastRun != nil {
		in, out := &in.LastRun, &out.LastRun
		*out = new(LastRunInfo)
		(*in).DeepCopyInto(*out)
	}
	if in.ActiveReentrantThrottle != nil {
		in, out := &in.ActiveReentrantThrottle, &out.ActiveReentrantThrottle
		*out = new(ActiveReentrantThrottle)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ThrottlingRuleStatus.
func (in *ThrottlingRuleStatus) DeepCopy() *ThrottlingRuleStatus {
	if in == nil {
		return nil
	}
	out := new(ThrottlingRuleStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ThrottlingSchedule) DeepCopyInto(out *ThrottlingSchedule) {
	*out = *in
	out.ReentrantSchedule = in.ReentrantSchedule
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ThrottlingSchedule.
func (in *ThrottlingSchedule) DeepCopy() *ThrottlingSchedule {
	if in == nil {
		return nil
	}
	out := new(ThrottlingSchedule)
	in.DeepCopyInto(out)
	return out
}
