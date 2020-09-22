package types

// Struct for global configuration of the client
type Config struct {
	Paths struct {
		GoClientBin  string
		Bin          string
		BinCtl       string
		ConfigFile   string
		SocketFile   string
		Cert         string
		DefaultCert  string
		ErrorFile414 string
		ErrorFile500 string
		ErrorFile501 string
		ErrorFile503 string
	}
	Client struct {
		ControllerName      string
		ConfigMapName       string
		AnnotationPrefix    string
		DaemonsCheckTimeout int
		ClientLogsLevel     int
		ClientStartGraceTme int
	}
	Global struct {
		ProxyLogsLevel int
		DHFile         string
		TotalTO        int
		ConnTO         int
		AliveTO        int
		ClientTO       int
	}
	Listener struct {
		ListenerIP string
		HTTPPort   int
		HTTPSPort  int

		Err414               string
		Err500               string
		Err501               string
		Err503               string
		XHTTP                int
		RewriteLocation      int
		RemoveRequestHeader  []string
		RemoveResponseHeader []string
		AddRequestHeader     []string
		AddResponseHeader    []string

		DefaultCert         string
		Ciphers             string
		DisableSSLProtocol  []string
		SSLHonorCipherOrder int
	}
	Service struct {
		HTTPSBackends           bool
		StrictTransportSecurity int
		CookieName              string
		CookiePath              string
		CookieDomain            string
		CookieTTL               int
		RedirectURL             string
		RedirectCode            int
		RedirectType            string
		SessionType             string
		SessionTTL              int
		SessionID               string
	}
}
