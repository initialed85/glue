#!/usr/bin/env bash

set -e

go test -v ./pkg/discovery
go test -v ./pkg/endpoint
go test -v ./pkg/fragmentation
go test -v ./pkg/network
go test -v ./pkg/serialization
go test -v ./pkg/topics
go test -v ./pkg/transport
go test -v ./pkg/types
go test -v ./pkg/utils
go test -v ./pkg/worker
