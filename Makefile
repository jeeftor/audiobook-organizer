# Build variables
GIT_TAG ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

VERSION_FLAGS := \
	-X github.com/jeeftor/audiobook-organizer/cmd.buildVersion=$(GIT_TAG) \
	-X github.com/jeeftor/audiobook-organizer/cmd.buildCommit=$(GIT_COMMIT) \
	-X github.com/jeeftor/audiobook-organizer/cmd.buildTime=$(BUILD_TIME)

# LDFLAGS for CLI-only builds (no CGO)
LDFLAGS := -ldflags "-s -w $(VERSION_FLAGS)"

# Test packages
UNIT_TEST_PKGS = ./...
INTEGRATION_TEST_PKGS = $(shell go list ./... | grep -v '/integration$$')
ABS_TEST_RUN ?= Test(MetadataJSONMode|EmbeddedAlreadyIndexed|EmbeddedMetadataImport|FlatModeImport|RESTHarness_MetadataJSONModeLifecycle)

.PHONY: all build clean dev dev-linux-amd64 web-install web-build web-dev gui-rest-test gui-test gui-test-headed gui-test-ui abs-dev-seed abs-dev-init abs-dev-configure abs-dev-up abs-dev-down abs-dev-reset abs-dev-reset-all abs-dev-scan abs-dev-reset-scan abs-ci-smoke abs-test-metadata abs-test-rest abs-test-matrix abs-test-e2e abs-dev-capture-baseline abs-dev-restore-baseline abs-dev-wait release test test-unit test-integration coverage coverage-html lint fmt fmt-check vet help scp-dev

# Default target - show help
all: help

# Show available targets
help:
	@echo "Available targets:"
	@echo ""
	@echo "  Development:"
	@printf "    %-26s %s\n" "dev" "Build CLI/TUI binary (native platform)"
	@printf "    %-26s %s\n" "dev-linux-amd64" "Build for Linux AMD64 (cross-compile)"
	@printf "    %-26s %s\n" "web-install" "Install web frontend dependencies"
	@printf "    %-26s %s\n" "web-build" "Build embedded web frontend assets"
	@printf "    %-26s %s\n" "web-dev" "Run the web frontend dev server"
	@printf "    %-26s %s\n" "gui-rest-test" "Run local web UI REST endpoint tests"
	@printf "    %-26s %s\n" "gui-test" "Run local web UI Playwright tests"
	@printf "    %-26s %s\n" "gui-test-headed" "Run local web UI Playwright tests headed"
	@printf "    %-26s %s\n" "gui-test-ui" "Open Playwright UI runner for local web UI tests"
	@printf "    %-26s %s\n" "scp-dev" "Copy linux-amd64 binary to remote server"
	@echo ""
	@echo "  ABS Development:"
	@printf "    %-26s %s\n" "abs-dev-seed" "Download public-domain ABS test media"
	@printf "    %-26s %s\n" "abs-dev-init" "Reset ABS and start with empty libraries for setup"
	@printf "    %-26s %s\n" "abs-dev-configure" "Configure empty ABS test servers through the API"
	@printf "    %-26s %s\n" "abs-dev-up" "Start local Audiobookshelf test server"
	@printf "    %-26s %s\n" "abs-dev-wait" "Start ABS and wait until services respond"
	@printf "    %-26s %s\n" "abs-dev-down" "Stop local Audiobookshelf test server"
	@printf "    %-26s %s\n" "abs-dev-reset" "Restore baseline, stage books, start ABS"
	@printf "    %-26s %s\n" "abs-dev-reset-all" "Reset ABS state and clear staged media"
	@printf "    %-26s %s\n" "abs-dev-scan" "Trigger scans for configured ABS libraries"
	@printf "    %-26s %s\n" "abs-dev-reset-scan" "Reset ABS, start it, and trigger scans"
	@printf "    %-26s %s\n" "abs-dev-capture-baseline" "Capture ABS baseline config fixture"
	@printf "    %-26s %s\n" "abs-dev-restore-baseline" "Restore ABS baseline config fixture"
	@echo ""
	@echo "  Build:"
	@printf "    %-26s %s\n" "build" "Build for distribution (goreleaser)"
	@printf "    %-26s %s\n" "release" "Create a release (requires GITHUB_TOKEN)"
	@printf "    %-26s %s\n" "clean" "Remove build artifacts"
	@echo ""
	@echo "  Testing:"
	@printf "    %-26s %s\n" "test" "Run unit tests (default)"
	@printf "    %-26s %s\n" "test-unit" "Run unit tests only"
	@printf "    %-26s %s\n" "test-integration" "Run integration tests"
	@printf "    %-26s %s\n" "test-all" "Run all tests"
	@printf "    %-26s %s\n" "coverage" "Run tests with coverage"
	@printf "    %-26s %s\n" "coverage-html" "Generate HTML coverage report"
	@echo ""
	@echo "  ABS Testing:"
	@printf "    %-26s %s\n" "abs-ci-smoke" "CI-style seed, restore baseline, and scan"
	@printf "    %-26s %s\n" "abs-test-metadata" "Run ABS metadata.json E2E tests"
	@printf "    %-26s %s\n" "abs-test-rest" "Run Docker-backed REST ABS E2E tests"
	@printf "    %-26s %s\n" "abs-test-matrix" "Run implemented ABS matrix E2E tests"
	@printf "    %-26s %s\n" "abs-test-e2e" "Run all ABS E2E tests"
	@echo ""
	@echo "  Code Quality:"
	@printf "    %-26s %s\n" "lint" "Run all linting (vet + fmt-check)"
	@printf "    %-26s %s\n" "vet" "Run go vet"
	@printf "    %-26s %s\n" "fmt" "Format Go code"
	@printf "    %-26s %s\n" "fmt-check" "Check code formatting"

# Development build (CLI/TUI only, no CGO needed)
dev:
	go build $(LDFLAGS) -o bin/audiobook-organizer

# Cross-compile for Linux AMD64 (for remote ABS servers)
# CGO disabled for cross-compilation compatibility
dev-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/audiobook-organizer-linux-amd64

# Install web frontend dependencies
web-install:
	cd web && npm install

# Build embedded web frontend assets
web-build:
	cd web && npm run build

# Run the web frontend development server
web-dev:
	cd web && npm run dev

# Run local web UI REST endpoint tests
gui-rest-test:
	GOCACHE=$${TMPDIR:-/tmp}/audiobook-organizer-go-build go test ./internal/app ./internal/server

# Run local web UI Playwright tests
gui-test:
	cd web && npm run test:e2e

# Run local web UI Playwright tests in headed mode
gui-test-headed:
	cd web && npm run test:e2e:headed

# Open the Playwright UI runner for local web UI tests
gui-test-ui:
	cd web && npm run test:e2e:ui

# Download public-domain media fixtures for the local ABS test server
abs-dev-seed:
	@test/abs/scripts/seed-public-domain.sh

# Reset ABS and start with empty mounted libraries for initial setup
abs-dev-init:
	@test/abs/scripts/reset.sh --empty-runtime
	@docker compose -f test/abs/docker-compose.yml up -d --remove-orphans
	@test/abs/scripts/wait-for-abs.sh

# Configure empty ABS test servers through the ABS API
abs-dev-configure:
	@test/abs/scripts/configure-from-api.sh

# Start the local ABS test server
abs-dev-up:
	@docker compose -f test/abs/docker-compose.yml up -d --remove-orphans

# Stop the local ABS test server
abs-dev-down:
	@docker compose -f test/abs/docker-compose.yml down --remove-orphans

# Reset ABS to baseline config, restore runtime books, start, and wait
abs-dev-reset:
	@test/abs/scripts/reset.sh
	@test/abs/scripts/restore-baseline.sh
	@test/abs/scripts/wait-for-restored-config.sh
	@docker compose -f test/abs/docker-compose.yml up -d --remove-orphans
	@test/abs/scripts/wait-for-abs.sh

# Reset ABS config, metadata, mounted books, and staged/downloaded books
abs-dev-reset-all:
	@test/abs/scripts/reset.sh --clear-staging

# Trigger scans for all configured ABS libraries in both instances
abs-dev-scan:
	@test/abs/scripts/scan-libraries.sh

# Reset ABS to baseline, start it, and trigger scans
abs-dev-reset-scan: abs-dev-reset abs-dev-scan

# Run the ABS harness the way GitHub Actions should: restore baseline fixture.
abs-ci-smoke:
	@test/abs/scripts/seed-public-domain.sh
	@test/abs/scripts/reset.sh
	@test/abs/scripts/restore-baseline.sh
	@test/abs/scripts/wait-for-restored-config.sh
	@docker compose -f test/abs/docker-compose.yml up -d --remove-orphans
	@test/abs/scripts/wait-for-abs.sh
	@test/abs/scripts/scan-libraries.sh

# Run the metadata.json ABS E2E tests. Tests reset ABS before each case.
abs-test-metadata:
	@test/abs/scripts/seed-public-domain.sh
	go test -tags=abs_e2e ./test/abs/e2e -run TestMetadataJSONMode -count=1 -v

# Run Docker-backed REST E2E tests. Tests reset ABS before each case.
abs-test-rest:
	@test/abs/scripts/seed-public-domain.sh
	go test -tags=abs_e2e ./test/abs/e2e -run TestRESTHarness_MetadataJSONModeLifecycle -count=1 -v

# Run implemented ABS matrix E2E tests. Tests reset ABS before each case.
abs-test-matrix:
	@test/abs/scripts/seed-public-domain.sh
	go test -tags=abs_e2e ./test/abs/e2e -run '$(ABS_TEST_RUN)' -count=1 -v

# Run all ABS E2E tests. Tests reset ABS before each case.
abs-test-e2e:
	@test/abs/scripts/seed-public-domain.sh
	go test -tags=abs_e2e ./test/abs/e2e -count=1 -v

# Capture configured ABS instances into the baseline config fixture
abs-dev-capture-baseline:
	@test/abs/scripts/capture-baseline.sh

# Restore configured ABS instances from the baseline config fixture
abs-dev-restore-baseline: abs-dev-reset

# Start the local ABS test servers and wait until both respond
abs-dev-wait: abs-dev-up
	@test/abs/scripts/wait-for-abs.sh

# Build using goreleaser for distribution
build: web-build
	goreleaser build --snapshot --clean

# Clean all build artifacts including the embedded frontend dist
clean:
	rm -rf ./dist ./bin ./coverage.out ./coverage.html
	rm -rf ./internal/server/static/assets

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

# Run go vet
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

# Deployment to remote ABS server (for testing)
# Usage: make scp-dev REMOTE_HOST=user@nas.local REMOTE_PATH=/usr/local/bin
.PHONY: scp-dev
scp-dev: dev-linux-amd64
	@if [ -z "$(REMOTE_HOST)" ]; then \
		echo "Error: REMOTE_HOST not set"; \
		echo "Usage: make scp-dev REMOTE_HOST=user@server REMOTE_PATH=/usr/local/bin"; \
		exit 1; \
	fi
	@if [ -z "$(REMOTE_PATH)" ]; then \
		echo "Error: REMOTE_PATH not set"; \
		echo "Usage: make scp-dev REMOTE_HOST=user@server REMOTE_PATH=/usr/local/bin"; \
		exit 1; \
	fi
	@echo "Copying linux-amd64 binary to $(REMOTE_HOST):$(REMOTE_PATH)/audiobook-organizer..."
	scp bin/audiobook-organizer-linux-amd64 $(REMOTE_HOST):$(REMOTE_PATH)/audiobook-organizer
	@echo "Done! Test with: ssh $(REMOTE_HOST) '$(REMOTE_PATH)/audiobook-organizer version'"
