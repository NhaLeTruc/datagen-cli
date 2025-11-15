.PHONY: build test lint coverage clean install help

# Go parameters
GOCMD=/usr/local/go/bin/go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=datagen
BINARY_PATH=./bin/$(BINARY_NAME)

# Build information
VERSION?=1.0.0
BUILD_TIME=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GO_VERSION=$(shell $(GOCMD) version | awk '{print $$3}')

# Linker flags
LDFLAGS=-ldflags "\
	-X 'github.com/NhaLeTruc/datagen-cli/internal/cli.Version=$(VERSION)' \
	-X 'github.com/NhaLeTruc/datagen-cli/internal/cli.BuildTime=$(BUILD_TIME)' \
	-X 'github.com/NhaLeTruc/datagen-cli/internal/cli.GitCommit=$(GIT_COMMIT)' \
	-X 'github.com/NhaLeTruc/datagen-cli/internal/cli.GoVersion=$(GO_VERSION)'"

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_PATH) ./cmd/datagen
	@echo "Built: $(BINARY_PATH)"

install: build ## Install the binary to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	$(GOCMD) install $(LDFLAGS) ./cmd/datagen
	@echo "Installed to $$($(GOCMD) env GOPATH)/bin/$(BINARY_NAME)"

test: ## Run all tests
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.txt -covermode=atomic ./...

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	$(GOTEST) -v -race ./tests/unit/...

test-integration: ## Run integration tests only
	@echo "Running integration tests..."
	$(GOTEST) -v -race ./tests/integration/...

test-e2e: ## Run end-to-end tests only
	@echo "Running e2e tests..."
	$(GOTEST) -v ./tests/e2e/...

coverage: test ## Run tests and generate coverage report
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report: coverage.html"

lint: ## Run linters
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install from https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi

fmt: ## Format code
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	$(GOCMD) vet ./...

tidy: ## Tidy go.mod
	@echo "Tidying go.mod..."
	$(GOMOD) tidy

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.txt coverage.html
	@rm -f *.dump *.sql
	@echo "Cleaned"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) verify

run: build ## Build and run the binary
	$(BINARY_PATH)

dev: ## Run in development mode
	$(GOCMD) run ./cmd/datagen

.DEFAULT_GOAL := help