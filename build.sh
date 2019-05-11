#!/usr/bin/env bash
export GO111MODULE=on
CGO_ENABLED=0 GOOS=linux go build -o ./bin/main .
