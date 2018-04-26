#!/bin/bash

if [ "$#" -ne 2 ]; then
       echo "please input image tag and port"
       exit
fi


#需要两个参数:
#1:镜像版本号 staticmanagement:$1
tag=$1

#2:本机开放的端口号$2:8000
port=$2

if [ -f "/opt/staticmanagement/conf/app.conf" ]; then
	rm -rf /opt/staticmanagement/conf/app.conf
else
	mkdir -p /opt/staticmanagement/conf
fi

if [ -d "/opt/staticmanagement/logs/" ]; then
        rm -rf /opt/staticmanagement/logs
fi

\cp conf/app.conf /opt/staticmanagement/conf/

#删除运行中的容器
containerid=`docker ps -a | grep static | awk '{print $1}'`
docker rm -f ${containerid}

#删除老版本的镜像
imageid=`docker images | grep static | awk '{print $3}'`
docker rmi ${imageid}

#加载镜像
docker load -i staticmanagement_${tag}.tar

#删除以前版本的镜像(这个要通过 docker ps -a | grep staticmanagement 得出结果去获得相应的containerId)

docker run -d -p ${port}:8000 --name=staticmanagement -v /opt/staticmanagement/logs:/logs -v /opt/staticmanagement/conf:/conf staticmanagement:${tag}

