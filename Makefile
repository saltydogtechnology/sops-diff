# Makefile for sops-diff

# Variables
BINARY_NAME=sops-diff
VERSION=$(shell grep -o 'Version = "[^"]*"' main.go | cut -d'"' -f2 || echo "dev")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT=$(shell git rev-parse --short HEAD || echo "unknown")
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.CommitSHA=${COMMIT}"

# Go related variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GOFILES=$(wildcard *.go)
GOOS_LINUX=linux
GOOS_DARWIN=darwin
GOOS_WINDOWS=windows
GOARCH_AMD64=amd64
GOARCH_ARM64=arm64

# Targets
.PHONY: all build clean test lint vet fmt check install uninstall release help test-coverage test-integration test-all test-with-sops

all: check build test ## Run checks, build binary, and run tests

build: ## Build the binary for the current platform
	@echo "Building ${BINARY_NAME} for $(shell go env GOOS)/$(shell go env GOARCH)..."
	go build ${LDFLAGS} -o ${GOBIN}/${BINARY_NAME} .

build-all: build-linux build-darwin build-windows ## Build binaries for all platforms

build-linux: ## Build binaries for Linux
	@echo "Building ${BINARY_NAME} for linux/amd64..."
	GOOS=${GOOS_LINUX} GOARCH=${GOARCH_AMD64} go build ${LDFLAGS} -o ${GOBIN}/${BINARY_NAME}-${GOOS_LINUX}-${GOARCH_AMD64} .
	@echo "Building ${BINARY_NAME} for linux/arm64..."
	GOOS=${GOOS_LINUX} GOARCH=${GOARCH_ARM64} go build ${LDFLAGS} -o ${GOBIN}/${BINARY_NAME}-${GOOS_LINUX}-${GOARCH_ARM64} .

build-darwin: ## Build binaries for macOS
	@echo "Building ${BINARY_NAME} for darwin/amd64..."
	GOOS=${GOOS_DARWIN} GOARCH=${GOARCH_AMD64} go build ${LDFLAGS} -o ${GOBIN}/${BINARY_NAME}-${GOOS_DARWIN}-${GOARCH_AMD64} .
	@echo "Building ${BINARY_NAME} for darwin/arm64..."
	GOOS=${GOOS_DARWIN} GOARCH=${GOARCH_ARM64} go build ${LDFLAGS} -o ${GOBIN}/${BINARY_NAME}-${GOOS_DARWIN}-${GOARCH_ARM64} .

build-windows: ## Build binaries for Windows
	@echo "Building ${BINARY_NAME} for windows/amd64..."
	GOOS=${GOOS_WINDOWS} GOARCH=${GOARCH_AMD64} go build ${LDFLAGS} -o ${GOBIN}/${BINARY_NAME}-${GOOS_WINDOWS}-${GOARCH_AMD64}.exe .
	@echo "Building ${BINARY_NAME} for windows/arm64..."
	GOOS=${GOOS_WINDOWS} GOARCH=${GOARCH_ARM64} go build ${LDFLAGS} -o ${GOBIN}/${BINARY_NAME}-${GOOS_WINDOWS}-${GOARCH_ARM64}.exe .

clean: ## Remove build artifacts
	@echo "Cleaning..."
	rm -rf ${GOBIN}
	rm -f coverage.out coverage.html sops-diff-test
	go clean

test: ## Run tests
	@echo "Running tests..."
	go test -v ./... -short -coverprofile=coverage.out

test-coverage: test ## Generate and view coverage report
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated in coverage.html"
	@go tool cover -func=coverage.out

# Integration tests (requires compiled binary)
test-integration: ## Run integration tests
	go build -o sops-diff-test
	TEST_BINARY=./sops-diff-test go test -v ./... -run=TestCommandLineInterface
	rm -f sops-diff-test

# Full test suite including integration tests
test-all: test test-integration ## Run all tests

# Test with actual SOPS (requires GPG setup)
test-with-sops: ## Run tests with actual SOPS decryption
	@echo "Running tests with actual SOPS decryption"
	@echo "This requires a valid GPG key set in TEST_SOPS_PGP_KEY"
	@if [ -z "$$TEST_SOPS_PGP_KEY" ]; then \
		echo "TEST_SOPS_PGP_KEY environment variable not set"; \
		exit 1; \
	fi
	go build -o sops-diff-test
	TEST_BINARY=./sops-diff-test TEST_WITH_SOPS=1 go test -v ./... -run=TestActualDecryption
	rm -f sops-diff-test

lint: ## Run linter
	@echo "Running linter..."
	golint ./...

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

fmt: ## Run go fmt
	@echo "Formatting code..."
	go fmt ./...

check: fmt vet lint ## Run all static checks

install: build ## Install binary to GOPATH
	@echo "Installing ${BINARY_NAME} to $(shell go env GOPATH)/bin..."
	cp ${GOBIN}/${BINARY_NAME} $(shell go env GOPATH)/bin/

uninstall: ## Uninstall binary
	@echo "Uninstalling ${BINARY_NAME}..."
	rm -f $(shell go env GOPATH)/bin/${BINARY_NAME}

# Release creates release artifacts in the current directory
release: build-all ## Prepare release artifacts
	@echo "Creating release archives..."
	mkdir -p releases
	
	# Linux archives
	tar czf releases/${BINARY_NAME}-${VERSION}-${GOOS_LINUX}-${GOARCH_AMD64}.tar.gz -C ${GOBIN} ${BINARY_NAME}-${GOOS_LINUX}-${GOARCH_AMD64}
	tar czf releases/${BINARY_NAME}-${VERSION}-${GOOS_LINUX}-${GOARCH_ARM64}.tar.gz -C ${GOBIN} ${BINARY_NAME}-${GOOS_LINUX}-${GOARCH_ARM64}
	
	# MacOS archives
	tar czf releases/${BINARY_NAME}-${VERSION}-${GOOS_DARWIN}-${GOARCH_AMD64}.tar.gz -C ${GOBIN} ${BINARY_NAME}-${GOOS_DARWIN}-${GOARCH_AMD64}
	tar czf releases/${BINARY_NAME}-${VERSION}-${GOOS_DARWIN}-${GOARCH_ARM64}.tar.gz -C ${GOBIN} ${BINARY_NAME}-${GOOS_DARWIN}-${GOARCH_ARM64}
	
	# Windows archives
	zip -j releases/${BINARY_NAME}-${VERSION}-${GOOS_WINDOWS}-${GOARCH_AMD64}.zip ${GOBIN}/${BINARY_NAME}-${GOOS_WINDOWS}-${GOARCH_AMD64}.exe
	zip -j releases/${BINARY_NAME}-${VERSION}-${GOOS_WINDOWS}-${GOARCH_ARM64}.zip ${GOBIN}/${BINARY_NAME}-${GOOS_WINDOWS}-${GOARCH_ARM64}.exe
	
	# Generate checksums
	cd releases && \
	sha256sum ${BINARY_NAME}-${VERSION}-*.tar.gz ${BINARY_NAME}-${VERSION}-*.zip > ${BINARY_NAME}-${VERSION}-checksums.txt

help: ## Display help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
