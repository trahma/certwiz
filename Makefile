# CertWiz Makefile

BINARY_NAME=cert
GO=go
GOFLAGS=-v

.PHONY: all build clean install test run help

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
	$(GO) clean

## install: Install the binary to GOPATH/bin
install:
	$(GO) install $(GOFLAGS)

## test: Run tests
test:
	$(GO) test -v ./...

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
	GOOS=darwin GOARCH=amd64 go build -o cert-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o cert-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build -o cert-linux-amd64 .
	GOOS=windows GOARCH=amd64 go build -o cert-windows-amd64.exe .