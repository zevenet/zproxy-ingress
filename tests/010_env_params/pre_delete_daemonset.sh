#!/bin/bash

# delete controller with default cfg
kubectl delete -f ../../yaml/03_zproxy-controller.yaml

echo "Waiting 50 sec"
sleep 50
