# Build variables
GIT_TAG ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# LDFLAGS for embedding version information
LDFLAGS := -ldflags "-s -w \
	-X audiobook-organizer/cmd.buildVersion=$(GIT_TAG) \
	-X audiobook-organizer/cmd.buildCommit=$(GIT_COMMIT) \
	-X audiobook-organizer/cmd.buildTime=$(BUILD_TIME)"

.PHONY: build clean dev release test test-verbose

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
	rm -rf ./dist ./bin

# Create a release (requires GITHUB_TOKEN)
release:
	goreleaser release --clean

# Run tests
test:
	go test ./...

# Run tests with verbose output and coverage
test-verbose:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
