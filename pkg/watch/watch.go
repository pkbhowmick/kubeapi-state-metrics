package watch

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

// ListWatchMetrics stores the pointers of kubeapi_state_metrics_[list|watch]_total metrics.
type ListWatchMetrics struct {
	WatchTotal *prometheus.CounterVec
	ListTotal  *prometheus.CounterVec
}

// NewListWatchMetrics takes in a prometheus registry and initializes
// and registers the kubeapi_state_metrics_list_total and
// kubeapi_state_metrics_watch_total metrics. It returns those registered metrics.
func NewListWatchMetrics(r prometheus.Registerer) *ListWatchMetrics {
	return &ListWatchMetrics{
		WatchTotal: promauto.With(r).NewCounterVec(
			prometheus.CounterOpts{
				Name: "kubeapi_state_metrics_watch_total",
				Help: "Number of total resource watches in kubeapi-state-metrics",
			},
			[]string{"result", "resource"},
		),
		ListTotal: promauto.With(r).NewCounterVec(
			prometheus.CounterOpts{
				Name: "kubeapi_state_metrics_list_total",
				Help: "Number of total resource list in kubeapi-state-metrics",
			},
			[]string{"result", "resource"},
		),
	}
}

// InstrumentedListerWatcher provides the kubeapi_state_metrics_watch_total metric
// with a cache.ListerWatcher obj and the related resource.
type InstrumentedListerWatcher struct {
	lw       cache.ListerWatcher
	metrics  *ListWatchMetrics
	resource string
}

// NewInstrumentedListerWatcher returns a new InstrumentedListerWatcher.
func NewInstrumentedListerWatcher(lw cache.ListerWatcher, metrics *ListWatchMetrics, resource string) cache.ListerWatcher {
	return &InstrumentedListerWatcher{
		lw:       lw,
		metrics:  metrics,
		resource: resource,
	}
}

// List is a wrapper func around the cache.ListerWatcher.List func. It increases the success/error
// / counters based on the outcome of the List operation it instruments.
func (i *InstrumentedListerWatcher) List(options metav1.ListOptions) (res runtime.Object, err error) {
	res, err = i.lw.List(options)
	if err != nil {
		i.metrics.ListTotal.WithLabelValues("error", i.resource).Inc()
		return
	}

	i.metrics.ListTotal.WithLabelValues("success", i.resource).Inc()
	return
}

// Watch is a wrapper func around the cache.ListerWatcher.Watch func. It increases the success/error
// counters based on the outcome of the Watch operation it instruments.
func (i *InstrumentedListerWatcher) Watch(options metav1.ListOptions) (res watch.Interface, err error) {
	res, err = i.lw.Watch(options)
	if err != nil {
		i.metrics.WatchTotal.WithLabelValues("error", i.resource).Inc()
		return
	}

	i.metrics.WatchTotal.WithLabelValues("success", i.resource).Inc()
	return
}
