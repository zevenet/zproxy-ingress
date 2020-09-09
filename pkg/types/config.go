package types

// Struct for global configuration of the client
type Config struct {
	Client struct {
		ControllerName          string
		ConfigMapName           string
		AnnotationPrefix        string
		DaemonCheckTimeout      int
		ConfigurationReloadTime int
		ClientLogsLevel         int
		ClientStartGraceTme     int
		ErrorFile414            string
		ErrorFile500            string
		ErrorFile501            string
		ErrorFile503            string
		DefaultCert             string
	}
	Global struct {
		BinPath           string
		BinCtlPath        string
		ConfigFile        string
		SocketFile        string
		LogsLevel         int
		DHFile            string
		ECDHCurve         string
		TotalTO           int
		ConnTO            int
		AliveTO           int
		ClientTO          int
		Ignore100Continue int
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
		Priority                bool
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
