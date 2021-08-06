FROM alpine
RUN apk add netcat-openbsd curl

EXPOSE 9030

ARG WORKDIR=/usr/yunji/cloudiac/
WORKDIR ${WORKDIR}

COPY configs/config-portal.yml.sample config-portal.yml
COPY configs/dotenv.sample .
COPY build/iac-portal .
COPY build/iac-tool .
COPY repos  repos
CMD  ./iac-portal
