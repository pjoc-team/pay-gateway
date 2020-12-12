#!/usr/bin/env bash

find . -name "*.go" | while read f; do echo "${f%/*}"; done | sort | uniq | xargs docker run -t --rm -v $(pwd):/app -v ${GOPATH}/pkg/mod:/go/pkg/mod -w /app golangci/golangci-lint:v1.31.0 golangci-lint run --timeout 120s --allow-parallel-runners
#find . -name "*.go" -type f -exec bash -c "golangci-lint run --disable=typecheck {} --allow-parallel-runners " \;
