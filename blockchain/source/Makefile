#!/usr/bin/env make
ifeq ($(GO), )
    GO=go
endif
TRUFFLE=./node_modules/truffle/build/cli.bundled.js
TESTRPC=./node_modules/ethereumjs-testrpc/build/cli.node.js
SOLIUM=./node_modules/solium/bin/solium.js
ESLINT=./node_modules/.bin/eslint

.PHONY: all test node_modules coverage api

all: node_modules lint compile test

compile:
	@echo "+ $@"
	@${TRUFFLE} compile

node_modules:
	@echo "+ $@"
	@npm ci

test:
	@echo "+ $@"
	@scripts/test.sh

lint: lint_sol lint_js

lint_sol:
	@echo "+ $@"
	@${SOLIUM} --dir ./contracts

lint_js:
	@echo "+ $@"
	@${ESLINT} .

migrate:
	${TRUFFLE} migrate --network=$(NETWORK)

coverage:
	scripts/test_coverage.sh

deploy:
	${TRUFFLE} migrate --network private

liveMasterDeploy:
	mv migrations tmp_dir
	mv live-deploy-scripts migrations
	${TRUFFLE} migrate --reset --network master
	mv migrations live-deploy-scripts
	mv tmp_dir migrations
	echo "DONT FORGET TO SET KEEPER ADDRESS"

liveSideDeploy:
	mv migrations tmp_dir
	mv live-deploy-scripts migrations
	${TRUFFLE} migrate --network privateLive
	mv migrations live-deploy-scripts
	mv tmp_dir migrations

generate_api: compile
	${GO} build utils/generate_api.go
	./generate_api
	rm generate_api
