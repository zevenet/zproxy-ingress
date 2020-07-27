package watchers

import (
	types "github.com/zevenet/zproxy-ingress/pkg/types"
	v1 "k8s.io/api/core/v1"
	fields "k8s.io/apimachinery/pkg/fields"
	kubernetes "k8s.io/client-go/kubernetes"
	cache "k8s.io/client-go/tools/cache"
)

var (
	resourceName   = "secrets"
	resourceStruct = v1.Secret{}
)

// GetServiceListWatch makes a ListWatch of every Service in the cluster.
func GetSecretListWatch(clientset *kubernetes.Clientset) *cache.ListWatch {

	listwatch := cache.NewListWatchFromClient(
		clientset.CoreV1().RESTClient(), // REST interface
		resourceName,                    // Resource to watch for
		v1.NamespaceAll,                 // Resource can be found in ALL namespaces
		fields.Everything(),             // Get objects controlled by zproxy-ingress controller
	)

	return listwatch
}

// GetServiceController returns a Controller based on listWatch.
// Exports every message into logChannel.
func GetSecretController(listWatch *cache.ListWatch, logChannel chan string, globalConfig *types.Config) cache.Controller {
	return getController(listWatch, &resourceStruct, resourceName, logChannel, globalConfig)
}
