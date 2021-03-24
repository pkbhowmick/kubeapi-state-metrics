package types

import (
	"context"

	clientset "github.com/pkbhowmick/k8s-crd/pkg/client/clientset/versioned"
	generator "github.com/pkbhowmick/kubeapi-state-metrics/pkg/metric_generator"
	"github.com/pkbhowmick/kubeapi-state-metrics/pkg/options"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/tools/cache"
)

type BuilderInterface interface {
	WithMetrics(r prometheus.Registerer)
	WithEnableResources(c []string) error
	WithNamespaces(n options.NamespaceList)
	WithContext(ctx context.Context)
	WithKubeClient(c clientset.Interface)
	WithGeneratedStoreFunc(f BuildStoreFunc)
	WithAllowDenyList(l AllowDenyLister)
}

type BuildStoreFunc func(metricsFamilies []generator.FamilyGenerator,
	expectedType interface{},
	listWatchFunc func(kubeapiClient clientset.Interface, ns string) cache.ListerWatcher,
) cache.Store

// AllowDenyLister interface for AllowDeny lister that can allow or exclude metrics by there names
type AllowDenyLister interface {
	IsIncluded(string) bool
	IsExcluded(string) bool
}
