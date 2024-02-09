package kubernetes

import (
	"context"
	"log"

	// Kubernetes clients
	// Ref: https://pkg.go.dev/k8s.io/client-go/discovery
	"k8s.io/client-go/dynamic"            // Ref: https://pkg.go.dev/k8s.io/client-go/dynamic
	ctrl "sigs.k8s.io/controller-runtime" // Ref: https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client/config

	// Kubernetes types
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	//
	"tekton-exporter/internal/globals"
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

// GetAllPipelineRuns TODO
func GetAllPipelineRuns(ctx context.Context, client *dynamic.DynamicClient) (resources []unstructured.Unstructured, err error) {

	globals.ExecContext.Logger.Info("pepe")

	resourceId := schema.GroupVersionResource{
		Group:    "tekton.dev",
		Version:  "v1",
		Resource: "pipelineruns",
	}

	list, err := client.Resource(resourceId).Namespace("freeclip").List(ctx, metav1.ListOptions{})

	log.Print(list.Items[0])

	//client := &http.Client{}
	//req, err := http.NewRequest("GET", opts.Url+"/api/search?query=&", nil)
	//if err != nil {
	//	return nil, err
	//}
	//
	//req.Header = opts.Headers
	//
	//resp, err := client.Do(req)
	//if err != nil {
	//	return nil, err
	//}
	//defer resp.Body.Close()
	//
	//body, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	return nil, err
	//}
	//
	//var dashboards DashboardList
	//err = json.Unmarshal(body, &dashboards.Dashboards)
	//if err != nil {
	//	return nil, err
	//}

	return resources, nil
}
