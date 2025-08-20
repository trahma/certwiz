#!/bin/bash

set -e

echo "================================================"
echo "Testing CI Build Locally"
echo "================================================"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test 1: Build the project
echo -e "${YELLOW}[TEST 1]${NC} Building project..."
if go build -v -o cert .; then
    echo -e "${GREEN}✓${NC} Build successful"
else
    echo -e "${RED}✗${NC} Build failed"
    exit 1
fi

# Test 2: Run tests
echo -e "${YELLOW}[TEST 2]${NC} Running tests..."
if go test ./... ; then
    echo -e "${GREEN}✓${NC} Tests passed"
else
    echo -e "${RED}✗${NC} Tests failed"
    exit 1
fi

# Test 3: Test --version flag
echo -e "${YELLOW}[TEST 3]${NC} Testing --version flag..."
if ./cert --version | grep -q "cert version"; then
    echo -e "${GREEN}✓${NC} Version flag works"
else
    echo -e "${RED}✗${NC} Version flag failed"
    exit 1
fi

# Test 4: Check for common issues
echo -e "${YELLOW}[TEST 4]${NC} Checking for common issues..."

# Check go.mod version
GO_VERSION=$(grep "^go " go.mod | awk '{print $2}')
echo "  Go version in go.mod: $GO_VERSION"

if [[ "$GO_VERSION" == "1.20" || "$GO_VERSION" == "1.21" ]]; then
    echo -e "  ${GREEN}✓${NC} Go version is compatible"
else
    echo -e "  ${RED}✗${NC} Go version may cause issues in CI"
fi

# Check for Go 1.21+ features
echo -e "${YELLOW}[TEST 5]${NC} Checking for Go 1.21+ features..."
if grep -r "slices\." . --include="*.go" 2>/dev/null | grep -v "^Binary"; then
    echo -e "${RED}✗${NC} Found usage of slices package (requires Go 1.21+)"
    exit 1
else
    echo -e "${GREEN}✓${NC} No Go 1.21+ specific features found"
fi

echo ""
echo -e "${GREEN}================================================"
echo "All local CI tests passed!"
echo "================================================${NC}"
echo ""
echo "To test with Docker and Go 1.20:"
echo "  docker build -f test-go-1.20.dockerfile -t certwiz-test ."
echo "  docker run --rm certwiz-test"