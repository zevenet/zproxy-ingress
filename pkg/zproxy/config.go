package zproxy

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	config "github.com/zevenet/zproxy-ingress/pkg/config"
	log "github.com/zevenet/zproxy-ingress/pkg/logs"
	types "github.com/zevenet/zproxy-ingress/pkg/types"
	v1beta "k8s.io/api/networking/v1beta1"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
)

// Global variables
var IngressesList []*v1beta.Ingress

type service struct {
	backendList []v1beta.IngressBackend
	path        string
}

// configuration to set up zproxy
var globalCfg *types.Config

func LoadConfig() {
	globalCfg = config.Settings
}

func printConfigError(cfgFile string, errString string) {
	f, err := os.Open(cfgFile)

	if err != nil {
		log.Print(0, "Error reading the confg file: %s", cfgFile)
		return
	}

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	log.Print(0, ">> Syntax error in config file: ")
	log.Print(0, errString)

	msg := ""
	ind := 1
	for scanner.Scan() {
		msg = fmt.Sprintf("> %d:  %s", ind, scanner.Text())
		log.Print(0, msg)
		ind++
	}

	f.Close()
}

func checkProxyConfig() int {

	cmdLine := globalCfg.Global.BinPath + " -f " + globalCfg.Global.ConfigFile + " -c"

	log.Print(1, cmdLine)

	cmd := exec.Command("sh", "-c", cmdLine)

	out, err := cmd.CombinedOutput()

	if err != nil {
		log.Print(0, "Latest reload was: error")
		printConfigError(globalCfg.Global.ConfigFile, string(out))
		return 1
	} else {
		log.Print(0, "Latest reload was: success")
	}

	return 0
}

func ReloadDaemon() int {

	if checkProxyConfig() != 0 {
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

func CreateProxyConfig(ingressesList []*v1beta.Ingress) int {

	var buffFile string //cfg buffer

	// global cfg
	addProxyConfigGlobal(&buffFile)

	// listeners
	addProxyConfigListener(&buffFile, ingressesList)

	// save file
	if writeProxyConfig(&buffFile, globalCfg.Global.ConfigFile) != 0 {
		return 1
	}

	return 0
}

func addProxyConfigGlobal(buff *string) {

	*buff += fmt.Sprintf("Daemon\t%d\n", 0) +
		fmt.Sprintf("LogLevel\t%d\n", globalCfg.Global.LogsLevel) +
		fmt.Sprintf("logfacility\t%c\n", '-') +
		fmt.Sprintf("Timeout\t%d\n", globalCfg.Global.TotalTO) +
		fmt.Sprintf("ConnTO\t%d\n", globalCfg.Global.ConnTO) +
		fmt.Sprintf("Alive\t%d\n", globalCfg.Global.AliveTO) +
		fmt.Sprintf("Client\t%d\n", globalCfg.Global.ClientTO) +
		fmt.Sprintf("Control\t\"%s\"\n", globalCfg.Global.SocketFile) +

		// SSL settings
		fmt.Sprintf("DHParams\t\"%s\"\n", globalCfg.Global.DHFile) +
		fmt.Sprintf("ECDHCurve\t\"%s\"\n", globalCfg.Global.ECDHCurve) +
		fmt.Sprintf("Ignore100Continue\t%d\n", globalCfg.Global.Ignore100Continue) +

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

func addServiceRedirect(buff *string, ingress *v1beta.Ingress, redirectFlag *int) {
	redURL := globalCfg.Service.RedirectURL
	redCode := globalCfg.Service.RedirectCode
	redType := globalCfg.Service.RedirectType

	if ingress.Annotations[globalCfg.Client.AnnotationPrefix+"/service-redirect-url"] != "" {
		redURL = ingress.Annotations[globalCfg.Client.AnnotationPrefix+"/service-redirect-url"]
	}
	if ingress.Annotations[globalCfg.Client.AnnotationPrefix+"/service-redirect-code"] != "" {
		redCode, _ = strconv.Atoi(ingress.Annotations[globalCfg.Client.AnnotationPrefix+"/service-redirect-code"])
	}
	if ingress.Annotations[globalCfg.Client.AnnotationPrefix+"/service-redirect-type"] != "" {
		redType = ingress.Annotations[globalCfg.Client.AnnotationPrefix+"/service-redirect-type"]
	}

	if redURL != "" {
		// [Redirect | RedirectAppend | RedirectDynamic] [code] "url"
		if redType == "append" {
			redType = "RedirectAppend"
		} else {
			redType = "Redirect"
		}
		*redirectFlag = 1
		*buff += fmt.Sprintf("\t\t%s %d \"%s\"\n", redType, redCode, redURL)
	}
}

func addServiceSession(buff *string, ingress *v1beta.Ingress) {
	sessionType := globalCfg.Service.SessionType
	sessionTTL := globalCfg.Service.SessionTTL
	sessionID := globalCfg.Service.SessionID

	if ingress.Annotations[globalCfg.Client.AnnotationPrefix+"/service-session-type"] != "" {
		sessionType = ingress.Annotations[globalCfg.Client.AnnotationPrefix+"/service-session-type"]
	}
	if ingress.Annotations[globalCfg.Client.AnnotationPrefix+"/service-session-ttl"] != "" {
		sessionTTL, _ = strconv.Atoi(ingress.Annotations[globalCfg.Client.AnnotationPrefix+"/service-session-ttl"])
	}
	if ingress.Annotations[globalCfg.Client.AnnotationPrefix+"/service-session-id"] != "" {
		sessionID = ingress.Annotations[globalCfg.Client.AnnotationPrefix+"/service-session-id"]
	}

	if sessionType != "" {
		idStr := ""
		/*
		*	Session
		*	  Type    URL
		*	  TTL     120
		*	  ID      "sessid"
		*	End
		 */
		if sessionType == "PARAM" || sessionType == "URL" || sessionType == "COOKIE" {
			idStr = fmt.Sprintf("\t\t\tID\t\"%s\"\n", sessionID)
		}
		*buff += fmt.Sprintf("\t\tSession\n\t\t\tType\t%s\n\t\t\tTTL\t%d\n%s\t\tEnd\n",
			sessionType, sessionTTL, idStr)
	}
}

func addServiceCookie(buff *string, ingress *v1beta.Ingress) {

	cookieName := globalCfg.Service.CookieName
	cookieTTL := globalCfg.Service.CookieTTL
	cookiePath := globalCfg.Service.CookiePath
	cookieDomain := globalCfg.Service.CookieDomain

	if ingress.Annotations[globalCfg.Client.AnnotationPrefix+"/service-cookie-name"] != "" {
		cookieName = ingress.Annotations[globalCfg.Client.AnnotationPrefix+"/service-cookie-name"]
	}
	if ingress.Annotations[globalCfg.Client.AnnotationPrefix+"/service-cookie-ttl"] != "" {
		cookieTTL, _ = strconv.Atoi(ingress.Annotations[globalCfg.Client.AnnotationPrefix+"/service-cookie-ttl"])
	}
	if ingress.Annotations[globalCfg.Client.AnnotationPrefix+"/service-cookie-path"] != "" {
		cookiePath = ingress.Annotations[globalCfg.Client.AnnotationPrefix+"/service-cookie-path"]
	}
	if ingress.Annotations[globalCfg.Client.AnnotationPrefix+"/service-cookie-domain"] != "" {
		cookieDomain = ingress.Annotations[globalCfg.Client.AnnotationPrefix+"/service-cookie-domain"]
	}

	if cookieName != "" {
		// BackendCookie "ZENSESSIONIDw" "domainname.com" "/" 233
		*buff += fmt.Sprintf("\t\tBackendCookie \"%s\" \"%s\" \"%s\" %d\n", cookieName, cookieDomain, cookiePath, cookieTTL)
	}
}

func addServiceTransportSecurity(buff *string, ingress *v1beta.Ingress) {

	strictTransport := globalCfg.Service.StrictTransportSecurity

	if ingress.Spec.TLS == nil {
		return
	}

	if ingress.Annotations[globalCfg.Client.AnnotationPrefix+"/service-strict-transport-security-ttl"] != "" {
		strictTransport, _ = strconv.Atoi(ingress.Annotations[globalCfg.Client.AnnotationPrefix+"/service-strict-transport-security-ttl"])
	}

	if strictTransport != 0 {
		// StrictTransportSecurity 21600000
		*buff += fmt.Sprintf("\t\tStrictTransportSecurity %d\n", strictTransport)
	}
}

//
func getServiceBackendHTTPS(ingress *v1beta.Ingress) string {

	flag := globalCfg.Service.HTTPSBackends

	if ingress.Annotations[globalCfg.Client.AnnotationPrefix+"/backend-https"] != "" {
		flag, _ = strconv.ParseBool(ingress.Annotations[globalCfg.Client.AnnotationPrefix+"/backend-https"])
	}

	backendHttps := ""
	if flag {
		backendHttps = "\t\t\tHTTPS\n"
	}

	return backendHttps
}

// This function is used to configure:
//	* the HTTP/HTTPS services
// 	* the service to forward HTTPS request to HTTP service (in HTTPS listener)
// 	* the service with the default backend (in HTTP listener)
// The function sets the svc name as svc-default when the svcId parameter is set to 0
func genProxyConfigService(buff *string, svcId int, host string, path string, ingressObj *v1beta.Ingress, backendList *[]v1beta.IngressBackend, namespace string, ssl bool) {

	redirectFlag := 0
	backendHttps := ""

	if svcId == 0 {
		*buff += fmt.Sprintf("\tService \"svc%s\"\n", "-default")
	} else {
		*buff += fmt.Sprintf("\tService \"svc%d\"\n", svcId)
	}

	// Create matching rules
	if host != "" {
		*buff += fmt.Sprintf("\t\tHeadRequire \"Host: %s\"\n", host)
	}
	if path != "" {
		*buff += fmt.Sprintf("\t\tUrl \"%s\"\n", path)
	}

	// customized params
	if globalCfg != nil && ingressObj != nil {

		addServiceRedirect(buff, ingressObj, &redirectFlag)

		addServiceSession(buff, ingressObj)

		addServiceCookie(buff, ingressObj)

		addServiceTransportSecurity(buff, ingressObj)

		// backends
		backendHttps = getServiceBackendHTTPS(ingressObj)
	}

	if len(*backendList) > 0 && redirectFlag == 0 {
		for _, bck := range *backendList {
			backendName := bck.ServiceName
			if namespace != "" {
				backendName += "."
				backendName += namespace
			}

			*buff += fmt.Sprintf("\t\tBackEnd\n%s\t\t\tAddress %s\n\t\t\tPort %d\n\t\tEnd\n",
				backendHttps, backendName, bck.ServicePort.IntVal)
		}
	}

	*buff += fmt.Sprintf("\tEnd\n\n")
}

// Set up the parameters for HTTP(S) traffic
func addProxyConfigListenerParamsPlain(buff *string) {

	if globalCfg.Listener.Err414 != "" {
		*buff += fmt.Sprintf("\tErr414\t\"%s\"\n", globalCfg.Listener.Err414)
	}
	if globalCfg.Listener.Err500 != "" {
		*buff += fmt.Sprintf("\tErr500\t\"%s\"\n", globalCfg.Listener.Err500)
	}
	if globalCfg.Listener.Err501 != "" {
		*buff += fmt.Sprintf("\tErr501\t\"%s\"\n", globalCfg.Listener.Err501)
	}
	if globalCfg.Listener.Err503 != "" {
		*buff += fmt.Sprintf("\tErr503\t\"%s\"\n", globalCfg.Listener.Err503)
	}
	*buff += fmt.Sprintf("\txHTTP\t%d\n", globalCfg.Listener.XHTTP) +
		fmt.Sprintf("\tRewriteLocation\t%d\n", globalCfg.Listener.RewriteLocation)

	for _, directive := range globalCfg.Listener.RemoveRequestHeader {
		*buff += fmt.Sprintf("\tHeadRemove\t\"%s\"\n", directive)
	}
	for _, directive := range globalCfg.Listener.AddRequestHeader {
		*buff += fmt.Sprintf("\tAddHeader\t\"%s\"\n", directive)
	}
	for _, directive := range globalCfg.Listener.RemoveResponseHeader {
		*buff += fmt.Sprintf("\tRemoveResponseHeader\t\"%s\"\n", directive)
	}
	for _, directive := range globalCfg.Listener.AddResponseHeader {
		*buff += fmt.Sprintf("\tAddResponseHeader\t\"%s\"\n", directive)
	}
}

// Set up the parameters for SSL listeners
// this function fills the sslIndex slice
func addProxyConfigListenerParamsSSL(buff *string, ingressList []*v1beta.Ingress, sslIndex *[]int) {
	certList := make(map[string]int)

	addProxyConfigListenerParamsPlain(buff)

	for ind, ing := range ingressList {
		if ing.Spec.TLS != nil {
			addProxyCerts(buff, &ing.Spec.TLS, ing.ObjectMeta.Namespace, certList)
			*sslIndex = append(*sslIndex, ind)
		}
	}

	// add default SSL certificate
	*buff += fmt.Sprintf("\tCert\t\"%s\"\n", globalCfg.Listener.DefaultCert) +
		fmt.Sprintf("\tCiphers\t\"%s\"\n", globalCfg.Listener.Ciphers)

	for _, proto := range globalCfg.Listener.DisableSSLProtocol {
		*buff += fmt.Sprintf("\tDisable %s\n", proto)
	}
	*buff += fmt.Sprintf("\tSSLHonorCipherOrder\t%d\n\n", globalCfg.Listener.SSLHonorCipherOrder)
}

// This backend is to forward the request to HTTP listener, for HTTPS request that are defined without TLS configuration
func addRedirectToHTTPService(buff *string) {
	var localBackend *v1beta.IngressBackend = new(v1beta.IngressBackend)
	localBackend.ServiceName = "127.0.0.1"
	localBackend.ServicePort = intstr.FromInt(globalCfg.Listener.HTTPPort)
	bckList := []v1beta.IngressBackend{*localBackend}

	genProxyConfigService(buff, 0, "", "", nil, &bckList, "", false)
}

// Add two listener, one HTTP and another HTTPS, the configuration depend on the tls configuration
func addProxyConfigListener(buff *string, ingressList []*v1beta.Ingress) {

	var sslIndex []int
	svcId := 1

	// https listener
	*buff += fmt.Sprintf("ListenHTTPS\n") +
		fmt.Sprintf("\tAddress\t%s\n", globalCfg.Listener.ListenerIP) +
		fmt.Sprintf("\tPort\t%d\n", globalCfg.Listener.HTTPSPort)

	addProxyConfigListenerParamsSSL(buff, ingressList, &sslIndex)

	// add ssl svc to HTTPS listener
	for _, ind := range sslIndex {
		addProxyConfigServices(buff, ingressList[ind], true, &svcId)
	}

	// create default bck with local HTTP svc
	addRedirectToHTTPService(buff)

	// http listener
	*buff += "End\n\nListenHTTP\n" +
		fmt.Sprintf("\tAddress\t%s\n", globalCfg.Listener.ListenerIP) +
		fmt.Sprintf("\tPort\t%d\n", globalCfg.Listener.HTTPPort)

	addProxyConfigListenerParamsPlain(buff)

	for _, ingObj := range ingressList {
		// add service to listener HTTP without
		if ingObj.Spec.TLS == nil {
			addProxyConfigServices(buff, ingObj, false, &svcId)
			// TODO: implement redirect to https svc for HTTP requests if ingress has TLS cfg.
			// Now, the respose is not service available.
			//~ else if ingObj.Annotation.redirectToSSL != nil {
			//~ addProxyConfigServices(buff, ingObj, true, &svcId)
		}
	}

	// set default backend. Set only one, the first one
	for _, ingressObj := range ingressList {
		if ingressObj.Spec.Backend != nil {
			bckList := []v1beta.IngressBackend{*ingressObj.Spec.Backend}
			genProxyConfigService(buff, 0, "", "", ingressObj, &bckList, ingressObj.ObjectMeta.Namespace, false)
			break
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
				genProxyConfigService(buff, *svcId, host, svc.path, ingressObj, &svc.backendList, ingressObj.ObjectMeta.Namespace, ssl)
				*svcId += 1
			}
		}
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

		fileName = config.GetCertificateFileName(tlsInfo.SecretName, namespace)
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

func UpdateIngressCfg() bool {

	start := time.Now()

	LoadConfig()

	if CreateProxyConfig(IngressesList) != 0 {
		log.Print(0, "Error creating the config file")
		return false
	}

	if ReloadDaemon() != 0 {
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
