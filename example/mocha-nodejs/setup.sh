#!/bin/bash
set -e

docker build -t mocha-nodejs-example:latest .
kind load docker-image mocha-nodejs-example:latest
kubectl apply -f manifests/
kubectl wait --for=condition=available --timeout=60s deployment/express-server
