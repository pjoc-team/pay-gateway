#!/usr/bin/env bash

cur_script_dir="$(cd $(dirname "$0") && pwd)"
WORK_HOME="${cur_script_dir}/.."
cd "${WORK_HOME}"

go run ${WORK_HOME}/cmd/pay-gateway/ --listen=9090 --listen-http=8080 --listen-internal=8081 --enable-pprof=true --listen-pprof=61616
