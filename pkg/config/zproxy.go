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

func CreateProxyConfig(fileName string, ingressObj *v1beta.Ingress, globalCfg *types.Config) int {

	var buffFile string //cfg buffer

	// manages tls
	tls := false
	if ingressObj.Spec.TLS != nil {
		tls = true
	}

	msg := fmt.Sprintf("Ingress obj: %+v", ingressObj)
	log.Print(1, msg)

	// global cfg
	addProxyConfigGlobal(&buffFile, ingressObj, globalCfg, tls)

	// listeners
	addProxyConfigListener(&buffFile, ingressObj, globalCfg, tls)

	// save file
	if writeProxyConfig(&buffFile, fileName) != 0 {
		return 1
	}

	return 0
}

func addProxyConfigGlobal(buff *string, ingressObj *v1beta.Ingress, globalCfg *types.Config, ssl bool) {

	*buff = fmt.Sprintf("LogLevel\t%d\n", globalCfg.Global.LogsLevel) +
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

// Add two listener, one HTTP and another HTTPS, the configuration depend on the tls configuration
func addProxyConfigListener(buff *string, ingressObj *v1beta.Ingress, globalCfg *types.Config, ssl bool) {

	// https listener
	*buff += fmt.Sprintf("ListenHTTPS\n") +
		fmt.Sprintf("\tAddress\t%s\n", globalCfg.Global.DefaultIp) +
		fmt.Sprintf("\tPort\t%s\n", globalCfg.Global.DefaultPortHTTPS)

	if ssl {
		addProxyCerts(buff, &ingressObj.Spec.TLS, ingressObj.ObjectMeta.Namespace)
	}
	// add default SSL certificate
	*buff += fmt.Sprintf("\tCert\t\"%s\"\n", globalCfg.Global.DefaultCert) +
		fmt.Sprintf("\tCiphers\t\"%s\"\n", "ALL") +
		fmt.Sprintf("\tDisable SSLv3\n") +
		fmt.Sprintf("\tSSLHonorCipherOrder\t%d\n\n", 1)
	if ssl {
		addProxyConfigServices(buff, ingressObj, ssl)
	} else {
		// create default bck with local HTTP svc
		var localBackend *v1beta.IngressBackend = new(v1beta.IngressBackend)
		localBackend.ServiceName = globalCfg.Global.DefaultIp
		port, _ := strconv.Atoi(globalCfg.Global.DefaultPortHTTP)
		localBackend.ServicePort = intstr.FromInt(port)

		// the local HTTP service will manage the request
		genProxyConfigService(buff, 0, "", "", localBackend, "", false)
	}

	// http listener
	*buff += "End\n\n" +
		fmt.Sprintf("ListenHTTP\n") +
		fmt.Sprintf("\tAddress\t%s\n", globalCfg.Global.DefaultIp) +
		fmt.Sprintf("\tPort\t%s\n", globalCfg.Global.DefaultPortHTTP) +
		fmt.Sprintf("\txHTTP\t%d\n", 4) +
		fmt.Sprintf("\tRewriteLocation\t%d\n\n", 1)

	if !ssl {
		addProxyConfigServices(buff, ingressObj, ssl)
		// } else {
		// TODO: implement redirect to https svc
	}
	*buff += "End\n"
}

func addProxyConfigServices(buff *string, ingressObj *v1beta.Ingress, ssl bool) int {

	host := ""
	path := ""
	svcId := 1

	for _, rule := range ingressObj.Spec.Rules {

		// Create host header
		host = ""
		if rule.Host != "" {
			host = rule.Host
		}

		if rule.IngressRuleValue.HTTP != nil {
			for _, http := range rule.IngressRuleValue.HTTP.Paths {

				// Create path
				path = ""
				if http.Path != "" {
					pathType := *http.PathType
					// Adding path type
					if pathType != "" {
						if pathType == "Exact" {
							path = "^" + http.Path + "$"
						} else if pathType == "Prefix" {
							path = "^" + http.Path
							// else implementations, the request is accepted if the pattern matches in any part of the URI
						} else {
							path = http.Path
						}
					}
				}

				backend := &http.Backend
				genProxyConfigService(buff, svcId, host, path, backend, ingressObj.ObjectMeta.Namespace, ssl)
				svcId += 1
			}
		}
	}

	// manage the default service
	if ingressObj.Spec.Backend != nil {
		genProxyConfigService(buff, 0, "", "", ingressObj.Spec.Backend, ingressObj.ObjectMeta.Namespace, ssl)
	}

	return 0
}

// the function set the svc name as svc-default when the svcId parameter is set to 0
func genProxyConfigService(buff *string, svcId int, host, path string, backend *v1beta.IngressBackend, namespace string, ssl bool) {

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

	if namespace != "" {
		backend.ServiceName += "."
		backend.ServiceName += namespace
	}

	if backend != nil {
		*buff += fmt.Sprintf("\t\tBackEnd\n\t\t\tAddress %s\n\t\t\tPort %d\n\t\tEnd\n",
			backend.ServiceName, backend.ServicePort.IntVal)
	}

	*buff += fmt.Sprintf("\tEnd\n\n")
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
func addProxyCerts(buff *string, tlsList *[]v1beta.IngressTLS, namespace string) {

	fileName := ""
	msg := ""
	for _, tlsInfo := range *tlsList {

		// {"hosts":["sslexample.foo.com","sslexample2.foo.com"],"secretName":"testsecret-tls"},
		// The "hosts" field is not managed

		msg = fmt.Sprintf("Adding the \"%s\" certificate", tlsInfo.SecretName)
		log.Print(1, msg)

		fileName = GetCertificateFileName(tlsInfo.SecretName, namespace)

		_, err := os.Stat(fileName)
		if os.IsNotExist(err) {
			msg := fmt.Sprintf("The certificate \"%s\" was not found", fileName)
			log.Print(0, msg)
		} else {
			// Add directive to HTTPS listener
			*buff += fmt.Sprintf("\tCert\t\"%s\"\n", fileName)
		}
	}
}
