#!/usr/bin/env make
VER = v0.2.1.1
BUILD = $(shell git rev-parse --short HEAD)
FULL_VER = $(VER)-$(BUILD)

GOCMD=./cmd
ifeq ($(GO), )
    GO=go
endif
INSTALLDIR=${GOPATH}/bin

BOOTNODE=sonmbootnode
MINER=sonmminer
HUB=sonmhub
CLI=sonmcli
LOCATOR=sonmlocator
MARKET=sonmmarketplace


TAGS=nocgo

GPU_SUPPORT?=false
ifeq ($(GPU_SUPPORT),true)
    TAGS+=cl
endif

.PHONY: fmt vet test

all: mock vet fmt build test install

build/locator:
	@echo "+ $@"
	${GO} build -tags "$(TAGS)" -ldflags "-s -X main.version=$(FULL_VER)" -o ${LOCATOR} ${GOCMD}/locator

build/bootnode:
	@echo "+ $@"
	${GO} build -tags "$(TAGS)" -ldflags "-s -X main.version=$(FULL_VER)" -o ${BOOTNODE} ${GOCMD}/bootnode

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

build/cli_win32:
	@echo "+ $@"
	GOOS=windows GOARCH=386 go build -tags nocgo -ldflags "-s -X github.com/sonm-io/core/cmd/cli/commands.version=$(FULL_VER).win32" -o ${CLI}_win32.exe ${GOCMD}/cli

build/blockchain:
	@echo "+ $@"
	$(MAKE) -C blockchain build_contract_wrappers

build/insomnia: build/hub build/miner build/cli

build/aux: build/locator build/marketplace

build: build/blockchain build/bootnode build/insomnia build/aux

install/bootnode: build/bootnode
	@echo "+ $@"
	cp ${BOOTNODE} ${INSTALLDIR}

install/miner: build/miner
	@echo "+ $@"
	cp ${MINER} ${INSTALLDIR}

install/hub: build/hub
	@echo "+ $@"
	cp ${HUB} ${INSTALLDIR}

install/cli: build/cli
	@echo "+ $@"
	cp ${CLI} ${INSTALLDIR}

install: install/bootnode install/miner install/hub install/cli

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
	$(MAKE) -C blockchain test


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
	mockgen -package config -destination cmd/cli/config/config_mock.go  -source cmd/cli/config/config.go
	mockgen -package commands -destination cmd/cli/commands/interactor_mock.go  -source cmd/cli/commands/interactor.go
	mockgen -package task_config -destination cmd/cli/task_config/config_mock.go  -source cmd/cli/task_config/config.go

clean:
	rm -f ${MINER} ${HUB} ${CLI} ${BOOTNODE} ${MARKET}

deb:
	debuild --no-lintian --preserve-env -uc -us -i -I
