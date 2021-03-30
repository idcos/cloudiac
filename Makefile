#GOCMD=CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go
GOCMD=CGO_ENABLED=0 go
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

VERSION=$(shell git describe --tags --abbrev=0 --always)
COMMIT=$(shell git rev-parse --short HEAD)

GOLDFLAGS="-X cloudiac/consts.VERSION=$(VERSION) -X cloudiac/consts.BUILD=$(COMMIT)"
GOBUILD=$(GOCMD) build -v -ldflags $(GOLDFLAGS)
GORUN=IMS_NO_AUTH=1 $(GOCMD) run -v -ldflags $(GOLDFLAGS)
PB_PROTOC=protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative

BUILD_DIR=$(PWD)/builds

all: portal runner

build-dir:
	mkdir -p $(BUILD_DIR)

portal: build-dir
	$(GOBUILD) -o $(BUILD_DIR)/iac-portal ./cmds/portal

runner: build-dir
	$(GOBUILD) -o $(BUILD_DIR)/ct-runner ./cmds/runner

run: run-portal

run-portal:
	$(GORUN) ./cmds/portal -v

run-runner:
	$(GORUN) ./cmds/runner -v

clean:
	$(GOCLEAN) ./cmds/portal
	$(GOCLEAN) ./cmds/runner
