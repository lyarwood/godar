.PHONY: build test lint clean install release vendor

# Build variables
BINARY_NAME=godar
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

# Default target
.DEFAULT_GOAL := build

build: ## Build the application
	go mod tidy
	go build ${LDFLAGS} -o ${BINARY_NAME} ./

install: ## Install the application
	go install ${LDFLAGS} ./

test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-benchmark: ## Run benchmark tests
	go test -bench=. ./...

lint: ## Run linter
	golangci-lint run ./...

lint-fix: ## Run linter with auto-fix
	golangci-lint run --fix ./...

clean: ## Clean build artifacts
	go clean
ifeq ($(OS),Windows_NT)
	-del /f /q ${BINARY_NAME}.exe coverage.out coverage.html ${BINARY_NAME}-*
else
	-rm -f ${BINARY_NAME} coverage.out coverage.html ${BINARY_NAME}-*
endif

vendor: ## Vendor all dependencies
	go mod tidy
	go mod vendor
	@echo "Dependencies vendored successfully"

release: ## Build for multiple platforms
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY_NAME}-linux-amd64 ./
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY_NAME}-darwin-amd64 ./
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY_NAME}-windows-amd64.exe ./

deps: ## Download dependencies
	go mod download
	go mod tidy

deps-update: ## Update dependencies
	go get -u ./...
	go mod tidy

check: ## Run all checks (lint, test, build)
	make lint
	make test
	make build