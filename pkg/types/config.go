package types

// Struct for global configuration of the client
type Config struct {
	Global struct {
		ControllerName   string
		LogsLevel        int
		BinPath          string
		BinCtlPath       string
		ConfigFile       string
		SocketFile       string
		DhFile           string
		DefaultIp        string
		DefaultPortHTTP  string
		DefaultPortHTTPS string
		DefaultCert      string
		ClientLogsLevel  int
	}
}
