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
COPY ./container_params.conf /container_params.conf
COPY ./ingress.tpl /ingress.tpl

# After compiling the Golang client, copy the binary to / and make it the entrypoint
#COPY ./app /goclient

COPY ./start.sh /start.sh
COPY ./ingress.tpl /ingress.cfg
RUN source /container_params.conf && \
	sed -Ei "s|#CONFIGFILE#|$configFile|g" /start.sh && \
	sed -Ei "s|#DAEMONCHECKTIMEOUT#|$daemonCheckTimeout|g" /start.sh && \
	sed -Ei "s|#BINPATH#|$binPath|g" /start.sh && \
	sed -Ei "s|#SOCKETFILE#|$socketFile|g" /ingress.cfg && \
	sed -Ei "s|#DEFAULTIP#|$defaultIp|g" /ingress.cfg && \
	sed -Ei "s|#DEFAULTPORT#|$defaultPort|g" /ingress.cfg && \
	sed -Ei "s|#LOGSLEVEL#|$logsLevel|g" /ingress.cfg

CMD [ "/bin/bash", "-c", "/start.sh" ]
