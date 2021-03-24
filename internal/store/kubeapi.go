package store

import (
	"context"
	"fmt"

	"github.com/pkbhowmick/k8s-crd/pkg/apis/stable.example.com/v1alpha1"
	clientset "github.com/pkbhowmick/k8s-crd/pkg/client/clientset/versioned"
	"github.com/pkbhowmick/kubeapi-state-metrics/pkg/metric"
	generator "github.com/pkbhowmick/kubeapi-state-metrics/pkg/metric_generator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

var (
	descKubeapiLabelsHelp          = "Kubernetes labels converted to Prometheus labels."
	descKubeapiLabelsDefaultLabels = []string{"namespace", "deployment"}
)

func kubeapiMetricFamilies() []generator.FamilyGenerator {
	return []generator.FamilyGenerator{
		*generator.NewFamilyGenerator(
			"kubeapi_resource_replicas",
			"The number of replicas of kubeapi",
			metric.Gauge,
			"",
			wrapKubeapiFunc(func(k *v1alpha1.KubeApi) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(k.Status.Replicas),
						},
					},
				}
			}),
		)}
}

func wrapKubeapiFunc(f func(*v1alpha1.KubeApi) *metric.Family) func(interface{}) *metric.Family {
	return func(obj interface{}) *metric.Family {
		kubeapi := obj.(*v1alpha1.KubeApi)

		metricFamily := f(kubeapi)

		fmt.Println(metricFamily)

		for _, m := range metricFamily.Metrics {
			m.LabelKeys = append(descKubeapiLabelsDefaultLabels, m.LabelKeys...)
			m.LabelValues = append([]string{kubeapi.Namespace, kubeapi.Name}, m.LabelValues...)
		}

		return metricFamily
	}
}

func createKubeapiListWatch(kubeapiClient clientset.Interface, ns string) cache.ListerWatcher {
	return &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return kubeapiClient.StableV1alpha1().KubeApis(ns).List(context.TODO(), opts)
		},
		WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
			return kubeapiClient.StableV1alpha1().KubeApis(ns).Watch(context.TODO(), opts)
		},
	}
}
