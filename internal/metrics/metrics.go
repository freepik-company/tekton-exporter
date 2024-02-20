package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/exp/maps"
	"regexp"
)

const (

	// MetricsPrefix
	MetricsPrefix = "tekton_exporter_"
)

var (
	Pool = PoolSpec{}
)

// GetProcessedLabels accept a list of strings representing an object's labels and return a map
// whose keys are the input labels, and the values are the same labels with a Prometheus-ready syntax
func GetProcessedLabels(labelNames []string) (promLabelNames map[string]string, err error) {

	promLabelNames = make(map[string]string)

	// Make a regex to say we only want lower + uppercase letters, numbers and underscore
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		return promLabelNames, err
	}

	promLabelNames = make(map[string]string)

	for _, labelName := range labelNames {
		processedlabelName := reg.ReplaceAllString(labelName, "_")
		promLabelNames[labelName] = processedlabelName
	}

	return promLabelNames, err
}

// RegisterMetrics register declared metrics with their labels on Prometheus SDK
func RegisterMetrics(extraLabelNames []string) {

	parsedLabelsMap, _ := GetProcessedLabels(extraLabelNames) // TODO: Handle error
	parsedLabels := maps.Values(parsedLabelsMap)

	// Metrics for _status on PipelineRun resources
	pipelineRunStatusLabels := []string{"name", "namespace", "status", "reason"}
	pipelineRunStatusLabels = append(pipelineRunStatusLabels, parsedLabels...)

	Pool.PipelineRunStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: MetricsPrefix + "pipelinerun_status",
		Help: "tbd",
	}, pipelineRunStatusLabels)

	// Metrics for _status on TaskRun resources
	taskRunStatusLabels := []string{"name", "namespace", "status", "reason"}
	taskRunStatusLabels = append(taskRunStatusLabels, parsedLabels...)

	Pool.TaskRunStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: MetricsPrefix + "taskrun_status",
		Help: "tbd",
	}, taskRunStatusLabels)

	// Metrics for _duration on PipelineRun resources
	pipelineRunDurationLabels := []string{"name", "namespace", "start_timestamp", "completion_timestamp"}
	pipelineRunDurationLabels = append(pipelineRunDurationLabels, parsedLabels...)

	Pool.PipelineRunDuration = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: MetricsPrefix + "pipelinerun_duration_seconds",
		Help: "tbd",
	}, pipelineRunDurationLabels)

	// Metrics for _duration on TaskRun resources
	taskRunDurationLabels := []string{"name", "namespace", "start_timestamp", "completion_timestamp"}
	taskRunDurationLabels = append(taskRunDurationLabels, parsedLabels...)

	Pool.TaskRunDuration = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: MetricsPrefix + "taskrun_duration_seconds",
		Help: "tbd",
	}, taskRunDurationLabels)
}
