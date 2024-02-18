package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	"maps"
	"reflect"
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

// WatchPipelineRuns TODO
// Hey!, this function is intended to be executed as a go routine
func WatchPipelineRuns(ctx *context.Context, client *dynamic.DynamicClient) (err error) {

	commonLabels := map[string]string{}
	populatedLabels := map[string]string{}
	statusLabels := map[string]string{}

	//globals.ExecContext.Logger.Info("Watching PipelineRun objects")
	resourceId := schema.GroupVersionResource{
		Group:    "tekton.dev",
		Version:  "v1",
		Resource: "pipelineruns",
	}

	// TODO
	pipelineRunWatcher, err := client.Resource(resourceId).Watch(*ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for pipelineRunEvent := range pipelineRunWatcher.ResultChan() {

		// 1. Craft common labels
		objectBasicData, err := GetObjectBasicData(&pipelineRunEvent.Object)
		commonLabels = map[string]string{
			"name":      objectBasicData["name"].(string),
			"namespace": objectBasicData["namespace"].(string),
		}

		populatedLabels, _ = GetRunPopulatedLabels(ctx, &pipelineRunEvent.Object)
		maps.Copy(commonLabels, populatedLabels)

		// Conversion to a Prometheus SDK Labels type will be needed later
		// Maps in golang are ReferenceTypes, so we need to iterate to copy
		commonLabelsProm := prometheus.Labels{}
		for k, v := range commonLabels {
			commonLabelsProm[k] = v
		}

		// 2. Craft status-related labels
		statusLabels, err = GetRunStatusLabels(&pipelineRunEvent.Object)
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
		durationLabels, err := GetRunDurationLabels(&pipelineRunEvent.Object)
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

		switch pipelineRunEvent.Type {
		case watch.Added:
			globals.ExecContext.Logger.With(zap.Any("labels", statusLabelMap)).
				Info("a PipelineRun resource has been created. Exposing...")

			metrics.Pool.PipelineRunStatus.With(statusLabelMap).Set(float64(runStatusLabelStatusValue))
			metrics.Pool.PipelineRunDuration.With(durationLabelMap).Set(float64(runDurationValue))

		case watch.Modified:
			globals.ExecContext.Logger.With(zap.Any("labels", statusLabelMap)).
				Info("a PipelineRun resource has been modified. Exchanging it...")

			// Delete metrics that partially match labels
			_ = metrics.Pool.PipelineRunStatus.DeletePartialMatch(commonLabelsProm)
			_ = metrics.Pool.PipelineRunDuration.DeletePartialMatch(commonLabelsProm)

			// Regenerate the metric with newer labels
			metrics.Pool.PipelineRunStatus.With(statusLabelMap).Set(float64(runStatusLabelStatusValue))
			metrics.Pool.PipelineRunDuration.With(durationLabelMap).Set(float64(runDurationValue))

		case watch.Deleted:
			globals.ExecContext.Logger.With(zap.Any("labels", commonLabelsProm)).
				Info("a PipelineRun resource has been deleted. Cleaning...")

			_ = metrics.Pool.PipelineRunStatus.DeletePartialMatch(commonLabelsProm)
			_ = metrics.Pool.PipelineRunDuration.DeletePartialMatch(commonLabelsProm)
		}
	}

	return nil
}

// WatchTaskRuns TODO
// Hey!, this function is intended to be executed as a go routine
func WatchTaskRuns(ctx *context.Context, client *dynamic.DynamicClient) (err error) {

	commonLabels := map[string]string{}
	populatedLabels := map[string]string{}
	statusLabels := map[string]string{}

	//globals.ExecContext.Logger.Info("Watching PipelineRun objects")
	resourceId := schema.GroupVersionResource{
		Group:    "tekton.dev",
		Version:  "v1",
		Resource: "taskruns",
	}

	// TODO: Delete the namespace once the controller is fully working
	taskRunWatcher, err := client.Resource(resourceId).Namespace("freeclip").Watch(*ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for taskRunEvent := range taskRunWatcher.ResultChan() {

		// 1. Craft common labels
		objectBasicData, err := GetObjectBasicData(&taskRunEvent.Object)
		commonLabels = map[string]string{
			"name":      objectBasicData["name"].(string),
			"namespace": objectBasicData["namespace"].(string),
		}

		populatedLabels, _ = GetRunPopulatedLabels(ctx, &taskRunEvent.Object)
		maps.Copy(commonLabels, populatedLabels)

		// Conversion to a Prometheus SDK Labels type will be needed later
		// Maps in golang are ReferenceTypes, so we need to iterate to copy
		commonLabelsProm := prometheus.Labels{}
		for k, v := range commonLabels {
			commonLabelsProm[k] = v
		}

		// 2. Craft status-related labels
		statusLabels, err = GetRunStatusLabels(&taskRunEvent.Object)
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
		durationLabels, err := GetRunDurationLabels(&taskRunEvent.Object)
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

		///////

		switch taskRunEvent.Type {
		case watch.Added:
			globals.ExecContext.Logger.With(zap.Any("labels", statusLabelMap)).
				Info("a TaskRun resource has been created. Exposing...")

			metrics.Pool.TaskRunStatus.With(statusLabelMap).Set(float64(runStatusLabelStatusValue))
			metrics.Pool.TaskRunDuration.With(durationLabelMap).Set(float64(runDurationValue))

		case watch.Modified:
			globals.ExecContext.Logger.With(zap.Any("labels", statusLabelMap)).
				Info("a TaskRun resource has been modified. Exchanging it...")

			// Delete metrics that partially match labels
			_ = metrics.Pool.TaskRunStatus.DeletePartialMatch(commonLabelsProm)
			_ = metrics.Pool.TaskRunDuration.DeletePartialMatch(commonLabelsProm)

			// Regenerate the metric with newer labels
			metrics.Pool.TaskRunStatus.With(statusLabelMap).Set(float64(runStatusLabelStatusValue))
			metrics.Pool.TaskRunDuration.With(durationLabelMap).Set(float64(runDurationValue))

		case watch.Deleted:
			globals.ExecContext.Logger.With(zap.Any("labels", commonLabelsProm)).
				Info("a TaskRun resource has been deleted. Cleaning...")

			_ = metrics.Pool.TaskRunStatus.DeletePartialMatch(commonLabelsProm)
			_ = metrics.Pool.TaskRunDuration.DeletePartialMatch(commonLabelsProm)
		}
	}

	return nil
}

// GetRunPopulatedLabels return only user's desired labels from an object of type runtime.Object
// Desired labels are defined by flag "--populated-labels"
func GetRunPopulatedLabels(ctx *context.Context, object *runtime.Object) (labelsMap map[string]string, err error) {

	// Read labels from event's resource
	objectLabels, err := GetObjectLabels(object)
	if err != nil {
		return labelsMap, err
	}

	// No existing labels on object, quit
	if len(objectLabels) == 0 {
		return labelsMap, nil
	}

	// Recover flag 'populated-labels' from context
	populatedLabels := map[string]string{}
	populatedLabelsFlag := (*ctx).Value("flag-populated-labels").([]string)

	//
	parsedLabelsMap, _ := metrics.GetProcessedLabels(populatedLabelsFlag) // TODO: Handle error
	for _, populatedLabelName := range populatedLabelsFlag {

		// Populated labels are dynamic, but labels must be pre-registered on Prometheus SDK
		// This is a mechanism to avoid crashes if the labels are not present in the object
		populatedLabels[parsedLabelsMap[populatedLabelName]] = "#"

		// Fill only user's requested labels
		// Label names will be changed to a Prometheus compatible syntax
		if _, objectLabelsFound := objectLabels[populatedLabelName]; objectLabelsFound {
			populatedLabels[parsedLabelsMap[populatedLabelName]] = objectLabels[populatedLabelName]
		}
	}

	return populatedLabels, nil
}

// GetRunStatusLabels obtains the status-related labels for a pipeline based on the 'Succeeded' condition type and
// returns a map containing the 'status' and 'reason' labels.
// If the 'Succeeded' condition is not found, it populates a default condition with status 'False' and reason 'Unknown'.
func GetRunStatusLabels(object *runtime.Object) (labelsMap map[string]string, err error) {

	// Obtain the status of 'Succeeded' condition type
	condition, err := GetObjectCondition(object, "Succeeded")
	if err != nil {
		return labelsMap, err
	}

	// TODO: Should we manage this or make it fail??
	if condition == nil {
		condition = map[string]interface{}{
			"type":   "Succeeded",
			"status": "False",
			"reason": "Unknown",
		}
	}

	// Make the 'status' label understandable in metrics that are using it
	runStatusLabelStatus := "failed"
	if strings.ToLower(condition["status"].(string)) == "true" {
		runStatusLabelStatus = "success"
	}

	statusLabels := map[string]string{
		"status": runStatusLabelStatus,
		"reason": condition["reason"].(string),
	}

	return statusLabels, nil
}

// GetRunDurationLabels return a map with 'start_timestamp' and 'completion_timestamp'
// from the object representing a run's status.
// If any timestamp is missing, it populates a default value set to '#'
func GetRunDurationLabels(object *runtime.Object) (labelsMap map[string]string, err error) {

	// Obtain the status of 'Succeeded' condition type
	status, err := GetObjectStatus(object)
	if err != nil {
		return labelsMap, err
	}

	// TODO: Should we manage this or make it fail??
	if status == nil {
		return labelsMap, nil
	}

	// TODO
	timestampLabels := map[string]string{
		"start_timestamp":      "#",
		"completion_timestamp": "#",
	}

	// Check whether we have start time
	if reflect.TypeOf(status["startTime"]) != nil {
		// TODO: Extract DateToTimestamp code to a function
		// Tekton uses RFC3339 for dates. Convert it to timestamp
		parsedTime, err := time.Parse(time.RFC3339, status["startTime"].(string))
		if err != nil {
			return labelsMap, errors.New(fmt.Sprintf("impossible to parse timestamp: %s", err))
		}

		timestampLabels["start_timestamp"] = strconv.Itoa(int(parsedTime.Unix()))
	}

	// Check whether we have completion time
	if reflect.TypeOf(status["completionTime"]) != nil {
		// TODO: Extract DateToTimestamp code to a function
		// Tekton uses RFC3339 for dates. Convert it to timestamp
		parsedTime, err := time.Parse(time.RFC3339, status["completionTime"].(string))
		if err != nil {
			return labelsMap, errors.New(fmt.Sprintf("impossible to parse timestamp: %s", err))
		}

		timestampLabels["completion_timestamp"] = strconv.Itoa(int(parsedTime.Unix()))
	}

	return timestampLabels, nil
}
