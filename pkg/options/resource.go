package options

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// DefaultNamespaces is the default namespace selector for selecting and filtering across all namespaces.
	DefaultNamespaces = NamespaceList{metav1.NamespaceAll}

	// DefaultResources represents the default set of resources in kubeapi-state-metrics.
	DefaultResources = ResourceSet{
		"kubeapis": struct{}{},
	}
)
