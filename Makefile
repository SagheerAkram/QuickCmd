.PHONY: help build test lint clean install deps docker

# Variables
BINARY_NAME=quickcmd
MAIN_PATH=./cmd/quickcmd
BUILD_DIR=./bin
GO=go
GOFLAGS=-v

# Build information
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Install dependencies
	$(GO) mod download
	$(GO) mod verify

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

install: build ## Install the binary to GOPATH/bin
	$(GO) install $(LDFLAGS) $(MAIN_PATH)

test: ## Run unit tests
	$(GO) test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

test-integration: ## Run integration tests (requires Docker)
	$(GO) test -v -tags=integration ./...

test-security: ## Run security tests
	@echo "Running gosec security scanner..."
	@which gosec > /dev/null || (echo "Installing gosec..." && $(GO) install github.com/securego/gosec/v2/cmd/gosec@latest)
	gosec -quiet ./...

test-all: test test-integration test-security ## Run all tests

coverage: test ## Generate coverage report
	$(GO) tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report: coverage.html"

lint: ## Run linters
	@echo "Running golangci-lint..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && $(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

fmt: ## Format code
	$(GO) fmt ./...
	@which goimports > /dev/null || $(GO) install golang.org/x/tools/cmd/goimports@latest
	goimports -w .

vet: ## Run go vet
	$(GO) vet ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.txt coverage.html
	@$(GO) clean

docker: ## Build Docker image
	docker build -t quickcmd:$(VERSION) .

docker-test: ## Run tests in Docker
	docker build -f Dockerfile.test -t quickcmd-test .
	docker run --rm quickcmd-test

run: build ## Build and run
	$(BUILD_DIR)/$(BINARY_NAME)

dev: ## Run in development mode with auto-reload
	@which air > /dev/null || (echo "Installing air..." && $(GO) install github.com/cosmtrek/air@latest)
	air

.DEFAULT_GOAL := help
