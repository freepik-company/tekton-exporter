package kubernetes

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	"maps"
	"strings"

	// Kubernetes clients
	// Ref: https://pkg.go.dev/k8s.io/client-go/dynamic
	"k8s.io/client-go/dynamic"
	// Ref: https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client/config
	ctrl "sigs.k8s.io/controller-runtime"

	// Kubernetes types
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"

	//
	"tekton-exporter/internal/globals"
	"tekton-exporter/internal/metrics"
)

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

// GetNamespaces get a list of all namespaces existing in the cluster
// TODO: Evaluate if this method is needed
func GetNamespaces(ctx context.Context, client *dynamic.DynamicClient) (namespaces *unstructured.UnstructuredList, err error) {

	resourceId := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "namespaces",
	}

	namespaceList, err := client.Resource(resourceId).List(ctx, metav1.ListOptions{})

	if err != nil {
		return namespaces, err
	}

	return namespaceList, err
}

// GetAllPipelineRuns
// TODO: Evaluate if this method is needed
func GetAllPipelineRuns(ctx context.Context, client *dynamic.DynamicClient) (resources []unstructured.Unstructured, err error) {

	globals.ExecContext.Logger.Info("pepe")

	resourceId := schema.GroupVersionResource{
		Group:    "tekton.dev",
		Version:  "v1",
		Resource: "pipelineruns",
	}

	list, err := client.Resource(resourceId).Namespace("freeclip").List(ctx, metav1.ListOptions{})
	_ = list
	//log.Print(list.Items[0])

	return resources, nil
}

// WatchPipelineRuns TODO
func WatchPipelineRuns(ctx context.Context, client *dynamic.DynamicClient) (err error) {

	populatedLabels := map[string]string{}
	calculatedLabels := map[string]string{}

	//globals.ExecContext.Logger.Info("Watching PipelineRun objects")
	resourceId := schema.GroupVersionResource{
		Group:    "tekton.dev",
		Version:  "v1",
		Resource: "pipelineruns",
	}

	// TODO
	pipelineRunWatcher, err := client.Resource(resourceId).Watch(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for pipelineRunEvent := range pipelineRunWatcher.ResultChan() {

		// Obtain the status of 'Succeeded' condition type
		condition, err := GetObjectCondition(&pipelineRunEvent.Object, "Succeeded")
		if err != nil {
			return err
		}

		// Make the status label understandable on the metrics using it
		pipelineRunStatusMetricLabelStatus := "failed"
		pipelineRunStatusMetricLabelStatusValue := 0
		if strings.ToLower(condition["status"].(string)) == "true" {
			pipelineRunStatusMetricLabelStatus = "success"
			pipelineRunStatusMetricLabelStatusValue = 1
		}

		objectBasicData, err := GetObjectBasicData(&pipelineRunEvent.Object)

		// Read labels from event's resource and merge them with
		populatedLabels, _ = GetObjectPopulatedLabels(ctx, &pipelineRunEvent.Object)
		calculatedLabels = map[string]string{
			"name":   objectBasicData["name"].(string),
			"status": pipelineRunStatusMetricLabelStatus,
			"reason": condition["reason"].(string),
		}

		// Prepare labels for Promauto SDK
		maps.Copy(populatedLabels, calculatedLabels)
		metricsLabels := prometheus.Labels(populatedLabels)

		switch pipelineRunEvent.Type {
		case watch.Added, watch.Modified:
			globals.ExecContext.Logger.With(zap.Any("labels", metricsLabels)).
				Info("a PipelineRun resource has been created. Exposing...")
			metrics.Pool.PipelineRunStatus.With(metricsLabels).Set(float64(pipelineRunStatusMetricLabelStatusValue))
		case watch.Deleted:
			collectorDeleted := metrics.Pool.PipelineRunStatus.Delete(metricsLabels)
			if collectorDeleted {
				globals.ExecContext.Logger.With(zap.Any("labels", metricsLabels)).
					Info("a PipelineRun resource has been deleted. Cleaning...")
			}
		}
	}

	return nil
}

// WatchTaskRuns TODO
func WatchTaskRuns(ctx context.Context, client *dynamic.DynamicClient) (err error) {

	populatedLabels := map[string]string{}
	calculatedLabels := map[string]string{}

	//globals.ExecContext.Logger.Info("Watching PipelineRun objects")
	resourceId := schema.GroupVersionResource{
		Group:    "tekton.dev",
		Version:  "v1",
		Resource: "taskruns",
	}

	// TODO: Delete the namespace once the controller is fully working
	taskRunWatcher, err := client.Resource(resourceId).Namespace("freeclip").Watch(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for taskRunEvent := range taskRunWatcher.ResultChan() {

		// Obtain the status of 'Succeeded' condition type
		condition, err := GetObjectCondition(&taskRunEvent.Object, "Succeeded")
		if err != nil {
			return err
		}

		// Make the status label understandable on the metrics using it
		taskRunStatusMetricLabelStatus := "failed"
		taskRunStatusMetricLabelStatusValue := 0
		if strings.ToLower(condition["status"].(string)) == "true" {
			taskRunStatusMetricLabelStatus = "success"
			taskRunStatusMetricLabelStatusValue = 1
		}

		objectBasicData, err := GetObjectBasicData(&taskRunEvent.Object)

		// Read labels from event's resource and merge them with
		populatedLabels, _ = GetObjectPopulatedLabels(ctx, &taskRunEvent.Object)
		calculatedLabels = map[string]string{
			"name":   objectBasicData["name"].(string),
			"status": taskRunStatusMetricLabelStatus,
			"reason": condition["reason"].(string),
		}

		// Prepare labels for Promauto SDK
		maps.Copy(populatedLabels, calculatedLabels)
		metricsLabels := prometheus.Labels(populatedLabels)

		switch taskRunEvent.Type {
		case watch.Added, watch.Modified:
			globals.ExecContext.Logger.With(zap.Any("labels", metricsLabels)).
				Info("a TaskRun resource has been created. Exposing...")
			metrics.Pool.TaskRunStatus.With(metricsLabels).Set(float64(taskRunStatusMetricLabelStatusValue))
		case watch.Deleted:
			collectorDeleted := metrics.Pool.TaskRunStatus.Delete(metricsLabels)
			if collectorDeleted {
				globals.ExecContext.Logger.With(zap.Any("labels", metricsLabels)).
					Info("a TaskRun resource has been deleted. Cleaning...")
			}
		}
	}

	return nil
}

// GetObjectLabels return all the labels from an object of type runtime.Object
func GetObjectLabels(obj *runtime.Object) (labelsMap map[string]string, err error) {

	// Convert the runtime.Object to unstructured.Unstructured for convenience
	pipelineRunObject, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return labelsMap, err
	}

	// Read labels from event's resource
	objectMetadata := pipelineRunObject["metadata"]
	objectLabelsOriginal := objectMetadata.(map[string]interface{})["labels"]
	objectLabels := objectLabelsOriginal.(map[string]interface{})

	labelsMap = make(map[string]string)

	// Iterar sobre el mapa original y hacer el "casting" de los valores
	for key, value := range objectLabels {
		strValue, ok := value.(string)
		if !ok {
			// Manejo del error si el valor no es un string
			fmt.Printf("El valor para '%s' no es un string y no puede ser convertido.\n", key)
			continue // Opcionalmente, puedes decidir cómo manejar este caso.
		}
		labelsMap[key] = strValue
	}

	return labelsMap, err
}

// GetObjectPopulatedLabels return only user's desired labels from an object of type runtime.Object
// Desired labels are defined by flag "--populated-labels"
func GetObjectPopulatedLabels(ctx context.Context, object *runtime.Object) (labelsMap map[string]string, err error) {

	// Read labels from event's resource
	objectLabels, err := GetObjectLabels(object)
	if err != nil {
		return labelsMap, err
	}

	// No existing labels on object, quit
	if len(objectLabels) == 0 {
		return labelsMap, nil
	}

	// TODO
	populatedLabels := map[string]string{}
	populatedLabelsFlag := ctx.Value("flag-populated-labels").([]string)

	//
	parsedLabelsMap, _ := metrics.GetProcessedLabels(populatedLabelsFlag) // TODO: Handle error
	for _, labelName := range populatedLabelsFlag {

		// Fill only user's requested labels
		// Label names will be changed to a Prometheus compatible syntax
		if _, objectLabelsFound := objectLabels[labelName]; objectLabelsFound {
			populatedLabels[parsedLabelsMap[labelName]] = objectLabels[labelName]
		}
	}

	return populatedLabels, nil
}

// GetObjectCondition return selected condition type from an object of type runtime.Object
func GetObjectCondition(obj *runtime.Object, conditionType string) (condition map[string]interface{}, err error) {

	// Convert the runtime.Object to unstructured.Unstructured for convenience
	pipelineRunObject, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return condition, err
	}

	prObjectStatus := pipelineRunObject["status"].(map[string]interface{})
	prObjectStatusConditions := prObjectStatus["conditions"].([]interface{})

	for _, currentCondition := range prObjectStatusConditions {
		currentConditionMap := currentCondition.(map[string]interface{})

		if currentConditionMap["type"] == conditionType {
			condition = currentConditionMap
			break
		}
	}

	return condition, nil
}

// GetObjectBasicData return basic data (name, namespace) from an object of type runtime.Object
func GetObjectBasicData(obj *runtime.Object) (objectData map[string]interface{}, err error) {

	// Convert the runtime.Object to unstructured.Unstructured for convenience
	pipelineRunObject, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return objectData, err
	}

	objectData = make(map[string]interface{})

	objectMetadata := pipelineRunObject["metadata"]
	objectData["name"] = objectMetadata.(map[string]interface{})["name"]
	objectData["namespace"] = objectMetadata.(map[string]interface{})["namespace"]

	return objectData, nil
}
