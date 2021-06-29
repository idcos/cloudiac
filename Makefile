#GOCMD=CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go
GOCMD=CGO_ENABLED=0 go
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

RM=/bin/rm -f

## VERSION=$(shell git describe --tags --abbrev=0 --always)
VERSION=$(shell cat VERSION)
DATE_VER=$(shell date "+%Y%m%d")
COMMIT=$(shell git rev-parse --short HEAD)

GOLDFLAGS="-X cloudiac/consts.VERSION=$(VERSION) -X cloudiac/consts.BUILD=$(COMMIT)"
GOBUILD=$(GOCMD) build -v -ldflags $(GOLDFLAGS)
GORUN=$(GOCMD) run -v -ldflags $(GOLDFLAGS)
PB_PROTOC=protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative

WORKDIR?=/usr/yunji/cloudiac
DOCKER_BUILD=docker build --build-arg WORKDIR=$(WORKDIR)

BUILD_DIR=$(PWD)/targets

.PHONY: all build portal runner run run-portal ru-runner clean package repos providers package-release

all: build
build: portal runner tool

build-linux-amd64: 
	GOOS=linux GOARCH=amd64 $(MAKE) build

reset-build-dir:
	$(RM) -r $(BUILD_DIR)
	mkdir -p $(BUILD_DIR)/assets/

portal: reset-build-dir
	swag init -g web/api/v1/route.go
	$(GOBUILD) -o $(BUILD_DIR)/iac-portal ./cmds/portal
	cp ./configs/config-portal.yml.sample $(BUILD_DIR)/config-portal.yml.sample
	cp ./configs/dotenv.sample $(BUILD_DIR)/dotenv.sample

runner: reset-build-dir
	$(GOBUILD) -o $(BUILD_DIR)/ct-runner ./cmds/runner
	cp ./configs/config-runner.yml.sample $(BUILD_DIR)/config-runner.yml.sample

tool: 
	$(GOBUILD) -o $(BUILD_DIR)/iac-tool ./cmds/tool



run: run-portal

run-portal:
	swag init -g web/api/v1/route.go
	$(GORUN) ./cmds/portal -v -c config-portal.yml

run-runner:
	$(GORUN) ./cmds/runner -v -c config-runner.yml

run-tool:
	$(GORUN) ./cmds/tool -v -c config-portal.yml

clean: reset-build-dir
	$(GOCLEAN) ./cmds/portal
	$(GOCLEAN) ./cmds/runner
	$(GOCLEAN) ./cmds/tool


PACKAGE_NAME=cloudiac-$(VERSION).tar.gz
package-local: reset-build-dir clean build
	cp -a ./assets/terraform.py $(BUILD_DIR)/assets/
	cp ./scripts/iac-portal.service ./scripts/ct-runner.service $(BUILD_DIR)/
	@cd $(BUILD_DIR) && tar -czf ../$(PACKAGE_NAME) ./ && echo "Package: $(PACKAGE_NAME)"

package-linux-amd64:
	GOOS=linux GOARCH=amd64 $(MAKE) package-local

package: package-linux-amd64


repos: repos.list
	mkdir -p ./repos/cloud-iac && \
	cd ./repos/cloud-iac && bash ../../scripts/clone-repos.sh

REPOS_PACKAGE_NAME=cloudiac-repos-$(VERSION)-$(DATE_VER).tar.gz
repos-package: repos
	@tar -czf $(REPOS_PACKAGE_NAME) ./repos && echo Package: $(REPOS_PACKAGE_NAME)


providers: 
	bash scripts/generate-providers-mirror.sh

PROVIDERS_PACKAGE_NAME=cloudiac-providers-$(VERSION)-$(DATE_VER).tar.gz
providers-package: providers
	@tar -czf $(PROVIDERS_PACKAGE_NAME) ./assets/providers && echo Package: $(PROVIDERS_PACKAGE_NAME)


docker-image: build-linux-amd64
	$(DOCKER_BUILD) -t cloudiac/iac-portal:$(VERSION) -f docker/portal/Dockerfile . && \
	$(DOCKER_BUILD) -t cloudiac/ct-runner:$(VERSION) -f docker/runner/Dockerfile . && \
	$(DOCKER_BUILD) -t cloudiac/ct-worker:$(VERSION) -f docker/worker/Dockerfile .

docker-push: 
	for NAME in iac-portal ct-runner ct-worker; do \
	  docker push cloudiac/$${NAME}:$(VERSION) || exit $$?; \
	done

docker-push-latest:
	for NAME in iac-portal ct-runner ct-worker; do \
	  docker tag cloudiac/$${NAME}:$(VERSION) cloudiac/$${NAME}:latest && \
	  docker push cloudiac/$${NAME}:latest || exit $$?; \
	done
