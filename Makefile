# Build variables
GIT_TAG ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# LDFLAGS for embedding version information
LDFLAGS := -ldflags "-s -w \
	-X github.com/jeeftor/audiobook-organizer/cmd.buildVersion=$(GIT_TAG) \
	-X github.com/jeeftor/audiobook-organizer/cmd.buildCommit=$(GIT_COMMIT) \
	-X github.com/jeeftor/audiobook-organizer/cmd.buildTime=$(BUILD_TIME)"

# Test packages
UNIT_TEST_PKGS = ./...
INTEGRATION_TEST_PKGS = $(shell go list ./... | grep -v '/integration$$')

.PHONY: all build clean dev gui-dev gui-dev1 gui-dev2 gui-build gui-install release test test-unit test-integration coverage coverage-html lint fmt fmt-check vet help

# Default target - show help
all: help

# Show available targets
help:
	@echo "Available targets:"
	@echo ""
	@echo "  Development:"
	@echo "    dev              Build CLI binary with version info"
	@echo "    gui-dev          Start GUI in dev mode (copies books to gui-books)"
	@echo "    gui-dev1         Start GUI with ./books as input"
	@echo "    gui-dev2         Start GUI with ./books-meta as input"
	@echo "    gui-install      Install GUI frontend dependencies"
	@echo ""
	@echo "  Build:"
	@echo "    build            Build for distribution (goreleaser)"
	@echo "    gui-build        Build GUI for production"
	@echo "    release          Create a release (requires GITHUB_TOKEN)"
	@echo "    clean            Remove build artifacts"
	@echo ""
	@echo "  Testing:"
	@echo "    test             Run unit tests (default)"
	@echo "    test-unit        Run unit tests only"
	@echo "    test-integration Run integration tests"
	@echo "    test-all         Run all tests"
	@echo "    coverage         Run tests with coverage"
	@echo "    coverage-html    Generate HTML coverage report"
	@echo ""
	@echo "  Code Quality:"
	@echo "    lint             Run all linting (vet + fmt-check)"
	@echo "    vet              Run go vet"
	@echo "    fmt              Format Go code"
	@echo "    fmt-check        Check code formatting"

# Development build with version info (CLI)
dev:
	go build $(LDFLAGS) -o bin/audiobook-organizer

# Start GUI in development mode
gui-dev:
	@echo "Starting GUI in development mode..."
	@rm -rf gui-books
	@rm -rf output
	@mkdir -p output
	@cp -r books gui-books
	cd audiobook-organizer-gui && wails dev -appargs "--dir=../gui-books --out=../output"


# Start GUI with ./books as input and . as output
gui-dev1:
	cd audiobook-organizer-gui && wails dev -appargs "--dir=../books --out=.."

# Start GUI with books-meta as input
gui-dev2:
	cd audiobook-organizer-gui && wails dev -appargs "--dir=../books-meta"

# Build GUI for production
gui-build:
	cd audiobook-organizer-gui && wails build

# Install GUI dependencies
gui-install:
	cd audiobook-organizer-gui/frontend && npm install

# Build using goreleaser for distribution
build:
	goreleaser build --snapshot --clean

# Clean build artifacts
clean:
	rm -rf ./dist ./bin ./coverage.out ./coverage.html

# Check and install gotestsum if needed
GOTESTSUM := $(shell command -v gotestsum 2> /dev/null)

# Run all tests (unit tests only by default)
test: test-unit

# Run unit tests (fast, no external dependencies)
test-unit: ensure-gotestsum
	@echo "Running unit tests..."
	gotestsum --format testname -- -short $(UNIT_TEST_PKGS)

# Run integration tests (slower, may require external dependencies)
test-integration: ensure-gotestsum
	@echo "Running integration tests..."
	gotestsum --format testname -- -tags=integration -v $(INTEGRATION_TEST_PKGS)

# Run all tests (both unit and integration)
test-all: ensure-gotestsum
	@echo "Running all tests..."
	gotestsum --format testname -- -v $(UNIT_TEST_PKGS)

# Run tests with coverage reporting
coverage: ensure-gotestsum
	@echo "Running tests with coverage..."
	gotestsum --format testname -- -coverprofile=coverage.out -covermode=count $(UNIT_TEST_PKGS)
	go tool cover -func=coverage.out

# Generate HTML coverage report
coverage-html: coverage
	go tool cover -html=coverage.out -o coverage.html
	@echo "\nCoverage report generated at coverage.html"
	@echo "Open it with: open coverage.html"

# Ensure gotestsum is installed
ensure-gotestsum:
	@if [ -z "$(GOTESTSUM)" ]; then \
		echo "gotestsum not found, installing..."; \
		go install gotest.tools/gotestsum@latest; \
	fi

# Create a release (requires GITHUB_TOKEN)
release:
	goreleaser release --clean

# Run go vet to check for suspicious code
vet:
	@echo "Running go vet..."
	go vet ./...

# Format all Go code
fmt:
	@echo "Formatting Go code..."
	gofmt -s -w .

# Check if code is properly formatted (non-destructive)
fmt-check:
	@echo "Checking code formatting..."
	@UNFORMATTED=$$(gofmt -l .); \
	if [ -n "$$UNFORMATTED" ]; then \
		echo "The following files are not properly formatted:"; \
		echo "$$UNFORMATTED"; \
		echo ""; \
		echo "Run 'make fmt' to format them."; \
		exit 1; \
	else \
		echo "All files are properly formatted."; \
	fi

# Run all linting checks (vet + fmt-check)
lint: vet fmt-check
	@echo "All linting checks passed!"
