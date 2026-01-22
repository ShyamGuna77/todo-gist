.PHONY: build run test clean fmt lint help deps tidy vet

# Variables
APP_NAME=api
BINARY_NAME=todo-gist
GO=go
GOFLAGS=-v
CMD_DIR=cmd/$(APP_NAME)

# Targets
build: ## Build the Go binary
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) ./$(CMD_DIR)

run: ## Build and run the application
	$(GO) run ./$(CMD_DIR)

test: ## Run tests
	$(GO) test $(GOFLAGS) ./...

test-verbose: ## Run tests with verbose output
	$(GO) test -v -race -coverprofile=coverage.out ./...

coverage: test-verbose ## Generate coverage report
	$(GO) tool cover -html=coverage.out

clean: ## Remove built binaries and temporary files
	$(GO) clean
	rm -f $(BINARY_NAME)
	rm -f coverage.out

deps: ## Download dependencies
	$(GO) mod download

tidy: ## Tidy module dependencies
	$(GO) mod tidy

fmt: ## Format code
	$(GO) fmt ./...

lint: ## Run linter (requires golangci-lint)
	golangci-lint run ./...

vet: ## Run go vet
	$(GO) vet ./...

help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
	awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
