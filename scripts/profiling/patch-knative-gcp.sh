#!/usr/bin/env bash

BASEDIR=$(dirname "$0")
KN_GCP_DIR="$GOPATH/src/github.com/google/knative-gcp"

# Copy patch to knative-gcp
cp "$BASEDIR/broker.patch" "$KN_GCP_DIR"

# Change working dir
pushd "$KN_GCP_DIR"

# Apply patch
git apply broker.patch

# Update deps
go get "cloud.google.com/go/profiler"
./hack/update-deps.sh

popd
