package watchers

import (
	"fmt"

	types "github.com/zevenet/zproxy-ingress/pkg/types"
	funcs "github.com/zevenet/zproxy-ingress/pkg/watchers/funcs"
	v1 "k8s.io/api/core/v1"
	v1beta "k8s.io/api/networking/v1beta1"
	fields "k8s.io/apimachinery/pkg/fields"
	runtime "k8s.io/apimachinery/pkg/runtime"
	kubernetes "k8s.io/client-go/kubernetes"
	cache "k8s.io/client-go/tools/cache"
)

// getListWatch makes a ListWatch of every resource in the cluster.
func getListWatch(clientset *kubernetes.Clientset, resource string) *cache.ListWatch {
	listwatch := cache.NewListWatchFromClient(
		clientset.NetworkingV1beta1().RESTClient(), // REST interface
		resource,            // Resource to watch for
		v1.NamespaceAll,     // Resource can be found in ALL namespaces
		fields.Everything(), // Get objects controlled by zproxy-ingress controller
	)

	return listwatch
}

// getController returns a Controller based on listWatch.
// Exports every message into logChannel.
func getController(listWatch *cache.ListWatch, resourceStruct runtime.Object, resourceName string, logChannel chan string, globalCfg *types.Config) cache.Controller {
	_, controller := cache.NewInformer(
		listWatch,      // Resources to watch for
		resourceStruct, // Resource struct
		0,

		// Event handler: new, deleted or updated resource
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				switch tp := obj.(type) {
				case *v1beta.Ingress:
					funcs.AddIngress(obj.(*v1beta.Ingress), globalCfg)
				case *v1.Secret:
					funcs.CreateCertificateFile(obj.(*v1.Secret))
				default:
					err := fmt.Sprintf("Object not recognised of type %t", tp)
					panic(err)
				}
				//~ logChannel <- fmt.Sprintf("\nNew %s:\n%s", resourceName, obj)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				switch tp := oldObj.(type) {
				case *v1beta.Ingress:
					funcs.UpdateIngress(oldObj.(*v1beta.Ingress), newObj.(*v1beta.Ingress), globalCfg)
				case *v1.Secret:
					funcs.UpdateCertificate(newObj.(*v1.Secret), globalCfg)
				default:
					err := fmt.Sprintf("Object not recognised of type %t", tp)
					panic(err)
				}
				//~ logChannel <- fmt.Sprintf("\nUpdated %s:\n* BEFORE: %s\n* NOW: %s", resourceName, oldObj, newObj)
			},
			DeleteFunc: func(obj interface{}) {
				switch tp := obj.(type) {
				case *v1beta.Ingress:
					funcs.DeleteIngress(obj.(*v1beta.Ingress), globalCfg)
				case *v1.Secret:
					funcs.DeleteCertificate(obj.(*v1.Secret), globalCfg)
				default:
					err := fmt.Sprintf("Object not recognised of type %t", tp)
					panic(err)
				}

				//~ logChannel <- fmt.Sprintf("\nDelete %s:\n%s", resourceName, obj)
			},
		},
	)
	return controller
}
