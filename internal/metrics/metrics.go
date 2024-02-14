package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (

	// MetricsPrefix
	MetricsPrefix = "tekton_exporter_"
)

var (
	Pool = PoolSpec{}
)

// RegisterMetrics register declared metrics with their labels on Prometheus SDK
func RegisterMetrics(extraLabelNames []string) {

	// Metrics for PipelineRun resources
	pipelineRunStatusLabels := []string{"status", "reason"}
	pipelineRunStatusLabels = append(pipelineRunStatusLabels, extraLabelNames...)

	Pool.PipelineRunStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: MetricsPrefix + "pipelinerun_status",
		Help: "number of nodes ready to schedule pods",
	}, pipelineRunStatusLabels)

}

// TODO
//func upgradePrometheusMetrics(nodeGroups *NodeGroups) (err error) {
//
//	for _, nodegroup := range *nodeGroups {
//
//		// Convert all the parsed strings into float64
//		// TODO abstract this section to a function, oh dirty Diana
//		healthReady, err := strconv.ParseFloat(nodegroup.Health.Ready, 64)
//		if err != nil {
//			return err
//		}
//
//		healthUnready, err := strconv.ParseFloat(nodegroup.Health.Unready, 64)
//		if err != nil {
//			return err
//		}
//
//		// Update all the metrics for this nodegroup
//		mReady.WithLabelValues(nodegroup.Name).Set(healthReady)
//		mUnready.WithLabelValues(nodegroup.Name).Set(healthUnready)
//	}
//
//	return nil
//}
