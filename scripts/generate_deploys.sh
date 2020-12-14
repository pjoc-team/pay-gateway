#!/usr/bin/env bash

params="${1:-version}"
echo "replace params: ${params}"

function sedFile() {
    pattern="${1}"
    file="${2}"
    [[ -z "${pattern}" ]] && echo "pattern is null!" && exit 1
    [[ -z "${file}" ]] && echo "file is null!" && exit 1
    now=$(date +'%Y%m%d%H%M%S')
    sed "${pattern}" "${file}" > "${file}.${now}.tmp";
    mv "${file}.${now}.tmp" "${file}";
}

CUR_SCRIPT_DIR="$(dirname "$0")"
WORK_HOME="./${CUR_SCRIPT_DIR}/.."
if [[ -z "${CUR_SCRIPT_DIR}" ]];then
  WORK_HOME="./"
fi

ls ${WORK_HOME}/cmd/channels | while read channel; do
  mkdir -p ${WORK_HOME}/generated-deployments;
  ls ${WORK_HOME}/deployments/kubernetes/channel-tpl | while read f; do
    filename="${f%.*}-$channel"
    ext="${f##*.}"
    file="${filename}.${ext}"
    sed "s/{{channel}}/$channel/g" ${WORK_HOME}/deployments/kubernetes/channel-tpl/$f > ${WORK_HOME}/generated-deployments/$file;
  done;
done

# 生成服务器部署模板
cp -R ${WORK_HOME}/deployments/kubernetes/deploy-tpl/* ${WORK_HOME}/generated-deployments/

# 渠道服务部署模板
ls ${WORK_HOME}/generated-deployments/ | while read f; do
  echo "$params" | tr " " "\n" | while read param; do
    echo "replace {{${param}}} to ${!param}"
    sedFile "s/{{${param}}}/${!param}/g" ${WORK_HOME}/generated-deployments/$f;
  done
done

