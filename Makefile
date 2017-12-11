#!/usr/bin/env make
VER = v0.2.1.1
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

MINER=sonmminer
HUB=sonmhub
CLI=sonmcli
LOCATOR=sonmlocator
MARKET=sonmmarketplace
LOCAL_NODE=sonmnode

TAGS=nocgo

GPU_SUPPORT?=false
ifeq ($(GPU_SUPPORT),true)
    TAGS+=cl
endif

UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
SED=sed -i 's/github\.com\/sonm-io\/core\/vendor\///g' insonmnia/node/hub_mock.go
endif

ifeq ($(UNAME_S),Darwin)
SED=sed -i "" 's/github\.com\/sonm-io\/core\/vendor\///g' insonmnia/node/hub_mock.go
endif

.PHONY: fmt vet test

all: mock vet fmt build test

build/locator:
	@echo "+ $@"
	${GO} build -tags "$(TAGS)" -ldflags "-s -X main.version=$(FULL_VER)" -o ${LOCATOR} ${GOCMD}/locator

build/miner:
	@echo "+ $@"
	${GO} build -tags "$(TAGS)" -ldflags "-s -X main.version=$(FULL_VER)" -o ${MINER} ${GOCMD}/miner

build/hub:
	@echo "+ $@"
	${GO} build -tags "$(TAGS)" -ldflags "-s -X main.version=$(FULL_VER)" -o ${HUB} ${GOCMD}/hub

build/marketplace:
	@echo "+ $@"
	${GO} build -tags "$(TAGS)" -ldflags "-s -X main.version=$(FULL_VER)" -o ${MARKET} ${GOCMD}/marketplace

build/cli:
	@echo "+ $@"
	${GO} build -tags "$(TAGS)" -ldflags "-s -X github.com/sonm-io/core/cmd/cli/commands.version=$(FULL_VER)" -o ${CLI} ${GOCMD}/cli

build/node:
	@echo "+ $@"
	${GO} build -tags "$(TAGS)" -ldflags "-s -X main.version=$(FULL_VER)" -o ${LOCAL_NODE} ${GOCMD}/node

build/cli_win32:
	@echo "+ $@"
	GOOS=windows GOARCH=386 ${GO} build -tags "$(TAGS)" -ldflags "-s -X github.com/sonm-io/core/cmd/cli/commands.version=$(FULL_VER).win32" -o ${CLI}_win32.exe ${GOCMD}/cli

build/node_win32:
	@echo "+ $@"
	GOOS=windows GOARCH=386 ${GO} build -tags "$(TAGS)" -ldflags "-s -X main.version=$(FULL_VER).win32" -o ${LOCAL_NODE}_win32.exe ${GOCMD}/node


build/insomnia: build/hub build/miner build/cli build/node

build/aux: build/locator build/marketplace

build: build/insomnia build/aux

install: all
	@echo "+ $@"
	mkdir -p ${INSTALLDIR}
	cp ${MINER} ${HUB} ${CLI} ${LOCATOR} ${MARKET} ${LOCAL_NODE} ${INSTALLDIR}

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
	@if ! which protoc-gen-go > /dev/null; then echo "protoc-gen-go protobuf  plugin required for build.\nRun \`go get -u github.com/golang/protobuf/protoc-gen-go\`"; exit 1; fi;
	@protoc -I proto proto/*.proto --go_out=plugins=grpc:proto/

mock:
	@echo "+ $@"
	@if ! which mockgen > /dev/null; then \
	echo "mockgen is required."; \
	echo "Run \`go get github.com/golang/mock/gomock\`"; \
	echo "\`go get github.com/golang/mock/mockgen\`"; \
	echo "and add your go bin directory to PATH"; exit 1; fi;
	mockgen -package miner -destination insonmnia/miner/overseer_mock.go -source insonmnia/miner/overseer.go
	mockgen -package miner -destination insonmnia/miner/config_mock.go -source insonmnia/miner/config.go
	mockgen -package hardware -destination insonmnia/hardware/hardware_mock.go -source insonmnia/hardware/hardware.go
	mockgen -package commands -destination cmd/cli/commands/interactor_mock.go  -source cmd/cli/commands/interactor.go
	mockgen -package task_config -destination cmd/cli/task_config/config_mock.go  -source cmd/cli/task_config/config.go
	mockgen -package accounts -destination accounts/keys_mock.go  -source accounts/keys.go
	mockgen -package blockchain -destination blockchain/api_mock.go  -source blockchain/api.go
	mockgen -package sonm -destination proto/locator_mock.go  -source proto/locator.pb.go
	mockgen -package sonm -destination proto/marketplace_mock.go  -source proto/marketplace.pb.go
	mockgen -package hub -destination insonmnia/hub/cluster_mock.go  -source insonmnia/hub/cluster.go
	mockgen -package config -destination cmd/cli/config/config_mock.go  -source cmd/cli/config/config.go \
		-aux_files accounts=accounts/keys.go
	mockgen -package node -destination insonmnia/node/config_mock.go -source insonmnia/node/config.go \
		-aux_files accounts=accounts/keys.go
	mockgen -imports "context=golang.org/x/net/context" -package node -destination insonmnia/node/hub_mock.go \
		"github.com/sonm-io/core/proto" HubClient && ${SED}

clean:
	rm -f ${MINER} ${HUB} ${CLI} ${MARKET}

deb:
	debuild --no-lintian --preserve-env -uc -us -i -I
