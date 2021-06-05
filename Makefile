#GOCMD=CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go
GOCMD=CGO_ENABLED=0 go
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

RM=/bin/rm -fv

VERSION=$(shell git describe --tags --abbrev=0 --always)
COMMIT=$(shell git rev-parse --short HEAD)

GOLDFLAGS="-X cloudiac/consts.VERSION=$(VERSION) -X cloudiac/consts.BUILD=$(COMMIT)"
GOBUILD=$(GOCMD) build -v -ldflags $(GOLDFLAGS)
GORUN=$(GOCMD) run -v -ldflags $(GOLDFLAGS)
PB_PROTOC=protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative

BUILD_DIR=$(PWD)/targets

all: portal runner
build: portal runner

build-dir:
	mkdir -p $(BUILD_DIR)

portal: build-dir
	$(GOBUILD) -o $(BUILD_DIR)/iac-portal ./cmds/portal
	cp ./configs/config-portal.yaml.sample $(BUILD_DIR)/config-portal.yaml.sample
	cp ./configs/dotenv.sample $(BUILD_DIR)/dotenv.sample

runner: build-dir
	$(GOBUILD) -o $(BUILD_DIR)/ct-runner ./cmds/runner
	cp ./configs/config-runner.yaml.sample $(BUILD_DIR)/config-runner.yaml.sample

run: run-portal

run-portal:
	$(GORUN) ./cmds/portal -v -c config-portal.yaml

run-runner:
	$(GORUN) ./cmds/runner -v -c config-runner.yaml

clean:
	$(GOCLEAN) ./cmds/portal
	$(GOCLEAN) ./cmds/runner
	$(RM) $(BUILD_DIR)/*

package: clean build
	cd $(BUILD_DIR) && tar -czvf ../cloud-iac-$(VERSION).tar.gz ./

