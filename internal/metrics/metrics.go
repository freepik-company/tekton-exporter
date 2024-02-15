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

func GetProcessedLabels(labelNames []string) (promLabelNames map[string]string, err error) {

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

	// Metrics for PipelineRun resources
	pipelineRunStatusLabels := []string{"name", "status", "reason"}
	pipelineRunStatusLabels = append(pipelineRunStatusLabels, parsedLabels...)

	Pool.PipelineRunStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: MetricsPrefix + "pipelinerun_status",
		Help: "number of nodes ready to schedule pods",
	}, pipelineRunStatusLabels)

	// Metrics for TaskRun resources
	taskRunStatusLabels := []string{"name", "status", "reason"}
	taskRunStatusLabels = append(taskRunStatusLabels, parsedLabels...)

	Pool.TaskRunStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: MetricsPrefix + "taskrun_status",
		Help: "number of nodes ready to schedule pods",
	}, taskRunStatusLabels)
}
