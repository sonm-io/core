#!/usr/bin/env make
VER = v0.3.3
BUILD = $(shell git rev-parse --short HEAD)
FULL_VER = $(VER)-$(BUILD)

GOCMD=./cmd
ifeq ($(GO), )
    GO=go
endif

ifeq ($(GOPATH), )
    GOPATH=$(shell ls -d ~/go)
endif

INSTALLDIR=${GOPATH}/bin/


TARGETDIR=target
ARCH := $(shell uname -m)
OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
OS_ARCH := $(OS)_$(ARCH)

MINER=${TARGETDIR}/sonmworker_$(OS_ARCH)
HUB=${TARGETDIR}/sonmhub_$(OS_ARCH)
CLI=${TARGETDIR}/sonmcli_$(OS_ARCH)
LOCATOR=${TARGETDIR}/sonmlocator_$(OS_ARCH)
LOCAL_NODE=${TARGETDIR}/sonmnode_$(OS_ARCH)
AUTOCLI=${TARGETDIR}/autocli_$(OS_ARCH)
RENDEZVOUS=${TARGETDIR}/sonmrendezvous_$(OS_ARCH)
LSGPU=${TARGETDIR}/lsgpu_$(OS_ARCH)

TAGS=nocgo

GPU_SUPPORT?=false
ifeq ($(GPU_SUPPORT),true)
    GPU_TAGS=cl
    # required for build nvidia-docker libs with NVML included via cgo
    NV_CGO=vendor/github.com/sshaman1101/nvidia-docker/build
    CGO_LDFLAGS=-L$(shell pwd)/${NV_CGO}/lib
    CGO_CFLAGS=-I$(shell pwd)/${NV_CGO}/include
    CGO_LDFLAGS_ALLOW='-Wl,--unresolved-symbols=ignore-in-object-files'
endif


ifeq ($(OS),linux)
SED=sed -i 's/github\.com\/sonm-io\/core\/vendor\///g' insonmnia/node/hub_mock.go
endif

ifeq ($(OS),darwin)
SED=sed -i "" 's/github\.com\/sonm-io\/core\/vendor\///g' insonmnia/node/hub_mock.go
endif

LDFLAGS = -X main.appVersion=$(FULL_VER)

.PHONY: fmt vet test

all: mock vet fmt build test

build/locator:
	@echo "+ $@"
	${GO} build -tags "$(TAGS)" -ldflags "-s $(LDFLAGS)" -o ${LOCATOR} ${GOCMD}/locator

build/miner:
	@echo "+ $@"
	CGO_LDFLAGS_ALLOW=${CGO_LDFLAGS_ALLOW} CGO_LDFLAGS=${CGO_LDFLAGS} CGO_CFLAGS=${CGO_CFLAGS} ${GO} build -tags "$(TAGS) $(GPU_TAGS)" -ldflags "-s $(LDFLAGS)" -o ${MINER} ${GOCMD}/miner

build/hub:
	@echo "+ $@"
	${GO} build -tags "$(TAGS)" -ldflags "-s $(LDFLAGS)" -o ${HUB} ${GOCMD}/hub

build/rv:
	@echo "+ $@"
	${GO} build -tags "$(TAGS)" -o ${RENDEZVOUS} ${GOCMD}/rv

build/rendezvous: build/rv

build/cli:
	@echo "+ $@"
	${GO} build -tags "$(TAGS)" -ldflags "-s $(LDFLAGS)" -o ${CLI} ${GOCMD}/cli

build/node:
	@echo "+ $@"
	${GO} build -tags "$(TAGS)" -ldflags "-s $(LDFLAGS)" -o ${LOCAL_NODE} ${GOCMD}/node

build/lsgpu:
	@echo "+ $@"
	${GO} build -tags "$(TAGS)" -ldflags "-s $(LDFLAGS)" -o ${LSGPU} ${GOCMD}/lsgpu

build/cli_win32:
	@echo "+ $@"
	GOOS=windows GOARCH=386 ${GO} build -tags "$(TAGS)" -ldflags "-s $(LDFLAGS).win32" -o ${CLI}_win32.exe ${GOCMD}/cli

build/node_win32:
	@echo "+ $@"
	GOOS=windows GOARCH=386 ${GO} build -tags "$(TAGS)" -ldflags "-s $(LDFLAGS).win32" -o ${LOCAL_NODE}_win32.exe ${GOCMD}/node

build/autocli:
	@echo "+ $@"
	${GO} build -tags "$(TAGS)" -ldflags "-s $(LDFLAGS)" -o ${AUTOCLI} ${GOCMD}/autocli


build/insomnia: build/hub build/miner build/cli build/node build/rv

build/aux: build/locator

build: build/insomnia build/aux

install: all
	@echo "+ $@"
	mkdir -p ${INSTALLDIR}
	cp ${MINER} ${HUB} ${CLI} ${LOCATOR} ${LOCAL_NODE} ${INSTALLDIR}

vet:
	@echo "+ $@"
	@go tool vet $(shell ls -1 -d */ | grep -v -e vendor -e contracts)

fmt:
	@echo "+ $@"
	@test -z "$$(gofmt -s -l . 2>&1 | grep -v ^vendor/ | tee /dev/stderr)" || \
		(echo >&2 "+ please format Go code with 'gofmt -s'" && false)

test: mock
	@echo "+ $@"
	${GO} test -tags nocgo $(shell go list ./... | grep -vE 'vendor|blockchain')

grpc:
	@echo "+ $@"
	@if ! which protoc > /dev/null; then echo "protoc protobuf compiler required for build"; exit 1; fi;
	@if ! which protoc-gen-grpccmd > /dev/null; then echo "protoc-gen-grpccmd protobuf  plugin required for build.\nRun \`go get -u github.com/sshaman1101/grpccmd/cmd/protoc-gen-grpccmd\`"; exit 1; fi;
	@protoc -I proto proto/*.proto --grpccmd_out=proto/

mock:
	@echo "+ $@"
	@if ! which mockgen > /dev/null; then \
	echo "mockgen is required."; \
	echo "Run \`go get github.com/golang/mock/gomock\`"; \
	echo "\`go get github.com/golang/mock/mockgen\`"; \
	echo "and add your go bin directory to PATH"; exit 1; fi;
	mockgen -package miner -destination insonmnia/miner/overseer_mock.go -source insonmnia/miner/overseer.go
	mockgen -package miner -destination insonmnia/miner/config_mock.go -source insonmnia/miner/config.go \
		-aux_files logging=insonmnia/logging/logging.go
	mockgen -package hardware -destination insonmnia/hardware/hardware_mock.go -source insonmnia/hardware/hardware.go
	mockgen -package commands -destination cmd/cli/commands/interactor_mock.go  -source cmd/cli/commands/interactor.go
	mockgen -package task_config -destination cmd/cli/task_config/config_mock.go  -source cmd/cli/task_config/config.go
	mockgen -package accounts -destination accounts/keys_mock.go  -source accounts/keys.go
	mockgen -package blockchain -destination blockchain/api_mock.go  -source blockchain/api.go
	mockgen -package sonm -destination proto/locator_mock.go  -source proto/locator.pb.go
	mockgen -package sonm -destination proto/marketplace_mock.go  -source proto/marketplace.pb.go
	mockgen -package hub -destination insonmnia/hub/cluster_mock.go  -source insonmnia/hub/cluster.go
	mockgen -package config -destination cmd/cli/config/config_mock.go  -source cmd/cli/config/config.go \
		-aux_files accounts=accounts/keys.go,logging=insonmnia/logging/logging.go
	mockgen -package node -destination insonmnia/node/config_mock.go -source insonmnia/node/config.go \
		-aux_files accounts=accounts/keys.go,logging=insonmnia/logging/logging.go
	mockgen -imports "context=golang.org/x/net/context" -package node -destination insonmnia/node/hub_mock.go \
		"github.com/sonm-io/core/proto" HubClient && ${SED}

clean:
	rm -f ${MINER} ${HUB} ${CLI} ${LOCATOR} ${LOCAL_NODE} ${AUTOCLI} ${RENDEZVOUS}

deb:
	debuild --no-lintian --preserve-env -uc -us -i -I -b
	debuild clean

coverage:
	.ci/coverage.sh

release:
	@echo "preparing $(VER) to release..."
	@ .ci/release.sh $(VER) $(GH_TOKEN)
