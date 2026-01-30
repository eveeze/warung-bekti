.PHONY: help build run dev test lint clean docker-up docker-down migrate-up migrate-down

# Variables
BINARY_NAME=warung-api
MAIN_PATH=./cmd/api
DOCKER_COMPOSE=docker-compose

# Help
help:
	@echo "WarungOS Backend - Available Commands:"
	@echo ""
	@echo "  make build          Build the application binary"
	@echo "  make run            Run the application locally"
	@echo "  make dev            Run with hot reload (requires air)"
	@echo "  make test           Run all tests"
	@echo "  make test-coverage  Run tests with coverage report"
	@echo "  make lint           Run linter"
	@echo "  make clean          Clean build artifacts"
	@echo ""
	@echo "  make docker-up      Start all Docker services"
	@echo "  make docker-down    Stop all Docker services"
	@echo "  make docker-logs    View Docker logs"
	@echo "  make docker-build   Build Docker image"
	@echo ""
	@echo "  make migrate-up     Run database migrations"
	@echo "  make migrate-down   Rollback last migration"
	@echo "  make migrate-status Show migration status"
	@echo ""
	@echo "  make deps           Download dependencies"
	@echo "  make tidy           Tidy go modules"

# Build
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -ldflags="-w -s" -o bin/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: bin/$(BINARY_NAME)"

# Run locally
run: build
	@./bin/$(BINARY_NAME)

# Development with hot reload
dev:
	@air -c .air.toml

# Tests
test:
	@go test -v ./...

test-coverage:
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Lint
lint:
	@golangci-lint run ./...

# Clean
clean:
	@rm -rf bin/
	@rm -rf tmp/
	@rm -f coverage.out coverage.html
	@echo "Cleaned build artifacts"

# Docker commands
docker-up:
	@$(DOCKER_COMPOSE) up -d
	@echo "Services started. API available at http://localhost:8080"

docker-down:
	@$(DOCKER_COMPOSE) down

docker-logs:
	@$(DOCKER_COMPOSE) logs -f

docker-build:
	@$(DOCKER_COMPOSE) build

docker-dev:
	@$(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.dev.yml up

# Database migrations
migrate-up:
	@go run ./cmd/api migrate up

migrate-down:
	@go run ./cmd/api migrate down

migrate-status:
	@go run ./cmd/api migrate status

# Dependencies
deps:
	@go mod download

tidy:
	@go mod tidy
	
# Data Seeding
seed:
	@go run cmd/seeder/main.go

# Install development tools
tools:
	@go install github.com/air-verse/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Development tools installed"

# Generate (for future use with code generators)
generate:
	@go generate ./...

# Quick infrastructure setup (PostgreSQL, Redis, Minio only)
infra-up:
	@$(DOCKER_COMPOSE) up -d postgres redis minio
	@echo "Infrastructure ready:"
	@echo "  PostgreSQL: localhost:5432"
	@echo "  Redis:      localhost:6379"
	@echo "  Minio:      localhost:9000 (console: localhost:9001)"

infra-down:
	@$(DOCKER_COMPOSE) stop postgres redis minio

# Performance & Load Testing
test-seed:
	@echo "Seeding 2000 products and 100 customers..."
	@python scripts/seed_data.py --products 2000 --customers 100

test-quick:
	@echo "Running quick stress test (3 minutes)..."
	@k6 run tests/load/quick_test.js

test-load:
	@echo "Running full load test (1000 users, ~22 minutes)..."
	@k6 run tests/load/load_test.js

test-db-stress:
	@echo "Running database stress test (5 minutes)..."
	@k6 run tests/load/db_stress_test.js

test-api:
	@echo "Running API integration tests..."
	@python tests/integration/run_api_tests.py

test-bench:
	@echo "Running Go benchmark tests..."
	@go test -bench=. -benchmem ./tests/benchmark/

test-all: test-api test-bench test-quick

# Production-Realistic Load Tests
test-prod-quick:
	@echo "Running quick production test (5 minutes, 200 users)..."
	@k6 run tests/load/production_quick.js

test-prod-full:
	@echo "Running full production test (20+ minutes, 500 users)..."
	@k6 run tests/load/production_test.js

# Combined test suites
test-validate: test-api test-prod-quick
	@echo "✅ Validation complete!"

test-production: test-api test-prod-full
	@echo "✅ Production readiness test complete!"

	@echo "All tests completed!"
