#!/usr/bin/env make

REPO=github.com/sonm-io/insonmnia
NAME=sonmd

BUILDDT=$(shell date -u +%F@%H:%M:%S)
VERSION=$(shell git show-ref --head --hash head)
TAG=$(shell git describe --tags --always)
DEBVER=$(shell dpkg-parsechangelog | sed -n -e 's/^Version: //p')
VERSIONPKG=${REPO}/version
DEPSPKG=${REPO}/deps

# encode godeps into base64 to pass to a linker
ifeq ($(shell uname),Linux)
	BASE64FLAGS=-w0
endif
GODEPS=$(shell base64 ${BASE64FLAGS} ./Godeps/Godeps.json)

GOMAIN=${REPO}/cmd
GO=go
PKGS := $(shell ${GO} list ./... | grep -v ^${REPO}/vendor/ | grep -v ^${REPO}/version)

LDFLAGS=-ldflags "-X ${VERSIONPKG}.GitTag=${TAG} -X ${VERSIONPKG}.Version=${DEBVER} -X ${VERSIONPKG}.Build=${BUILDDT} -X ${VERSIONPKG}.GitHash=${VERSION} -X ${DEPSPKG}.godeps=${GODEPS}"


.PHONY: fmt vet test


binaries:
	@echo "+ $@"
	${GO} build ${LDFLAGS} -o ${NAME} ${GOMAIN}

all: vet fmt test buildbinary

vet:
	@echo "+ $@"
	@go vet $(PKGS)

fmt:
	@echo "+ $@"
	@test -z "$$(gofmt -s -l . 2>&1 | grep -v ^vendor/ | tee /dev/stderr)" || \
		(echo >&2 "+ please format Go code with 'gofmt -s'" && false)

test:
	@echo "+ $@"
	@echo > coverage.txt
	@set -e; for pkg in $(PKGS); do ${GO} test -coverprofile=profile.out $$pkg; \
	if [ -f profile.out ]; then \
		cat profile.out >> coverage.txt; rm  profile.out; \
	fi done;
	@sed -ie '2!s/mode: set//;/^$$/d' coverage.txt

coverage:
	${GO} tool cover -func=coverage.txt
	${GO} tool cover -func=coverage.txt -o funccoverage.txt
	${GO} tool cover -html=coverage.txt -o coverage.html

clean:
	rm sonmd || true
	rm coverage.txt || true
	rm coverage.html || true
	rm funccoverage.txt || true
