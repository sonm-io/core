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

migrate_test:
ifeq (${PRV_KEY},)
	@echo "ERROR: PRV_KEY is required for livenet migration"
	@exit 1
endif
	MIGRATION=true ${TRUFFLE} migrate --network=rinkeby
	MIGRATION=true ${TRUFFLE} migrate --network=private

migrate_live:
ifeq (${PRV_KEY},)
	@echo "ERROR: PRV_KEY is required for livenet migration"
	@exit 1
endif
	MIGRATION=true ${TRUFFLE} migrate --network=master
	MIGRATION=true ${TRUFFLE} migrate --network=privateLive

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
