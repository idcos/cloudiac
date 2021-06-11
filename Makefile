#GOCMD=CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go
GOCMD=CGO_ENABLED=0 go
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

RM=/bin/rm -f

VERSION=$(shell git describe --tags --abbrev=0 --always)
DATE_VER=$(shell date "+%Y%m%d")
COMMIT=$(shell git rev-parse --short HEAD)

GOLDFLAGS="-X cloudiac/consts.VERSION=$(VERSION) -X cloudiac/consts.BUILD=$(COMMIT)"
GOBUILD=$(GOCMD) build -v -ldflags $(GOLDFLAGS)
GORUN=$(GOCMD) run -v -ldflags $(GOLDFLAGS)
PB_PROTOC=protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative

BUILD_DIR=$(PWD)/targets

.PHONY: all build portal runner run run-portal ru-runner clean package repos providers package-release

all: portal runner
build: portal runner

reset-build-dir:
	$(RM) -r $(BUILD_DIR)
	mkdir -p $(BUILD_DIR)/assets/

portal: reset-build-dir
	swag init -g web/api/v1/route.go
	$(GOBUILD) -o $(BUILD_DIR)/iac-portal ./cmds/portal
	cp ./configs/config-portal.yaml.sample $(BUILD_DIR)/config-portal.yaml.sample
	cp ./configs/dotenv.sample $(BUILD_DIR)/dotenv.sample

runner: reset-build-dir
	$(GOBUILD) -o $(BUILD_DIR)/ct-runner ./cmds/runner
	cp ./configs/config-runner.yaml.sample $(BUILD_DIR)/config-runner.yaml.sample

run: run-portal

run-portal:
	swag init -g web/api/v1/route.go
	$(GORUN) ./cmds/portal -v -c config-portal.yaml

run-runner:
	$(GORUN) ./cmds/runner -v -c config-runner.yaml

clean: reset-build-dir
	$(GOCLEAN) ./cmds/portal
	$(GOCLEAN) ./cmds/runner


PACKAGE_NAME=cloud-iac-$(VERSION).tar.gz
package-local: reset-build-dir clean build
	cp -a ./assets/terraform.py $(BUILD_DIR)/assets/
	@cd $(BUILD_DIR) && tar -czf ../cloud-iac-$(VERSION).tar.gz ./ && echo "Package: $(PACKAGE_NAME)"

package-linux-amd64:
	GOOS=linux GOARCH=amd64 $(MAKE) package-local

package: package-linux-amd64


repos: repos.list
	mkdir -p ./repos/cloud-iac && \
	cd ./repos/cloud-iac && bash ../../scripts/clone-repos.sh

REPOS_PACKAGE_NAME=cloud-iac-repos-$(VERSION)-$(DATE_VER).tar.gz
#repos-package: repos
repos-package:
	@tar -czf $(REPOS_PACKAGE_NAME) ./repos && echo Package: $(REPOS_PACKAGE_NAME)


providers: 
	bash scripts/generate-providers-mirror.sh

PROVIDERS_PACKAGE_NAME=cloud-iac-providers-$(VERSION)-$(DATE_VER).tar.gz
providers-package: providers
	@tar -czf $(PROVIDERS_PACKAGE_NAME) ./assets/providers && echo Package: $(PROVIDERS_PACKAGE_NAME)

