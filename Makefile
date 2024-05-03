ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
TOOLS_DIR := .tools
GOOS ?= linux
GOARCH ?= amd64

GOLANGCI_LINT_REPO = github.com/golangci/golangci-lint
GOLANGCI_LINT_VERSION = v1.56.1

HELM_DOCS_REPO = github.com/norwoodj/helm-docs
HELM_DOCS_VERSION = v1.13.1

all: test build
PHONY: test coverage lint docs

test: | lint
	@echo Running tests...
	@go test -mod=readonly -race -coverprofile=coverage.out -covermode=atomic ./...

lint: $(TOOLS_DIR)/golangci-lint
	@echo Linting Go files...
	@$(TOOLS_DIR)/golangci-lint run --modules-download-mode=readonly

build:
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -mod=readonly -v

docs: $(TOOLS_DIR)/helm-docs
	$(TOOLS_DIR)/helm-docs --chart-search-root ./chart/

go-dependencies:
	@go mod download
	@go mod tidy

$(TOOLS_DIR):
	mkdir -p $(TOOLS_DIR)

$(TOOLS_DIR)/golangci-lint: | $(TOOLS_DIR)
	@echo "Installing $(GOLANGCI_LINT_REPO)/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)"
	@GOBIN=$(ROOT_DIR)/$(TOOLS_DIR) go install $(GOLANGCI_LINT_REPO)/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

$(TOOLS_DIR)/helm-docs: | $(TOOLS_DIR)
	@echo "Installing $(HELM_DOCS_REPO)/cmd/helm-docs@$(HELM_DOCS_VERSION)"
	@GOBIN=$(ROOT_DIR)/$(TOOLS_DIR) go install $(HELM_DOCS_REPO)/cmd/helm-docs@$(HELM_DOCS_VERSION)
