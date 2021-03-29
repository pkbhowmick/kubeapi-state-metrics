module github.com/pkbhowmick/kubeapi-state-metrics

go 1.15

require (
	github.com/oklog/run v1.1.0
	github.com/pkbhowmick/k8s-crd v0.0.0-20210304112549-69a20ee433df
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.10.0
	github.com/prometheus/common v0.19.0
	github.com/prometheus/exporter-toolkit v0.5.1
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/autoscaler/vertical-pod-autoscaler v0.9.2
	k8s.io/client-go v0.20.4
	k8s.io/klog/v2 v2.8.0
	k8s.io/kube-state-metrics/v2 v2.0.0-rc.1
)
