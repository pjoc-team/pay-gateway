#!/usr/bin/env bash

docker run --rm -v $(pwd):/app -w /app golangci/golangci-lint:v1.31.0 golangci-lint run -v
#find . -name "*.go" -type f -exec bash -c "golangci-lint run --disable=typecheck {} --allow-parallel-runners " \;
