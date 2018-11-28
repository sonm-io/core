#!/usr/bin/env make

# Version of the entire package. Do not forget to update this when it's time
# to bump the version.
VERSION = v0.4.17

# Build tag. Useful to distinguish between same-version builds, but from
# different commits.
BUILD = $(shell git rev-parse --short HEAD)

# Full version includes both semantic version and git ref.
FULL_VERSION = $(VERSION)-$(BUILD)

# NOTE: variables defined with := in GNU make are expanded when they are
# defined rather than when they are used.
GOCMD := ./cmd

# NOTE: variables defined with ?= sets the default value, which can be
# overriden using env.
GO ?= go
GOPATH ?= $(shell ls -d ~/go)

TARGETDIR := target
INSTALLDIR := ${GOPATH}/bin/

HOSTOS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
HOSTARCH := $(shell uname -m)

GOOS ?= ${HOSTOS}
GOARCH ?= ${HOSTARCH}

# Set the execution extension for Windows.
ifeq (${GOOS},windows)
    EXE := .exe
endif

OS_ARCH := $(GOOS)_$(GOARCH)$(EXE)

WORKER     := ${TARGETDIR}/sonmworker_$(OS_ARCH)
NODE       := ${TARGETDIR}/sonmnode_$(OS_ARCH)
CLI        := ${TARGETDIR}/sonmcli_$(OS_ARCH)
AUTOCLI    := ${TARGETDIR}/autocli_$(OS_ARCH)
DWH        := ${TARGETDIR}/sonmdwh_$(OS_ARCH)
RENDEZVOUS := ${TARGETDIR}/sonmrendezvous_$(OS_ARCH)
RELAY      := ${TARGETDIR}/sonmrelay_$(OS_ARCH)
OPTIMUS    := ${TARGETDIR}/sonmoptimus_$(OS_ARCH)
LSGPU      := ${TARGETDIR}/lsgpu_$(OS_ARCH)
PANDORA    := ${TARGETDIR}/pandora_$(OS_ARCH)
ORACLE     := ${TARGETDIR}/sonmoracle_$(OS_ARCH)
CONNOR     := ${TARGETDIR}/sonmconnor_$(OS_ARCH)
SONMMON    := ${TARGETDIR}/sonmmon_$(OS_ARCH)

TAGS = nocgo

GPU_SUPPORT ?= true
ifeq ($(GPU_SUPPORT),true)
    GPU_TAGS := cl
endif

# This can be set to "false" to enable cross-compilation on non-linux planforms for linux.
ifeq (${GOOS},linux)
    WITH_NL ?= true
else
    WITH_NL ?= false
endif

ifeq ($(WITH_NL),true)
    NL_TAGS := nl
endif

LDFLAGS = -X github.com/sonm-io/core/insonmnia/version.Version=$(FULL_VERSION)

.PHONY: fmt vet test

all: mock vet fmt build test

build/worker:
	@echo "+ $@"
    ifneq (${GOOS},linux)
        ifeq (${WITH_NL},true)
			@echo "ERROR: Building with netlink support on non-linux platforms is not allowed"
			@exit 1
        endif
    endif
	${GO} build -tags "$(TAGS) $(GPU_TAGS) ${NL_TAGS}" -ldflags "-s $(LDFLAGS)" -o ${WORKER} ${GOCMD}/worker

build/dwh:
	@echo "+ $@"
	${GO} build -tags "$(TAGS)" -ldflags "$(LDFLAGS)" -o ${DWH} ${GOCMD}/dwh

build/rv:
	@echo "+ $@"
	${GO} build -tags "$(TAGS)" -ldflags "$(LDFLAGS)" -o ${RENDEZVOUS} ${GOCMD}/rv

build/rendezvous: build/rv

build/relay:
	@echo "+ $@"
	${GO} build -tags "$(TAGS)" -ldflags "$(LDFLAGS)" -o ${RELAY} ${GOCMD}/relay

build/cli:
	@echo "+ $@"
	${GO} build -tags "$(TAGS)" -ldflags "$(LDFLAGS)" -o ${CLI} ${GOCMD}/cli

build/node:
	@echo "+ $@"
	${GO} build -tags "$(TAGS)" -ldflags "$(LDFLAGS)" -o ${NODE} ${GOCMD}/node

build/lsgpu:
	@echo "+ $@"
	${GO} build -tags "$(TAGS)" -ldflags "$(LDFLAGS)" -o ${LSGPU} ${GOCMD}/lsgpu

build/autocli:
	@echo "+ $@"
	@if ! which protoc-gen-grpccmd > /dev/null; then echo "protoc-gen-grpccmd protobuf plugin required for build.\nRun \`go get -u github.com/sshaman1101/grpccmd/cmd/protoc-gen-grpccmd\`"; exit 1; fi;
	@protoc -I proto proto/*.proto --grpccmd_out=cmd/autocli/proto/
	${GO} build -tags "$(TAGS)" -ldflags "$(LDFLAGS)" -o ${AUTOCLI} ${GOCMD}/autocli
	@rm -rf cmd/autocli/proto/*.pb.go

build/pandora:
	@echo "+ $@"
	${GO} build -tags "$(TAGS)" -ldflags "$(LDFLAGS)" -o ${PANDORA} ${GOCMD}/pandora

build/optimus:
	@echo "+ $@"
	${GO} build -tags "$(TAGS)" -ldflags "$(LDFLAGS)" -o ${OPTIMUS} ${GOCMD}/optimus

build/oracle:
	@echo "+ $@"
	${GO} build -tags "$(TAGS)" -ldflags "$(LDFLAGS)" -o ${ORACLE} ${GOCMD}/oracle

build/connor:
	@echo "+ $@"
	${GO} build -tags "$(TAGS)" -ldflags "$(LDFLAGS)" -o ${CONNOR} ${GOCMD}/connor

build/sonmmon:
ifeq ($(GOOS),linux)
	@echo "+ $@"
	${GO} build -tags "$(TAGS)" -ldflags "$(LDFLAGS)" -o ${SONMMON} ${GOCMD}/sonmmon
else
	@echo "Skipping build of sonmmon for non-linux target"
endif

build/insomnia: build/worker build/cli build/node

build/aux: build/relay build/rv build/dwh build/pandora build/optimus build/oracle build/connor build/sonmmon

build: build/insomnia build/aux

install: all
	@echo "+ $@"
	mkdir -p ${INSTALLDIR}
	cp ${WORKER} ${CLI} ${NODE} ${INSTALLDIR}

vet:
	@echo "+ $@"
	@go tool vet $(shell ls -1 -d */ | grep -v -e vendor -e contracts)

fmt:
	@echo "+ $@"
	@test -z "$$(gofmt -s -l . 2>&1 | grep -v ^vendor/ | tee /dev/stderr)" || \
		(echo >&2 "+ please format Go code with 'gofmt -s'" && false)

test: mock
	@echo "+ $@"
	${GO} test -tags nocgo $(shell go list ./... | grep -vE 'vendor')

# Everything except DWH tests.
test/lite: mock
	@echo "+ $@"
	${GO} test -tags nocgo $(shell go list ./... | grep -vE 'vendor|blockchain|dwh')

contracts:
	@$(MAKE) -C blockchain/source all

grpc:
	@echo "+ $@"
	@if ! which protoc > /dev/null; then echo "protoc protobuf compiler required for build"; exit 1; fi;
	@protoc -I proto proto/*.proto --go_out=plugins=grpc:proto/

build_mockgen:
	@go get github.com/golang/mock/mockgen@v1.0.0

mock: build_mockgen
	mockgen -package worker -destination insonmnia/worker/overseer_mock.go -source insonmnia/worker/overseer.go
	mockgen -package worker -destination insonmnia/worker/acl_mock.go -source insonmnia/worker/acl.go
	mockgen -package benchmarks -destination insonmnia/benchmarks/benchmarks_mock.go -source insonmnia/benchmarks/benchmarks.go
	mockgen -package benchmarks -destination insonmnia/benchmarks/mapping_mock.go -source insonmnia/benchmarks/mapping.go
	mockgen -package blockchain -destination blockchain/api_mock.go  -source blockchain/api.go
	mockgen -package sonm -destination proto/marketplace_mock.go  -source proto/marketplace.pb.go
	mockgen -package sonm -destination proto/dwh_mock.go  -source proto/dwh.pb.go
	mockgen -package node -destination insonmnia/node/server_mock.go -source insonmnia/node/server.go
	mockgen -package version -destination insonmnia/version/version_mock.go -source insonmnia/version/version.go

clean:
	rm -f ${WORKER} ${CLI} ${NODE} ${AUTOCLI} ${RENDEZVOUS}
	find . -name "*_mock.go" | xargs rm -f

deb:
	go mod download
	debuild --no-lintian --preserve-env -uc -us -i -I -b
	debuild clean

coverage:
	.ci/coverage.sh
