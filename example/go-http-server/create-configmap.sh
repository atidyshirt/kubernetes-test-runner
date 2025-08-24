#!/bin/bash
set -e

TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

cp -r . $TEMP_DIR/

rm -rf $TEMP_DIR/node_modules $TEMP_DIR/.git $TEMP_DIR/dist $TEMP_DIR/build

kubectl create configmap source-code --from-file=$TEMP_DIR --dry-run=client -o yaml > manifest/source-code-configmap.yaml
