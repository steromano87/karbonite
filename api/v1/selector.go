package v1

import (
	"context"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"regexp"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Selector struct {
	MatchNamespaces []string          `json:"matchNamespaces,omitempty"`
	MatchKinds      []string          `json:"matchKinds,omitempty"`
	MatchNames      []string          `json:"matchNames,omitempty"`
	MatchLabels     map[string]string `json:"matchLabels,omitempty"`
}

func (in *Selector) FindMatchingResources(kubeClient client.Client) ([]unstructured.Unstructured, error) {
	namespaces, err := in.findMatchingNamespaces(kubeClient)
	if err != nil {
		return nil, err
	}

	// Get matching resources list from all matching namespaces
	rawResources, err := in.findAllMatchingResourcesInNamespaces(kubeClient, namespaces)
	if err != nil {
		return nil, err
	}

	// Filter by resource name
	resourceList, err := in.filterResourcesByName(rawResources)
	if err != nil {
		return nil, err
	}

	return resourceList, nil
}

func (in *Selector) findMatchingNamespaces(kubeClient client.Client) ([]string, error) {
	matchedNamespaces := make([]string, 0)

	// Get all namespaces
	allNamespaces := &v1.NamespaceList{}
	err := kubeClient.List(context.Background(), allNamespaces)
	if err != nil {
		return nil, err
	}

	// Compile the matchers into regexes
	matcherRegexps := make([]*regexp.Regexp, 0)
	for _, matcher := range in.MatchNamespaces {
		matcherRegex, err := regexp.Compile(matcher)
		if err != nil {
			return nil, err
		}

		matcherRegexps = append(matcherRegexps, matcherRegex)
	}

	for _, ns := range allNamespaces.Items {
		for _, nsRegex := range matcherRegexps {
			if nsRegex.MatchString(ns.Name) {
				matchedNamespaces = append(matchedNamespaces, ns.Name)
				break
			}
		}
	}

	return matchedNamespaces, nil
}

// findAllMatchingResourcesInNamespaces retrieves all the resources whose type matches one of the given regexps.
// Thanks to this link for the correct approach when using unstructured objects to gather a dynamic resource type:
// https://github.com/kubernetes-sigs/controller-runtime/issues/1823
func (in *Selector) findAllMatchingResourcesInNamespaces(kubeClient client.Client, namespaces []string) ([]unstructured.Unstructured, error) {
	output := make([]unstructured.Unstructured, 0)

	// Compile the matchers into regexes
	matcherRegexps := make([]*regexp.Regexp, 0)
	for _, matcher := range in.MatchKinds {
		matcherRegex, err := regexp.Compile("(?i)" + matcher)
		if err != nil {
			return nil, err
		}

		matcherRegexps = append(matcherRegexps, matcherRegex)
	}

	// Compile the regex to verify that the matching resource kind contains the word "List" in it
	listMatcher := regexp.MustCompile(`(?i).+List$`)

	// Find resource kinds that match the given regexps
	allKnownTypes := kubeClient.Scheme().AllKnownTypes()
	matchingTypes := make([]schema.GroupVersionKind, 0)

	for gvk := range allKnownTypes {
		for _, regex := range matcherRegexps {
			// Ensure that the Kind contains the word "List" when filtering
			if regex.MatchString(gvk.Kind) && listMatcher.MatchString(gvk.Kind) {
				matchingTypes = append(matchingTypes, gvk)
				break
			}
		}
	}

	// Now iterate over matching types and manually set UnstructuredList's kind to the current one
	for _, matchingType := range matchingTypes {
		for _, namespace := range namespaces {
			matchingResources := unstructured.UnstructuredList{}
			matchingResources.SetGroupVersionKind(matchingType)
			listOptions := []client.ListOption{client.InNamespace(namespace)}

			err := kubeClient.List(context.Background(), &matchingResources, listOptions...)
			if err != nil {
				// If no match has been found, simply continue with the next type
				if _, ok := err.(*meta.NoKindMatchError); ok {
					continue
				}
				return nil, err
			}

			output = append(output, matchingResources.Items...)
		}
	}

	return output, nil
}

func (in *Selector) filterResourcesByName(input []unstructured.Unstructured) ([]unstructured.Unstructured, error) {
	output := make([]unstructured.Unstructured, 0)

	// Trivial case: if no matcher is defined, return all available resources
	if len(in.MatchNames) == 0 {
		return input, nil
	}

	// Compile the matchers into regexes
	matcherRegexps := make([]*regexp.Regexp, 0)
	for _, matcher := range in.MatchNames {
		matcherRegex, err := regexp.Compile(matcher)
		if err != nil {
			return nil, err
		}

		matcherRegexps = append(matcherRegexps, matcherRegex)
	}

	for _, resource := range input {
		for _, resourceRegex := range matcherRegexps {
			if resourceRegex.MatchString(resource.GetName()) {
				output = append(output, resource)
				break
			}
		}
	}

	return output, nil
}
