package watcher

import (
	v1 "k8s.io/api/core/v1"
	v1beta "k8s.io/api/networking/v1beta1"
	fields "k8s.io/apimachinery/pkg/fields"
	kubernetes "k8s.io/client-go/kubernetes"
	cache "k8s.io/client-go/tools/cache"
)

var (
	resourceNameIngress   = "ingresses"
	resourceStructIngress = v1beta.Ingress{}
)

// getListWatch makes a ListWatch of every resource in the cluster.
func GetIngressListWatch(clientset *kubernetes.Clientset) *cache.ListWatch {

	listwatch := cache.NewListWatchFromClient(
		clientset.NetworkingV1beta1().RESTClient(), // REST interface
		resourceNameIngress,                        // Resource to watch for
		v1.NamespaceAll,                            // Resource can be found in ALL namespaces
		fields.Everything(),                        // Get objects controlled by zproxy-ingress controller
	)

	return listwatch
}
