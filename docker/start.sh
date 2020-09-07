#!/bin/bash

### set only unset variables

ENVFILE="/env.conf"
function overwriteEnvVariables
{
	for V in `env`
	do
		# split key and value
		out=( $(grep -Eo '[^=]+|[^=]+' <<<"$V") )
		sed -Ei "s|${out[0]}=(\"?).*|${out[0]}=\1${out[1]}\1|" $ENVFILE
	done

	source $ENVFILE
}


function createCfgFile ()
{
	sed -Ei "s|#SOCKETFILE#|$SocketFile|g" /ingress.cfg && \
	sed -Ei "s|#LISTENERIP#|$ListenerIP|g" /ingress.cfg && \
	sed -Ei "s|#HTTPPORT#|$HTTPPort|g" /ingress.cfg && \
	sed -Ei "s|#TOTALTO#|$TotalTO|g" /ingress.cfg && \
	sed -Ei "s|#CONNTO#|$ConnTO|g" /ingress.cfg && \
	sed -Ei "s|#ALIVETO#|$AliveTO|g" /ingress.cfg && \
	sed -Ei "s|#CLIENTTO#|$ClientTO|g" /ingress.cfg && \
	sed -Ei "s|#ECDHCURVE#|$ECDHCurve|g" /ingress.cfg && \
	sed -Ei "s|#DHFILE#|$DHFile|g" /ingress.cfg && \
	sed -Ei "s|#IGNORE100CONTINUE#|$Ignore100Continue|g" /ingress.cfg && \
	sed -Ei "s|#LOGSLEVEL#|$LogsLevel|g" /ingress.cfg
}

# configure the app
overwriteEnvVariables
createCfgFile


# Start the first process
$BinPath -v -f $ConfigFile &
status=$?
if [ $status -ne 0 ]; then
  echo "Failed to start zproxy: $status"
  exit $status
fi

# waiting a grace time
sleep 3

# Start the second process
/goclient /container_params.conf &
status=$?
if [ $status -ne 0 ]; then
  echo "Failed to start GO client: $status"
  exit $status
fi


while sleep $DaemonCheckTimeout; do
  ps aux |grep zproxy |grep -q -v grep
  PROCESS_ZPROXY_STATUS=$?
  if [ $PROCESS_ZPROXY_STATUS -ne 0 ]; then
    echo "The zproxy process exited with error."
  fi

  ps aux |grep goclient |grep -q -v grep
  PROCESS_GO_STATUS=$?
  if [ $PROCESS_GO_STATUS -ne 0 ]; then
    echo "The GO client exited with error."
  fi

  if [ $PROCESS_ZPROXY_STATUS -ne 0 -o $PROCESS_GO_STATUS -ne 0 ]; then
    exit 1
  fi
done

