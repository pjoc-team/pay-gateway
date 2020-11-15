#!/usr/bin/env bash

cur_script_dir="$(cd $(dirname "$0") && pwd)"
WORK_HOME="${cur_script_dir}/.."

go run ${WORK_HOME}/cmd/database-service/ --listen=9091 --listen-http=8082 --listen-internal=8083