package metrics

import "github.com/prometheus/client_golang/prometheus"

type PoolSpec struct {
	PipelineRunStatus   *prometheus.GaugeVec
	TaskRunStatus       *prometheus.GaugeVec
	PipelineRunDuration *prometheus.GaugeVec
	TaskRunDuration     *prometheus.GaugeVec
}
