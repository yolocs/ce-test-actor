#!/usr/bin/env bash

BASEDIR=$(dirname "$0")
KN_GCP_DIR="$GOPATH/src/github.com/google/knative-gcp"

# Copy patch to knative-gcp
cp "$BASEDIR/profile.patch" "$KN_GCP_DIR"

sed "s/{{.PROJECT}}/$PROJECT/" "$BASEDIR/profile.patch" > "$KN_GCP_DIR/profile.patch"

# Change working dir
pushd "$KN_GCP_DIR"

# Apply patch
git apply profile.patch

popd
