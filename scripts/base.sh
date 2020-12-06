#!/usr/bin/env bash

echo "baseName: $(basename $0)"
function echoGreen() {
  echo -e "\033[32m$1\033[0m"
}

function echoRed() {
  echo -e "\033[31m$1\033[0m"
}

function echoYellow() {
  echo -e "\033[33m$1\033[0m"
}

cur_script_dir="$(cd $(dirname $0) && pwd)"
WORK_HOME="${cur_script_dir}/.."