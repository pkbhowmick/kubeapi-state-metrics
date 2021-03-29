/*
Copyright 2018 The Kubernetes Authors All rights reserved.

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

package store

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/pkbhowmick/k8s-crd/pkg/apis/stable.example.com/v1alpha1"

	clientset "github.com/pkbhowmick/k8s-crd/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	vpaclientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	ksmtypes "github.com/pkbhowmick/kubeapi-state-metrics/pkg/builder/types"
	"github.com/pkbhowmick/kubeapi-state-metrics/pkg/options"
	"k8s.io/kube-state-metrics/v2/pkg/listwatch"
	generator "k8s.io/kube-state-metrics/v2/pkg/metric_generator"
	metricsstore "k8s.io/kube-state-metrics/v2/pkg/metrics_store"
	"k8s.io/kube-state-metrics/v2/pkg/sharding"
	"k8s.io/kube-state-metrics/v2/pkg/watch"
)

func Debug(val string) {
	fmt.Println("*********************** Debug point *************************")
	fmt.Printf("                         %s                              \n", val)
	fmt.Println("*********************** Debug point *************************")
}

// Builder helps to build store. It follows the builder pattern
// (https://en.wikipedia.org/wiki/Builder_pattern).
type Builder struct {
	kubeClient       clientset.Interface
	vpaClient        vpaclientset.Interface //skip
	namespaces       options.NamespaceList
	ctx              context.Context
	enabledResources []string
	allowDenyList    ksmtypes.AllowDenyLister
	listWatchMetrics *watch.ListWatchMetrics
	shard            int32
	totalShards      int
	buildStoreFunc   ksmtypes.BuildStoreFunc
	allowLabelsList  map[string][]string
}

// NewBuilder returns a new builder.
func NewBuilder() *Builder {
	b := &Builder{}
	return b
}

// WithMetrics sets the metrics property of a Builder.
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

// WithNamespaces sets the namespaces property of a Builder.
func (b *Builder) WithNamespaces(n options.NamespaceList) {
	b.namespaces = n
}

// WithContext sets the ctx property of a Builder.
func (b *Builder) WithContext(ctx context.Context) {
	b.ctx = ctx
}

// WithKubeClient sets the kubeClient property of a Builder.
func (b *Builder) WithKubeClient(c clientset.Interface) {
	b.kubeClient = c
}

// WithVPAClient sets the vpaClient property of a Builder so that the verticalpodautoscaler collector can query VPA objects.
func (b *Builder) WithVPAClient(c vpaclientset.Interface) {
	b.vpaClient = c
}

// WithAllowDenyList configures the allow or denylisted metric to be exposed
// by the store build by the Builder.
func (b *Builder) WithAllowDenyList(l ksmtypes.AllowDenyLister) {
	b.allowDenyList = l
}

// WithGenerateStoreFunc configures a custom generate store function
func (b *Builder) WithGenerateStoreFunc(f ksmtypes.BuildStoreFunc) {
	b.buildStoreFunc = f
}

// DefaultGenerateStoreFunc returns default buildStore function
func (b *Builder) DefaultGenerateStoreFunc() ksmtypes.BuildStoreFunc {
	return b.buildStore
}

// WithAllowLabels configures which labels can be returned for metrics
func (b *Builder) WithAllowLabels(labels map[string][]string) {
	if len(labels) > 0 {
		b.allowLabelsList = labels
	}
}

// Build initializes and registers all enabled stores.
func (b *Builder) Build() []cache.Store {
	fmt.Println("In buillder.go Build() function")
	if b.allowDenyList == nil {
		panic("allowDenyList should not be nil")
	}

	stores := []cache.Store{}
	activeStoreNames := []string{}

	for _, c := range b.enabledResources {
		constructor, ok := availableStores[c]
		if ok {
			store := constructor(b)
			activeStoreNames = append(activeStoreNames, c)
			stores = append(stores, store)
		}
	}

	klog.Infof("Active resources: %s", strings.Join(activeStoreNames, ","))

	return stores
}

var availableStores = map[string]func(f *Builder) cache.Store{
	"kubeapis": func(b *Builder) cache.Store { return b.buildKubeApiStore() },
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

func (b *Builder) buildKubeApiStore() cache.Store {
	return b.buildStoreFunc(kubeapiMetricsFamily(b.allowLabelsList["kubeapis"]), &v1alpha1.KubeApi{}, createKubeApiListWatch)
}

func (b *Builder) buildStore(
	metricFamilies []generator.FamilyGenerator,
	expectedType interface{},
	listWatchFunc func(kubeClient clientset.Interface, ns string) cache.ListerWatcher,
) cache.Store {
	fmt.Println("In builder.go buildStore() function")
	metricFamilies = generator.FilterMetricFamilies(b.allowDenyList, metricFamilies)
	composedMetricGenFuncs := generator.ComposeMetricGenFuncs(metricFamilies)
	familyHeaders := generator.ExtractMetricFamilyHeaders(metricFamilies)

	store := metricsstore.NewMetricsStore(
		familyHeaders,
		composedMetricGenFuncs,
	)
	b.reflectorPerNamespace(expectedType, store, listWatchFunc)
	Debug("store")
	fmt.Println(store)

	return store
}

// reflectorPerNamespace creates a Kubernetes client-go reflector with the given
// listWatchFunc for each given namespace and registers it with the given store.
func (b *Builder) reflectorPerNamespace(
	expectedType interface{},
	store cache.Store,
	listWatchFunc func(kubeClient clientset.Interface, ns string) cache.ListerWatcher,
) {
	lwf := func(ns string) cache.ListerWatcher { return listWatchFunc(b.kubeClient, ns) }
	lw := listwatch.MultiNamespaceListerWatcher(b.namespaces, nil, lwf)
	instrumentedListWatch := watch.NewInstrumentedListerWatcher(lw, b.listWatchMetrics, reflect.TypeOf(expectedType).String())
	reflector := cache.NewReflector(sharding.NewShardedListWatch(b.shard, b.totalShards, instrumentedListWatch), expectedType, store, 0)
	go reflector.Run(b.ctx.Done())
}
