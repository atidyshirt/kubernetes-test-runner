#!/bin/bash
set -e

if ! kind get clusters | grep -q "kind"; then
    kind create cluster
fi

kubectl apply -f manifest/
kubectl wait --for=condition=ready pod -l app=go-http-server --timeout=60s
