package config

import (
	"fmt"
	"gopkg.in/gcfg.v1"
	"os"

	log "github.com/zevenet/zproxy-ingress/pkg/logs"
	types "github.com/zevenet/zproxy-ingress/pkg/types"
)

// Set the global configuration from a config file received for arguments
func Init() *types.Config {
	if len(os.Args) != 2 {
		err := fmt.Sprintf("Error: only the configuration file is expected as argument")
		panic(err)
	}

	var cfg types.Config
	err := gcfg.ReadFileInto(&cfg, os.Args[1])
	if err != nil {
		panic(err)
	}

	if cfg.Client.ClientLogsLevel > 0 {
		log.SetLevel(cfg.Client.ClientLogsLevel)
		msg := fmt.Sprintf("%+v\n", cfg)
		log.Print(1, msg)
	}

	return &cfg
}
