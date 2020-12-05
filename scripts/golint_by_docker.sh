#!/usr/bin/env bash

find . -name "*.go" | while read f; do echo "${f%/*}"; done | sort | uniq | xargs docker run --rm -t --entrypoint="" -v $(pwd):/app -w /app pjoc/go-action golint
