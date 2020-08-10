#!/bin/bash

#~ set -e

cd "$(dirname "${BASH_SOURCE[0]}")"
BASEDIR="$(dirname "$(pwd)")"
TMP_FILE="/tmp/proxy_rules.cfg"
OUTPUT_FILE="ingress.cfg"
PROXY_FILE="ingress.cfg"
PROXY_NAME="zproxy-ingress"
PROXY_NAMESPACE="zproxy-ingress"
CFG_YAML_DIR="../yaml"
GLOBAL_CFG_YAML="000_GLOBAL_CFG/"
LOG_ERR_REGEXP='err'

TEST_GRACE_TIME=5	# time to wait before checking the proxy cfg
TEST_NAME=""
REPORT_FILE=""
WRITE_FLAG=0
DEBUG_FLAG=0
EXIT_FLAG=0
ENV_FLAG=1
GLOBAL_ERROR=0


function die ()
{
	echo "## $1"
	exit 2
}

function printHelp ()
{
  echo "Usage: \"$0 [Options...]\""
  echo "Options:"
  echo -e "-h, --helo\tIt displays the program help."
  echo -e "-w, --write\tIt writes the output of the tests to be used in future checks."
  echo -e "-t, --test [test_dir]\tIt executes only a test."
  echo -e "-d, --debug\tIt stops the test process until the user wants to continue."
  echo -e "-e, --error-exit\tIt exits from the program if an error exists."
  echo -e "-ne, --no-set-envirovement\tIt does not init and delete the yamls that build the testint env."
  #~ echo -e "-o, --output-file [file]\tIt creates a report file with the error tests."

  exit
}

function stop
{
	read "Press the key ENTER to continue"
}

function getProxyPodName
{
	PROXY_PODNAME=`kubectl get pods -n $PROXY_NAMESPACE | grep $PROXY_NAME | cut -d " " -f1`

	if [[ -z $PROXY_PODNAME ]]; then die "Zproxy pod name was not found"; fi
}

function getProxyCfg
{
	kubectl exec -it $PROXY_PODNAME -n $PROXY_NAMESPACE -- cat $PROXY_FILE | sed 's/\r//g' > $1
}

function checkLogs
{

	return		## TMP, complete this function

	LOGS=`kubectl logs $PROXY_PODNAME -n $PROXY_NAMESPACE | grep -i "$LOG_ERR_REGEXP"`
	if [[ $? -eq 0 ]]; then
		echo "Test '$test' - FAILED. Some errors were found in logs:"
		echo "<< $LOGS"
		TEST_ERR=1
		GLOBAL_ERROR=1
	fi
}

function waitGraceTime
{
	echo "Waiting $TEST_GRACE_TIME"
	sleep $TEST_GRACE_TIME
}

function setEnv
{
	echo "## Configuring controller and services"
	# check if kubectl is configured
	kubectl cluster-info >/dev/null
	if [[ $? -ne 0 ]]; then die "kubectl is not configured"; fi

	# default cfg
	kubectl apply -f $CFG_YAML_DIR
	if [[ $? -ne 0 ]]; then die "Error setting zproxy-ingresses"; fi

	# svc for tests
	kubectl apply -f $GLOBAL_CFG_YAML
	if [[ $? -ne 0 ]]; then die "Error creating global services"; fi

	# svc for tests
	sh $GLOBAL_CFG_YAML/*.sh

	waitGraceTime
}

function unsetEnv
{
	echo ""
	echo "## Cleaning envirovement"

	kubectl delete -f $GLOBAL_CFG_YAML
	kubectl delete -f $CFG_YAML_DIR
}

function execTest
{
	local TEST=$1

	echo ""
	echo "## Executing test '$TEST'"
	cd $TEST

	# execute some scripts before than k8s cfg
	if [[ -e pre_*.sh ]]; then
		sh pre_*.sh
		if [[ $? -ne 0 ]]; then die "Error executing pre scripts"; fi
	fi

	# execute the k8s cfg
	kubectl apply -f ./
	if [[ $? -ne 0 ]]; then die "Error applying yalm files"; fi

	# execute some script before checking the output
	if [[ -e post_*.sh ]]; then
		sh post_*.sh
		if [[ $? -ne 0 ]]; then die "Error executing post scripts"; fi
	fi

	waitGraceTime

	if [[ $WRITE_FLAG -eq 1 ]]; then
		getProxyCfg $OUTPUT_FILE
		echo ">> saved file $TEST/$OUTPUT_FILE"
	else
		getProxyCfg $TMP_FILE
		MSG=`diff $TMP_FILE $OUTPUT_FILE`
		TEST_ERR=$?

		if [[ $TEST_ERR -ne 0 ]]; then
			GLOBAL_ERROR=1
			echo "Test '$TEST' - FAILED"
			echo "<< $MSG"

			if [[ $DEBUG_FLAG -eq 1 ]]; then stop; fi

			#finish?
			if [[ $EXIT_FLAG -eq 1 ]]; then exit 1; fi
		else
			# check logs
			checkLogs
		fi

		if [[ $TEST_ERR -eq 0 ]]; then echo "Test '$TEST' - success"; fi
	fi

	# clean env
	if [[ -e clean_*.sh ]]; then
		sh clean_*.sh
		if [[ $? -ne 0 ]]; then echo "Error executing cleanning scripts"; fi
	fi

	# remove yaml configurations
	kubectl delete -f ./
	if [[ $? -ne 0 ]]; then echo "Error deleting yaml configurations"; fi

	cd ..
}

# INIT

while [[ $# -gt 0 ]]; do
  ARG="$1"
  case $ARG in
	"-h"|"--help")
	  printHelp
	  shift
	  ;;
	"-d"|"--debug")
	  DEBUG_FLAG=1
	  shift
	  ;;
	"-e"|"--error-exit")
	  EXIT_FLAG=1
	  shift
	  ;;
	"-w"|"--write")
	  WRITE_FLAG=1
	  shift
	  ;;
	"-t"|"--test")
	  TEST_NAME=$2
	  shift
	  shift
	  ;;
	"-ne"|"--no-set-envirovement")
	  ENV_FLAG=0
	  shift
	  ;;
	*)
	  echo "Try $0 -h or --help"
	  exit 1
	  ;;
  esac
done


if [[ $ENV_FLAG -eq 1 ]]; then
	setEnv
fi

# load controller pod name
getProxyPodName


if [[ $TEST_NAME != "" ]]; then
	execTest $TEST_NAME
else
	for TEST in `ls -d */`; do

		if [[ $TEST == $GLOBAL_CFG_YAML ]]; then continue; fi

		execTest $TEST
	done
fi


if [[ $ENV_FLAG -eq 1 ]]; then
	unsetEnv
fi

exit $GLOBAL_ERROR
