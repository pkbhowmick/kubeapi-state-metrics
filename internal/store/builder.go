package store

import (
	"context"
	"reflect"
	"sort"
	"strings"

	"github.com/kubernetes/kube-state-metrics/pkg/listwatch"
	metricsstore "github.com/kubernetes/kube-state-metrics/pkg/metrics_store"
	"github.com/pkbhowmick/k8s-crd/pkg/apis/stable.example.com/v1alpha1"
	clientset "github.com/pkbhowmick/k8s-crd/pkg/client/clientset/versioned"
	ksmtypes "github.com/pkbhowmick/kubeapi-state-metrics/pkg/builder/types"
	generator "github.com/pkbhowmick/kubeapi-state-metrics/pkg/metric_generator"
	"github.com/pkbhowmick/kubeapi-state-metrics/pkg/options"
	"github.com/pkbhowmick/kubeapi-state-metrics/pkg/watch"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/tools/cache"
)

type Builder struct {
	kubeapiClient    clientset.Interface
	namespaces       options.NamespaceList
	ctx              context.Context
	enabledResources []string
	listWatchMetrics *watch.ListWatchMetrics
	buildStoreFunc   ksmtypes.BuildStoreFunc
	allowLabelsList  map[string][]string
	allowDenyList    ksmtypes.AllowDenyLister
}

func NewBuilder() *Builder {
	b := &Builder{}
	return b
}

func (b *Builder) WithGenerateStoreFunc(f ksmtypes.BuildStoreFunc) {
	b.buildStoreFunc = f
}

func (b *Builder) WithMetrics(r prometheus.Registerer) {
	b.listWatchMetrics = watch.NewListWatchMetrics(r)
}

// WithEnabledResources sets the enabledResources property of a Builder.
func (b *Builder) WithEnabledResources(r []string) error {
	for _, col := range r {
		if !resourceExists(col) {
			return errors.Errorf("resource %s does not exist. Available resources: %s", col, strings.Join(availableResources(), ","))
		}
	}

	var copy []string
	copy = append(copy, r...)

	sort.Strings(copy)

	b.enabledResources = copy
	return nil
}

func (b *Builder) WithNamespaces(n options.NamespaceList) {
	b.namespaces = n
}

var availableStores = map[string]func(f *Builder) cache.Store{
	"kubeapis": func(b *Builder) cache.Store { return b.buildKubeapiStore() },
}

func resourceExists(name string) bool {
	_, ok := availableStores[name]
	return ok
}

func availableResources() []string {
	c := []string{}
	for name := range availableStores {
		c = append(c, name)
	}
	return c
}

func (b *Builder) buildKubeapiStore() cache.Store {
	return b.buildStoreFunc(kubeapiMetricFamilies(), &v1alpha1.KubeApi{}, createKubeapiListWatch)
}

func (b *Builder) buildStore(
	metricFamilies []generator.FamilyGenerator,
	expectedType interface{},
	listWatchFunc func(kubeClient clientset.Interface, ns string) cache.ListerWatcher,
) cache.Store {
	metricFamilies = generator.FilterMetricFamilies(b.allowDenyList, metricFamilies)
	composedMetricGenFuncs := generator.ComposeMetricGenFuncs(metricFamilies)
	familyHeaders := generator.ExtractMetricFamilyHeaders(metricFamilies)

	store := metricsstore.NewMetricsStore(
		familyHeaders,
		composedMetricGenFuncs,
	)
	b.reflectorPerNamespace(expectedType, store, listWatchFunc)
	return store
}

// DefaultGenerateStoreFunc returns default buildStore function
func (b *Builder) DefaultGenerateStoreFunc() ksmtypes.BuildStoreFunc {
	return b.buildStore
}

// reflectorPerNamespace creates a Kubernetes client-go reflector with the given
// listWatchFunc for each given namespace and registers it with the given store.
func (b *Builder) reflectorPerNamespace(
	expectedType interface{},
	store cache.Store,
	listWatchFunc func(kubeClient clientset.Interface, ns string) cache.ListerWatcher,
) {
	lwf := func(ns string) cache.ListerWatcher { return listWatchFunc(b.kubeapiClient, ns) }
	lw := listwatch.MultiNamespaceListerWatcher(b.namespaces, nil, lwf)
	instrumentedListWatch := watch.NewInstrumentedListerWatcher(lw, b.listWatchMetrics, reflect.TypeOf(expectedType).String())
	reflector := cache.NewReflector(sharding.NewShardedListWatch(b.shard, b.totalShards, instrumentedListWatch), expectedType, store, 0)
	go reflector.Run(b.ctx.Done())
}
