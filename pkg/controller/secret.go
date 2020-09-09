package controller

import (
	"fmt"
	"os"

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
func GetSecretController(clientset *kubernetes.Clientset) cache.Controller {

	listWatch := watcher.GetSecretListWatch(clientset)

	eventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc:    createCertificate,
		DeleteFunc: deleteCertificate,
		UpdateFunc: updateCertificate,
	}

	_, controller := cache.NewInformer(
		listWatch,
		&v1.Secret{},
		0,
		eventHandler,
	)

	return controller
}

// overwrite the certificate file
// if the certificate is not valid, does not overwrite the old one

func createCertificateFile(obj interface{}) bool {

	cert := obj.(*v1.Secret)

	msg := fmt.Sprintf("Checking certificate \"%s\" from the \"%s\" namespace \n", cert.ObjectMeta.Name, cert.ObjectMeta.Namespace)
	log.Print(2, msg)

	if cert.Data["pem"] == nil && !(cert.Data["tls.crt"] != nil && cert.Data["tls.key"] != nil) {
		msg = fmt.Sprintf("Skipping the certificate \"%s\" from the \"%s\" namespace because it has not contain the key \"pem\" or \"tls\"", cert.ObjectMeta.Name, cert.ObjectMeta.Namespace)
		log.Print(2, msg)
		return false
	}

	fileName := config.GetCertificateFileName(cert.ObjectMeta.Name, cert.ObjectMeta.Namespace)

	// write cfg
	f, err := os.Create(fileName)
	if err != nil {
		log.Print(0, err.Error())
		return false
	}

	if cert.Data["pem"] != nil {
		_, err = f.Write(cert.Data["pem"])
		if err != nil {
			log.Print(0, err.Error())
			f.Close()
			return false
		}
	} else { // saving tls format
		_, err = f.Write(cert.Data["tls.key"])
		if err != nil {
			log.Print(0, err.Error())
			f.Close()
			return false
		}
		_, err = f.Write(cert.Data["tls.crt"])
		if err != nil {
			log.Print(0, err.Error())
			f.Close()
			return false
		}
	}

	err = f.Close()
	if err != nil {
		log.Print(0, err.Error())
		return false
	}

	msg = fmt.Sprintf("The \"%s\" certificate was created properly", fileName)
	log.Print(0, msg)

	return true
}

func createCertificate(obj interface{}) {

	cert := obj.(*v1.Secret)

	createCertificateFile(cert)
}

// overwrite the certificate file and reload the daemon
func updateCertificate(_, obj interface{}) {

	cert := obj.(*v1.Secret)

	certFile := createCertificateFile(cert)

	if certFile {
		zproxy.ReloadDaemon()
	}
}

func deleteCertificate(obj interface{}) {

	cert := obj.(*v1.Secret)

	if config.DeleteCertificateFile(cert.ObjectMeta.Name, cert.ObjectMeta.Namespace) {
		zproxy.ReloadDaemon()
	}
}
