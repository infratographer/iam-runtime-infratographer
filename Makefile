all: lint test
PHONY: test coverage lint golint clean vendor docker-up docker-down unit-test
GOOS=linux
# use the working dir as the app name, this should be the repo name
APP_NAME=$(shell basename $(CURDIR))

test: | lint
	@echo Running tests...
	@go test -mod=readonly -race -coverprofile=coverage.out -covermode=atomic ./...

lint:
	@echo Linting Go files...
	@golangci-lint run --modules-download-mode=readonly

build:
	@CGO_ENABLED=0 GOOS=linux go build -mod=readonly -v -o bin/${APP_NAME}

go-dependencies:
	@go mod download
	@go mod tidy
