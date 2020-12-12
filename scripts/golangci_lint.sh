#!/usr/bin/env bash

find . -name "*.go" | while read f; do echo "${f%/*}"; done | sort | uniq | xargs golangci-lint run --timeout 120s --allow-parallel-runners
#find . -name "*.go" -type f -exec bash -c "golangci-lint run --disable=typecheck {} --allow-parallel-runners " \;
