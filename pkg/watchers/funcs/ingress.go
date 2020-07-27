package funcs

import (
	"fmt"

	config "github.com/zevenet/zproxy-ingress/pkg/config"
	log "github.com/zevenet/zproxy-ingress/pkg/logs"
	types "github.com/zevenet/zproxy-ingress/pkg/types"
	v1beta "k8s.io/api/networking/v1beta1"
)

//~ var IngessesList = []*v1beta.Ingress

//
func UpdateIngressCfg(ingressObj *v1beta.Ingress, globalCfg *types.Config) {

	if ingressObj.Spec.IngressClassName == nil {
		message := fmt.Sprintf("Object %s without ingressClass\n", ingressObj.ObjectMeta.Name)
		log.Print(1, message)
		return
	}
	if *ingressObj.Spec.IngressClassName != "zproxy-ingress" {
		message := fmt.Sprintf("Skipt object %s, ingressClass = %s\n", ingressObj.ObjectMeta.Name, ingressObj.Spec.IngressClassName)
		log.Print(1, message)
		return
	}

	fileName := globalCfg.Global.ConfigFile // file name
	msg := "cfg file: " + fileName
	log.Print(1, msg)

	if config.CreateProxyConfig(fileName, ingressObj, globalCfg) != 0 {
		log.Print(0, "Error creating the config file")
		return
	}

	// Reload
	if config.ReloadDaemon(globalCfg) != 0 {
		log.Print(0, "Error reloading zproxy daemon")
		return
	}

	// Prints info
	message := fmt.Sprintf("Updated object: %+v\n", ingressObj)
	log.Print(1, message)
}
