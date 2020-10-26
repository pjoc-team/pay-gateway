#!/usr/bin/env bash

find . -name "*.go" -type f -exec bash -c "golangci-lint run --disable=typecheck {} --allow-parallel-runners " \;
