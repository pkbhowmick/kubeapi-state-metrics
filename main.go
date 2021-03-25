package main

import (
	"context"

	"github.com/oklog/run"

	clientset "github.com/pkbhowmick/k8s-crd/pkg/client/clientset/versioned"
	"github.com/pkbhowmick/kubeapi-state-metrics/internal/store"
	"github.com/pkbhowmick/kubeapi-state-metrics/pkg/options"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

const (
	metricsPath = "/metrics"
)

// promLogger implements promhttp.Logger
type promLogger struct{}

func (pl promLogger) Println(v ...interface{}) {
	klog.Error(v...)
}

// promLogger implements the Logger interface
func (pl promLogger) Log(v ...interface{}) error {
	klog.Info(v...)
	return nil
}

func main() {
	opts := options.NewOptions()
	opts.AddFlags()

	promLogger := promLogger{}
	ctx := context.Background()

	err := opts.Parse()

	if err != nil {
		klog.Fatalf("Error: %s", err)
	}

	storeBuilder := store.NewBuilder()

	ksmMetricsRegistry := prometheus.NewRegistry()
	ksmMetricsRegistry.MustRegister(version.NewCollector("kubeapi_state_metrics"))
	storeBuilder.WithMetrics(ksmMetricsRegistry)

	var resources []string
	if len(opts.Resources) == 0 {
		klog.Info("Using default resources")
		resources = options.DefaultResources.AsSlice()
	} else {
		klog.Infof("Using resources %s", opts.Resources.String())
		resources = opts.Resources.AsSlice()
	}

	if err := storeBuilder.WithEnabledResources(resources); err != nil {
		klog.Fatalf("Failed to set up resources: %v", err)
	}

	if len(opts.Namespaces) == 0 {
		klog.Info("Using all namespace")
		storeBuilder.WithNamespaces(options.DefaultNamespaces)
	} else {
		if opts.Namespaces.IsAllNamespaces() {
			klog.Info("Using all namespace")
		} else {
			klog.Infof("Using %s namespaces", opts.Namespaces)
		}
		storeBuilder.WithNamespaces(opts.Namespaces)
	}

	storeBuilder.WithGenerateStoreFunc(storeBuilder.DefaultGenerateStoreFunc())

	kubeapiClient, err := createKubeapiClient(opts.Apiserver, opts.Kubeconfig)

	if err != nil {
		klog.Fatalf("Failed to create client: %v", err)
	}
	storeBuilder.WithKubeapiClient(kubeapiClient)
	ksmMetricsRegistry.MustRegister(
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
		prometheus.NewGoCollector(),
	)
	var g run.Group

	m := metricshandler.New(opts, kubeapiClient, storeBuilder, opts.EnableGZIPEncoding)

}

func createKubeapiClient(apiserver string, kubeconfig string) (clientset.Interface, error) {
	config, err := clientcmd.BuildConfigFromFlags(apiserver, kubeconfig)

	if err != nil {
		return nil, err
	}

	kubeapiClient, err := clientset.NewForConfig(config)

	return kubeapiClient, nil

}
