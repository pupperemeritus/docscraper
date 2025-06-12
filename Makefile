# Makefile for docscraper

.PHONY: build test clean install run help

# Default target
help:
	@echo "Available targets:"
	@echo "  build     - Build the application"
	@echo "  test      - Run all tests"
	@echo "  test-unit - Run unit tests only"
	@echo "  test-integration - Run integration tests only"
	@echo "  clean     - Clean build artifacts"
	@echo "  install   - Install the application"
	@echo "  run       - Run the application (requires ROOT_URL)"
	@echo "  fmt       - Format Go code"
	@echo "  vet       - Run go vet"
	@echo "  lint      - Run golint (if available)"

# Build the application
build:
	@echo "Building docscraper..."
	go build -o bin/docscraper ./cmd/docscraper

# Install the application
install:
	@echo "Installing docscraper..."
	go install ./cmd/docscraper

# Run all tests
test:
	@echo "Running all tests..."
	go test ./...

# Run unit tests only (excluding integration tests)
test-unit:
	@echo "Running unit tests..."
	go test ./config ./scraper ./utils ./output ./devtools

# Run integration tests only
test-integration:
	@echo "Running integration tests..."
	go test ./tests/integration/...

# Run main package tests
test-main:
	@echo "Running main package tests..."
	go test ./tests/main/...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf output/
	rm -f *.log

# Format Go code
fmt:
	@echo "Formatting Go code..."
	go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Run golint if available
lint:
	@echo "Running golint..."
	@which golint > /dev/null && golint ./... || echo "golint not installed"

# Run the application (example usage)
run:
	@if [ -z "$(ROOT_URL)" ]; then \
		echo "Usage: make run ROOT_URL=https://example.com"; \
		exit 1; \
	fi
	./bin/docscraper --root-url $(ROOT_URL)

# Development workflow
dev: fmt vet test build

# Release build
release: clean fmt vet test build
