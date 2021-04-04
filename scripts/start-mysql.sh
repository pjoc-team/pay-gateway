#!/usr/bin/env bash

export name="pay-mysql"

[ -n "$(docker ps -a -f name=${name} | grep ${name} | awk '{print $NF}' | grep -w ${name})" ] && echo "${name} is running!" && exit 0

cur_script_dir="$(cd $(dirname "$0") && pwd)"
WORK_HOME="${cur_script_dir}/.."
sqlDir="${WORK_HOME}/ddl"
echo "Sql dir: $sqlDir"
echo "Should init sql scripts: $(ls "$sqlDir")"
docker run -d --name ${name} -v "$sqlDir":/docker-entrypoint-initdb.d -v "$WORK_HOME/mysql/data":/var/lib/mysql -e MYSQL_ROOT_PASSWORD=111 -e MYSQL_USER=pjoc -e MYSQL_PASSWORD=111 -e MYSQL_DATABASE=pay_gateway -e LANG=C.UTF-8  -p 3306:3306 mysql:latest --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci --transaction_isolation=READ-COMMITTED --default-authentication-plugin=mysql_native_password
