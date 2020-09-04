#!/bin/bash

# create controller with default cfg
#~ kubectl delete -f ../../yaml/

echo "Waiting 50 sec"
sleep 50

# create controller with default cfg
kubectl apply -f ../../yaml/

echo "Waiting 5 sec"
sleep 5
