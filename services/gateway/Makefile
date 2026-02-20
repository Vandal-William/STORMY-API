.PHONY: build run test lint clean help

# Variables
BINARY_NAME=gateway
MAIN_PATH=./cmd/api/main.go
BUILD_DIR=./build

# Default target
help:
	@echo "Gateway Service - Available targets:"
	@echo "  build       - Build the binary"
	@echo "  run         - Run the service locally"
	@echo "  test        - Run unit tests"
	@echo "  test-v      - Run unit tests with verbose output"
	@echo "  lint        - Run linter"
	@echo "  clean       - Clean build artifacts"
	@echo "  docker-build - Build Docker image"

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run locally
run: build
	@echo "Running $(BINARY_NAME)..."
	@$(BUILD_DIR)/$(BINARY_NAME)

# Run tests
test:
	@echo "Running tests..."
	@go test ./...

# Run tests with verbose output
test-v:
	@echo "Running tests (verbose)..."
	@go test -v ./...

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@go clean

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t gateway:dev .

# Development mode with hot reload (requires air)
dev:
	@echo "Running in development mode..."
	@air
