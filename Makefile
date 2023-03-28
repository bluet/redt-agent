.PHONY: all build test lint install-tools

all:	test build

build:	build-amd64 build-arm64

build-amd64:	lint install-tools
	GOOS=linux GOARCH=amd64 go build -o bin/redt-agent-linux-amd64 ./cmd/redt-agent

build-arm64:	lint install-tools
	GOOS=linux GOARCH=arm64 go build -o bin/redt-agent-linux-arm64 ./cmd/redt-agent

test:
	go test -v ./...

lint:
	golangci-lint run

install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
