package config

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	log "github.com/zevenet/zproxy-ingress/pkg/logs"
	types "github.com/zevenet/zproxy-ingress/pkg/types"
	v1beta "k8s.io/api/networking/v1beta1"

	intstr "k8s.io/apimachinery/pkg/util/intstr"
)

type service struct {
	backendList []v1beta.IngressBackend
	path        string
}

func checkProxyConfig(globalCfg *types.Config) int {

	cmdLine := globalCfg.Global.BinPath + " -f " + globalCfg.Global.ConfigFile + " -c"

	log.Print(1, cmdLine)

	cmd := exec.Command("sh", "-c", cmdLine)

	out, err := cmd.CombinedOutput()

	if err != nil {
		log.Print(0, "Syntax error in config file")
		log.Print(0, string(out))
		return 1
	}

	return 0
}

func ReloadDaemon(globalCfg *types.Config) int {

	if checkProxyConfig(globalCfg) != 0 {
		return 1
	}

	cmdLine := globalCfg.Global.BinCtlPath + " -c " + globalCfg.Global.SocketFile + " -R 0"

	log.Print(1, cmdLine)

	output, err := exec.Command("sh", "-c", cmdLine).CombinedOutput()

	if err != nil {
		log.Print(0, err.Error())
		log.Print(0, string(output))
		return 1
	}

	return 0
}

func printUpdated(object string, json string, response string) {
	message := fmt.Sprintf("\nUpdated %s:\n%s\n%s", object, json, response)
	log.Print(0, message)
}

func CreateProxyConfig(ingressesList []*v1beta.Ingress, globalCfg *types.Config) int {

	var buffFile string //cfg buffer

	// global cfg
	addProxyConfigGlobal(&buffFile, globalCfg)

	// listeners
	addProxyConfigListener(&buffFile, ingressesList, globalCfg)

	// save file
	if writeProxyConfig(&buffFile, globalCfg.Global.ConfigFile) != 0 {
		return 1
	}

	return 0
}

func addProxyConfigGlobal(buff *string, globalCfg *types.Config) {

	*buff = fmt.Sprintf("Daemon\t%d\n", 0) +
		fmt.Sprintf("LogLevel\t%d\n", globalCfg.Global.LogsLevel) +
		fmt.Sprintf("logfacility\t%c\n", '-') +
		fmt.Sprintf("Timeout\t%d\n", 45) +
		fmt.Sprintf("ConnTO\t%d\n", 20) +
		fmt.Sprintf("Alive\t%d\n", 10) +
		fmt.Sprintf("Client\t%d\n", 30) +
		fmt.Sprintf("Control\t\"%s\"\n", globalCfg.Global.SocketFile) +

		// SSL settings
		fmt.Sprintf("DHParams\t\"%s\"\n", globalCfg.Global.DhFile) +
		fmt.Sprintf("ECDHCurve\t\"%s\"\n", "prime256v1") +

		fmt.Sprintf("\n")
}

// get index of ingresses services with SSL settings
func getSslServicesIndex(ingressesList []*v1beta.Ingress, ssl *[]int, nossl *[]int) {
	for it, ingressObj := range ingressesList {
		if ingressObj.Spec.TLS != nil {
			*ssl = append(*ssl, it)
		} else {
			*nossl = append(*nossl, it)
		}
	}
}

// the function set the svc name as svc-default when the svcId parameter is set to 0
func genProxyConfigService(buff *string, svcId int, host string, path string, backendList *[]v1beta.IngressBackend, namespace string, ssl bool) {

	// Creating svc
	if svcId == 0 {
		*buff += fmt.Sprintf("\tService \"svc%s\"\n", "-default")
	} else {
		*buff += fmt.Sprintf("\tService \"svc%d\"\n", svcId)
	}

	if host != "" {
		*buff += fmt.Sprintf("\t\tHeadRequire \"Host: %s\"\n", host)
	}
	if path != "" {
		*buff += fmt.Sprintf("\t\tUrl \"%s\"\n", path)
	}

	if len(*backendList) > 0 {
		for _, bck := range *backendList {
			backendName := bck.ServiceName
			if namespace != "" {
				backendName += "."
				backendName += namespace
			}

			*buff += fmt.Sprintf("\t\tBackEnd\n\t\t\tAddress %s\n\t\t\tPort %d\n\t\tEnd\n",
				backendName, bck.ServicePort.IntVal)
		}
	}

	*buff += fmt.Sprintf("\tEnd\n\n")
}

// Add two listener, one HTTP and another HTTPS, the configuration depend on the tls configuration
func addProxyConfigListener(buff *string, ingressList []*v1beta.Ingress, globalCfg *types.Config) {

	var sslIndex []int
	var nosslIndex []int
	var bckList []v1beta.IngressBackend
	svcId := 1
	certList := make(map[string]int)

	// https listener
	*buff += fmt.Sprintf("ListenHTTPS\n") +
		fmt.Sprintf("\tAddress\t%s\n", globalCfg.Global.DefaultIp) +
		fmt.Sprintf("\tPort\t%s\n", globalCfg.Global.DefaultPortHTTPS)

	getSslServicesIndex(ingressList, &sslIndex, &nosslIndex) // manages tls
	for _, ind := range sslIndex {
		addProxyCerts(buff, &ingressList[ind].Spec.TLS, ingressList[ind].ObjectMeta.Namespace, certList)
	}
	// add default SSL certificate
	*buff += fmt.Sprintf("\tCert\t\"%s\"\n", globalCfg.Global.DefaultCert) +
		fmt.Sprintf("\tCiphers\t\"%s\"\n", "ALL") +
		fmt.Sprintf("\tDisable SSLv3\n") +
		fmt.Sprintf("\tSSLHonorCipherOrder\t%d\n\n", 1)

	// add ssl svc to HTTPS listener
	for _, ind := range sslIndex {
		addProxyConfigServices(buff, ingressList[ind], true, &svcId)
	}

	// create default bck with local HTTP svc
	var localBackend *v1beta.IngressBackend = new(v1beta.IngressBackend)
	localBackend.ServiceName = "127.0.0.1"
	port, _ := strconv.Atoi(globalCfg.Global.DefaultPortHTTP)
	localBackend.ServicePort = intstr.FromInt(port)
	bckList = append(bckList, *localBackend)
	genProxyConfigService(buff, 0, "", "", &bckList, "", false)

	// http listener
	*buff += "End\n\n" +
		fmt.Sprintf("ListenHTTP\n") +
		fmt.Sprintf("\tAddress\t%s\n", globalCfg.Global.DefaultIp) +
		fmt.Sprintf("\tPort\t%s\n", globalCfg.Global.DefaultPortHTTP) +
		fmt.Sprintf("\txHTTP\t%d\n", 4) +
		fmt.Sprintf("\tRewriteLocation\t%d\n\n", 1)

	for _, ingObj := range ingressList {

		// add service to listener HTTP without
		if ingObj.Spec.TLS == nil {
			addProxyConfigServices(buff, ingObj, false, &svcId)
			// TODO: implement redirect to https svc for HTTP requests if ingress has TLS cfg.
			// Now, the respose is not service available.
			//~ else if ingObj.MetaObj.Annotation.redirectToSSL != nil {
			//~ addProxyConfigServices(buff, ingObj, true, &svcId)
		}
	}

	*buff += "End\n"
}

func createPath(pathStr string, pathType v1beta.PathType) string {
	// Create path
	path := ""
	if pathStr != "" {
		// Adding path type
		if pathType != "" {
			if pathType == "Exact" {
				path = "^" + pathStr + "$"
			} else if pathType == "Prefix" {
				path = "^" + pathStr + "(?:/|$)"
			} else {
				// the request is accepted if the pattern matches in any part of the URI
				path = pathStr
			}
		}
	}
	return path
}

func createService(svcList *[]service, path string, backend *v1beta.IngressBackend) {
	var svc service
	svc.path = path
	svc.backendList = append(svc.backendList, *backend)
	*svcList = append(*svcList, svc)
}

func addBackendToService(svcList *[]service, path string, backendOri *v1beta.IngressBackend) {
	found := 0
	backend := backendOri.DeepCopy()

	if len(*svcList) == 0 {
		createService(svcList, path, backend)
	} else {
		for it, svcIt := range *svcList {
			if path == svcIt.path {
				(*svcList)[it].backendList = append(svcIt.backendList, *backend)
				found = 1
				break
			}
		}
		// New entry
		if found == 0 {
			createService(svcList, path, backend)
		}
	}

}

func addProxyConfigServices(buff *string, ingressObj *v1beta.Ingress, ssl bool, svcId *int) int {

	var svcList []service
	var host string

	for _, rule := range ingressObj.Spec.Rules {
		svcList = nil

		// Create host header
		host = ""
		if rule.Host != "" {
			host = rule.Host
		}

		if rule.IngressRuleValue.HTTP != nil {
			// creating svc structs
			for _, http := range rule.IngressRuleValue.HTTP.Paths {
				path := createPath(http.Path, *http.PathType)
				addBackendToService(&svcList, path, &http.Backend)
			}

			// print
			for _, svc := range svcList {
				genProxyConfigService(buff, *svcId, host, svc.path, &svc.backendList, ingressObj.ObjectMeta.Namespace, ssl)
				*svcId += 1
			}
		}
	}

	// manage the default service
	if ingressObj.Spec.Backend != nil {
		var bckList []v1beta.IngressBackend
		bckList = append(bckList, *ingressObj.Spec.Backend)
		genProxyConfigService(buff, 0, "", "", &bckList, ingressObj.ObjectMeta.Namespace, ssl)
	}

	return 0
}

func writeProxyConfig(buff *string, fileName string) int {
	f, err := os.Create(fileName)

	if err != nil {
		fmt.Print(0, err)
		return 1
	}

	_, err = f.WriteString(*buff)
	if err != nil {
		fmt.Print(0, err)
		f.Close()
		return 1
	}

	err = f.Close()
	if err != nil {
		log.Print(0, err.Error())
		return 1
	}
	return 0
}

// add the certficate directive to the listener. It does not read the secret.
// the secret event update the certificate file
func addProxyCerts(buff *string, tlsList *[]v1beta.IngressTLS, namespace string, certList map[string]int) {

	fileName := ""
	msg := ""
	for _, tlsInfo := range *tlsList {

		// {"hosts":["sslexample.foo.com","sslexample2.foo.com"],"secretName":"testsecret-tls"},
		// The "hosts" field is not managed

		fileName = GetCertificateFileName(tlsInfo.SecretName, namespace)
		if certList[fileName] != 1 {
			// Add to the list of uploaded certificates
			certList[fileName] = 1

			_, err := os.Stat(fileName)
			if os.IsNotExist(err) {
				msg := fmt.Sprintf("The certificate \"%s\" was not found", fileName)
				log.Print(0, msg)
			} else {
				// Add directive to HTTPS listener
				*buff += fmt.Sprintf("\tCert\t\"%s\"\n", fileName)

				msg = fmt.Sprintf("Adding the \"%s\" certificate", tlsInfo.SecretName)
				log.Print(1, msg)
			}
		}
	}
}
