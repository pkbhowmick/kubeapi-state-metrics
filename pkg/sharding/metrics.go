package sharding

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	LabelOrdinal = "shard_ordinal"
)

type Metrics struct {
	Ordinal *prometheus.GaugeVec
	Total   prometheus.Gauge
}

func NewShardingMetrics(r prometheus.Registerer) *Metrics {
	return &Metrics{
		Ordinal: promauto.With(r).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "kubeapi_state_metrics_shard_ordinal",
				Help: "Current sharding ordinal/index of this instance",
			}, []string{LabelOrdinal},
		),
		Total: promauto.With(r).NewGauge(
			prometheus.GaugeOpts{
				Name: "kubeapi_state_metrics_total_shards",
				Help: "Number of total shards this instance is aware of",
			},
		),
	}
}
