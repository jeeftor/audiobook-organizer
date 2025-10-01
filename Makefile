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

.PHONY: all build clean dev release test test-unit test-integration coverage coverage-html

# Default target
all: build

# Development build with version info
dev:
	go build $(LDFLAGS) -o bin/audiobook-organizer

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
