BINARY ?= tick
GORELEASER ?= goreleaser

.PHONY: build test smoke release release-check

build:
	mkdir -p bin
	go build -o bin/$(BINARY) ./cmd/tick

test:
	go test ./...

smoke: build
	./scripts/smoke.sh

release:
	@command -v $(GORELEASER) >/dev/null 2>&1 || { echo "goreleaser is required: https://goreleaser.com/install/"; exit 1; }
	$(GORELEASER) release --snapshot --clean

release-check:
	@command -v $(GORELEASER) >/dev/null 2>&1 || { echo "goreleaser is required: https://goreleaser.com/install/"; exit 1; }
	$(GORELEASER) check
