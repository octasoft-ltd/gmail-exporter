# Gmail Exporter Makefile

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS = -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

# Binary name
BINARY = gmail-exporter

# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
GOMOD = $(GOCMD) mod

# Default target
.PHONY: all
all: test build

# Build the binary
.PHONY: build
build:
	$(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BINARY) ./cmd/gmail-exporter

# Build for multiple platforms
.PHONY: build-all
build-all: clean
	mkdir -p dist
	GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o dist/$(BINARY)-linux-amd64 ./cmd/gmail-exporter
	GOOS=linux GOARCH=arm64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o dist/$(BINARY)-linux-arm64 ./cmd/gmail-exporter
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o dist/$(BINARY)-darwin-amd64 ./cmd/gmail-exporter
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o dist/$(BINARY)-darwin-arm64 ./cmd/gmail-exporter
	GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o dist/$(BINARY)-windows-amd64.exe ./cmd/gmail-exporter
	GOOS=windows GOARCH=arm64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o dist/$(BINARY)-windows-arm64.exe ./cmd/gmail-exporter

# Run tests
.PHONY: test
test:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

# Run tests with coverage report
.PHONY: test-coverage
test-coverage: test
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter
.PHONY: lint
lint:
	golangci-lint run

# Clean build artifacts
.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY)
	rm -rf dist/
	rm -f coverage.out coverage.html

# Download dependencies
.PHONY: deps
deps:
	$(GOMOD) download
	$(GOMOD) verify

# Tidy dependencies
.PHONY: tidy
tidy:
	$(GOMOD) tidy

# Install the binary
.PHONY: install
install: build
	sudo mv $(BINARY) /usr/local/bin/

# Uninstall the binary
.PHONY: uninstall
uninstall:
	sudo rm -f /usr/local/bin/$(BINARY)

# Run the application
.PHONY: run
run: build
	./$(BINARY)

# Development setup
.PHONY: dev-setup
dev-setup:
	@echo "Setting up development environment..."
	$(GOMOD) download
	@echo "Installing golangci-lint..."
	@which golangci-lint > /dev/null || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin
	@echo "Development setup complete!"

# Security scan
.PHONY: security
security:
	@which gosec > /dev/null || go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	gosec ./...

# Format code
.PHONY: fmt
fmt:
	$(GOCMD) fmt ./...
	goimports -w .

# Check for vulnerabilities
.PHONY: vuln-check
vuln-check:
	@which govulncheck > /dev/null || go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

# Release preparation
.PHONY: pre-release
pre-release: clean fmt lint test security vuln-check build-all
	@echo "Pre-release checks complete!"

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all          - Run tests and build"
	@echo "  build        - Build the binary"
	@echo "  build-all    - Build for all platforms"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage report"
	@echo "  lint         - Run linter"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Download dependencies"
	@echo "  tidy         - Tidy dependencies"
	@echo "  install      - Install binary to /usr/local/bin"
	@echo "  uninstall    - Remove binary from /usr/local/bin"
	@echo "  run          - Build and run the application"
	@echo "  dev-setup    - Set up development environment"
	@echo "  security     - Run security scan"
	@echo "  fmt          - Format code"
	@echo "  vuln-check   - Check for vulnerabilities"
	@echo "  pre-release  - Run all pre-release checks"
	@echo "  help         - Show this help"
