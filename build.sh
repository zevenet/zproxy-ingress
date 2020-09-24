#!/bin/sh

set -e

BASEDIR=$(dirname "$0")
DOCKER_PATH="${BASEDIR}/docker"

# Optionally, use a nftlb devel package
if [ -s "$1" ]; then
	cp -v "$1" $DOCKER_PATH/zproxy.deb
else
	# Use empty file to avoid docker COPY directive failure
	touch $DOCKER_PATH/zproxy.deb
fi

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
rm -fv $DOCKER_PATH/app $DOCKER_PATH/zproxy.deb
