package config

import (
	"fmt"
	"gopkg.in/gcfg.v1"
	"os"
	"strconv"
	"strings"

	log "github.com/zevenet/zproxy-ingress/pkg/logs"
	types "github.com/zevenet/zproxy-ingress/pkg/types"
)

var Settings *types.Config // struct with current global cfg

// Set the global configuration from a config file received for arguments
func Init() {
	if len(os.Args) != 2 {
		err := fmt.Sprintf("Error: only the configuration file is expected as argument")
		panic(err)
	}

	var Default types.Config // default cfg loaded when the client is executed
	err := gcfg.ReadFileInto(&Default, os.Args[1])
	if err != nil {
		panic(err)
	}

	if Default.Client.ClientLogsLevel > 0 {
		log.SetLevel(Default.Client.ClientLogsLevel)
		msg := fmt.Sprintf("%+v\n", Default)
		log.Print(1, msg)
	}

	Settings = &Default
}

func writeFile(buff *string, fileName string) int {
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

func removeEmpty(s []string) []string {
	var ret []string
	for _, str := range s {
		if str != "" {
			ret = append(ret, str)
		}
	}
	return ret
}

func ReplaceParams(data map[string]string) {

	// Listener Parameters
	if val, ok := data["listener-ip"]; ok {
		Settings.Listener.ListenerIP = val
	}
	if val, ok := data["listener-http-port"]; ok {
		Settings.Listener.HTTPPort, _ = strconv.Atoi(val)
	}
	if val, ok := data["listener-https-port"]; ok {
		Settings.Listener.HTTPSPort, _ = strconv.Atoi(val)
	}
	if val, ok := data["listener-error-414"]; ok {
		if val == "" {
			Settings.Listener.Err414 = ""
		} else {
			writeFile(&val, Settings.Paths.ErrorFile414)
			Settings.Listener.Err414 = Settings.Paths.ErrorFile414
		}
	}
	if val, ok := data["listener-error-500"]; ok {
		if val == "" {
			Settings.Listener.Err414 = ""
		} else {
			writeFile(&val, Settings.Paths.ErrorFile500)
			Settings.Listener.Err500 = Settings.Paths.ErrorFile500
		}
	}
	if val, ok := data["listener-error-501"]; ok {
		if val == "" {
			Settings.Listener.Err414 = ""
		} else {
			writeFile(&val, Settings.Paths.ErrorFile501)
			Settings.Listener.Err501 = Settings.Paths.ErrorFile501
		}
	}
	if val, ok := data["listener-error-503"]; ok {
		if val == "" {
			Settings.Listener.Err414 = ""
		} else {
			writeFile(&val, Settings.Paths.ErrorFile503)
			Settings.Listener.Err503 = Settings.Paths.ErrorFile503
		}
	}
	if val, ok := data["listener-xhttp"]; ok {
		Settings.Listener.XHTTP, _ = strconv.Atoi(val)
	}
	if val, ok := data["listener-rewrite-location"]; ok {
		Settings.Listener.RewriteLocation, _ = strconv.Atoi(val)
	}
	if val, ok := data["listener-remove-request-header"]; ok {
		Settings.Listener.RemoveRequestHeader = removeEmpty(strings.Split(val, "\n"))
	}
	if val, ok := data["listener-remove-response-header"]; ok {
		Settings.Listener.RemoveResponseHeader = removeEmpty(strings.Split(val, "\n"))
	}
	if val, ok := data["listener-add-request-header"]; ok {
		Settings.Listener.AddRequestHeader = removeEmpty(strings.Split(val, "\n"))
	}
	if val, ok := data["listener-add-response-header"]; ok {
		Settings.Listener.AddResponseHeader = removeEmpty(strings.Split(val, "\n"))
	}
	if val, ok := data["listener-default-cert"]; ok {

		if val == "" { // reset value
			Settings.Listener.DefaultCert = Settings.Paths.DefaultCert
		} else {
			certfile := GetCertificateFileName(val, os.Getenv("POD_NAMESPACE"))
			if _, err := os.Stat(certfile); os.IsNotExist(err) {
				msg := fmt.Sprintf("Error, the default certificate '%s' (%s) was not found\n", val, certfile)
				log.Print(0, msg)
			} else {
				Settings.Listener.DefaultCert = certfile
			}
		}
	}
	if val, ok := data["listener-ciphers"]; ok {
		Settings.Listener.Ciphers = val
	}
	if val, ok := data["listener-disable-ssl-protocol"]; ok {
		Settings.Listener.DisableSSLProtocol = strings.Split(val, "|")
	}
	if val, ok := data["listener-ssl-honor-cipher-order"]; ok {
		Settings.Listener.SSLHonorCipherOrder, _ = strconv.Atoi(val)
	}

	// Service Parameters
	if val, ok := data["service-https-backends"]; ok {
		Settings.Service.HTTPSBackends, _ = strconv.ParseBool(val)
	}
	if val, ok := data["service-strict-transport-security"]; ok {
		Settings.Service.StrictTransportSecurity, _ = strconv.Atoi(val)
	}
	if val, ok := data["service-cookie-name"]; ok {
		Settings.Service.CookieName = val
	}
	if val, ok := data["service-cookie-path"]; ok {
		Settings.Service.CookiePath = val
	}
	if val, ok := data["service-cookie-domain"]; ok {
		Settings.Service.CookieDomain = val
	}
	if val, ok := data["service-cookie-ttl"]; ok {
		Settings.Service.CookieTTL, _ = strconv.Atoi(val)
	}
	// this cannot be set, because it would replace all backends
	//~ if val, ok := data["Service.RedirectURL"]; ok {
	//~ Settings.Service.RedirectURL = val
	//~ }
	if val, ok := data["service-redirect-code"]; ok {
		Settings.Service.RedirectCode, _ = strconv.Atoi(val)
	}
	if val, ok := data["service-redirect-type"]; ok {
		Settings.Service.RedirectType = val
	}
	if val, ok := data["service-session-type"]; ok {
		Settings.Service.SessionType = val
	}
	if val, ok := data["service-session-ttl"]; ok {
		Settings.Service.SessionTTL, _ = strconv.Atoi(val)
	}
	if val, ok := data["service-session-id"]; ok {
		Settings.Service.SessionID = val
	}
}
