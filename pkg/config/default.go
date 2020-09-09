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
	if val, ok := data["Listener.ListenerIP"]; ok {
		Settings.Listener.ListenerIP = val
	}
	if val, ok := data["Listener.HTTPPort"]; ok {
		Settings.Listener.HTTPPort, _ = strconv.Atoi(val)
	}
	if val, ok := data["Listener.HTTPSPort"]; ok {
		Settings.Listener.HTTPSPort, _ = strconv.Atoi(val)
	}
	if val, ok := data["Listener.Err414"]; ok {
		if val == "" {
			Settings.Listener.Err414 = ""
		} else {
			writeFile(&val, Settings.Client.ErrorFile414)
			Settings.Listener.Err414 = Settings.Client.ErrorFile414
		}
	}
	if val, ok := data["Listener.Err500"]; ok {
		if val == "" {
			Settings.Listener.Err414 = ""
		} else {
			writeFile(&val, Settings.Client.ErrorFile500)
			Settings.Listener.Err500 = Settings.Client.ErrorFile500
		}
	}
	if val, ok := data["Listener.Err501"]; ok {
		if val == "" {
			Settings.Listener.Err414 = ""
		} else {
			writeFile(&val, Settings.Client.ErrorFile501)
			Settings.Listener.Err501 = Settings.Client.ErrorFile501
		}
	}
	if val, ok := data["Listener.Err503"]; ok {
		if val == "" {
			Settings.Listener.Err414 = ""
		} else {
			writeFile(&val, Settings.Client.ErrorFile503)
			Settings.Listener.Err503 = Settings.Client.ErrorFile503
		}
	}
	if val, ok := data["Listener.XHTTP"]; ok {
		Settings.Listener.XHTTP, _ = strconv.Atoi(val)
	}
	if val, ok := data["Listener.RewriteLocation"]; ok {
		Settings.Listener.RewriteLocation, _ = strconv.Atoi(val)
	}
	if val, ok := data["Listener.RemoveRequestHeader"]; ok {
		Settings.Listener.RemoveRequestHeader = removeEmpty(strings.Split(val, "\n"))
	}
	if val, ok := data["Listener.RemoveResponseHeader"]; ok {
		Settings.Listener.RemoveResponseHeader = removeEmpty(strings.Split(val, "\n"))
	}
	if val, ok := data["Listener.AddRequestHeader"]; ok {
		Settings.Listener.AddRequestHeader = removeEmpty(strings.Split(val, "\n"))
	}
	if val, ok := data["Listener.AddResponseHeader"]; ok {
		Settings.Listener.AddResponseHeader = removeEmpty(strings.Split(val, "\n"))
	}
	if val, ok := data["Listener.DefaultCert"]; ok {

		if val == "" { // reset value
			Settings.Listener.DefaultCert = Settings.Client.DefaultCert
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
	if val, ok := data["Listener.Ciphers"]; ok {
		Settings.Listener.Ciphers = val
	}
	if val, ok := data["Listener.DisableSSLProtocol"]; ok {
		Settings.Listener.DisableSSLProtocol = strings.Split(val, "|")
	}
	if val, ok := data["Listener.SSLHonorCipherOrder"]; ok {
		Settings.Listener.SSLHonorCipherOrder, _ = strconv.Atoi(val)
	}

	// Service Parameters
	if val, ok := data["Service.HTTPSBackends"]; ok {
		Settings.Service.HTTPSBackends, _ = strconv.ParseBool(val)
	}
	if val, ok := data["Service.StrictTransportSecurity"]; ok {
		Settings.Service.StrictTransportSecurity, _ = strconv.Atoi(val)
	}
	if val, ok := data["Service.Priority"]; ok {
		Settings.Service.Priority, _ = strconv.ParseBool(val)
	}
	if val, ok := data["Service.CookieName"]; ok {
		Settings.Service.CookieName = val
	}
	if val, ok := data["Service.CookiePath"]; ok {
		Settings.Service.CookiePath = val
	}
	if val, ok := data["Service.CookieDomain"]; ok {
		Settings.Service.CookieDomain = val
	}
	if val, ok := data["Service.CookieTTL"]; ok {
		Settings.Service.CookieTTL, _ = strconv.Atoi(val)
	}
	// this cannot be set, because it would replace all backends
	//~ if val, ok := data["Service.RedirectURL"]; ok {
	//~ Settings.Service.RedirectURL = val
	//~ }
	if val, ok := data["Service.RedirectCode"]; ok {
		Settings.Service.RedirectCode, _ = strconv.Atoi(val)
	}
	if val, ok := data["Service.RedirectType"]; ok {
		Settings.Service.RedirectType = val
	}
	if val, ok := data["Service.SessionType"]; ok {
		Settings.Service.SessionType = val
	}
	if val, ok := data["Service.SessionTTL"]; ok {
		Settings.Service.SessionTTL, _ = strconv.Atoi(val)
	}
	if val, ok := data["Service.SessionID"]; ok {
		Settings.Service.SessionID = val
	}
}
