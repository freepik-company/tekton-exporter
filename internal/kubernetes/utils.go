package kubernetes

import (
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"

	//
	"tekton-exporter/internal/globals"
)

// GetObjectBasicData return basic data (name, namespace) from an object of type runtime.Object
func GetObjectBasicData(obj *runtime.Object) (objectData map[string]interface{}, err error) {

	// Convert the runtime.Object to unstructured.Unstructured for convenience
	object, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return objectData, err
	}

	objectData = make(map[string]interface{})

	objectMetadata := object["metadata"]
	objectData["name"] = objectMetadata.(map[string]interface{})["name"]
	objectData["namespace"] = objectMetadata.(map[string]interface{})["namespace"]

	return objectData, nil
}

// GetObjectLabels return all the labels from an object of type runtime.Object
func GetObjectLabels(obj *runtime.Object) (labelsMap map[string]string, err error) {

	// Convert the runtime.Object to unstructured.Unstructured for convenience
	object, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return labelsMap, err
	}

	// Read labels from event's resource
	objectMetadata := object["metadata"].(map[string]interface{})

	// If there is no label, just quit
	objectLabelsOriginal := objectMetadata["labels"]
	if reflect.TypeOf(objectLabelsOriginal) == nil {
		return labelsMap, nil
	}
	objectLabels := objectLabelsOriginal.(map[string]interface{})

	labelsMap = make(map[string]string)

	// Iterate over the original map and cast its values
	for key, value := range objectLabels {
		strValue, ok := value.(string)
		if !ok {
			globals.ExecContext.Logger.Infof("Value of label '%s' is not a string. Ignoring it", key)
			continue
		}
		labelsMap[key] = strValue
	}

	return labelsMap, err
}

// GetObjectCondition return selected condition type from an object of type runtime.Object
func GetObjectCondition(obj *runtime.Object, conditionType string) (condition map[string]interface{}, err error) {

	objectStatus, err := GetObjectStatus(obj)

	if reflect.TypeOf(objectStatus) == nil || reflect.TypeOf(objectStatus["conditions"]) == nil {
		return condition, nil
	}

	//
	objectStatusConditions := objectStatus["conditions"].([]interface{})
	for _, currentCondition := range objectStatusConditions {
		currentConditionMap := currentCondition.(map[string]interface{})

		if currentConditionMap["type"] == conditionType {
			condition = currentConditionMap
			break
		}
	}

	return condition, nil
}

// GetObjectStatus return status from an object of type runtime.Object
func GetObjectStatus(obj *runtime.Object) (objectStatus map[string]interface{}, err error) {

	// Convert the runtime.Object to unstructured.Unstructured for convenience
	object, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return objectStatus, err
	}

	// TODO: Evaluate whether this condition is what we want to do
	if reflect.TypeOf(object["status"]) == nil {
		return objectStatus, nil
	}

	objectStatus = object["status"].(map[string]interface{})
	return objectStatus, nil
}
