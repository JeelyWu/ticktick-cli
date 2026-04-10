BINARY ?= tick

.PHONY: build test release

build:
	mkdir -p bin
	go build -o bin/$(BINARY) ./cmd/tick

test:
	go test ./...

release:
	mkdir -p dist
	GOOS=darwin GOARCH=arm64 go build -o dist/tick-darwin-arm64 ./cmd/tick
	GOOS=darwin GOARCH=amd64 go build -o dist/tick-darwin-amd64 ./cmd/tick
	GOOS=linux GOARCH=arm64 go build -o dist/tick-linux-arm64 ./cmd/tick
	GOOS=linux GOARCH=amd64 go build -o dist/tick-linux-amd64 ./cmd/tick
