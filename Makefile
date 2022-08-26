.PHONY: build all test clean

all: build build-websocket-server

build:
	go build -o bin/illa-backend cmd/http-server/main.go cmd/http-server/wire_gen.go cmd/http-server/server.go

build-websocket-server:
	go build -o bin/illa-backend-ws cmd/websocket-server/main.go

test:
	PROJECT_PWD=$(shell pwd) go test -race ./...

test-cover:
	go test -cover --count=1 ./...
	
cov:
	PROJECT_PWD=$(shell pwd) go test -coverprofile cover.out ./...
	go tool cover -html=cover.out -o cover.html

fmt:
	@gofmt -w $(shell find . -type f -name '*.go' -not -path './*_test.go')

fmt-check:
	@gofmt -l $(shell find . -type f -name '*.go' -not -path './*_test.go')

init-database:
	/bin/bash scripts/postgres-init.sh

clean:
	@ro -fR bin
