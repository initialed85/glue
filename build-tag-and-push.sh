#!/bin/bash

set -e

mkdir -p ./bin

GOOS=linux GOARCH=amd64 go build -x -o ./bin/simple_endpoint ./cmd/simple_endpoint

docker build --platform=linux/amd64 -t kube-registry:5000/glue:latest -f ./Dockerfile .

docker image push kube-registry:5000/glue:latest
