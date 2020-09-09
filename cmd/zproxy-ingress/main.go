package main

import (
	auth "github.com/zevenet/zproxy-ingress/pkg/auth"
	config "github.com/zevenet/zproxy-ingress/pkg/config"
	controller "github.com/zevenet/zproxy-ingress/pkg/controller"
	log "github.com/zevenet/zproxy-ingress/pkg/logs"
	zproxy "github.com/zevenet/zproxy-ingress/pkg/zproxy"
	wait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"time"
)

func main() {

	// Read input configuration
	config.Init()
	log.Print(0, "Read input configuration...")

	zproxy.LoadConfig()

	clientIngress := auth.GetClienset()
	log.Print(0, "Authentication successful...")

	// Get controllers
	controllers := []cache.Controller{
		controller.GetSecretController(clientIngress),
		controller.GetIngressController(clientIngress),
		controller.GetConfigMapController(clientIngress),
	}

	for _, controller := range controllers {
		go controller.Run(wait.NeverStop)
		time.Sleep(time.Duration(config.Settings.Client.ClientStartGraceTme) * time.Second)
	}

	select {}
	// This line is unreachable: working as intended
}
