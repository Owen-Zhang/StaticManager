FROM frolvlad/alpine-glibc:alpine-3.7
MAINTAINER Owen

ADD . /

RUN apk add --no-cache tzdata \
   &&  /bin/cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
   && echo 'Asia/Shanghai' >/etc/timezone \

EXPOSE 8000
CMD ["/home"]

