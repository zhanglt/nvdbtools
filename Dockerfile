FROM centos:centos8
MAINTAINER zhanglt


ENV LANG=C.UTF-8 \
    PYTHONUNBUFFERED=1

COPY exec /

RUN sed -i 's/mirrorlist/#mirrorlist/g' /etc/yum.repos.d/CentOS-*
RUN sed -i 's|#baseurl=http://mirror.centos.org|baseurl=http://vault.centos.org|g' /etc/yum.repos.d/CentOS-*
RUN yum makecache && yum install python39 -y
RUN pip3 install -r ./requirements.txt 

