/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package metricshandler

import (
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"strings"
	"sync"

	kubernetes "github.com/pkbhowmick/k8s-crd/pkg/client/clientset/versioned"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	"github.com/pkbhowmick/kubeapi-state-metrics/internal/store"
	"github.com/pkbhowmick/kubeapi-state-metrics/pkg/options"
	metricsstore "k8s.io/kube-state-metrics/v2/pkg/metrics_store"
)

// MetricsHandler is a http.Handler that exposes the main kube-state-metrics
// /metrics endpoint. It allows concurrent reconfiguration at runtime.
type MetricsHandler struct {
	opts               *options.Options
	kubeClient         kubernetes.Interface
	storeBuilder       *store.Builder
	enableGZIPEncoding bool

	cancel func()

	// mtx protects stores, curShard, and curTotalShards
	mtx            *sync.RWMutex
	stores         []cache.Store
	curShard       int32
	curTotalShards int
}

// New creates and returns a new MetricsHandler with the given options.
func New(opts *options.Options, kubeClient kubernetes.Interface, storeBuilder *store.Builder, enableGZIPEncoding bool) *MetricsHandler {
	//store.Debug("Creating New Metrics Handler")
	return &MetricsHandler{
		opts:               opts,
		kubeClient:         kubeClient,
		storeBuilder:       storeBuilder,
		enableGZIPEncoding: enableGZIPEncoding,
		mtx:                &sync.RWMutex{},
	}
}

// ConfigureSharding (re-)configures sharding. Re-configuration can be done
// concurrently.
func (m *MetricsHandler) ConfigureSharding(ctx context.Context, shard int32, totalShards int) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if m.cancel != nil {
		m.cancel()
	}
	if totalShards != 1 {
		klog.Infof("configuring sharding of this instance to be shard index %d (zero-indexed) out of %d total shards", shard, totalShards)
	}
	ctx, m.cancel = context.WithCancel(ctx)
	m.storeBuilder.WithContext(ctx)
	m.stores = m.storeBuilder.Build()
	m.curShard = shard
	m.curTotalShards = totalShards
}

// ServeHTTP implements the http.Handler interface. It writes the metrics in
// its stores to the response body.
func (m *MetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//store.Debug("Metrics Handler Serve HTTP function called")
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	resHeader := w.Header()
	var writer io.Writer = w

	resHeader.Set("Content-Type", `text/plain; version=`+"0.0.4")

	if m.enableGZIPEncoding {
		// Gzip response if requested. Taken from
		// github.com/prometheus/client_golang/prometheus/promhttp.decorateWriter.
		reqHeader := r.Header.Get("Accept-Encoding")
		parts := strings.Split(reqHeader, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "gzip" || strings.HasPrefix(part, "gzip;") {
				writer = gzip.NewWriter(writer)
				resHeader.Set("Content-Encoding", "gzip")
			}
		}
	}

	for _, s := range m.stores {
		ms := s.(*metricsstore.MetricsStore)
		ms.WriteAll(writer)
	}

	// In case we gzipped the response, we have to close the writer.
	if closer, ok := writer.(io.Closer); ok {
		closer.Close()
	}
}
