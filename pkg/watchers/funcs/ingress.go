package funcs

import (
	"fmt"
	"github.com/juju/fslock"
	"time"

	config "github.com/zevenet/zproxy-ingress/pkg/config"
	log "github.com/zevenet/zproxy-ingress/pkg/logs"
	types "github.com/zevenet/zproxy-ingress/pkg/types"
	v1beta "k8s.io/api/networking/v1beta1"
)

var lf = "/tmp/proxy_cfg.lock" // lock file

var ingressesList []*v1beta.Ingress

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

func updateIngressCfg(ingressObjList []*v1beta.Ingress, globalCfg *types.Config) bool {

	start := time.Now()

	if config.CreateProxyConfig(ingressObjList, globalCfg) != 0 {
		log.Print(0, "Error creating the config file")
		return false
	}

	if config.ReloadDaemon(globalCfg) != 0 {
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

func AddIngress(ingressObj *v1beta.Ingress, globalCfg *types.Config) {
	if !validateIngress(ingressObj) {
		return
	}

	lock := fslock.New(lf)

	ingressesList = append(ingressesList, ingressObj)

	if !updateIngressCfg(ingressesList, globalCfg) {
		ingressesList = ingressesList[:len(ingressesList)-1]
	}

	lock.Unlock()
}

func DeleteIngress(ingressObj *v1beta.Ingress, globalCfg *types.Config) {
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

	updateIngressCfg(ingressesList, globalCfg)

	lock.Unlock()
}

func UpdateIngress(oldIngressObj *v1beta.Ingress, ingressObj *v1beta.Ingress, globalCfg *types.Config) {
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

	if !updateIngressCfg(ingressesList, globalCfg) {
		ingressesList = oldList
	}

	lock.Unlock()
}
