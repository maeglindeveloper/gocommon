.PHONY: build

PROJECT = bunnybank
APP_NAME = service-common

PATH := $(GOPATH)/bin:$(PATH)
VERSION = $(shell git describe --tags --always --dirty)
BRANCH = $(shell git rev-parse --abbrev-ref HEAD)
REVISION = $(shell git rev-parse HEAD)
REVSHORT = $(shell git rev-parse --short HEAD)
USER = $(shell whoami)
DOCKER_IMAGE_NAME = $(PROJECT)/$(APP_NAME)
DOCKER_IMAGE_TAG = ${REVSHORT}

ifneq ($(OS), Windows_NT)
	BUILD_GOOS = linux

	# If on macOS, set the shell to bash explicitly
	ifeq ($(shell uname), Darwin)
		BUILD_GOOS = darwin
		
		SHELL := /bin/bash
	endif

	# The output binary name is different on Windows, so we're explicit here
	OUTPUT = main

	# To populate version metadata, we use unix tools to get certain data
	GOVERSION = $(shell go version | awk '{print $$3}')
	NOW	= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
else
	BUILD_GOOS = windows

	# The output binary name is different on Windows, so we're explicit here
	OUTPUT = main.exe

	# To populate version metadata, we use windows tools to get the certain data
	GOVERSION_CMD = "(go version).Split()[2]"
	GOVERSION = $(shell powershell $(GOVERSION_CMD))
	NOW	= $(shell powershell Get-Date -format s)
endif

KIT_VERSION = "\
	-X ${DOCKER_IMAGE_NAME}/version.appName=${APP_NAME} \
	-X ${DOCKER_IMAGE_NAME}/version.version=${VERSION} \
	-X ${DOCKER_IMAGE_NAME}/version.branch=${BRANCH} \
	-X ${DOCKER_IMAGE_NAME}/version.revision=${REVISION} \
	-X ${DOCKER_IMAGE_NAME}/version.buildDate=${NOW} \
	-X ${DOCKER_IMAGE_NAME}/version.buildUser=${USER} \
	-X ${DOCKER_IMAGE_NAME}/version.goVersion=${GOVERSION}"

all: build

define HELP_TEXT

  Makefile commands

	make deps         - Install dependent programs and libraries
	make generate     - Generate and bundle required code
	make generate-dev - Generate and bundle required code in a watch loop
	make distclean    - Delete all build artifacts

	make build        - Build the code
	make package 	  - Build rpm and deb packages for linux

	make test         - Run the full test suite
	make coverage     - Run test coverage
	make lint         - Run all linters

endef

help:
	$(info $(HELP_TEXT))

.prefix:
ifeq ($(OS), Windows_NT)
	if not exist build mkdir build
else
	mkdir build
endif

.pre-build:
	$(eval GOGC = off)
	$(eval CGO_ENABLED = 0)
	$(eval GOOS = $(BUILD_GOOS))

.pre-run:
	$(eval GOOS = $(BUILD_GOOS))

build: .prefix .pre-build
	go build -i -o build/${OUTPUT} -ldflags ${KIT_VERSION} main.go

run: .pre-run
	go run -ldflags ${KIT_VERSION} main.go

lint:
	go vet ./...

test: lint
	go test ./...

coverage:
	go test -race -cover ./...

protobuf:
	protoc -I pb service.proto --go_out=plugins=grpc:pb

generate: protobuf

deps:
	go mod -sync

vendor:
	go mod -vendor

distclean:
ifeq ($(OS), Windows_NT)
	if exist build rmdir /s/q build
	if exist vendor rmdir /s/q vendor
else
	rm -rf build vendor
endif

docker-build:
	docker build -t "${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}" .
	docker tag "${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}" "${DOCKER_IMAGE_NAME}:latest"
