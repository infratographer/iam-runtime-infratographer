ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
TOOLS_DIR := .tools

GOLANGCI_LINT_REPO = github.com/golangci/golangci-lint
GOLANGCI_LINT_VERSION = v1.56.1

all: test build
PHONY: test coverage lint golint clean vendor docker-up docker-down unit-test

test: | lint
	@echo Running tests...
	@go test -mod=readonly -race -coverprofile=coverage.out -covermode=atomic ./...

lint: $(TOOLS_DIR)/golangci-lint
	@echo Linting Go files...
	@$(TOOLS_DIR)/golangci-lint run --modules-download-mode=readonly

build:
	@CGO_ENABLED=0 go build -mod=readonly -v -o bin/${APP_NAME}

go-dependencies:
	@go mod download
	@go mod tidy

$(TOOLS_DIR):
	mkdir -p $(TOOLS_DIR)

$(TOOLS_DIR)/golangci-lint: | $(TOOLS_DIR)
	@echo "Installing $(GOLANGCI_LINT_REPO)/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)"
	@GOBIN=$(ROOT_DIR)/$(TOOLS_DIR) go install $(GOLANGCI_LINT_REPO)/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
