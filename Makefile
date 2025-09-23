# Makefile for Barbershop API

.PHONY: help run build test clean db-seed db-reset docker-up docker-down

# Variables
APP_NAME=barbershop-api
MAIN_PATH=./cmd/server
DB_URL?=postgres://postgres:password@localhost:5432/barbershop?sslmode=disable

# Colors for output
COLOR_RESET=\033[0m
COLOR_BOLD=\033[1m
COLOR_GREEN=\033[32m
COLOR_YELLOW=\033[33m
COLOR_BLUE=\033[34m

help: ## Show this help message
	@echo "$(COLOR_BOLD)Available commands:$(COLOR_RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_GREEN)%-15s$(COLOR_RESET) %s\n", $$1, $$2}'

install: ## Install dependencies
	@echo "$(COLOR_BLUE)📦 Installing dependencies...$(COLOR_RESET)"
	go mod download
	go mod tidy
	@echo "$(COLOR_GREEN)✅ Dependencies installed$(COLOR_RESET)"

run: ## Run the application
	@echo "$(COLOR_BLUE)🚀 Starting server...$(COLOR_RESET)"
	go run $(MAIN_PATH)/main.go $(MAIN_PATH)/routes.go

build: ## Build the application
	@echo "$(COLOR_BLUE)🔨 Building application...$(COLOR_RESET)"
	go build -o bin/$(APP_NAME) $(MAIN_PATH)/main.go $(MAIN_PATH)/routes.go
	@echo "$(COLOR_GREEN)✅ Build complete: bin/$(APP_NAME)$(COLOR_RESET)"

test: ## Run tests
	@echo "$(COLOR_BLUE)🧪 Running tests...$(COLOR_RESET)"
	go test -v -cover ./...

test-coverage: ## Run tests with coverage
	@echo "$(COLOR_BLUE)📊 Running tests with coverage...$(COLOR_RESET)"
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(COLOR_GREEN)✅ Coverage report: coverage.html$(COLOR_RESET)"

clean: ## Clean build artifacts
	@echo "$(COLOR_YELLOW)🧹 Cleaning...$(COLOR_RESET)"
	rm -rf bin/
	rm -f coverage.out coverage.html
	@echo "$(COLOR_GREEN)✅ Clean complete$(COLOR_RESET)"

db-seed: ## Seed the database
	@echo "$(COLOR_BLUE)🌱 Seeding database...$(COLOR_RESET)"
	@if [ -f "./scripts/seed.sh" ]; then \
		chmod +x ./scripts/seed.sh && ./scripts/seed.sh; \
	else \
		psql $(DB_URL) -f ./scripts/seeds/001_barbershop_seeds.sql; \
	fi
	@echo "$(COLOR_GREEN)✅ Database seeded$(COLOR_RESET)"

db-reset: ## Reset and seed database
	@echo "$(COLOR_YELLOW)⚠️  Resetting database...$(COLOR_RESET)"
	@read -p "Are you sure you want to reset the database? [y/N] " -n 1 -r; \
	echo ""; \
	if [[ $REPLY =~ ^[Yy]$ ]]; then \
		psql $(DB_URL) -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"; \
		echo "$(COLOR_BLUE)Running migrations...$(COLOR_RESET)"; \
		make db-migrate; \
		make db-seed; \
	fi

db-migrate: ## Run database migrations
	@echo "$(COLOR_BLUE)🔄 Running migrations...$(COLOR_RESET)"
	@for file in ./pkg/database/migrations/*.sql; do \
		echo "Applying $file..."; \
		psql $(DB_URL) -f $file; \
	done
	@echo "$(COLOR_GREEN)✅ Migrations complete$(COLOR_RESET)"

docker-up: ## Start Docker containers
	@echo "$(COLOR_BLUE)🐳 Starting Docker containers...$(COLOR_RESET)"
	docker-compose up -d
	@echo "$(COLOR_GREEN)✅ Containers started$(COLOR_RESET)"

docker-down: ## Stop Docker containers
	@echo "$(COLOR_YELLOW)🐳 Stopping Docker containers...$(COLOR_RESET)"
	docker-compose down
	@echo "$(COLOR_GREEN)✅ Containers stopped$(COLOR_RESET)"

docker-logs: ## Show Docker logs
	docker-compose logs -f

lint: ## Run linter
	@echo "$(COLOR_BLUE)🔍 Running linter...$(COLOR_RESET)"
	golangci-lint run ./...

fmt: ## Format code
	@echo "$(COLOR_BLUE)✨ Formatting code...$(COLOR_RESET)"
	go fmt ./...
	@echo "$(COLOR_GREEN)✅ Code formatted$(COLOR_RESET)"

dev: ## Run in development mode with hot reload
	@echo "$(COLOR_BLUE)🔥 Starting development server with hot reload...$(COLOR_RESET)"
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "$(COLOR_YELLOW)Air not installed. Running without hot reload...$(COLOR_RESET)"; \
		make run; \
	fi

api-test: ## Test API endpoints with curl
	@echo "$(COLOR_BLUE)🧪 Testing API endpoints...$(COLOR_RESET)"
	@echo "\n$(COLOR_BOLD)1. Health Check:$(COLOR_RESET)"
	@curl -s http://localhost:8080/health | jq '.'
	@echo "\n$(COLOR_BOLD)2. List Barbers:$(COLOR_RESET)"
	@curl -s http://localhost:8080/api/v1/barbers | jq '.data | length'
	@echo "\n$(COLOR_BOLD)3. Get Barber by ID (1):$(COLOR_RESET)"
	@curl -s http://localhost:8080/api/v1/barbers/1 | jq '.data.shop_name'
	@echo "$(COLOR_GREEN)✅ API tests complete$(COLOR_RESET)"

setup: install db-migrate db-seed ## Complete setup (install, migrate, seed)
	@echo "$(COLOR_GREEN)🎉 Setup complete! Run 'make run' to start the server$(COLOR_RESET)"