package controller

import (
	"fmt"
	"github.com/juju/fslock"
	"time"

	log "github.com/zevenet/zproxy-ingress/pkg/logs"
	watcher "github.com/zevenet/zproxy-ingress/pkg/watcher"
	zproxy "github.com/zevenet/zproxy-ingress/pkg/zproxy"
	v1beta "k8s.io/api/networking/v1beta1"
	kubernetes "k8s.io/client-go/kubernetes"
	cache "k8s.io/client-go/tools/cache"
)

// Global variables
var lf = "/tmp/proxy_cfg.lock" // lock file
var ingressesList []*v1beta.Ingress

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

	for i := range ingressesList {
		if ingressesList[i].ObjectMeta.Name == oldIngressObj.ObjectMeta.Name &&
			ingressesList[i].ObjectMeta.Namespace == oldIngressObj.ObjectMeta.Namespace {
			return i
		}
	}

	return -1
}

func updateIngressCfg(ingressObjList []*v1beta.Ingress) bool {

	start := time.Now()

	if zproxy.CreateProxyConfig(ingressObjList) != 0 {
		log.Print(0, "Error creating the config file")
		return false
	}

	if zproxy.ReloadDaemon() != 0 {
		log.Print(0, "Error reloading zproxy daemon")
		return false
	}

	if log.GetLevel() > 0 {
		elapsed := time.Since(start)
		msg := fmt.Sprintf("The reloading took \"%s\"", elapsed)
		log.Print(1, msg)
	}

	log.Print(1, "Ingress configuration was reloaded properly")

	return true
}

func addIngress(obj interface{}) {

	ingressObj := obj.(*v1beta.Ingress)

	if ingressObj == nil {
		return
	}

	if !validateIngress(ingressObj) {
		return
	}

	lock := fslock.New(lf)

	ingressesList = append(ingressesList, ingressObj)

	if !updateIngressCfg(ingressesList) {
		ingressesList = ingressesList[:len(ingressesList)-1]
	}

	lock.Unlock()
}

func deleteIngress(obj interface{}) {

	ingressObj := obj.(*v1beta.Ingress)

	if !validateIngress(ingressObj) {
		return
	}

	lock := fslock.New(lf)

	index := getIngressIndex(ingressObj)
	if index == -1 {
		log.Print(0, "Ingress object was not found")
		return
	}
	ingressesList = append(ingressesList[:index], ingressesList[index+1:]...) // remove element

	updateIngressCfg(ingressesList)

	lock.Unlock()
}

func updateIngress(obj interface{}, newobj interface{}) {

	oldIngressObj := obj.(*v1beta.Ingress)
	ingressObj := newobj.(*v1beta.Ingress)

	if !validateIngress(ingressObj) {
		return
	}

	lock := fslock.New(lf)

	oldList := ingressesList

	// replace in the list
	index := getIngressIndex(oldIngressObj)
	if index == -1 {
		log.Print(0, "Old ingress object was not found")
		return
	}
	ingressesList[index] = ingressObj

	if !updateIngressCfg(ingressesList) {
		ingressesList = oldList
	}

	lock.Unlock()
}
