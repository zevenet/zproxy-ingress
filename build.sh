#!/bin/sh

set -e

DOCKER_PATH="./docker"

# Before running this script, you MUST BE ROOT and you need to have the following tools installed:
#   - Docker
#   - Docker-machine
#   - Minikube (Kubernetes local cluster, virtualized)
#   - Golang
#   - client-go libs

# STEP 1:
#   Compile cmd/app/main.go.
#   The binary will be called "app".
GOOS=linux go build -o $DOCKER_PATH/app ./cmd/zproxy-ingress

# STEP 2:
#   The client container will be created using its Dockerfile.
#   It will be made for Docker, not for Minikube (this will come later).
docker build -t zproxy-ingress $DOCKER_PATH

# STEP 3:
#   Clean residual files.
rm -f $DOCKER_PATH/app

