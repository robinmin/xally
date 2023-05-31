GOCMD=go
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
BINARY_NAME=xally
VERSION?=0.2.1
SERVICE_PORT?=3000
EXPORT_RESULT?=false # for CI please set EXPORT_RESULT to true

GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

FILES=$(shell find . -name "*.go")

PACKAGE_NAME          := github.com/robinmin/xally
GOLANG_CROSS_VERSION  ?= v1.20.1

SYSROOT_DIR     ?= sysroots
SYSROOT_ARCHIVE ?= sysroots.tar.bz2


.PHONY: all test build vendor

all: fmt test

## Initialize:
init: ## Initialize project layout
	mkdir -p cmd		#### 本项目的主干
	mkdir -p configs    #### 配置文件模板或默认配置
	mkdir -p build     	#### 打包和持续集成
	mkdir -p logs     	#### 日志目录
	goreleaser init

## Build:
build: ## Build your project and put the output binary in build/bin/
	mkdir -p build/bin
## GO111MODULE=on CGO_ENABLED=on $(GOCMD) build -o build/bin/$(BINARY_NAME) ./cmd/client/main.go
	GO111MODULE=on $(GOCMD) build -o build/bin/$(BINARY_NAME) ./cmd/client/main.go
	chmod u+x build/bin/$(BINARY_NAME)

##	GO111MODULE=on CGO_ENABLED=on $(GOCMD) build -o build/bin/$(BINARY_NAME)_server ./cmd/server/main.go
	GO111MODULE=on $(GOCMD) build -o build/bin/$(BINARY_NAME)_server ./cmd/server/main.go
	chmod u+x build/bin/$(BINARY_NAME)_server

release-check: ## check before release
	goreleaser check --config .goreleaser-for-darwin.yaml
	goreleaser check --config .goreleaser-for-linux.yaml
	goreleaser check --config .goreleaser-for-wind.yaml
	goreleaser check --config .goreleaser-release.yaml

release: ## check before release
	goreleaser release --clean

dep: ## donwload dependencies packages
	go mod download


run: build ## run x-ally client
	build/bin/$(BINARY_NAME)

run-svr: build ## run x-ally server
	build/bin/$(BINARY_NAME)-server

clean: ## Remove build related file
	go clean
	rm -fr build/bin
	rm -f ./junit-report.xml checkstyle-report.xml ./coverage.xml ./profile.cov yamllint-checkstyle.xml


## Test:
test: ## Run the tests of the project
	$(info ******************** running tests ********************)
	go test -v ./...

coverage: ## Run the tests of the project and export the coverage
	go test ./... -coverprofile=coverage.out

fmt: ## Format all code
	$(info ******************** checking formatting ********************)
	@test -z $(shell gofmt -l $(FILES)) || (gofmt -d $(FILES); exit 1)

check: ## run precke
	$(info ******************** checking before commit ********************)
	pre-commit run
	goreleaser --snapshot --skip-publish --clean

## Lint:
lint:  ## Run all available linters
	$(info ******************** running lint tools ********************)
	errcheck -ignoretests ./cmd/client/main.go
	go vet ./cmd/client/main.go
	golangci-lint run -v ./cmd/client/main.go

	errcheck -ignoretests ./cmd/server/main.go
	go vet ./cmd/server/main.go
	golangci-lint run -v ./cmd/server/main.go

# codegen: ## generate source code from protobuf
# 	# protoc \
# 	# 	-I proto \
# 	# 	-I vendor/protoc-gen-validate \
# 	# 	--go_out=. \
# 	# 	--go_opt=paths=source_relative \
# 	# 	--go-grpc_out=. \
# 	# 	--go-grpc_opt=paths=source_relative \
# 	# 	$(find proto -name '*.proto')
# 	buf generate

## show all help information
help: ## Show this help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "${YELLOW}%-16s${GREEN}%s${RESET}\n", $$1, $$2}' $(MAKEFILE_LIST)




# .PHONY: sysroot-pack
# sysroot-pack:
# 	@tar cf - $(SYSROOT_DIR) -P | pv -s $[$(du -sk $(SYSROOT_DIR) | awk '{print $1}') * 1024] | pbzip2 > $(SYSROOT_ARCHIVE)

# .PHONY: sysroot-unpack
# sysroot-unpack:
# 	@pv $(SYSROOT_ARCHIVE) | pbzip2 -cd | tar -xf -

# .PHONY: release-dry-run
# release-dry-run:
# 	@docker run \
# 		--rm \
# 		-e CGO_ENABLED=1 \
# 		-v /var/run/docker.sock:/var/run/docker.sock \
# 		-v `pwd`:/go/src/$(PACKAGE_NAME) \
# 		-v `pwd`/sysroot:/sysroot \
# 		-w /go/src/$(PACKAGE_NAME) \
# 		ghcr.io/goreleaser/goreleaser-cross:${GOLANG_CROSS_VERSION} \
# 		--rm-dist --skip-validate --skip-publish

# .PHONY: release
# release:
# 	@if [ ! -f ".release-env" ]; then \
# 		echo "\033[91m.release-env is required for release\033[0m";\
# 		exit 1;\
# 	fi
# 	docker run \
# 		--rm \
# 		-e CGO_ENABLED=1 \
# 		--env-file .release-env \
# 		-v /var/run/docker.sock:/var/run/docker.sock \
# 		-v `pwd`:/go/src/$(PACKAGE_NAME) \
# 		-v `pwd`/sysroot:/sysroot \
# 		-w /go/src/$(PACKAGE_NAME) \
# 		ghcr.io/goreleaser/goreleaser-cross:${GOLANG_CROSS_VERSION} \
# 		release --rm-dist
