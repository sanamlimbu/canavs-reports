#!/bin/bash

# Build Go binary 'bootstrap' file inside 'terraform/tf_generated' directory

set -e

export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0
export GOFLAGS=-trimpath

echo "Deleting 'bootstrap' binary..."
rm -rf ./terraform/tf_generated/bootstrap
mkdir -p ./terraform/tf_generated
echo "Deleted 'bootstrap' binary."

echo "Building 'bootstrap' binary..."
go build -tags lambda.norpc -mod=readonly -ldflags="-s -w" -o ./terraform/tf_generated/bootstrap ./cmd/lambda/main.go
echo "Built 'bootstrap' binary."