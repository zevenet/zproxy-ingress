package watcher

import (
	config "github.com/zevenet/zproxy-ingress/pkg/config"
	fields "k8s.io/apimachinery/pkg/fields"
	kubernetes "k8s.io/client-go/kubernetes"
	cache "k8s.io/client-go/tools/cache"
	"os"
)

var (
	resourceNameConfigMap = "configmaps"
)

// GetServiceListWatch makes a ListWatch of every Service in the cluster.
func GetConfigmapListWatch(clientset *kubernetes.Clientset) *cache.ListWatch {

	objectName := config.Settings.Client.ConfigMapName
	resourceNameSpaceConfigMap := os.Getenv("POD_NAMESPACE")

	listwatch := cache.NewListWatchFromClient(
		clientset.CoreV1().RESTClient(),                          // REST interface
		resourceNameConfigMap,                                    // Resource to watch for
		resourceNameSpaceConfigMap,                               // Resource can be found in ALL namespaces
		fields.OneTermEqualSelector("metadata.name", objectName), // Get objects controlled by zproxy-ingress controller
	)

	return listwatch
}
