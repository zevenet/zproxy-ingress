#################################################
# Temporal container: compile kube-nftlb-client #
#################################################
FROM golang:1.15.2-buster AS client-builder

# Default shell when executing RUN
SHELL ["/bin/bash", "-c"]

# Read issue and accepted answer: https://github.com/moby/moby/issues/34513#issuecomment-389467566
LABEL stage=intermediate

# Start at /kube-nftlb dir
WORKDIR /zproxy-ingress

# Copy everything to /zproxy-ingress
COPY . .

# Compile using local dependencies
RUN GOOS=linux CGO_ENABLED=1 go build -tags 'osusergo netgo' -mod=vendor ./cmd/zproxy-ingress



###############################################
# Main container: nftlb and kube-nftlb-client #
###############################################

# Download the latest Debian image
#FROM debian:stable
FROM debian:buster-slim

# Default shell when executing RUN
SHELL ["/bin/bash", "-c"]

ARG DOCKER_PATH
WORKDIR /

# Put Debian in not interactive mode
RUN echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections

RUN apt-get update
RUN apt-get install -y gnupg ca-certificates wget

# Install zproxy and dependencies
RUN wget -O - http://repo.zevenet.com/zevenet.com.gpg.key | apt-key add -
RUN echo "deb [arch=amd64] http://repo.zevenet.com/ce/v5 buster main" | tee -a /etc/apt/sources.list


RUN apt-get update
RUN apt-get install -y zproxy procps

# clean packaes
RUN apt-get remove -y gnupg ca-certificates wget
RUN apt-get autoremove -y
RUN apt-get clean

# After compiling the Golang client, copy the binary to / and make it the entrypoint
COPY  --from=client-builder /zproxy-ingress/zproxy-ingress /goclient

# load container parameters
COPY ${DOCKER_PATH}/ /

# Replace zproxy with a devel version if zproxy.deb exists in this directory
RUN if [ -s "zproxy.deb" ] ; then dpkg -i ./zproxy.deb; fi
RUN rm zproxy.deb
