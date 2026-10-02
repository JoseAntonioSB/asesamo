# ============================================================
# ASESAMO — Root Makefile
# ============================================================

.DEFAULT_GOAL := help
.PHONY: help docker-up docker-down docker-down-v migrate-up migrate-down migrate-version migrate-force seed db-shell test test-integration lint build run sqlc

# Load .env file if present
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Default variables
DB_USER        ?= asesamo_admin
DB_PASSWORD    ?= asesamo_secret_dev
DB_NAME        ?= asesamo_db
DB_HOST        ?= localhost
DB_PORT        ?= 5432
MIGRATE_IMAGE  ?= migrate/migrate:v4.17.0

export DB_USER DB_PASSWORD DB_NAME

## Display help menu
help:
	@echo "ASESAMO Phase 1 MVP — Automation Targets"
	@echo "=========================================="
	@echo "  make docker-up        - Start all docker-compose services in background"
	@echo "  make docker-down      - Stop all docker-compose services (preserves volumes)"
	@echo "  make docker-down-v    - Stop all services and destroy database volumes"
	@echo "  make migrate-up       - Apply all pending database migrations"
	@echo "  make migrate-down     - Roll back the latest migration step (down 1)"
	@echo "  make migrate-version  - Check the current migration version and dirty status"
	@echo "  make migrate-force V=1- Force migration version to clear dirty state"
	@echo "  make seed             - Populate database with deterministic dev seed records"
	@echo "  make db-shell         - Open an interactive psql session in the db container"
	@echo "  make test             - Run backend unit tests"
	@echo "  make test-integration - Run backend integration tests"
	@echo "  make lint             - Run linters (golangci-lint / ESLint)"
	@echo "  make build            - Compile backend binary"
	@echo "  make run              - Run backend service locally"
	@echo "  make sqlc             - Generate Go code from SQL queries via sqlc"

## Start local docker-compose stack
docker-up:
	@echo "==> Starting ASESAMO Docker services..."
	docker compose up -d
	@echo "==> Waiting for database readiness (max 30s)..."
	@ready=0; \
	for i in $$(seq 1 30); do \
		if docker compose exec -T db pg_isready -U "$$DB_USER" -d "$$DB_NAME" > /dev/null 2>&1; then \
			ready=1; \
			break; \
		fi; \
		sleep 1; \
	done; \
	if [ $$ready -ne 1 ]; then \
		echo "✖ Error: Database failed to become ready within 30 seconds."; \
		exit 1; \
	fi
	@echo "✔ Docker services started and database is ready."

## Stop docker-compose stack (preserves volumes)
docker-down:
	@echo "==> Stopping ASESAMO Docker services..."
	docker compose down --remove-orphans
	@echo "✔ Containers stopped cleanly."

## Stop docker-compose stack and purge persistent volumes
docker-down-v:
	@echo "==> Stopping ASESAMO services and purging volumes..."
	docker compose down -v --remove-orphans
	@echo "✔ Services stopped and volumes purged."

## Run pending migrations up
migrate-up:
	@echo "==> Executing database migrations UP..."
	docker compose run --rm migrate up
	@echo "✔ Migrations UP completed."

## Roll back the last migration step
migrate-down:
	@echo "==> Reverting last migration step (DOWN 1)..."
	docker compose run --rm migrate down 1
	@echo "✔ Migration DOWN 1 completed."

## Check current migration version
migrate-version:
	@echo "==> Checking migration version..."
	docker compose run --rm migrate version

## Clear dirty state by forcing specific version (Usage: make migrate-force V=1)
migrate-force: export V := $(V)
migrate-force:
	@if [ -z "$$V" ] || ! echo "$$V" | grep -qE '^[0-9]+$$'; then \
		echo "✖ Error: Please specify a numeric version, e.g. make migrate-force V=1"; \
		exit 1; \
	fi
	@echo "==> Forcing migration version to $$V..."
	docker compose run --rm migrate force "$$V"
	@echo "✔ Migration version forced to $$V."

## Seed database with deterministic test records
seed:
	@echo "==> Seeding database with deterministic dev/test data..."
	@if docker compose ps db --format '{{.State}}' 2>/dev/null | grep -q "running"; then \
		docker compose exec -T -e PGPASSWORD="$$DB_PASSWORD" db psql -U "$$DB_USER" -d "$$DB_NAME" < db/seeds/000001_seed_dev_data.sql; \
	else \
		echo "✖ Error: db container is not running. Run 'make docker-up' first."; \
		exit 1; \
	fi
	@echo "✔ Database seeded successfully."

## Open interactive psql shell
db-shell:
	docker compose exec -e PGPASSWORD="$$DB_PASSWORD" db psql -U "$$DB_USER" -d "$$DB_NAME"

## Run backend unit tests
test:
	@if [ -d "asesamo-api" ]; then \
		echo "==> Running backend unit tests..."; \
		(cd asesamo-api && go test -v -race -cover ./...); \
	else \
		echo "asesamo-api directory not found."; \
	fi

## Run backend integration tests
test-integration:
	@if [ -d "asesamo-api" ]; then \
		echo "==> Running backend integration tests..."; \
		(cd asesamo-api && go test -tags=integration -v ./...); \
	else \
		echo "asesamo-api directory not found."; \
	fi

## Run linters
lint:
	@if [ -d "asesamo-api" ]; then \
		echo "==> Linting asesamo-api..."; \
		(cd asesamo-api && if command -v golangci-lint >/dev/null 2>&1; then golangci-lint run ./...; else go vet ./...; fi); \
	fi
	@if [ -d "asesamo-web" ]; then \
		echo "==> Linting asesamo-web..."; \
		(cd asesamo-web && if [ -f "package.json" ]; then npm run lint || true; fi); \
	fi

## Compile backend binary
build:
	@if [ -d "asesamo-api" ]; then \
		echo "==> Building asesamo-api binary..."; \
		(cd asesamo-api && go build -o bin/api cmd/api/main.go); \
	else \
		echo "asesamo-api directory not found."; \
	fi

## Run backend locally
run:
	@if [ -d "asesamo-api" ]; then \
		echo "==> Running asesamo-api..."; \
		(cd asesamo-api && go run cmd/api/main.go); \
	else \
		echo "asesamo-api directory not found."; \
	fi

## Regenerate sqlc code
sqlc:
	@if [ -d "asesamo-api" ] && command -v sqlc >/dev/null 2>&1; then \
		echo "==> Generating sqlc code..."; \
		(cd asesamo-api && sqlc generate); \
	else \
		echo "sqlc not found or asesamo-api directory missing."; \
	fi
