#!/usr/bin/env bash

#BASEDIR=$(pwd)
#PRONAME=`basename $BASEDIR`
#echo "PRONAME=${PRONAME}"
startDir="`pwd`"
export curDir="`dirname $0`"
export fullDir="`pwd`/$curDir"
export appName="`cd $curDir && basename $(pwd)`"
echo "fullDir=$fullDir"
#appName=`basename $fullDir`

echo "appName=$appName"
go run $fullDir --listen-addr=":8080" --etcd-peers="127.0.0.1:2379" -s "$appName"
cd $startDir