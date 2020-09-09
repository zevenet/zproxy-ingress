#!/bin/bash

cd "$(dirname "${BASH_SOURCE[0]}")"

kubectl create secret tls tls-cert --key cert.key --cert ./cert.crt
kubectl create secret tls tls-cert --key cert.key --cert ./cert.crt -n app1-ns
kubectl create secret tls tls-cert --key cert.key --cert ./cert.crt -n app2-ns
kubectl create secret tls tls-cert-2 --key cert.key --cert ./cert.crt -n app2-ns

kubectl create secret generic pem-cert --from-file=pem=./cert.pem
kubectl create secret generic pem-cert --from-file=pem=./cert.pem -n app1-ns
kubectl create secret generic pem-cert --from-file=pem=./cert.pem -n app2-ns

kubectl create secret generic default-cert --from-file=pem=./cert.pem -n zproxy-ingress
