package watcher

import (
	v1 "k8s.io/api/core/v1"
	fields "k8s.io/apimachinery/pkg/fields"
	kubernetes "k8s.io/client-go/kubernetes"
	cache "k8s.io/client-go/tools/cache"
)

var (
	resourceNameSecret = "secrets"
)

// GetServiceListWatch makes a ListWatch of every Service in the cluster.
func GetSecretListWatch(clientset *kubernetes.Clientset) *cache.ListWatch {

	listwatch := cache.NewListWatchFromClient(
		clientset.CoreV1().RESTClient(), // REST interface
		resourceNameSecret,              // Resource to watch for
		v1.NamespaceAll,                 // Resource can be found in ALL namespaces
		fields.Everything(),             // Get objects controlled by zproxy-ingress controller
	)

	return listwatch
}
