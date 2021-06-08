#GOCMD=CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go
GOCMD=CGO_ENABLED=0 go
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

## REPO_BASE?=https://github.com/idcos
REPO_BASE?=http://10.0.3.124:3000

RM=/bin/rm -fv

VERSION=$(shell git describe --tags --abbrev=0 --always)
COMMIT=$(shell git rev-parse --short HEAD)

GOLDFLAGS="-X cloudiac/consts.VERSION=$(VERSION) -X cloudiac/consts.BUILD=$(COMMIT)"
GOBUILD=$(GOCMD) build -v -ldflags $(GOLDFLAGS)
GORUN=$(GOCMD) run -v -ldflags $(GOLDFLAGS)
PB_PROTOC=protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative

BUILD_DIR=$(PWD)/targets

.PHONY: all build build-dir portal runner run run-portal ru-runner clean package repos

all: portal runner
build: portal runner

build-dir:
	mkdir -p $(BUILD_DIR)

portal: build-dir
	swag init -g web/api/v1/route.go
	$(GOBUILD) -o $(BUILD_DIR)/iac-portal ./cmds/portal
	cp ./configs/config-portal.yaml.sample $(BUILD_DIR)/config-portal.yaml.sample
	cp ./configs/dotenv.sample $(BUILD_DIR)/dotenv.sample

runner: build-dir
	$(GOBUILD) -o $(BUILD_DIR)/ct-runner ./cmds/runner
	cp ./configs/config-runner.yaml.sample $(BUILD_DIR)/config-runner.yaml.sample

run: run-portal

run-portal:
	swag init -g web/api/v1/route.go
	$(GORUN) ./cmds/portal -v -c config-portal.yaml

run-runner:
	$(GORUN) ./cmds/runner -v -c config-runner.yaml

clean:
	$(GOCLEAN) ./cmds/portal
	$(GOCLEAN) ./cmds/runner
	$(RM) -r $(BUILD_DIR)

package: clean build repos
	cp -a ./repos $(BUILD_DIR) && \
	cd $(BUILD_DIR) && tar -czvf ../cloud-iac-$(VERSION).tar.gz ./

repos: repos.list
	$(RM) -r ./repos/iac
	mkdir -p ./repos/iac
	cd ./repos/iac/ && cat ../../repos.list | while read REPO_PATH; do \
		git clone --bare $(REPO_BASE)$${REPO_PATH} && REPO_NAME=`basename $${REPO_PATH}` && \
		cp $${REPO_NAME}/hooks/post-update.sample $${REPO_NAME}/hooks/post-update && \
		bash $${REPO_NAME}/hooks/post-update ;\
	done
