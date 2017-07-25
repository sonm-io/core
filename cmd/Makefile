VER = v0.1
BUILD = $(shell git rev-parse --short HEAD)
FULL_VER = $(VER).$(BUILD)

build:
	go build -tags nocgo -ldflags "-s -X main.version=$(FULL_VER)" -a -o bootnode_osx bootnode.go

run: build
	./bootnode_osx

linux:
	@echo "Building bootnode $(FULL_VER)"
	GOOS=linux GOARCH=amd64 go build -tags nocgo -ldflags "-s -X main.version=$(FULL_VER)" -o bootnode_linux64 bootnode.go