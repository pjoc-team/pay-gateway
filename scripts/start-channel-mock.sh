#!/usr/bin/env bash

cur_script_dir="$(cd $(dirname "$0") && pwd)"
WORK_HOME="${cur_script_dir}/.."
cd "${WORK_HOME}"

go run ${WORK_HOME}/cmd/channels/channel-mock --listen=9092 --listen-http=8084 --listen-internal=8085 --enable-pprof=true --listen-pprof=61618
