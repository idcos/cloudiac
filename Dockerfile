FROM golang:1.16-alpine AS builder
RUN apk add make git

WORKDIR /workspace
ENV GO111MODULE="on" GOPROXY="https://goproxy.io,direct"
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

COPY . .
RUN make all

############################
FROM alpine
RUN apk add netcat-openbsd curl

WORKDIR /usr/yunji/cloudiac/

COPY --from=builder /workspace/configs/config-portal.yaml.sample /usr/yunji/cloudiac/config-portal.yaml
COPY --from=builder /workspace/configs/config-runner.yaml.sample /usr/yunji/cloudiac/config-runner.yaml
COPY --from=builder /workspace/configs/dotenv.sample /usr/yunji/cloudiac/.env

COPY --from=builder /workspace/builds/iac-portal /usr/yunji/cloudiac/
COPY --from=builder /workspace/builds/ct-runner /usr/yunji/cloudiac/
RUN chmod a+x /usr/yunji/cloudiac/iac-portal /usr/yunji/cloudiac/ct-runner

