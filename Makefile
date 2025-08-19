# CertWiz Makefile

BINARY_NAME=cert
GO=go
GOFLAGS=-v

.PHONY: all build clean install test test-coverage test-coverage-html test-generate-certs run fmt vet help

## help: Display this help message
help:
	@echo "CertWiz - Certificate Management Tool"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## all: Build the binary
all: build

## build: Build the binary
build:
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) .

## clean: Remove build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	$(GO) clean

## install: Install the binary to GOPATH/bin
install:
	$(GO) install $(GOFLAGS)

## test: Run tests
test: test-generate-certs
	$(GO) test -v ./...

## test-coverage: Run tests with coverage
test-coverage: test-generate-certs
	@echo "Running tests with coverage..."
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out

## test-coverage-html: Generate HTML coverage report
test-coverage-html: test-coverage
	@echo "Generating HTML coverage report..."
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

## test-generate-certs: Generate test certificates
test-generate-certs:
	@echo "Generating test certificates..."
	@cd testdata && bash generate_test_certs.sh

## fmt: Format Go code
fmt:
	$(GO) fmt ./...

## vet: Run go vet
vet:
	$(GO) vet ./...

## run: Build and run the binary
run: build
	./$(BINARY_NAME)

# Development shortcuts
inspect: build
	./$(BINARY_NAME) inspect google.com

inspect-full: build
	./$(BINARY_NAME) inspect google.com --full

inspect-chain: build
	./$(BINARY_NAME) inspect google.com --chain

# Build all platforms
build-all:
	@echo "Building for all platforms..."
	@mkdir -p dist
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o dist/cert-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o dist/cert-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o dist/cert-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o dist/cert-linux-arm64 .
	GOOS=linux GOARCH=386 go build -ldflags "-s -w" -o dist/cert-linux-386 .
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o dist/cert-windows-amd64.exe .
	GOOS=windows GOARCH=arm64 go build -ldflags "-s -w" -o dist/cert-windows-arm64.exe .
	GOOS=windows GOARCH=386 go build -ldflags "-s -w" -o dist/cert-windows-386.exe .
	GOOS=freebsd GOARCH=amd64 go build -ldflags "-s -w" -o dist/cert-freebsd-amd64 .
	@echo "Build complete! Binaries are in ./dist/"

## release-test: Test release process locally with goreleaser
release-test:
	@echo "Testing release with goreleaser..."
	@goreleaser release --snapshot --clean

## release-local: Create a local release without publishing
release-local:
	@echo "Creating local release..."
	@goreleaser release --snapshot --skip=publish --clean