BINARY ?= tick

.PHONY: build test

build:
	mkdir -p bin
	go build -o bin/$(BINARY) ./cmd/tick

test:
	go test ./...
