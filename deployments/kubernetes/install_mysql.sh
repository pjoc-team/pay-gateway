#!/usr/bin/env bash
helm repo add incubator https://charts.helm.sh/incubator
export pass="$(LC_ALL=C tr -dc 'A-Za-z0-9' </dev/urandom | head -c 32 ; echo -n)"
helm -n pay install pay-mysqlha incubator/mysqlha --set "xtraBackupImage=yizhiyong/xtrabackup:latest" --set "mysqlDatabase=pay-gateway" --set "mysqlUser=pay-gateway"
( echo "cat <<EOF" ; cat mysql-cm.yaml.tpl ; echo EOF ) | sh > mysql-cm.yaml
kubectl apply -f mysql-cm.yaml