#!/usr/bin/env bash
# Test script for environment variable handling
# Can be run locally or in CI
# Tests issue: https://github.com/jeeftor/audiobook-organizer/issues/17

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

FAILED=0

echo "========================================"
echo "Environment Variable Test Suite"
echo "========================================"
echo ""

# Build the binary
echo "Building binary..."
go build -o ./test-audiobook-organizer .
echo "✓ Build complete"
echo ""

# Create test directories
TEST_INPUT=$(mktemp -d)
TEST_OUTPUT=$(mktemp -d)
echo "Test directories created:"
echo "  Input:  $TEST_INPUT"
echo "  Output: $TEST_OUTPUT"
echo ""

cleanup() {
    echo ""
    echo "Cleaning up..."
    rm -f ./test-audiobook-organizer
    rm -rf "$TEST_INPUT" "$TEST_OUTPUT"
}
trap cleanup EXIT

# Test 1: AO_ prefix environment variables
echo "========================================"
echo "Test 1: AO_ prefix environment variables"
echo "========================================"
export AO_INPUT="$TEST_INPUT"
export AO_OUTPUT="$TEST_OUTPUT"
export AO_VERBOSE=true
export AO_DRY_RUN=true
export AO_LAYOUT=author-series-title

OUTPUT=$(./test-audiobook-organizer 2>&1 || true)
echo "$OUTPUT"

if echo "$OUTPUT" | grep -q "either --dir or --input must be specified"; then
    echo -e "${RED}❌ FAILED: AO_ environment variables not recognized!${NC}"
    echo "   This reproduces issue #17"
    FAILED=1
elif echo "$OUTPUT" | grep -q "Resolving paths"; then
    echo -e "${GREEN}✅ PASSED: AO_ environment variables recognized${NC}"
else
    echo -e "${YELLOW}⚠️  UNKNOWN: Could not determine result${NC}"
    FAILED=1
fi
echo ""

# Cleanup env vars
unset AO_INPUT AO_OUTPUT AO_VERBOSE AO_DRY_RUN AO_LAYOUT

# Test 2: AUDIOBOOK_ORGANIZER_ prefix
echo "========================================"
echo "Test 2: AUDIOBOOK_ORGANIZER_ prefix"
echo "========================================"
export AUDIOBOOK_ORGANIZER_INPUT="$TEST_INPUT"
export AUDIOBOOK_ORGANIZER_OUTPUT="$TEST_OUTPUT"
export AUDIOBOOK_ORGANIZER_VERBOSE=true
export AUDIOBOOK_ORGANIZER_DRY_RUN=true

OUTPUT=$(./test-audiobook-organizer 2>&1 || true)
echo "$OUTPUT"

if echo "$OUTPUT" | grep -q "either --dir or --input must be specified"; then
    echo -e "${RED}❌ FAILED: AUDIOBOOK_ORGANIZER_ prefix not recognized!${NC}"
    FAILED=1
elif echo "$OUTPUT" | grep -q "Resolving paths"; then
    echo -e "${GREEN}✅ PASSED: AUDIOBOOK_ORGANIZER_ prefix recognized${NC}"
else
    echo -e "${YELLOW}⚠️  UNKNOWN: Could not determine result${NC}"
    FAILED=1
fi
echo ""

# Cleanup env vars
unset AUDIOBOOK_ORGANIZER_INPUT AUDIOBOOK_ORGANIZER_OUTPUT AUDIOBOOK_ORGANIZER_VERBOSE AUDIOBOOK_ORGANIZER_DRY_RUN

# Test 3: --dir alias environment variable
echo "========================================"
echo "Test 3: --dir alias (AO_DIR)"
echo "========================================"
export AO_DIR="$TEST_INPUT"
export AO_OUTPUT="$TEST_OUTPUT"
export AO_DRY_RUN=true

OUTPUT=$(./test-audiobook-organizer 2>&1 || true)
echo "$OUTPUT"

if echo "$OUTPUT" | grep -q "either --dir or --input must be specified"; then
    echo -e "${RED}❌ FAILED: AO_DIR not recognized!${NC}"
    FAILED=1
elif echo "$OUTPUT" | grep -q "Resolving paths"; then
    echo -e "${GREEN}✅ PASSED: AO_DIR recognized${NC}"
else
    echo -e "${YELLOW}⚠️  UNKNOWN: Could not determine result${NC}"
    FAILED=1
fi
echo ""

# Cleanup env vars
unset AO_DIR AO_OUTPUT AO_DRY_RUN

# Test 4: Config file (if it exists)
echo "========================================"
echo "Test 4: Config file override"
echo "========================================"
CONFIG_FILE=$(mktemp --suffix=.yaml)
cat > "$CONFIG_FILE" << EOF
input: $TEST_INPUT
output: $TEST_OUTPUT
verbose: true
dry-run: true
layout: author-series-title
EOF

OUTPUT=$(./test-audiobook-organizer --config "$CONFIG_FILE" 2>&1 || true)
echo "$OUTPUT"

if echo "$OUTPUT" | grep -q "either --dir or --input must be specified"; then
    echo -e "${RED}❌ FAILED: Config file not loaded!${NC}"
    FAILED=1
elif echo "$OUTPUT" | grep -q "Resolving paths"; then
    echo -e "${GREEN}✅ PASSED: Config file loaded${NC}"
else
    echo -e "${YELLOW}⚠️  UNKNOWN: Could not determine result${NC}"
    FAILED=1
fi

rm -f "$CONFIG_FILE"
echo ""

# Test 5: Docker test (if Docker is available)
if command -v docker &> /dev/null; then
    echo "========================================"
    echo "Test 5: Docker environment variables"
    echo "========================================"

    if [ -f Dockerfile ]; then
        echo "Building Docker image..."
        docker build -t audiobook-organizer:test-env . > /dev/null 2>&1

        OUTPUT=$(docker run --rm \
            -e AO_INPUT=/data/input \
            -e AO_OUTPUT=/data/output \
            -e AO_VERBOSE=true \
            -e AO_DRY_RUN=true \
            -v "$TEST_INPUT:/data/input" \
            -v "$TEST_OUTPUT:/data/output" \
            audiobook-organizer:test-env 2>&1 || true)

        echo "$OUTPUT"

        if echo "$OUTPUT" | grep -q "either --dir or --input must be specified"; then
            echo -e "${RED}❌ FAILED: Docker env vars not recognized! (Issue #17)${NC}"
            FAILED=1
        elif echo "$OUTPUT" | grep -q "Resolving paths"; then
            echo -e "${GREEN}✅ PASSED: Docker env vars recognized${NC}"
        else
            echo -e "${YELLOW}⚠️  UNKNOWN: Could not determine result${NC}"
            FAILED=1
        fi
        echo ""
    else
        echo "Dockerfile not found, skipping Docker test"
        echo ""
    fi
else
    echo "Docker not available, skipping Docker test"
    echo ""
fi

# Summary
echo "========================================"
echo "Test Summary"
echo "========================================"
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    echo ""
    echo "This likely indicates issue #17 is present:"
    echo "https://github.com/jeeftor/audiobook-organizer/issues/17"
    exit 1
fi
