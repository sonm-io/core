#!/usr/bin/env make

GO=go
TRUFFLE=./node_modules/truffle/build/cli.bundled.js
TESTRPC=./node_modules/ethereumjs-testrpc/build/cli.node.js

.PHONY: build test clean

build:
	@echo "+ $@"
	npm install
	${TRUFFLE} compile
	${GO} generate ./api.go
	${GO} build .

test:
	@echo "+ $@"
	TESTRPC="$(shell pwd)/${TESTRPC}" ${GO} test ./tests

clean:
	rm -rf api
	rm -rf build
