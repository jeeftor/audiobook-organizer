# Build variables
GIT_TAG ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# LDFLAGS for embedding version information
LDFLAGS := -ldflags "-s -w \
	-X audiobook-organizer/cmd.buildVersion=$(GIT_TAG) \
	-X audiobook-organizer/cmd.buildCommit=$(GIT_COMMIT) \
	-X audiobook-organizer/cmd.buildTime=$(BUILD_TIME)"

.PHONY: build clean dev release

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

# Run tests
test:
	@if [ -z "$(GOTESTSUM)" ]; then \
		echo "gotestsum not found, installing..."; \
		go install gotest.tools/gotestsum@latest; \
	fi
	gotestsum --format testname ./...

# Run tests with coverage reporting
coverage:
	@if [ -z "$(GOTESTSUM)" ]; then \
		echo "gotestsum not found, installing..."; \
		go install gotest.tools/gotestsum@latest; \
	fi
	gotestsum --format testname -- -coverprofile=coverage.out -covermode=count ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "\nCoverage report generated at coverage.html"
	@echo "Open it with: open coverage.html"

# Create a release (requires GITHUB_TOKEN)
release:
	goreleaser release --clean
