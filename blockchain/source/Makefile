#!/usr/bin/env make
ifeq ($(GO), )
    GO=go
endif
TRUFFLE=./node_modules/truffle/build/cli.bundled.js
TESTRPC=./node_modules/ethereumjs-testrpc/build/cli.node.js
SOLIUM=./node_modules/solium/bin/solium.js
ESLINT=./node_modules/.bin/eslint
DEV_KEY=4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d

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

check_keys:
ifeq (${PRV_KEY_MAIN},)
	@echo "ERROR: PRV_KEY_MAIN is required for migration"
endif
ifeq (${PRV_KEY_SIDE},)
	@echo "ERROR: PRV_KEY_SIDE is required for migration"
endif
ifeq (${MS_PRV_KEY},)
	@echo "ERROR: MS_PRV_KEY is required for migration"
endif

ifeq (${PRV_KEY_MAIN},)
	@exit 1
endif
ifeq (${PRV_KEY_SIDE},)
	@exit 1
endif
ifeq (${MS_PRV_KEY},)
	@exit 1
endif


migrate_dev:
	PRV_KEY=${DEV_KEY} MS_PRV_KEY=${DEV_KEY} ${TRUFFLE} migrate --network=dev_side
	PRV_KEY=${DEV_KEY} MS_PRV_KEY=${DEV_KEY} ${TRUFFLE} migrate --network=dev_main

migrate_test: check_keys
	MIGRATION=true PRV_KEY=${PRV_KEY_SIDE} ${TRUFFLE} migrate --network=private
	MIGRATION=true PRV_KEY=${PRV_KEY_MAIN} ${TRUFFLE} migrate --network=rinkeby

migrate_live: check_keys
	MIGRATION=true PRV_KEY=${PRV_KEY_SIDE} ${TRUFFLE} migrate --network=privateLive
	MIGRATION=true PRV_KEY=${PRV_KEY_MAIN} ${TRUFFLE} migrate --network=master

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
