package kubernetes

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/runtime"
	"log"
	"maps"

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
	log.Print(list.Items[0])

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

	// TODO: Delete the namespace once the controller is fully working
	pipelineRunWatcher, err := client.Resource(resourceId).Namespace("freeclip").Watch(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	counter := 0 // TODO: Debug Purposes
	for pipelineRunEvent := range pipelineRunWatcher.ResultChan() {

		log.Print("Event on the door") // TODO: Debug Purposes
		// Convert the runtime.Object to unstructured.Unstructured for convenience
		//pipelineRunObject, err := runtime.DefaultUnstructuredConverter.ToUnstructured(pipelineRunEvent.Object)
		//if err != nil {
		//	return err
		//}

		// LETS INSPECT THE STATUS GetObjectCondition
		condition, err := GetObjectCondition(&pipelineRunEvent.Object, "Succeeded")
		if err != nil {
			return err
		}
		log.Print(condition["type"])
		log.Print(condition["status"])
		log.Print(condition["reason"])

		/////////////////////////////////////////

		// Read labels from event's resource and merge them with
		populatedLabels, _ = GetObjectPopulatedLabels(ctx, &pipelineRunEvent.Object)

		switch pipelineRunEvent.Type {
		// TODO
		case watch.Added, watch.Modified:
			log.Print("Added or Modified") // TODO: Debug Purposes

			// TODO: Research the status and the reason to set the labels properly
			calculatedLabels = map[string]string{
				"status": "failed",
				"reason": "joe",
			}

			log.Print("/Added or Modified") // TODO: Debug Purposes

		// TODO
		case watch.Deleted:
			log.Print("Deleted")  // TODO: Debug Purposes
			log.Print("/Deleted") // TODO: Debug Purposes
		}

		// Prepare labels result
		maps.Copy(populatedLabels, calculatedLabels)
		metricsLabels := prometheus.Labels(populatedLabels)
		log.Print(populatedLabels)

		// TODO
		pepon := float64(counter) // TODO: Debug Purposes
		//prometheus.MPipelineRunStatus.WithLabelValues("pepito").Set(pepon)
		metrics.Pool.PipelineRunStatus.With(metricsLabels).Set(pepon)

		counter++                       // TODO: Debug Purposes
		log.Print("/Event on the door") // TODO: Debug Purposes
	}

	return nil
}

// GetObjectLabels TODO return all the labels from an object of type runtime.Object
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
	for _, labelName := range populatedLabelsFlag {

		// Honor all labels when wildcard is passed
		//if len(populatedLabelsFlag) == 1 && labelName == "*" {
		//	populatedLabels = objectLabels
		//	break
		//}

		// Fill only user's requested labels
		if _, objectLabelsFound := objectLabels[labelName]; objectLabelsFound {
			populatedLabels[labelName] = objectLabels[labelName]
		}
	}

	return populatedLabels, nil
}

// GetObjectCondition TODO
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
