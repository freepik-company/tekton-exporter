package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"maps"
	"strconv"
	"strings"
	"time"

	// Kubernetes clients
	// Ref: https://pkg.go.dev/k8s.io/client-go/dynamic
	"k8s.io/client-go/dynamic"
	// Ref: https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client/config
	ctrl "sigs.k8s.io/controller-runtime"

	// Kubernetes types
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"

	//
	"tekton-exporter/internal/globals"
	"tekton-exporter/internal/metrics"
)

const (
	watchPipelinerunMessage = "Watching PipelineRun objects"
	watchTaskrunMessage     = "Watching TaskRun objects"

	//
	timestampsPromLabelsRetrievalError = "placeholder: %s"
)

var (
	pipelineRunV1GVR = schema.GroupVersionResource{
		Group:    "tekton.dev",
		Version:  "v1",
		Resource: "pipelineruns",
	}

	taskRunV1GVR = schema.GroupVersionResource{
		Group:    "tekton.dev",
		Version:  "v1",
		Resource: "taskruns",
	}
)

// NewClient return a new Kubernetes Dynamic client from client-go SDK
func NewClient() (client *dynamic.DynamicClient, err error) {
	config, err := ctrl.GetConfig()
	if err != nil {
		return client, err
	}

	// Create the clients to do requests to out friend: Kubernetes
	client, err = dynamic.NewForConfig(config)
	if err != nil {
		return client, err
	}

	return client, err
}

// GetRunPopulatedPromLabels return only user's desired labels from an object
// Desired labels are defined by flag "--populated-labels"
func GetRunPopulatedPromLabels(ctx *context.Context, object *map[string]interface{}) (labelsMap map[string]string, err error) {
	labelsMap = make(map[string]string)

	// Read labels from event's resource
	objectLabels, err := GetObjectLabels(object)
	if err != nil {
		return labelsMap, err
	}

	// No existing labels on object, quit
	if len(objectLabels) == 0 {
		return labelsMap, nil
	}

	// Retrieve the 'populated-labels' flag from the context
	populatedLabels := map[string]string{}
	populatedLabelsFlag, ok := (*ctx).Value("flag-populated-labels").([]string)
	if !ok {
		return labelsMap, errors.New("populated labels flag not found in context")
	}

	//
	parsedLabelsMap, _ := metrics.GetProcessedLabels(populatedLabelsFlag)
	if err != nil {
		return labelsMap, fmt.Errorf("error processing populated labels: %v", err)
	}

	// Iterate over populated label names
	for _, populatedLabelName := range populatedLabelsFlag {

		// Populated labels are dynamic, but labels must be pre-registered in the Prometheus SDK.
		// This is a mechanism to avoid crashes if the labels are not present in the object.
		populatedLabels[parsedLabelsMap[populatedLabelName]] = "#"

		// Fill only the labels requested by the user.
		// Label names will be changed to a Prometheus-compatible syntax.
		if _, objectLabelsFound := objectLabels[populatedLabelName]; objectLabelsFound {
			populatedLabels[parsedLabelsMap[populatedLabelName]] = objectLabels[populatedLabelName]
		}
	}

	return populatedLabels, nil
}

// GetRunStatusPromLabels obtains the status-related labels for a pipeline based on the 'Succeeded' condition type and
// returns a map containing the 'status' and 'reason' labels.
// If the 'Succeeded' condition is not found, it populates a default condition with status 'False' and reason 'Unknown'.
func GetRunStatusPromLabels(object *map[string]interface{}) (labelsMap map[string]string, err error) {
	labelsMap = make(map[string]string)

	// Default condition if 'Succeeded' condition is not found
	defaultCondition := map[string]string{
		"status": "False",
		"reason": "Unknown",
	}

	// Obtain the status of 'Succeeded' condition type
	condition, err := GetObjectCondition(object, "Succeeded")
	if err != nil {
		// Use the default condition
		labelsMap["status"] = defaultCondition["status"]
		labelsMap["reason"] = defaultCondition["reason"]
		return labelsMap, nil
	}

	// Make the 'status' label understandable in metrics that are using it
	runStatusLabelStatus := "failed"
	if status, ok := condition["status"].(string); ok {
		if strings.ToLower(status) == "true" {
			runStatusLabelStatus = "success"
		}
	}

	statusLabels := map[string]string{
		"status": runStatusLabelStatus,
		"reason": condition["reason"].(string),
	}

	return statusLabels, nil
}

// GetRunDurationPromLabels return a map with 'start_timestamp' and 'completion_timestamp'
// from the object representing a run's status.
// If any timestamp is missing, it populates a default value set to '#'
func GetRunDurationPromLabels(object *map[string]interface{}) (labelsMap map[string]string, err error) {
	labelsMap = make(map[string]string)

	// Obtain the status of 'Succeeded' condition type
	status, err := GetObjectStatus(object)
	if err != nil {
		return nil, err
	}

	// If status is nil, return an error indicating missing status
	if status == nil {
		return nil, errors.New("status is missing in the object")
	}

	// Parse timestamps and populate labels
	timestampLabels := make(map[string]string)
	parseTimestamp := func(timestamp interface{}) (string, error) {
		if timestamp == nil {
			return "#", nil // Default value if timestamp is missing
		}
		parsedTime, err := time.Parse(time.RFC3339, timestamp.(string))
		if err != nil {
			return "", err // Return parsing error
		}
		return strconv.Itoa(int(parsedTime.Unix())), nil
	}

	// Parse start timestamp
	startTimestamp, err := parseTimestamp(status["startTime"])
	if err != nil {
		return nil, err
	}
	timestampLabels["start_timestamp"] = startTimestamp

	// Parse completion timestamp
	completionTimestamp, err := parseTimestamp(status["completionTime"])
	if err != nil {
		return nil, err
	}
	timestampLabels["completion_timestamp"] = completionTimestamp

	return timestampLabels, nil
}

// WatchPipelineRuns TODO
// Hey!, this function is intended to be executed as a go routine
func WatchPipelineRuns(ctx *context.Context, client *dynamic.DynamicClient) (err error) {
	globals.ExecContext.Logger.Info(watchPipelinerunMessage)

	// Create a watcher for PipelineRun resources
	pipelineRunWatcher, err := client.Resource(pipelineRunV1GVR).Watch(*ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	defer pipelineRunWatcher.Stop()

	for pipelineRunEvent := range pipelineRunWatcher.ResultChan() {
		// Extract the unstructured object from the event
		unstructuredObject, err := GetUnstructuredFromRuntimeObject(&pipelineRunEvent.Object)
		if err != nil {
			globals.ExecContext.Logger.Errorf("failed to parse object: %v", err)
			continue
		}

		// Process the PipelineRun event
		err = ProcessPipelineRunEvent(ctx, &unstructuredObject, pipelineRunEvent.Type)
		if err != nil {
			globals.ExecContext.Logger.Errorf("failed to process PipelineRun event: %v", err)
			continue
		}
	}

	return nil
}

// TODO
func ProcessPipelineRunEvent(ctx *context.Context, object *map[string]interface{}, eventType watch.EventType) error {

	// 1. Craft common labels
	objectBasicData, err := GetObjectBasicData(object)
	if err != nil {
		return err
	}

	commonLabels := map[string]string{
		"name":      objectBasicData["name"].(string),
		"namespace": objectBasicData["namespace"].(string),
	}

	// 2. Craft populated labels from PipelineRun object labels and merge them
	populatedLabels, err := GetRunPopulatedPromLabels(ctx, object)
	if err != nil {
		return err
	}

	maps.Copy(commonLabels, populatedLabels)

	// Conversion to a Prometheus SDK Labels type will be needed later
	// Maps in golang are ReferenceTypes, so we need to iterate to copy
	commonLabelsProm := prometheus.Labels{}
	for k, v := range commonLabels {
		commonLabelsProm[k] = v
	}

	// 3. Craft status-related labels
	statusLabels, err := GetRunStatusPromLabels(object)
	if err != nil {
		return err
	}

	runStatusLabelStatusValue := 0
	if statusLabels["status"] == "success" {
		runStatusLabelStatusValue = 1
	}

	// Prepare labels for '_status' metric
	maps.Copy(statusLabels, commonLabels)
	statusLabelMap := prometheus.Labels(statusLabels)

	// 3. Craft duration-related labels
	durationLabels, err := GetRunDurationPromLabels(object)
	if err != nil {
		return err
	}

	// Calculate duration for the Run object
	runDurationValue := 0
	if durationLabels["start_timestamp"] != "#" && durationLabels["completion_timestamp"] != "#" {
		runStartTime, _ := strconv.Atoi(durationLabels["start_timestamp"])
		runCompletionTime, _ := strconv.Atoi(durationLabels["completion_timestamp"])
		runDurationValue = runCompletionTime - runStartTime
	}

	// Prepare labels for '_duration' metric
	maps.Copy(durationLabels, commonLabels)
	durationLabelMap := prometheus.Labels(durationLabels)

	///////////////////////////////////////////////////////

	switch eventType {
	case watch.Added:
		globals.ExecContext.Logger.With(zap.Any("labels", statusLabelMap)).
			Info("PipelineRun resource created. Exposing metrics...")
		metrics.Pool.PipelineRunStatus.With(statusLabelMap).Set(float64(runStatusLabelStatusValue))
		metrics.Pool.PipelineRunDuration.With(durationLabelMap).Set(float64(runDurationValue))

	case watch.Modified:
		globals.ExecContext.Logger.With(zap.Any("labels", statusLabelMap)).
			Info("PipelineRun resource modified. Updating metrics...")

		// Delete metrics that partially match labels
		_ = metrics.Pool.PipelineRunStatus.DeletePartialMatch(commonLabelsProm)
		_ = metrics.Pool.PipelineRunDuration.DeletePartialMatch(commonLabelsProm)

		// Regenerate the metric with newer labels
		metrics.Pool.PipelineRunStatus.With(statusLabelMap).Set(float64(runStatusLabelStatusValue))
		metrics.Pool.PipelineRunDuration.With(durationLabelMap).Set(float64(runDurationValue))

	case watch.Deleted:
		globals.ExecContext.Logger.With(zap.Any("labels", commonLabelsProm)).
			Info("PipelineRun resource deleted. Cleaning up metrics...")
		_ = metrics.Pool.PipelineRunStatus.DeletePartialMatch(commonLabelsProm)
		_ = metrics.Pool.PipelineRunDuration.DeletePartialMatch(commonLabelsProm)
	}

	return nil
}

// WatchTaskRuns TODO
// Hey!, this function is intended to be executed as a go routine
func WatchTaskRuns(ctx *context.Context, client *dynamic.DynamicClient) (err error) {
	globals.ExecContext.Logger.Info(watchTaskrunMessage)

	// Create a watcher for TaskRun resources
	taskRunWatcher, err := client.Resource(taskRunV1GVR).Watch(*ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	defer taskRunWatcher.Stop()

	for taskRunEvent := range taskRunWatcher.ResultChan() {
		// Extract the unstructured object from the event
		unstructuredObject, err := GetUnstructuredFromRuntimeObject(&taskRunEvent.Object)
		if err != nil {
			globals.ExecContext.Logger.Errorf("failed to parse object: %v", err)
			continue
		}

		// Process the TaskRun event
		err = ProcessTaskRunEvent(ctx, &unstructuredObject, taskRunEvent.Type)
		if err != nil {
			globals.ExecContext.Logger.Errorf("failed to process TaskRun event: %v", err)
			continue
		}
	}

	return nil
}

// TODO
func ProcessTaskRunEvent(ctx *context.Context, object *map[string]interface{}, eventType watch.EventType) error {

	// 1. Craft common labels
	objectBasicData, err := GetObjectBasicData(object)
	if err != nil {
		return err
	}

	commonLabels := map[string]string{
		"name":      objectBasicData["name"].(string),
		"namespace": objectBasicData["namespace"].(string),
	}

	// 2. Craft populated labels from TaskRun object labels and merge them
	populatedLabels, err := GetRunPopulatedPromLabels(ctx, object)
	if err != nil {
		return err
	}

	maps.Copy(commonLabels, populatedLabels)

	// Conversion to a Prometheus SDK Labels type will be needed later
	// Maps in golang are ReferenceTypes, so we need to iterate to copy
	commonLabelsProm := prometheus.Labels{}
	for k, v := range commonLabels {
		commonLabelsProm[k] = v
	}

	// 3. Craft status-related labels
	statusLabels, err := GetRunStatusPromLabels(object)
	if err != nil {
		return err
	}

	runStatusLabelStatusValue := 0
	if statusLabels["status"] == "success" {
		runStatusLabelStatusValue = 1
	}

	// Prepare labels for '_status' metric
	maps.Copy(statusLabels, commonLabels)
	statusLabelMap := prometheus.Labels(statusLabels)

	// 3. Craft duration-related labels
	durationLabels, err := GetRunDurationPromLabels(object)
	if err != nil {
		return err
	}

	// Calculate duration for the Run object
	runDurationValue := 0
	if durationLabels["start_timestamp"] != "#" && durationLabels["completion_timestamp"] != "#" {
		runStartTime, _ := strconv.Atoi(durationLabels["start_timestamp"])
		runCompletionTime, _ := strconv.Atoi(durationLabels["completion_timestamp"])
		runDurationValue = runCompletionTime - runStartTime
	}

	// Prepare labels for '_duration' metric
	maps.Copy(durationLabels, commonLabels)
	durationLabelMap := prometheus.Labels(durationLabels)

	///////////////////////////////////////////////////////

	switch eventType {
	case watch.Added:
		globals.ExecContext.Logger.With(zap.Any("labels", statusLabelMap)).
			Info("TaskRun resource created. Exposing metrics...")
		metrics.Pool.TaskRunStatus.With(statusLabelMap).Set(float64(runStatusLabelStatusValue))
		metrics.Pool.TaskRunDuration.With(durationLabelMap).Set(float64(runDurationValue))

	case watch.Modified:
		globals.ExecContext.Logger.With(zap.Any("labels", statusLabelMap)).
			Info("TaskRun resource modified. Updating metrics...")

		// Delete metrics that partially match labels
		_ = metrics.Pool.TaskRunStatus.DeletePartialMatch(commonLabelsProm)
		_ = metrics.Pool.TaskRunDuration.DeletePartialMatch(commonLabelsProm)

		// Regenerate the metric with newer labels
		metrics.Pool.TaskRunStatus.With(statusLabelMap).Set(float64(runStatusLabelStatusValue))
		metrics.Pool.TaskRunDuration.With(durationLabelMap).Set(float64(runDurationValue))

	case watch.Deleted:
		globals.ExecContext.Logger.With(zap.Any("labels", commonLabelsProm)).
			Info("TaskRun resource deleted. Cleaning up metrics...")
		_ = metrics.Pool.TaskRunStatus.DeletePartialMatch(commonLabelsProm)
		_ = metrics.Pool.TaskRunDuration.DeletePartialMatch(commonLabelsProm)
	}

	return nil
}
