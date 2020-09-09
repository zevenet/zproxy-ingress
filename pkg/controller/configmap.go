package controller

import (
	"fmt"

	config "github.com/zevenet/zproxy-ingress/pkg/config"
	log "github.com/zevenet/zproxy-ingress/pkg/logs"
	watcher "github.com/zevenet/zproxy-ingress/pkg/watcher"
	zproxy "github.com/zevenet/zproxy-ingress/pkg/zproxy"
	v1 "k8s.io/api/core/v1"
	kubernetes "k8s.io/client-go/kubernetes"
	cache "k8s.io/client-go/tools/cache"
)

// GetServiceController returns a Controller based on listWatch.
// Exports every message into logChannel.
func GetConfigMapController(clientset *kubernetes.Clientset) cache.Controller {

	listWatch := watcher.GetConfigmapListWatch(clientset)

	eventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc:    addConfigMap,
		DeleteFunc: deleteConfigMap,
		UpdateFunc: updateConfigMap,
	}

	_, controller := cache.NewInformer(
		listWatch,
		&v1.ConfigMap{},
		0,
		eventHandler,
	)

	return controller
}

// overwrite the certificate file
// if the certificate is not valid, does not overwrite the old one

func updateDefaultConfig(cfgMap *v1.ConfigMap) bool {

	lock := zproxy.Lock()

	// reset cfg
	config.Init()

	if cfgMap != nil {
		config.ReplaceParams(cfgMap.Data)
		msg := fmt.Sprintf("replaced default params:  %+v \n", config.Settings)
		log.Print(2, msg)
	}

	zproxy.UpdateIngressCfg()

	zproxy.Unlock(lock)

	return true
}

func addConfigMap(obj interface{}) {

	cfgMap := obj.(*v1.ConfigMap)

	updateDefaultConfig(cfgMap)
}

func updateConfigMap(_, obj interface{}) {

	cfgMap := obj.(*v1.ConfigMap)

	updateDefaultConfig(cfgMap)
}

func deleteConfigMap(obj interface{}) {

	updateDefaultConfig(nil)
}
