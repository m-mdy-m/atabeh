SHELL := /bin/bash
set -euo pipefail

VERSION     ?= $(shell git describe --tags --match 'v[0-9]*' --always --dirty 2>/dev/null || echo dev)
GIT_COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE  ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
MODULE      := github.com/m-mdy-m/atabeh
BINARY      := atabeh
OUT_DIR     := bin
CMD_DIR     := ./cmd/atabeh

LDFLAGS     := -X '$(MODULE)/cmd/cli.Version=$(VERSION)' \
               -X '$(MODULE)/cmd/cli.GitCommit=$(GIT_COMMIT)' \
               -X '$(MODULE)/cmd/cli.BuildDate=$(BUILD_DATE)' \
               -s -w
.PHONY: help build test lint fmt clean install version tag release \
        docker docker-build docker-test deps tidy

help:
	@echo "atabeh — build & development commands"
	@echo ""
	@echo "  make build            Build the binary (bin/atabeh)"
	@echo "  make test             Run all unit tests"
	@echo "  make test-race        Run tests with the race detector"
	@echo "  make lint             Run go vet"
	@echo "  make fmt              Check & fix formatting"
	@echo "  make clean            Remove build artefacts"
	@echo "  make install          Install binary to \$GOPATH/bin"
	@echo "  make version          Print current version"
	@echo "  make tag              Create a new git tag (interactive)"
	@echo "  make release          Tag + push (interactive)"
	@echo "  make docker-build     Build the Docker image"
	@echo "  make docker-test      Run tests inside Docker"
	@echo "  make deps             Download dependencies"
	@echo "  make tidy             go mod tidy"

build: deps
	@echo "Building atabeh $(VERSION) …"
	go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(BINARY) $(CMD_DIR)
	@echo "  → $(OUT_DIR)/$(BINARY)"

build-linux-amd64:
	GOOS=linux  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(BINARY)-linux-amd64   $(CMD_DIR)

build-linux-arm64:
	GOOS=linux  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(BINARY)-linux-arm64   $(CMD_DIR)

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(BINARY)-darwin-arm64  $(CMD_DIR)

build-windows-amd64:
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(BINARY)-windows-amd64.exe $(CMD_DIR)

test: deps
	@echo "Running tests …"
	go test -count=1 -v ./tests/...

test-race: deps
	@echo "Running tests with race detector …"
	go test -race -count=1 -v ./tests/...

test-coverage: deps
	go test -count=1 -coverprofile=coverage.out ./tests/...
	go tool cover -func=coverage.out

# ---------------------------------------------------------------------------
# Lint & format
# ---------------------------------------------------------------------------
lint: deps
	@echo "Running go vet …"
	go vet ./...

fmt:
	@echo "Checking formatting …"
	gofmt -l -d .

# ---------------------------------------------------------------------------
# Clean
# ---------------------------------------------------------------------------
clean:
	@echo "Cleaning …"
	rm -rf $(OUT_DIR) coverage.out

# ---------------------------------------------------------------------------
# Install (copies binary to GOPATH/bin)
# ---------------------------------------------------------------------------
install: build
	@echo "Installing to \$GOPATH/bin …"
	cp $(OUT_DIR)/$(BINARY) $$(go env GOPATH)/bin/

# ---------------------------------------------------------------------------
# Version & tagging
# ---------------------------------------------------------------------------
version:
	@echo "version   : $(VERSION)"
	@echo "commit    : $(GIT_COMMIT)"
	@echo "build date: $(BUILD_DATE)"

tag:
	@echo "Current version: $(VERSION)"
	@read -p "New version (without 'v'): " ver; \
	if [ -z "$$ver" ]; then echo "Aborted."; exit 1; fi; \
	if git rev-parse "v$$ver" >/dev/null 2>&1; then \
		echo "Tag v$$ver already exists."; exit 1; \
	fi; \
	echo "Creating tag v$$ver …"; \
	git tag -a "v$$ver" -m "Release v$$ver"; \
	echo "Done. Push with: git push origin v$$ver"

release:
	@echo "Current version: $(VERSION)"
	@read -p "Release version (without 'v'): " ver; \
	if [ -z "$$ver" ]; then echo "Aborted."; exit 1; fi; \
	if git rev-parse "v$$ver" >/dev/null 2>&1; then \
		echo "Tag v$$ver already exists."; exit 1; \
	fi; \
	echo "Tagging v$$ver …"; \
	git tag -a "v$$ver" -m "Release v$$ver"; \
	echo "Pushing …"; \
	git push origin main; \
	git push origin "v$$ver"; \
	echo ""; \
	echo "Release v$$ver published."

# ---------------------------------------------------------------------------
# Docker
# ---------------------------------------------------------------------------
docker-build:
	docker build -t atabeh:$(VERSION) .

docker-test:
	docker build --target test -t atabeh-test:latest .
	docker run --rm atabeh-test:latest

# ---------------------------------------------------------------------------
# Module management
# ---------------------------------------------------------------------------
deps:
	go mod download

tidy:
	go mod tidy