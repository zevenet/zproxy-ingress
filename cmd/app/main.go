package main

import (
	auth "github.com/zevenet/zproxy-ingress/pkg/auth"
	config "github.com/zevenet/zproxy-ingress/pkg/config"
	log "github.com/zevenet/zproxy-ingress/pkg/logs"
	watchers "github.com/zevenet/zproxy-ingress/pkg/watchers"
	wait "k8s.io/apimachinery/pkg/util/wait"
	"time"
)

var controllerName string

func main() {

	// Read input configuration
	globalConfig := config.Init()
	controllerName = globalConfig.Client.ControllerName
	log.Print(0, "Read input configuration...")

	clientIngress := auth.GetClienset()
	log.Print(0, "Authentication successful...")

	// Make lists of resources to be observed
	listWatchSecret := watchers.GetSecretListWatch(clientIngress)
	listWatchIngress := watchers.GetIngressListWatch(clientIngress)
	log.Print(0, "Watchers ready...")

	// Make log channel before writing messages
	logChannel := make(chan string)

	// Notify every change into logChannel based on every list watch
	secretController := watchers.GetSecretController(listWatchSecret, logChannel, globalConfig)
	ingressController := watchers.GetIngressController(listWatchIngress, logChannel, globalConfig)
	log.Print(0, "Controllers ready...")

	// Event loop start, run them as background processes
	go secretController.Run(wait.NeverStop)
	time.Sleep(5 * time.Second) // Wait for farms to be created

	go ingressController.Run(wait.NeverStop)

	// Print every message received from the channel
	log.Print(0, "Init finished")
	for message := range logChannel {
		log.Print(0, message)
	}
	// This line is unreachable: working as intended
}
