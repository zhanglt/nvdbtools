#!/bin/bash
# 从nv scanner中提取cvddb文件
#如果文件夹不存在，创建文件夹
if [ ! -d "/tmp/nvdbtools" ]; then
  mkdir /tmp/nvdbtools
fi
# 删除已经存在的cvedb文件
if [ -f "/tmp/nvdbtools/cvedb" ];then
  rm  /tmp/nvdbtools/cvedb
fi

# 检查scanner容器镜像是否存在
scanner=`docker images neuvector/scanner:latest |wc -l`
if [ $scanner -lt 2 ]; then 
  echo "没有发现 neuvector/scanner:latest镜像,pull镜像文件"
  echo "执行命令docker pull neuvector/scanner 可pull镜像"
  docker pull neuvector/scanner
  exit 
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



