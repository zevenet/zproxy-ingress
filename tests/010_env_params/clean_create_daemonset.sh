#!/bin/bash

# create controller with default cfg
#~ kubectl delete -f ../../yaml/

echo "Waiting 50 sec"

TIME=0
while [[ $TIME -lt 50 ]]; do

	kubectl get pod -A | grep zproxy >/dev/null 2>&1
	if [[ $? -ne 0 ]]; then
		break
	fi

	TIME=${TIME}+5
done

exit 0

# create controller with default cfg
kubectl apply -f ../../yaml/

echo "Waiting 5 sec"
sleep 5
