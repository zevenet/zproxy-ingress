package config

import (
	"fmt"
	"gopkg.in/gcfg.v1"
	"os"

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
