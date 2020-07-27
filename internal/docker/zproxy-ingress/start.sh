#!/bin/bash

# Start the first process
#BINPATH# -v -f #CONFIGFILE# &
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

# Naive check runs checks once a minute to see if either of the processes exited.
# This illustrates part of the heavy lifting you need to do if you want to run
# more than one service in a container. The container exits with an error
# if it detects that either of the processes has exited.
# Otherwise it loops forever, waking up every 60 seconds

while sleep #DAEMONCHECKTIMEOUT#; do
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

