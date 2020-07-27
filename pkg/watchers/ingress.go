package watchers

import (
	types "github.com/zevenet/zproxy-ingress/pkg/types"
	v1beta "k8s.io/api/networking/v1beta1"
	kubernetes "k8s.io/client-go/kubernetes"
	cache "k8s.io/client-go/tools/cache"
)

var (
	resourceNameIngress   = "ingresses"
	resourceStructIngress = v1beta.Ingress{}
)

// GetServiceListWatch makes a ListWatch of every Service in the cluster.
func GetIngressListWatch(clientset *kubernetes.Clientset) *cache.ListWatch {
	return getListWatch(clientset, resourceNameIngress)
}

// GetServiceController returns a Controller based on listWatch.
// Exports every message into logChannel.
func GetIngressController(listWatch *cache.ListWatch, logChannel chan string, globalConfig *types.Config) cache.Controller {
	return getController(listWatch, &resourceStructIngress, resourceNameIngress, logChannel, globalConfig)
}
