#!/usr/bin/env make

GO=go
TRUFFLE=./node_modules/truffle/build/cli.bundled.js
TESTRPC=./node_modules/ethereumjs-testrpc/build/cli.node.js


SOL_SOURCES=$(wildcard contracts/*.sol)
GO_BUILDS=$(SOL_SOURCES:contracts/%.sol=api/%.go)


.PHONY: build test clean

node_modules:
	npm install

build/contracts/%.json: contracts/%.sol node_modules
	${TRUFFLE} compile

api/%.go: build/contracts/%.json
	${GO} generate ./api.go

build_contract_wrappers: $(GO_BUILDS)
	@echo "+ $@"
	${GO} build .

test: $(GO_BUILDS) node_modules
	@echo "+ $@"
	TESTRPC="$(shell pwd)/${TESTRPC}" ${GO} test ./tests

clean:
	rm -rf api
	rm -rf build
