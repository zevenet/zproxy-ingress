package config

import (
	"fmt"
	log "github.com/zevenet/zproxy-ingress/pkg/logs"
	"os"
)

var prePath = "/etc/zproxy/certs"
var fileExt = ".pem"

// use the subdirectory of the certificate to apply secrets namespaces
func GetCertificateFileName(secretName, namespace string) string {

	path := prePath + "/" + namespace

	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, os.ModeDir)
	}

	return path + "/" + secretName + fileExt
}

func DeleteCertificateFile(secretName, namespace string) bool {

	fileName := GetCertificateFileName(secretName, namespace)

	err := os.Remove(fileName)

	if err != nil {
		msg := fmt.Sprintf("Error deleting the file \"%s\": %s\n", fileName, err.Error)
		log.Print(0, msg)
		return false
	}

	return true
}
