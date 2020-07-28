package funcs

import (
	"fmt"
	"os"

	config "github.com/zevenet/zproxy-ingress/pkg/config"
	log "github.com/zevenet/zproxy-ingress/pkg/logs"
	types "github.com/zevenet/zproxy-ingress/pkg/types"
	v1 "k8s.io/api/core/v1"
)

// overwrite the certificate file
// if the certificate is not valid, does not overwrite the old one

func CreateCertificateFile(cert *v1.Secret) bool {

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

	msg = fmt.Sprintf("The \"%s\" certificate was created properly\n", fileName)
	log.Print(0, msg)

	return true
}

// overwrite the certificate file and reload the daemon
func UpdateCertificate(cert *v1.Secret, globalCfg *types.Config) bool {

	certFile := CreateCertificateFile(cert)

	if !certFile {
		return false
	}

	if config.ReloadDaemon(globalCfg) != 0 {
		return false
	}

	return true
}

func DeleteCertificate(cert *v1.Secret, globalCfg *types.Config) bool {

	if !config.DeleteCertificateFile(cert.ObjectMeta.Name, cert.ObjectMeta.Namespace) {
		return false
	}

	if config.ReloadDaemon(globalCfg) != 0 {
		return false
	}

	return true
}
