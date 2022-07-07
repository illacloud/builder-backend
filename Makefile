.PHONY: build all test clean

all: build build-websocket-server

build:
	go build -o bin/illa-backend main.go

build-websocket-server:
	go build -o bin/illa-websocker-server cmd/websocket-server/main.go

test:
	PROJECT_PWD=$(shell pwd) go test -race ./...

cov:
	PROJECT_PWD=$(shell pwd) go test -coverprofile cover.out ./...
	go tool cover -html=cover.out -o cover.html

fmt:
	@gofmt -w $(shell find . -type f -name '*.go' -not -path './*_test.go')

fmt-check:
	@gofmt -l $(shell find . -type f -name '*.go' -not -path './*_test.go')

clean:
	@ro -fR bin
