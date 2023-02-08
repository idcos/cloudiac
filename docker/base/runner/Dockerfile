#nolint
FROM alpine
RUN apk add --no-cache netcat-openbsd curl docker-cli tzdata 
ENV TZ="Asia/Shanghai"

ARG WORKDIR=/usr/yunji/cloudiac/
WORKDIR ${WORKDIR}

