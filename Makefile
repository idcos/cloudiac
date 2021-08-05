#GOCMD=CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go
GOCMD=CGO_ENABLED=0 go
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

RM=/bin/rm -f

#VERSION=$(shell git describe --tags --abbrev=0 --always)
VERSION=$(shell cat VERSION)
DATE_VER=$(shell date "+%Y%m%d")
COMMIT=$(shell git rev-parse --short HEAD)

GOLDFLAGS="-X cloudiac/common.VERSION=$(VERSION) -X cloudiac/common.BUILD=$(COMMIT)"
GOBUILD=$(GOCMD) build -v -ldflags $(GOLDFLAGS)
GORUN=$(GOCMD) run -v -ldflags $(GOLDFLAGS)
PB_PROTOC=protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative

WORKDIR?=/usr/yunji/cloudiac
DOCKER_BUILD=docker build --build-arg WORKDIR=$(WORKDIR)

BUILD_DIR=$(PWD)/build

.PHONY: all build portal runner run run-portal ru-runner clean package repos providers package-release

all: build
build: portal runner tool

build-linux-amd64-portal: 
	GOOS=linux GOARCH=amd64 $(MAKE) portal tool

build-linux-amd64-runner: 
	GOOS=linux GOARCH=amd64 $(MAKE) runner

reset-build-dir:
	$(RM) -r $(BUILD_DIR)
	mkdir -p $(BUILD_DIR)/assets/

swag-docs:
	swag init -g portal/web/api/v1/route.go

portal: reset-build-dir swag-docs
	$(GOBUILD) -o $(BUILD_DIR)/iac-portal ./cmds/portal

runner: reset-build-dir
	$(GOBUILD) -o $(BUILD_DIR)/ct-runner ./cmds/runner

tool: 
	$(GOBUILD) -o $(BUILD_DIR)/iac-tool ./cmds/tool



run: run-portal

run-portal: swag-docs
	$(GORUN) ./cmds/portal -v -c config-portal.yml

run-runner:
	$(GORUN) ./cmds/runner -v -c config-runner.yml

run-tool:
	$(GORUN) ./cmds/tool -v -c config-portal.yml

clean: reset-build-dir
	$(GOCLEAN) ./cmds/portal
	$(GOCLEAN) ./cmds/runner
	$(GOCLEAN) ./cmds/tool


PACKAGE_NAME=cloudiac_$(VERSION).tar.gz
package-local: reset-build-dir clean build
	cp -a ./assets/terraform.py $(BUILD_DIR)/assets/ && \
	cp ./scripts/iac-portal.service ./scripts/ct-runner.service $(BUILD_DIR)/ && \
	cp ./configs/config-portal.yml.sample $(BUILD_DIR)/config-portal.yml.sample && \
	cp ./configs/dotenv.sample $(BUILD_DIR)/dotenv.sample && \
	cp ./configs/config-runner.yml.sample $(BUILD_DIR)/config-runner.yml.sample && \
	cd $(BUILD_DIR) && tar -czf ../$(PACKAGE_NAME) ./ && echo "Package: $(PACKAGE_NAME)"

package-linux-amd64:
	GOOS=linux GOARCH=amd64 $(MAKE) package-local

package: package-linux-amd64


image-portal: build-linux-amd64-portal
	$(DOCKER_BUILD) -t cloudiac/iac-portal:$(VERSION) -f docker/portal/Dockerfile .

image-runner: build-linux-amd64-runner
	$(DOCKER_BUILD) -t cloudiac/ct-runner:$(VERSION) -f docker/runner/Dockerfile .

image-worker:
	$(DOCKER_BUILD) -t cloudiac/ct-worker:$(VERSION) -f docker/worker/Dockerfile .

image: image-portal image-runner image-worker

push-image:
	for NAME in iac-portal ct-runner ct-worker; do \
	  docker push cloudiac/$${NAME}:$(VERSION) || exit $$?; \
	done

push-image-latest:
	for NAME in iac-portal ct-runner ct-worker; do \
	  docker tag cloudiac/$${NAME}:$(VERSION) cloudiac/$${NAME}:latest && \
	  docker push cloudiac/$${NAME}:latest || exit $$?; \
	done


OSNAME=$(shell uname -s)
CMD_SHA1SUM=sha1sum | head -c8
ifeq ($(OSNAME),Darwin)
  CMD_SHA1SUM=shasum -a 1 | head -c8
endif

repos: repos.list
	mkdir -p ./repos/cloudiac && \
	cd ./repos/cloudiac && bash ../../scripts/clone-repos.sh

REPOS_SHA1SUM=$(shell tar -c ./repos | $(CMD_SHA1SUM))
REPOS_PACKAGE_NAME=cloudiac-repos_$(VERSION)_$(REPOS_SHA1SUM).tar.gz
repos-package:
	tar -czf $(REPOS_PACKAGE_NAME) ./repos


providers:
	bash scripts/generate-providers-mirror.sh

PROVIDERS_SHA1SUM=$(shell tar -c ./assets/providers | $(CMD_SHA1SUM))
PROVIDERS_PACKAGE_NAME=cloudiac-providers_$(VERSION)_$(PROVIDERS_SHA1SUM).tar.gz
providers-package:
	tar -czf $(PROVIDERS_PACKAGE_NAME) ./assets/providers

