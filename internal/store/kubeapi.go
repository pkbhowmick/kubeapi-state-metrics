package store

import (
	"context"

	"github.com/pkbhowmick/k8s-crd/pkg/apis/stable.example.com/v1alpha1"
	clientset "github.com/pkbhowmick/k8s-crd/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/kube-state-metrics/v2/pkg/metric"
	generator "k8s.io/kube-state-metrics/v2/pkg/metric_generator"
)

var (
	descKubeApiLabelsName          = "kubeApi_deployment_labels"
	descKubeApiLabelsHelp          = "Kubernetes labels converted to Prometheus labels."
	descKubeApiLabelsDefaultLabels = []string{"namespace", "kubeapi"}
)

func kubeapiMetricsFamily(allowLabelsList []string) []generator.FamilyGenerator {
	return []generator.FamilyGenerator{
		*generator.NewFamilyGenerator(
			"kubeapi_status_replicas",
			"The number of replicas per kubeapis",
			metric.Gauge,
			"",
			wrapKubeApiFunc(func(d *v1alpha1.KubeApi) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(d.Status.Replicas),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			descKubeApiLabelsName,
			descKubeApiLabelsHelp,
			metric.Gauge,
			"",
			wrapKubeApiFunc(func(d *v1alpha1.KubeApi) *metric.Family {
				labelKeys, labelValues := createLabelKeysValues(d.Labels, allowLabelsList)
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							LabelKeys:   labelKeys,
							LabelValues: labelValues,
							Value:       1,
						},
					},
				}
			}),
		),
	}
}

func wrapKubeApiFunc(f func(*v1alpha1.KubeApi) *metric.Family) func(interface{}) *metric.Family {
	return func(obj interface{}) *metric.Family {
		kubeapi := obj.(*v1alpha1.KubeApi)

		metricFamily := f(kubeapi)

		//fmt.Println(metricFamily)

		for _, m := range metricFamily.Metrics {
			m.LabelKeys = append(descKubeApiLabelsDefaultLabels, m.LabelKeys...)
			m.LabelValues = append([]string{kubeapi.Namespace, kubeapi.Name}, m.LabelValues...)
		}

		return metricFamily
	}
}

func createKubeApiListWatch(kubeClient clientset.Interface, ns string) cache.ListerWatcher {
	return &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return kubeClient.StableV1alpha1().KubeApis(ns).List(context.TODO(), opts)
		},
		WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
			return kubeClient.StableV1alpha1().KubeApis(ns).Watch(context.TODO(), opts)
		},
	}
}
