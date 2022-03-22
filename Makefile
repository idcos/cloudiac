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

DOCKER_REPO=cloudiac
ifneq ($(DOCKER_REGISTRY),)
  DOCKER_REPO=$(DOCKER_REGISTRY)/cloudiac
endif

# base image 不支持自定义 docker registry
BASE_IMAGE_DOCKER_REPO=cloudiac

DOCKER_BUILD=docker build --build-arg http_proxy="$(http_proxy)" --build-arg https_proxy="$(https_proxy)" --build-arg WORKDIR=$(WORKDIR) 

BUILD_DIR=$(PWD)/build

.PHONY: all build portal runner run run-portal ru-runner clean package repos providers package-release

all: build
build: portal runner tool

build-linux-amd64-portal: 
	GOOS=linux GOARCH=amd64 $(MAKE) portal tool

build-linux-amd64-runner: 
	GOOS=linux GOARCH=amd64 $(MAKE) runner

build-linux-arm64-portal: 
	GOOS=linux GOARCH=arm64 $(MAKE) portal tool

build-linux-arm64-runner: 
	GOOS=linux GOARCH=arm64 $(MAKE) runner

reset-build-dir:
	$(RM) -r $(BUILD_DIR)
	mkdir -p $(BUILD_DIR)/assets/


gen-lang:
	GOOS="" go run cmds/gen-lang/main.go docs/lang.csv portal/consts/e/lang.go

swag-docs: gen-lang
	swag init -g portal/web/api/v1/route.go

mkdocs: 
	GOOS="" GOARCH="" go run scripts/updatedocs/main.go

portal: reset-build-dir mkdocs swag-docs
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
	cp -a ./assets/terraformrc-* $(BUILD_DIR)/assets/ && \
	cp ./scripts/iac-portal.service ./scripts/ct-runner.service $(BUILD_DIR)/ && \
	cp ./configs/config-portal.yml.sample $(BUILD_DIR)/config-portal.yml.sample && \
	cp ./configs/dotenv.sample $(BUILD_DIR)/dotenv.sample && \
	cp ./configs/config-runner.yml.sample $(BUILD_DIR)/config-runner.yml.sample && \
	cp ./configs/demo-conf.yml.sample $(BUILD_DIR)/demo-conf.yml.sample && \
	cd $(BUILD_DIR) && tar -czf ../$(PACKAGE_NAME) ./ && echo "Package: $(PACKAGE_NAME)"

package-linux-amd64:
	GOOS=linux GOARCH=amd64 $(MAKE) package-local

package-linux-arm64:
	GOOS=linux GOARCH=arm64 $(MAKE) package-local

package: package-linux-amd64


BASE_IMAGE_VERSION=$(shell cat docker/base/VERSION)

base-image-portal:
	$(DOCKER_BUILD) -t ${BASE_IMAGE_DOCKER_REPO}/base-iac-portal:$(BASE_IMAGE_VERSION) -f docker/base/portal/Dockerfile .

base-image-portal-arm64:
	$(DOCKER_BUILD) -t ${BASE_IMAGE_DOCKER_REPO}/base-iac-portal:$(BASE_IMAGE_VERSION)-arm64 -f docker/base/portal/Dockerfile-arm64 .

base-image-runner:
	$(DOCKER_BUILD) -t ${BASE_IMAGE_DOCKER_REPO}/base-ct-runner:$(BASE_IMAGE_VERSION) -f docker/base/runner/Dockerfile .

base-image-runner-arm64:
	$(DOCKER_BUILD) -t ${BASE_IMAGE_DOCKER_REPO}/base-ct-runner:$(BASE_IMAGE_VERSION)-arm64 -f docker/base/runner/Dockerfile-arm64 .

base-image-worker: 
	$(DOCKER_BUILD) -t ${BASE_IMAGE_DOCKER_REPO}/base-ct-worker:$(BASE_IMAGE_VERSION) -f docker/base/worker/Dockerfile .

base-image: base-image-portal base-image-runner base-image-worker 
	@echo "Update base image version to $(BASE_IMAGE_VERSION)" && bash scripts/update-base-image-version.sh

base-image-arm64: base-image-portal-arm64 base-image-runner-arm64 base-image-worker
	bash scripts/update-base-image-version.sh

push-base-image:
	for NAME in iac-portal ct-runner ct-worker; do \
	  docker push cloudiac/base-$${NAME}:$(BASE_IMAGE_VERSION) || exit $$?; \
	done


image-portal: build-linux-amd64-portal
	$(DOCKER_BUILD) -t ${DOCKER_REPO}/iac-portal:$(VERSION) -f docker/portal/Dockerfile .

image-portal-arm64: build-linux-arm64-portal
	$(DOCKER_BUILD) -t ${DOCKER_REPO}/iac-portal:$(VERSION) -f docker/portal/Dockerfile-arm64 .

image-runner: build-linux-amd64-runner
	$(DOCKER_BUILD) --build-arg WORKER_IMAGE=${DOCKER_REPO}/ct-worker:$(VERSION) \
	  -t ${DOCKER_REPO}/ct-runner:$(VERSION) -f docker/runner/Dockerfile .

image-runner-arm64: build-linux-arm64-runner 
	$(DOCKER_BUILD) --build-arg WORKER_IMAGE=${DOCKER_REPO}/ct-worker:$(VERSION) \
	  -t ${DOCKER_REPO}/ct-runner:$(VERSION) -f docker/runner/Dockerfile-arm64 .

image-worker: build-linux-amd64-portal
	$(DOCKER_BUILD) -t ${DOCKER_REPO}/ct-worker:$(VERSION) -f docker/worker/Dockerfile .

image-worker-arm64: image-worker

image: image-portal image-worker image-runner 
image-arm64: image-portal-arm64 image-runner-arm64 image-worker-arm64

push-image:
	for NAME in iac-portal ct-runner ct-worker; do \
	  docker push ${DOCKER_REPO}/$${NAME}:$(VERSION) || exit $$?; \
	done

push-image-latest:
	for NAME in iac-portal ct-runner ct-worker; do \
	  docker tag cloudiac/$${NAME}:$(VERSION) cloudiac/$${NAME}:latest && \
	  docker push cloudiac/$${NAME}:latest || exit $$?; \
	done


OSNAME=$(shell uname -s)
CMD_MD5SUM=md5sum | head -c8
ifeq ($(OSNAME),Darwin)
  CMD_MD5SUM=md5 | head -c8
endif

repos: repos.list
	mkdir -p ./repos/cloudiac && \
	cd ./repos/cloudiac && bash ../../scripts/clone-repos.sh

REPOS_SHA1SUM=$(shell tar -c ./repos | $(CMD_MD5SUM))
REPOS_PACKAGE_NAME=cloudiac-repos_$(VERSION)_$(REPOS_SHA1SUM).tar.gz
repos-package:
	tar -czf $(REPOS_PACKAGE_NAME) ./repos


providers:
	bash scripts/generate-providers-mirror.sh

providers-arm64:
	PLATFORM=linux_arm64 bash scripts/generate-providers-mirror.sh

PROVIDERS_SHA1SUM=$(shell tar -c ./assets/providers | $(CMD_MD5SUM))
PROVIDERS_PACKAGE_NAME=cloudiac-providers_$(VERSION)_$(PROVIDERS_SHA1SUM).tar.gz
providers-package:
	@if [[ ! -e "$(PROVIDERS_PACKAGE_NAME)" ]]; then echo "Package $(PROVIDERS_PACKAGE_NAME)"; tar -czf $(PROVIDERS_PACKAGE_NAME) ./assets/providers; fi

