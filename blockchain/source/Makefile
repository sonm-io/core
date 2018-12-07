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

migrate_dev:
	${TRUFFLE} migrate --network=dev_main
	${TRUFFLE} migrate --network=dev_side

check_keys:
ifeq (${PRV_KEY_MAIN},)
	@echo "ERROR: PRV_KEY_MAIN is required for migration"
endif
ifeq (${PRV_KEY_SIDE},)
	@echo "ERROR: PRV_KEY_SIDE is required for migration"
endif
ifeq (${PRV_KEY_MAIN},)
	@exit 1
endif
ifeq (${PRV_KEY_SIDE},)
	@exit 1
endif

migrate_test: check_keys
	MIGRATION=true PRV_KEY=${PRV_KEY_MAIN} ${TRUFFLE} migrate --network=rinkeby
	MIGRATION=true PRV_KEY=${PRV_KEY_SIDE} ${TRUFFLE} migrate --network=private

migrate_live: check_keys
	MIGRATION=true PRV_KEY=${PRV_KEY_MAIN} ${TRUFFLE} migrate --network=master
	MIGRATION=true PRV_KEY=${PRV_KEY_SIDE} ${TRUFFLE} migrate --network=privateLive

migrate:
	MIGRATION=true ${TRUFFLE} migrate --network=$(NETWORK)

coverage:
	scripts/test_coverage.sh

deploy:
	${TRUFFLE} migrate --network private

generate_api: compile
	${GO} build utils/generate_api.go
	./generate_api
	rm generate_api
