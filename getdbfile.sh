#!/bin/bash
# 从nv scanner中提取cvddb文件

scanner=`docker images neuvector/scanner:latest |wc -l`
if [ $scanner -lt 2 ]; then 

  echo "没有发现 neuvector/scanner:latest镜像,请先pull镜像文件"
  exit 

fi

if [ -f "/tmp/nvdbtools/cvedb" ];then
  rm  /tmp/nvdbtools/cvedb
fi
ID=`docker ps -a  |grep cvedb |grep Created |awk '{print $1}'`
if [  -n "$ID" ]; then
 docker rm -f $ID
fi

docker container create --name cvedb neuvector/scanner

ID=`docker ps -a  |grep cvedb |awk '{print $1}'`


if [ ! -n "$ID" ]; then
    echo "scanner container is not created"
else
    docker cp   ${ID}":/etc/neuvector/db/cvedb" /tmp/nvdbtools/
fi
docker rm -f $ID
#docker cp   ${ID}":/etc/neuvector/db/cvedb" /tmp/ 


