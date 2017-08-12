#!/usr/bin/env make
VER = v0.1.0
BUILD = $(shell git rev-parse --short HEAD)
FULL_VER = $(VER).$(BUILD)

GOCMD=./cmd
GO=go
INSTALLDIR=${GOPATH}/bin

BOOTNODE=sonmbootnode
MINER=sonmminer
HUB=sonmhub
CLI=sonmcli

DOCKER_IMAGE_HUB="sonm/hub:latest"
DOCKER_IMAGE_MINER="sonm/miner:latest"
DOCKER_IMAGE_BOOTNODE="sonm/bootnode:latest"


.PHONY: fmt vet test

all: mock vet fmt build test install

build_bootnode:
	@echo "+ $@"
	${GO} build -tags nocgo -ldflags "-s -X main.version=$(FULL_VER)" -o ${BOOTNODE} ${GOCMD}/bootnode

build_miner:
	@echo "+ $@"
	${GO} build -tags nocgo -o ${MINER} ${GOCMD}/miner

build_hub:
	@echo "+ $@"
	${GO} build -tags nocgo -o ${HUB} ${GOCMD}/hub

build_cli:
	@echo "+ $@"
	${GO} build -tags nocgo -ldflags "-s -X github.com/sonm-io/core/cmd/cli/commands.version=$(FULL_VER)" -o ${CLI} ${GOCMD}/cli

build_blockchain:
	@echo "+ $@"
	$(MAKE) -C blockchain build

build: build_blockchain build_bootnode build_hub build_miner build_cli

install_bootnode: build_bootnode
	@echo "+ $@"
	cp ${BOOTNODE} ${INSTALLDIR}

install_miner: build_miner
	@echo "+ $@"
	cp ${MINER} ${INSTALLDIR}

install_hub: build_hub
	@echo "+ $@"
	cp ${HUB} ${INSTALLDIR}

install_cli: build_cli
	@echo "+ $@"
	cp ${CLI} ${INSTALLDIR}

install: install_bootnode install_miner install_hub install_cli

vet:
	@echo "+ $@"
	@go tool vet $(shell ls -1 -d */ | grep -v -e vendor -e contracts)

fmt:
	@echo "+ $@"
	@test -z "$$(gofmt -s -l . 2>&1 | grep -v ^vendor/ | tee /dev/stderr)" || \
		(echo >&2 "+ please format Go code with 'gofmt -s'" && false)

test: mock
	@echo "+ $@"
	@go test -tags nocgo $(shell go list ./... | grep -vE 'vendor|blockchain')
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

coverage:
	${GO} tool cover -func=coverage.txt
	${GO} tool cover -func=coverage.txt -o funccoverage.txt
	${GO} tool cover -html=coverage.txt -o coverage.html

clean:
	rm -f coverage.txt
	rm -f coverage.html
	rm -f funccoverage.txt
	rm -f ${MINER} ${HUB} ${CLI} ${BOOTNODE}
	$(MAKE) -C blockchain clean


docker_hub:
	docker build -t ${DOCKER_IMAGE_HUB} -f ./hub.Dockerfile .

docker_miner:
	docker build -t ${DOCKER_IMAGE_MINER} -f ./miner.Dockerfile .

docker_bootnode:
	docker build -t ${DOCKER_IMAGE_BOOTNODE} -f ./bootnode.Dockerfile .

docker_all: docker_hub docker_miner docker_bootnode
