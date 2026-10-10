.PHONY: help build run dev test lint clean docker-up docker-down migrate-up migrate-down

# Variables
BINARY_NAME=warung-api
MAIN_PATH=./cmd/api
DOCKER_COMPOSE=docker compose
AIR_VERSION=v1.61.7
AIR=$(CURDIR)/bin/air

# Help
help:
	@echo "WarungOS Backend - Available Commands:"
	@echo ""
	@echo "  make build          Build the application binary"
	@echo "  make run            Run the application locally"
	@echo "  make dev            Run with hot reload (requires air)"
	@echo "  make setup-local    Prepare local env, infrastructure, migrations, and dev users"
	@echo "  make seed-users     Create development users without importing Excel"
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
dev: $(AIR)
	@$(AIR) -c .air.toml

$(AIR):
	@mkdir -p bin
	@GOBIN=$(CURDIR)/bin go install github.com/air-verse/air@$(AIR_VERSION)

.PHONY: env setup-local seed-users dev-up dev-down infra-up infra-down
env:
	@if [ ! -f .env ]; then (umask 077; cp .env.example .env); fi

setup-local: env
	@$(MAKE) dev-up
	@$(MAKE) migrate-up
	@$(MAKE) seed-users
	@$(MAKE) $(AIR)
	@echo "Local setup ready. Run: make dev"

# Tests
test:
	@go test -v ./...

test-coverage:
	@go test -coverpkg=./internal/... -coverprofile=coverage.raw ./internal/...
	@python -c 'import sys; lines = open("coverage.raw").readlines(); blocks = {}; [blocks.update({l.rsplit(" ", 2)[0]: (int(l.rsplit(" ", 2)[1]), blocks.get(l.rsplit(" ", 2)[0], (0, 0))[1] + int(l.rsplit(" ", 2)[2]))}) for l in lines[1:] if len(l.rsplit(" ", 2)) == 3]; out = open("coverage.out", "w"); out.write(lines[0]); [out.write(f"{k} {v[0]} {v[1]}\n") for k, v in blocks.items()]; tot = sum(v[0] for v in blocks.values()); cov = sum(v[0] for v in blocks.values() if v[1] > 0); print(f"\n=========================================\nTOTAL STATEMENT COVERAGE: {cov}/{tot} ({cov/tot*100:.2f}%)\n=========================================\n")'
	@go tool cover -html=coverage.out -o coverage.html
	@rm -f coverage.raw
	@echo "HTML coverage report generated: coverage.html"

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

dev-up:
	@$(DOCKER_COMPOSE) -f docker-compose.dev.yml up -d --wait
	@echo "Development infrastructure started."

dev-down:
	@$(DOCKER_COMPOSE) -f docker-compose.dev.yml down
	@echo "Development infrastructure stopped."

test-auto:
	@set -eu; \
	trap '$(DOCKER_COMPOSE) -f docker-compose.test.yml down' EXIT; \
	$(DOCKER_COMPOSE) -f docker-compose.test.yml up -d --wait; \
	DB_HOST=127.0.0.1 DB_PORT=5436 DB_NAME=warung_test DB_USER=warung DB_PASSWORD=warung_secret DB_SSL_MODE=disable go run ./cmd/api migrate up; \
	REDIS_PORT=6382 TEST_DATABASE_URL="postgres://warung:warung_secret@127.0.0.1:5436/warung_test?sslmode=disable" \
	TEST_DB_HOST=127.0.0.1 TEST_DB_PORT=5436 TEST_DB_NAME=warung_test TEST_DB_USER=warung TEST_DB_PASSWORD=warung_secret \
	go test -v ./...

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

seed-users:
	@go run ./cmd/seeder --users-only

# Install development tools
tools: $(AIR)
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Development tools installed"

# Generate (for future use with code generators)
generate:
	@go generate ./...

# Local development infrastructure
infra-up: dev-up

infra-down:
	@$(DOCKER_COMPOSE) -f docker-compose.dev.yml stop

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

test-e2e:
	@echo "Running Full End-to-End API tests (all features)..."
	@python tests/integration/run_full_e2e.py

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
