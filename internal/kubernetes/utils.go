package kubernetes

import (
	"errors"
	"k8s.io/apimachinery/pkg/runtime"
	//
	"tekton-exporter/internal/globals"
)

// GetUnstructuredFromRuntimeObject converts the runtime.Object to unstructured.Unstructured
func GetUnstructuredFromRuntimeObject(obj *runtime.Object) (objectData map[string]interface{}, err error) {
	objectData, err = runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	return objectData, err
}

// GetObjectBasicData return basic data (name, namespace) from an object
func GetObjectBasicData(object *map[string]interface{}) (objectData map[string]interface{}, err error) {

	metadata, ok := (*object)["metadata"].(map[string]interface{})
	if !ok {
		err = errors.New("metadata not found or not in expected format")
		return
	}

	objectData = make(map[string]interface{})
	objectData["name"] = metadata["name"]
	objectData["namespace"] = metadata["namespace"]

	return objectData, nil
}

// GetObjectLabels return all the labels from an object
func GetObjectLabels(object *map[string]interface{}) (labelsMap map[string]string, err error) {
	labelsMap = make(map[string]string)

	objectMetadata, ok := (*object)["metadata"].(map[string]interface{})
	if !ok {
		err = errors.New("metadata not found or not in expected format")
		return
	}

	// If there is no label, just quit
	objectLabelsOriginal, exists := objectMetadata["labels"]
	if !exists {
		return labelsMap, nil
	}

	objectLabels, ok := objectLabelsOriginal.(map[string]interface{})
	if !ok {
		err = errors.New("labels not in expected format")
		return
	}

	// Iterate over the original map and cast its values
	for key, value := range objectLabels {
		strValue, ok := value.(string)
		if !ok {
			globals.ExecContext.Logger.Infof("value of label '%s' is not a string. Ignoring it", key)
			continue
		}
		labelsMap[key] = strValue
	}

	return labelsMap, err
}

// GetObjectStatus return status from an object
func GetObjectStatus(object *map[string]interface{}) (objectStatus map[string]interface{}, err error) {

	status, statusOk := (*object)["status"]
	if statusOk {

		objectStatus, objectStatusConverted := status.(map[string]interface{})
		if objectStatusConverted {
			return objectStatus, nil
		} else {
			return nil, errors.New("status field is not in the expected format")
		}
	}

	return nil, errors.New("status field is not present in the object")
}

// GetObjectCondition return selected condition type from an object
func GetObjectCondition(object *map[string]interface{}, conditionType string) (condition map[string]interface{}, err error) {

	objectStatus, err := GetObjectStatus(object)
	if err != nil {
		return nil, err
	}

	if objectStatus == nil {
		return nil, errors.New("status field is not present in the object")
	}

	conditions, ok := objectStatus["conditions"].([]interface{})
	if !ok {
		return nil, errors.New("conditions field is not in the expected format")
	}

	for _, currentCondition := range conditions {
		currentConditionMap, ok := currentCondition.(map[string]interface{})
		if !ok {
			return nil, errors.New("condition is not in the expected format")
		}

		if currentConditionMap["type"] == conditionType {
			return currentConditionMap, nil
		}
	}

	return nil, errors.New("condition type not found")
}
