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
COPY ./container_params.conf /tmp.conf
RUN grep -Ev '^\s*#' /tmp.conf > /container_params.conf
RUN grep -Ev '^\s*\[' /container_params.conf > /env.conf
COPY ./ingress.tpl /ingress.tpl

# After compiling the Golang client, copy the binary to / and make it the entrypoint
COPY ./app /goclient

RUN mkdir /certs
COPY ./zencert.pem /certs/zencert.pem
COPY ./dh2048.pem /dh2048.pem
COPY ./start.sh /start.sh
COPY ./ingress.tpl /ingress.cfg

RUN source /env.conf && \
	sed -Ei "s|#CONFIGFILE#|$ConfigFile|g" /start.sh && \
	sed -Ei "s|#DAEMONCHECKTIMEOUT#|$DaemonCheckTimeout|g" /start.sh && \
	sed -Ei "s|#BINPATH#|$BinPath|g" /start.sh && \
	sed -Ei "s|#SOCKETFILE#|$SocketFile|g" /ingress.cfg && \
	sed -Ei "s|#LISTENERIP#|$ListenerIP|g" /ingress.cfg && \
	sed -Ei "s|#HTTPPORT#|$HTTPPort|g" /ingress.cfg && \
	sed -Ei "s|#TOTALTO#|$TotalTO|g" /ingress.cfg && \
	sed -Ei "s|#CONNTO#|$ConnTO|g" /ingress.cfg && \
	sed -Ei "s|#ALIVETO#|$AliveTO|g" /ingress.cfg && \
	sed -Ei "s|#CLIENTTO#|$ClientTO|g" /ingress.cfg && \
	sed -Ei "s|#DHFILE#|$DHFile|g" /ingress.cfg && \
	sed -Ei "s|#ECDHCURVE#|$ECDHCurve|g" /ingress.cfg && \
	sed -Ei "s|#LOGSLEVEL#|$LogsLevel|g" /ingress.cfg


# remove tmp files
RUN rm /env.conf /tmp.conf

CMD [ "/bin/bash", "-c", "/start.sh" ]
