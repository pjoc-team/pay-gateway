#!/usr/bin/env bash

cur_script_dir="$(cd $(dirname "$0") && pwd)"
WORK_HOME="${cur_script_dir}/.."
cd "${WORK_HOME}"

go tool pprof -http=:61617 ./pay-gateway-cpu.prof

