FROM alpine
RUN apk add --no-cache netcat-openbsd curl tzdata 
ENV TZ="Asia/Shanghai"

ARG WORKDIR=/usr/yunji/cloudiac/
WORKDIR ${WORKDIR}
COPY repos repos

