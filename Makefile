# ==============================================================================
# VECAI - Victoria Eagles Construct AI
# Project Management & Automation Makefile
# ==============================================================================

# Ensure tools installed to GOPATH/bin (e.g. migrate, air, staticcheck) are on PATH
export PATH := $(shell go env GOPATH)/bin:$(PATH)

# Optional local environment file override
-include .env
export

# ==============================================================================
# Configuration & Variables
# ==============================================================================
APP_NAME        ?= vecai
MAIN_PATH       ?= ./backend/cmd/api/main.go
BIN_DIR         ?= bin
BINARY          ?= $(BIN_DIR)/$(APP_NAME)
MIGRATIONS_DIR  ?= backend/db/migrations
UPLOAD_DIR      ?= backend/uploads
PORT            ?= 8080
APP_ENV         ?= development
JWT_SECRET      ?= vecai-dev-secret-change-me-please

# Database Configuration
DB_CONTAINER    ?= vecai-db
DB_VOLUME       ?= vecai_postgres_data
DB_USER         ?= postgres
DB_PASSWORD     ?= postgres
DB_NAME         ?= vecai
DB_PORT         ?= 5433
DB_HOST         ?= localhost
DB_SSLMODE      ?= disable
DB_URL          ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

# Tool paths with automatic fallbacks
MIGRATE         ?= $(shell which migrate 2>/dev/null || echo "go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1")
AIR             ?= $(shell which air 2>/dev/null)
STATICCHECK     ?= $(shell which staticcheck 2>/dev/null)
GOLANGCI_LINT   ?= $(shell which golangci-lint 2>/dev/null)

# Default Goal
.DEFAULT_GOAL   := help

# ==============================================================================
# Help
# ==============================================================================

.PHONY: help
help: ## Display this help screen with available targets
	@echo "=================================================================="
	@echo "  VECAI Project Makefile"
	@echo "=================================================================="
	@echo "Usage: make [target] [VARIABLE=value]"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} \
		/^[a-zA-Z0-9_-]+:.*?## / {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2} \
		/^##@/ {printf "\n\033[1;33m%s\033[0m\n", substr($$0, 5)}' $(MAKEFILE_LIST)
	@echo ""

##@ Development Commands

.PHONY: run dev build clean

run: ## Start the backend API server directly
	@echo "[RUN] Starting $(APP_NAME) API server..."
	@APP_ENV="$(APP_ENV)" PORT="$(PORT)" DATABASE_URL="$(DB_URL)" MIGRATIONS_PATH="$(MIGRATIONS_DIR)" UPLOAD_DIR="$(UPLOAD_DIR)" JWT_SECRET="$(JWT_SECRET)" go run $(MAIN_PATH)

dev: ## Start backend with live reload (air)
	@if [ -n "$(AIR)" ]; then \
		echo "[DEV] Starting $(APP_NAME) with air live-reload..."; \
		APP_ENV="$(APP_ENV)" PORT="$(PORT)" DATABASE_URL="$(DB_URL)" MIGRATIONS_PATH="$(MIGRATIONS_DIR)" UPLOAD_DIR="$(UPLOAD_DIR)" JWT_SECRET="$(JWT_SECRET)" $(AIR); \
	else \
		echo "[DEV] 'air' not found. Falling back to 'go run'. Run 'make install-tools' to install live-reload."; \
		$(MAKE) run; \
	fi

build: ## Compile production backend binary into bin/
	@echo "[BUILD] Building $(APP_NAME) binary..."
	@mkdir -p $(BIN_DIR)
	@go build -ldflags "-s -w" -o $(BINARY) $(MAIN_PATH)
	@echo "[BUILD] Binary built successfully at $(BINARY)"

clean: ## Remove compiled binaries and temporary test artifacts
	@echo "[CLEAN] Removing binaries and test artifacts..."
	@rm -rf $(BIN_DIR) coverage.out coverage.html tmp/

##@ Testing & Code Quality

.PHONY: test test-cover fmt fmt-check vet lint check tidy

test: ## Run backend unit/integration tests with race detection
	@echo "[TEST] Running Go tests..."
	@go test -v -race ./backend/...

test-cover: ## Run backend tests and generate HTML coverage report
	@echo "[TEST] Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./backend/...
	@go tool cover -func=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "[TEST] Coverage report generated: coverage.html"

fmt: ## Format Go code in backend
	@echo "[FMT] Formatting Go files..."
	@gofmt -s -w backend/
	@go fmt ./backend/...

fmt-check: ## Verify Go code formatting (useful for CI)
	@echo "[FMT] Checking code formatting..."
	@test -z "$$(gofmt -l backend/)" || { echo "Code is not formatted. Run 'make fmt' to fix:"; gofmt -l backend/; exit 1; }
	@echo "[FMT] All Go code is properly formatted."

vet: ## Run go vet static analysis
	@echo "[VET] Running go vet..."
	@go vet ./backend/...

lint: ## Run linter (golangci-lint or staticcheck)
	@if [ -n "$(GOLANGCI_LINT)" ]; then \
		echo "[LINT] Running golangci-lint..."; \
		$(GOLANGCI_LINT) run ./backend/...; \
	elif [ -n "$(STATICCHECK)" ]; then \
		echo "[LINT] Running staticcheck..."; \
		$(STATICCHECK) ./backend/...; \
	else \
		echo "[LINT] No advanced linter found. Running go vet. Run 'make install-tools' for full linting."; \
		go vet ./backend/...; \
	fi

check: fmt-check vet test ## Run full quality checks (fmt-check, vet, and test)

tidy: ## Clean up and verify go.mod / go.sum dependencies
	@echo "[TIDY] Tidying and verifying Go module dependencies..."
	@go mod tidy
	@go mod verify

##@ Database & Migrations

.PHONY: migrate-up migrate-down migrate-down-all migrate-create migrate-version migrate-force migrate-drop db-reset

migrate-up: ## Apply all pending database migrations
	@echo "[MIGRATE] Running migrations up against $(DB_URL)..."
	@$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

migrate-down: ## Rollback the last database migration (use N=2 for multiple)
	@echo "[MIGRATE] Rolling back $(or $(N),1) migration(s)..."
	@$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down $(or $(N),1)

migrate-down-all: ## Rollback ALL database migrations
	@echo "[MIGRATE] Rolling back all migrations..."
	@$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down -all

migrate-create: ## Create new migration files (usage: make migrate-create NAME=my_migration)
	@if [ -z "$(NAME)" ]; then echo "Error: NAME is required. Example: make migrate-create NAME=add_suppliers_table"; exit 1; fi
	@echo "[MIGRATE] Creating migration: $(NAME)..."
	@$(MIGRATE) create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)

migrate-version: ## Print the current database migration version
	@$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" version

migrate-force: ## Force database version to clean dirty migration state (usage: make migrate-force VERSION=1)
	@if [ -z "$(VERSION)" ]; then echo "Error: VERSION is required. Example: make migrate-force VERSION=3"; exit 1; fi
	@echo "[MIGRATE] Forcing migration version to $(VERSION)..."
	@$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" force $(VERSION)

migrate-drop: ## Drop all database tables and schema
	@echo "[MIGRATE] Dropping all database tables..."
	@$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" drop -f

db-reset: migrate-drop migrate-up ## Drop all schema and re-run all migrations from scratch

##@ Docker & Database Service

.PHONY: docker-up docker-down docker-stop docker-start docker-restart docker-logs docker-status docker-clean db-shell db-wait

docker-up: ## Start PostgreSQL container (creates with volume if missing, or starts existing)
	@if [ ! "$$(docker ps -a -q -f name=$(DB_CONTAINER))" ]; then \
		echo "[DOCKER] Creating and starting $(DB_CONTAINER) on port $(DB_PORT)..."; \
		docker run --name $(DB_CONTAINER) \
			-e POSTGRES_USER=$(DB_USER) \
			-e POSTGRES_PASSWORD=$(DB_PASSWORD) \
			-e POSTGRES_DB=$(DB_NAME) \
			-v $(DB_VOLUME):/var/lib/postgresql/data \
			-p $(DB_PORT):5432 \
			-d postgres:16-alpine; \
	elif [ ! "$$(docker ps -q -f name=$(DB_CONTAINER))" ]; then \
		echo "[DOCKER] Starting existing $(DB_CONTAINER)..."; \
		docker start $(DB_CONTAINER); \
	else \
		echo "[DOCKER] Container $(DB_CONTAINER) is already running."; \
	fi

docker-stop: ## Stop the PostgreSQL container without removing it (preserves container)
	@echo "[DOCKER] Stopping $(DB_CONTAINER)..."
	@docker stop $(DB_CONTAINER) 2>/dev/null || echo "[DOCKER] $(DB_CONTAINER) is not running."

docker-start: ## Start existing stopped PostgreSQL container
	@echo "[DOCKER] Starting $(DB_CONTAINER)..."
	@docker start $(DB_CONTAINER) 2>/dev/null || echo "[DOCKER] $(DB_CONTAINER) does not exist. Run 'make docker-up'."

docker-restart: docker-stop docker-start ## Restart the PostgreSQL container

docker-down: ## Stop and remove the PostgreSQL container (keeps volume/data safe)
	@echo "[DOCKER] Stopping and removing $(DB_CONTAINER)..."
	@docker stop $(DB_CONTAINER) 2>/dev/null || true
	@docker rm $(DB_CONTAINER) 2>/dev/null || true

docker-logs: ## Follow PostgreSQL container logs
	@docker logs -f $(DB_CONTAINER)

docker-status: ## Show PostgreSQL container status
	@docker ps -a --filter name=$(DB_CONTAINER)

docker-clean: ## Remove PostgreSQL container AND volume (wipes all local data!)
	@echo "[DOCKER] Removing container $(DB_CONTAINER) and volume $(DB_VOLUME)..."
	@docker stop $(DB_CONTAINER) 2>/dev/null || true
	@docker rm $(DB_CONTAINER) 2>/dev/null || true
	@docker volume rm $(DB_VOLUME) 2>/dev/null || true
	@echo "[DOCKER] Container and volume deleted."

db-shell: ## Open interactive psql shell inside the database container
	@docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME)

db-wait: ## Wait until PostgreSQL container is ready to accept connections
	@echo "[DATABASE] Waiting for PostgreSQL on port $(DB_PORT)..."
	@until docker exec $(DB_CONTAINER) pg_isready -U $(DB_USER) >/dev/null 2>&1; do \
		sleep 1; \
	done
	@echo "[DATABASE] PostgreSQL is accepting connections!"

##@ Fullstack & Setup

.PHONY: setup frontend-install frontend-dev frontend-build frontend-lint install-tools

setup: docker-up db-wait migrate-up ## Bootstrap local dev environment (starts DB, waits for readiness, runs migrations)
	@echo "[SETUP] Local database and migrations ready!"
	@echo "[SETUP] Run 'make run' or 'make dev' to start the backend."

frontend-install: ## Install frontend npm dependencies
	@echo "[FRONTEND] Installing dependencies..."
	@npm --prefix frontend install

frontend-dev: ## Start frontend Next.js dev server
	@echo "[FRONTEND] Starting Next.js dev server..."
	@npm --prefix frontend run dev

frontend-build: ## Build frontend Next.js production bundle
	@echo "[FRONTEND] Building frontend bundle..."
	@npm --prefix frontend run build

frontend-lint: ## Run frontend ESLint checks
	@echo "[FRONTEND] Linting frontend..."
	@npm --prefix frontend run lint

install-tools: ## Install development tools (golang-migrate, air, staticcheck)
	@echo "[TOOLS] Installing golang-migrate CLI..."
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1
	@echo "[TOOLS] Installing air (live reload)..."
	@go install github.com/air-verse/air@latest
	@echo "[TOOLS] Installing staticcheck..."
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@echo "[TOOLS] Tools installed to $$(go env GOPATH)/bin"
