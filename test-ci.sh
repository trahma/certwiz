#!/bin/bash

set -e

echo "================================================"
echo "Testing CI Compatibility"
echo "================================================"
echo ""

# Save original go.mod
cp go.mod go.mod.backup

# Use CI-compatible go.mod
cp go.mod.ci go.mod

# Download dependencies
echo "Downloading dependencies..."
go mod download

echo "Building project..."
if go build -o cert .; then
    echo "✓ Build successful"
else
    echo "✗ Build failed"
    cp go.mod.backup go.mod
    exit 1
fi

echo "Testing --version flag..."
if ./cert --version | grep -q "cert version"; then
    echo "✓ Version flag works"
else
    echo "✗ Version flag failed"
    cp go.mod.backup go.mod
    exit 1
fi

echo "Running tests..."
if go test ./...; then
    echo "✓ Tests passed"
else
    echo "✗ Tests failed"
    cp go.mod.backup go.mod
    exit 1
fi

# Restore original go.mod
cp go.mod.backup go.mod
rm go.mod.backup

echo ""
echo "================================================"
echo "✓ All CI compatibility tests passed!"
echo "================================================"