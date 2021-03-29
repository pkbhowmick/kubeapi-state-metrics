module github.com/pkbhowmick/kubeapi-state-metrics

go 1.15

require (
	github.com/oklog/run v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.10.0
	github.com/prometheus/common v0.19.0
	github.com/prometheus/exporter-toolkit v0.5.1
	github.com/spf13/pflag v1.0.5
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba // indirect
	k8s.io/api v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/autoscaler/vertical-pod-autoscaler v0.9.2
	k8s.io/client-go v0.20.4
	k8s.io/klog/v2 v2.8.0
	k8s.io/kube-state-metrics/v2 v2.0.0-rc.1
	k8s.io/utils v0.0.0-20210111153108-fddb29f9d009 // indirect
)
