GOOS ?= linux
GOARCH ?= amd64

all: test build
PHONY: test coverage lint docs

test: | lint
	@echo Running tests...
	@go test -mod=readonly -race -coverprofile=coverage.out -covermode=atomic ./...

lint:
	@echo Linting Go files...
	@go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint run --modules-download-mode=readonly

fixlint:
	@echo Fixing go imports
	@find . -type f -iname '*.go' | xargs go run golang.org/x/tools/cmd/goimports -w -local go.infratographer.com/iam-runtime-infratographer

build:
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -mod=readonly -v

docs:
	@go run github.com/norwoodj/helm-docs/cmd/helm-docs --chart-search-root ./chart/

go-dependencies:
	@go mod download
	@go mod tidy
