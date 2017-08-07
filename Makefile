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

all: vet fmt test build install


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


build: build_bootnode build_hub build_miner build_cli

install_bootnode:
	@echo "+ $@"
	cp ${BOOTNODE} ${INSTALLDIR}

install_miner:
	@echo "+ $@"
	cp ${MINER} ${INSTALLDIR}

install_hub:
	@echo "+ $@"
	cp ${HUB} ${INSTALLDIR}

install_cli:
	@echo "+ $@"
	cp ${CLI} ${INSTALLDIR}

install: install_bootnode install_miner install_hub install_cli

vet:
	@echo "+ $@"
	@go vet $(PKGS)

fmt:
	@echo "+ $@"
	@test -z "$$(gofmt -s -l . 2>&1 | grep -v ^vendor/ | tee /dev/stderr)" || \
		(echo >&2 "+ please format Go code with 'gofmt -s'" && false)

test:
	@echo "+ $@"
	@go test -tags nocgo $(shell go list ./... | grep -v vendor)

grpc:
	protoc -I proto proto/hub/hub.proto --go_out=plugins=grpc,Mminer/miner.proto=github.com/sonm-io/core/proto/miner:proto/
	protoc -I proto proto/miner/miner.proto --go_out=plugins=grpc:proto/

coverage:
	${GO} tool cover -func=coverage.txt
	${GO} tool cover -func=coverage.txt -o funccoverage.txt
	${GO} tool cover -html=coverage.txt -o coverage.html

clean:
	rm -f coverage.txt
	rm -f coverage.html
	rm -f funccoverage.txt
	rm -f ${MINER} ${HUB} ${CLI} ${BOOTNODE}


docker_hub:
	docker build -t ${DOCKER_IMAGE_HUB} -f ./hub.Dockerfile .

docker_miner:
	docker build -t ${DOCKER_IMAGE_MINER} -f ./miner.Dockerfile .

docker_bootnode:
	docker build -t ${DOCKER_IMAGE_BOOTNODE} -f ./bootnode.Dockerfile .

docker_all: docker_hub docker_miner docker_bootnode
