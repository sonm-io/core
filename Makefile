#!/usr/bin/env make


GOCMD=./cmd
GO=go

.PHONY: fmt vet test


MINER=sonmminer
HUB=sonmhub
CLI=sonmcli

all: vet fmt test build

build:
	@echo "+ $@"
	${GO} build -tags nocgo -o ${MINER} ${GOCMD}/miner
	${GO} build -tags nocgo -o ${HUB} ${GOCMD}/hub
	${GO} build -o ${CLI} ${GOCMD}/cli

install: build
	cp ${MINER} /usr/bin/
	cp ${HUB} /usr/bin/
	cp ${CLI} /usr/bin/


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
	protoc -I proto/hub/ proto/hub/hub.proto --go_out=plugins=grpc:proto/hub/
	protoc -I proto/miner/ proto/miner/miner.proto --go_out=plugins=grpc:proto/miner/

coverage:
	${GO} tool cover -func=coverage.txt
	${GO} tool cover -func=coverage.txt -o funccoverage.txt
	${GO} tool cover -html=coverage.txt -o coverage.html

clean:
	rm coverage.txt || true
	rm coverage.html || true
	rm funccoverage.txt || true
	rm ${MINER} ${HUB} ${CLI}
