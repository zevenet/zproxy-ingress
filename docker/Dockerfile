# Download the latest Debian image
FROM debian:stable

# Default shell when executing RUN
SHELL ["/bin/bash", "-c"]


# Put Debian in not interactive mode
RUN echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections

RUN apt-get update
RUN apt-get install -y gnupg ca-certificates wget procps

# Install zproxy and dependencies
RUN wget -O - http://repo.zevenet.com/zevenet.com.gpg.key | apt-key add -
RUN echo "deb [arch=amd64] http://repo.zevenet.com/ce/v5 buster main" | tee -a /etc/apt/sources.list

RUN apt-get update
RUN apt-get install -y zproxy

# load container parameters
COPY ./cfg_tpl /cfg_tpl

# After compiling the Golang client, copy the binary to / and make it the entrypoint
COPY ./app /goclient

COPY ./etc/ /etc/
COPY ./start.sh /start.sh


# Replace zproxy with a devel version if zproxy.deb exists in this directory
COPY ./zproxy.deb .
RUN if [ -s "zproxy.deb" ] ; then dpkg -i ./zproxy.deb; fi
RUN rm zproxy.deb

