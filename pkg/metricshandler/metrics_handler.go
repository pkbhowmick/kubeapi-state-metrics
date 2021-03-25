package metricshandler

import (
	"sync"

	clientset "github.com/pkbhowmick/k8s-crd/pkg/client/clientset/versioned"
	"github.com/pkbhowmick/kubeapi-state-metrics/internal/store"
	"github.com/pkbhowmick/kubeapi-state-metrics/pkg/options"
)

type MetricsHandler struct {
	opts               *options.Options
	kubeapiClient      clientset.Interface
	storeBuilder       *store.Builder
	enableGZIPEncoding bool

	cancel func()

	mtx *sync.RWMutex
}
