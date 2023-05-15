#!/bin/bash
# 从nv scanner中提取cvddb文件

# 检查scanner容器镜像是否存在
scanner=`docker images neuvector/scanner:latest |wc -l`
if [ $scanner -lt 2 ]; then 
  echo "没有发现 neuvector/scanner:latest镜像,请先pull镜像文件"
  echo "执行命令docker pull neuvector/scanner 可pull镜像"
  exit 
fi

if [ -f "/tmp/nvdbtools/cvedb" ];then
  rm  /tmp/nvdbtools/cvedb
fi
#检查 scanner容器是否处于运行状态
ID=`docker ps -a  |grep cvedb |grep Created |awk '{print $1}'`
if [  -n "$ID" ]; then
 docker rm -f $ID
fi
#运行scanner容器
docker container create --name cvedb neuvector/scanner
#获取容器ID
ID=`docker ps -a  |grep cvedb |awk '{print $1}'`

if [ ! -n "$ID" ]; then
    echo "scanner container is not created"
else
#copy cvedb数据库文件 到本地
    docker cp   ${ID}":/etc/neuvector/db/cvedb" /tmp/nvdbtools/
fi
# 删除运行的容器
docker rm -f $ID



