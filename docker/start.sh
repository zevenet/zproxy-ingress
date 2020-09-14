#!/bin/bash

### set only unset variables

ENVFILE="/env.conf"
CFGDIR="/cfg_tpl"
CFGFILE="/container_params.conf"

# get variables from the envirovement and it compares with the file ones
function overwriteEnvVariables ()
{
	local FILE=$1
	for V in `env`
	do
		# split key and value
		out=( $(grep -Eo '[^=]+|[^=]+' <<<"$V") )
		sed -Ei "s|${out[0]}=(\"?).*|${out[0]}=\1${out[1]}\1|" $FILE
	done

	source $FILE
}


# it creates a config file using all configuration files
function createClientCfgFile ()
{
	$CFGFILE > "";	# Clean file

	for FILE in `ls $CFGDIR`;
	do
		PATHFILE="${CFGDIR}/${FILE}"
		overwriteEnvVariables $PATHFILE

		echo "[$FILE]" >> $CFGFILE
		grep -Ev '^\s*#' "${CFGDIR}/$FILE" >> $CFGFILE
		echo "" >> $CFGFILE
	done
}

# it creates a zproxy config file to start it
function createDaemonCfgFile ()
{
echo -n "
# Init template

Daemon			0
LogLevel        $ProxyLogsLevel
LogFacility		-
Timeout         $TotalTO
ConnTO          $ConnTO
Alive           $AliveTO
Client          $ClientTO
Control         \"$SocketFile\"
DHParams		\"$DHFile\"
ECDHCurve		\"$ECDHCurve\"
Ignore100Continue $Ignore100Continue


ListenHTTP
        Address 0.0.0.0
        Port 80
        xHTTP 4
        RewriteLocation 1
End" > $ConfigFile

}

# configure the app
createClientCfgFile
createDaemonCfgFile


# Start the first process
$Bin -v -f $ConfigFile &
status=$?
if [ $status -ne 0 ]; then
  echo "Failed to start zproxy: $status"
  exit $status
fi

# waiting a grace time
sleep 3

# Start the second process
$GoClientBin $CFGFILE &
status=$?
if [ $status -ne 0 ]; then
  echo "Failed to start GO client: $status"
  exit $status
fi


while sleep $DaemonsCheckTimeout; do
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

