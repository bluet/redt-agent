.PHONY: build test lint

build:	lint install-tools
	GOOS=linux GOARCH=amd64 go build -o bin/redt-agent ./cmd/redt-agent

test:
	go test -v ./...

lint:
	golangci-lint run

install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
