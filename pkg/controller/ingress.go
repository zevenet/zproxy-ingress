package controller

import (
	"fmt"

	log "github.com/zevenet/zproxy-ingress/pkg/logs"
	watcher "github.com/zevenet/zproxy-ingress/pkg/watcher"
	zproxy "github.com/zevenet/zproxy-ingress/pkg/zproxy"
	v1beta "k8s.io/api/networking/v1beta1"
	kubernetes "k8s.io/client-go/kubernetes"
	cache "k8s.io/client-go/tools/cache"
)

// GetServiceController returns a Controller based on listWatch.
// Exports every message into logChannel.
func GetIngressController(clientset *kubernetes.Clientset) cache.Controller {

	listWatch := watcher.GetIngressListWatch(clientset)

	eventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc:    addIngress,
		DeleteFunc: deleteIngress,
		UpdateFunc: updateIngress,
	}

	_, controller := cache.NewInformer(
		listWatch,
		&v1beta.Ingress{},
		0,
		eventHandler,
	)

	return controller
}

func validateIngress(ingressObj *v1beta.Ingress) bool {

	if ingressObj.Spec.IngressClassName == nil {
		message := fmt.Sprintf("Object %s without ingressClass\n", ingressObj.ObjectMeta.Name)
		log.Print(1, message)
		return false
	}
	if *ingressObj.Spec.IngressClassName != "zproxy-ingress" {
		message := fmt.Sprintf("Skipt object %s, ingressClass = %s\n", ingressObj.ObjectMeta.Name, ingressObj.Spec.IngressClassName)
		log.Print(1, message)
		return false
	}

	return true
}

func getIngressIndex(oldIngressObj *v1beta.Ingress) int {

	for i := range zproxy.IngressesList {
		if zproxy.IngressesList[i].ObjectMeta.Name == oldIngressObj.ObjectMeta.Name &&
			zproxy.IngressesList[i].ObjectMeta.Namespace == oldIngressObj.ObjectMeta.Namespace {
			return i
		}
	}

	return -1
}

func addIngress(obj interface{}) {

	ingressObj := obj.(*v1beta.Ingress)

	if ingressObj == nil {
		return
	}

	if !validateIngress(ingressObj) {
		return
	}

	msg := fmt.Sprintf("addIngress: %+v", obj)
	log.Print(2, msg)

	lock := zproxy.Lock()

	zproxy.IngressesList = append(zproxy.IngressesList, ingressObj)

	if !zproxy.UpdateIngressCfg() {
		zproxy.IngressesList = zproxy.IngressesList[:len(zproxy.IngressesList)-1]
	}

	zproxy.Unlock(lock)
}

func deleteIngress(obj interface{}) {

	ingressObj := obj.(*v1beta.Ingress)

	if !validateIngress(ingressObj) {
		return
	}

	msg := fmt.Sprintf("deleteIngress: %+v", obj)
	log.Print(2, msg)

	lock := zproxy.Lock()

	index := getIngressIndex(ingressObj)
	if index == -1 {
		log.Print(0, "Ingress object was not found")
		return
	}
	zproxy.IngressesList = append(zproxy.IngressesList[:index], zproxy.IngressesList[index+1:]...) // remove element

	zproxy.UpdateIngressCfg()

	zproxy.Unlock(lock)
}

func updateIngress(obj interface{}, newobj interface{}) {

	oldIngressObj := obj.(*v1beta.Ingress)
	ingressObj := newobj.(*v1beta.Ingress)

	if !validateIngress(ingressObj) {
		return
	}

	msg := fmt.Sprintf("updateIngress old: %+v", obj)
	log.Print(2, msg)
	msg = fmt.Sprintf("updateIngress new: %+v", newobj)
	log.Print(2, msg)

	lock := zproxy.Lock()

	oldList := zproxy.IngressesList

	// replace in the list
	index := getIngressIndex(oldIngressObj)
	if index == -1 {
		log.Print(0, "Old ingress object was not found")
		return
	}
	zproxy.IngressesList[index] = ingressObj

	if !zproxy.UpdateIngressCfg() {
		zproxy.IngressesList = oldList
	}

	zproxy.Unlock(lock)
}
