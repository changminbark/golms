# Go parameters
BINARY_NAME=golms
GO=go
GOFLAGS=
GOTEST=$(GO) test
GOVET=$(GO) vet
GOFMT=gofmt
GOBUILD=$(GO) build
GOCLEAN=$(GO) clean
GOGET=$(GO) get
GOMOD=$(GO) mod

# Build directories
BUILD_DIR=.

.PHONY: all build build-vendor install clean test test-coverage fmt vet check vendor help

# Default target
all: check build

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## build-vendor: Build the binary using vendor dependencies
build-vendor:
	@echo "Building $(BINARY_NAME) with vendor dependencies..."
	$(GOBUILD) -mod=vendor $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## install: Install the binary to $GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install
	@echo "Installed $(BINARY_NAME) to $$GOPATH/bin"

## clean: Remove binary and clean build cache
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	@echo "Clean complete"

## test: Run all tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

## test-coverage: Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## fmt: Format all Go files
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .
	@echo "Format complete"

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...
	@echo "Vet complete"

## check: Run fmt, vet, and test
check: fmt vet test
	@echo "All checks passed"

## vendor: Download and vendor dependencies
vendor:
	@echo "Vendoring dependencies..."
	$(GOMOD) tidy
	$(GOMOD) vendor
	@echo "Vendor complete"

## run: Build and run the binary
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME)

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	@echo "Dependencies downloaded"

## tidy: Tidy go.mod
tidy:
	@echo "Tidying go.mod..."
	$(GOMOD) tidy
	@echo "Tidy complete"

## help: Display this help message
help:
	@echo "Available targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
