#!/usr/bin/env bash

cur_script_dir="$(cd $(dirname "$0") && pwd)"
WORK_HOME="${cur_script_dir}/.."

go run ${WORK_HOME}/cmd/channels/mock --listen=9092 --listen-http=8084 --listen-internal=8085