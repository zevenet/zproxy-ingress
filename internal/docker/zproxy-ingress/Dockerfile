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


# remove tmp files
RUN rm /tmp.conf

#CMD [ "/bin/bash", "-c", "/start.sh" ]
