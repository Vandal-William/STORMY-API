.PHONY: build run test lint clean help db-up db-down db-init docker-compose-up docker-compose-down dev

# Variables
BINARY_NAME=message-service
MAIN_PATH=./cmd/api/main.go
BUILD_DIR=./build
CASSANDRA_HOST?=localhost
CASSANDRA_PORT?=9042

# Default target
help:
	@echo "Message Service - Available Targets"
	@echo ""
	@echo "Building & Running:"
	@echo "  build           - Build the binary"
	@echo "  run             - Run the service locally"
	@echo "  dev             - Run with hot reload (requires air)"
	@echo "  docker-build    - Build Docker image"
	@echo ""
	@echo "Database:"
	@echo "  db-up           - Start Cassandra + Redis + NATS with docker-compose"
	@echo "  db-down         - Stop and remove containers"
	@echo "  db-init         - Initialize Cassandra schema"
	@echo "  db-clean        - Remove Cassandra volumes and reset"
	@echo "  db-status       - Check Cassandra status"
	@echo ""
	@echo "Full Stack:"
	@echo "  docker-compose-up   - Start all services (DB + App)"
	@echo "  docker-compose-down - Stop all services"
	@echo ""
	@echo "Testing:"
	@echo "  test            - Run unit tests"
	@echo "  test-v          - Run unit tests (verbose)"
	@echo "  test-integration - Run integration tests"
	@echo ""
	@echo "Code Quality:"
	@echo "  lint            - Run linter"
	@echo "  fmt             - Format code"
	@echo "  clean           - Clean build artifacts"

# ============================================================
# BUILD & RUN
# ============================================================

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✓ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

run: build
	@echo "Running $(BINARY_NAME)..."
	@$(BUILD_DIR)/$(BINARY_NAME)

dev:
	@echo "Running in development mode with hot reload..."
	@if ! command -v air &> /dev/null; then \
		echo "Installing air..."; \
		go install github.com/cosmtrek/air@latest; \
	fi
	@air

docker-build:
	@echo "Building Docker image: $(BINARY_NAME):dev"
	@docker build -t $(BINARY_NAME):dev .
	@echo "✓ Docker image built successfully"

# ============================================================
# DATABASE
# ============================================================

db-up:
	@echo "Starting Cassandra + Redis + NATS..."
	@docker-compose up -d cassandra redis nats
	@echo "Waiting for Cassandra to be healthy (40-50s)..."
	@sleep 45
	@echo "✓ Services started"

db-down:
	@echo "Stopping database services..."
	@docker-compose down cassandra redis nats
	@echo "✓ Services stopped"

db-clean:
	@echo "Removing Cassandra volumes..."
	@docker-compose down -v cassandra
	@echo "✓ Cassandra volumes removed"

db-init:
	@echo "Initializing Cassandra schema..."
	@if command -v python3 &> /dev/null; then \
		python3 db/cassandra/setup.py $(CASSANDRA_HOST) $(CASSANDRA_PORT); \
	elif command -v cqlsh &> /dev/null; then \
		cqlsh $(CASSANDRA_HOST) $(CASSANDRA_PORT) -f db/cassandra/schema.cql; \
	elif command -v bash &> /dev/null; then \
		bash db/cassandra/setup.sh $(CASSANDRA_HOST) $(CASSANDRA_PORT); \
	else \
		echo "✗ Error: No initialization tool found"; \
		exit 1; \
	fi

db-status:
	@echo "Checking Cassandra status..."
	@docker-compose exec cassandra nodetool status 2>/dev/null || \
		cqlsh $(CASSANDRA_HOST) $(CASSANDRA_PORT) -e "SELECT now() FROM system.local LIMIT 1;" || \
		echo "✗ Cassandra is not running"

# ============================================================
# FULL STACK
# ============================================================

docker-compose-up:
	@echo "Starting full stack (DB + App)..."
	@docker-compose up -d
	@echo "Waiting for services to be ready..."
	@sleep 45
	@echo "✓ Full stack running"
	@echo "Services:"
	@echo "  - Message Service: http://localhost:3001"
	@echo "  - Cassandra: localhost:9042"
	@echo "  - Redis: localhost:6379"
	@echo "  - NATS: localhost:4222"

docker-compose-down:
	@echo "Stopping full stack..."
	@docker-compose down
	@echo "✓ Stack stopped"

# ============================================================
# TESTING
# ============================================================

test:
	@echo "Running unit tests..."
	@go test ./internal/... -v
	@echo "✓ Tests completed"

test-v:
	@echo "Running unit tests (verbose)..."
	@go test -v -race -coverprofile=coverage.out ./...
	@echo "✓ Coverage report: coverage.out"

test-integration:
	@echo "Running integration tests..."
	@if ! docker-compose ps cassandra | grep -q "Up"; then \
		echo "Starting Cassandra for integration tests..."; \
		make db-up; \
		make db-init; \
	fi
	@go test -v ./tests/integration/...
	@echo "✓ Integration tests completed"

# ============================================================
# CODE QUALITY
# ============================================================

lint:
	@echo "Running linter..."
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@golangci-lint run ./...
	@echo "✓ Lint completed"

fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✓ Code formatted"

# ============================================================
# CLEANUP
# ============================================================

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out
	@go clean
	@echo "✓ Cleanup completed"
