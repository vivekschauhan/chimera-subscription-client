.PHONY: all dep test lint build 

WORKSPACE ?= $$(pwd)

GO_PKG_LIST := $(shell go list ./... | grep -v /vendor/)

lint:
	@golint -set_exit_status ${GO_PKG_LIST}

dep:
	@echo "Resolving go package dependencies"
	@go mod tidy
	@go mod vendor
	@echo "Package dependencies completed"


${WORKSPACE}/chimera-subscription-client:
	@export GOARCH=amd64 && \
	go build -tags static_all \
		-a -o ${WORKSPACE}/bin/chimera-subscription-client ${WORKSPACE}/main.go

build:${WORKSPACE}/chimera-subscription-client
	@echo "Build complete"