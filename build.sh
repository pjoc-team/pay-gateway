#!/usr/bin/env bash
cur_script_dir="$(cd $(dirname $0) && pwd)"
WORK_HOME="${cur_script_dir}"
source "${WORK_HOME}/setup.sh"
docker build --build-arg REPOSITORY=$REPOSITORY --build-arg GOPROXY=${GOPROXY} --build-arg APP=${NAME} . -t image --file Dockerfile
