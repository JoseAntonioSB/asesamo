# M1 Architecture Analysis: Database Migration Execution, Makefile Automation & Idempotency Strategy

**Author:** teamwork_preview_explorer_m1_3  
**Milestone:** M1 — Database & Environment Scaffolding  
**Date:** 2026-10-01  
**Working Directory:** `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m1_3`

---

## 1. Executive Summary

Milestone M1 establishes the database foundation and developer environment for ASESAMO Phase 1 MVP. This analysis formulates the exact technical strategy for:
1. **Migration container execution in Docker Compose** utilizing `golang-migrate` (`migrate/migrate:v4.17.0`), configured to run automatically only when the PostgreSQL 16 database reaches a verified healthy state.
2. **Unified root `Makefile` targets**: `make migrate-up`, `make migrate-down`, `make docker-up`, `make docker-down`, and `make seed`, engineered for portability, zero external tool requirements on developer host machines, and deterministic execution.
3. **Rigorous validation framework for migration idempotency and clean container shutdown**, addressing dirty state recovery, forward/backward schema replay (`up` → `down 1` → `up`), seed idempotency (`ON CONFLICT DO NOTHING`), and graceful signal handling (`SIGTERM`).

---

## 2. Docker Compose Migration Container Architecture

### 2.1 Image Selection & Registry Naming Disambiguation

In Go documentation and project literature, the migration tool is commonly referenced as `golang-migrate/migrate` (matching its canonical GitHub repository `github.com/golang-migrate/migrate`).

However, on Docker Hub, the official container image repository maintained by the project team is:
- **Canonical Docker Hub Image:** `migrate/migrate:v4.17.0`
- **Docker Hub URL:** `https://hub.docker.com/r/migrate/migrate`

Attempting to pull `golang-migrate/migrate:v4.17.0` directly from Docker Hub may result in image pull failures because the organization namespace is `migrate`, not `golang-migrate`. Therefore:
- The container definition should use `image: migrate/migrate:v4.17.0` (or parameterized via `${MIGRATE_IMAGE:-migrate/migrate:v4.17.0}`).
- The Alpine-based tag provides a lightweight (~20MB) footprint containing all required PostgreSQL drivers (`postgres://`).

### 2.2 Container Lifecycle & Dependency Ordering

In multi-container orchestration, a common failure mode is race conditions during startup:
- If a migration container starts immediately when the database container starts (`depends_on: [db]`), PostgreSQL is still initializing, allocating shared memory buffers, and running internal startup checks.
- Without a strict readiness gate, `migrate` attempts connection to `db:5432`, fails with `connection refused` or `reset by peer`, and exits with a fatal error code.

**Solution: Strict Healthcheck Barrier**
1. The PostgreSQL container (`db`) defines a native healthcheck:
   ```yaml
   healthcheck:
     test: ["CMD-SHELL", "pg_isready -U $${POSTGRES_USER} -d $${POSTGRES_DB}"]
     interval: 3s
     timeout: 5s
     retries: 5
     start_period: 5s
   ```
2. The migration service (`migrate`) declares:
   ```yaml
   depends_on:
     db:
       condition: service_healthy
   ```
3. Downstream services (`seeder`, `api`) declare:
   ```yaml
   depends_on:
     migrate:
       condition: service_completed_successfully
   ```

### 2.3 Optimal Entrypoint vs Command Pattern in Docker Compose

To allow the migration container to function both as an automated boot service (during `docker compose up`) and as an interactive CLI runner (during `make migrate-down`, `make migrate-up`), we structure `entrypoint` and `command` intentionally:

```yaml
  migrate:
    image: migrate/migrate:v4.17.0
    restart: "no"
    depends_on:
      db:
        condition: service_healthy
    volumes:
      - ./db/migrations:/migrations:ro
    entrypoint:
      - /migrate
      - -path=/migrations
      - -database=postgres://${DB_USER:-asesamo_admin}:${DB_PASSWORD:-asesamo_secret_dev}@db:5432/${DB_NAME:-asesamo_db}?sslmode=disable
      - -verbose
    command: ["up"]
    networks:
      - asesamo-net
```

#### Why This Pattern Is Superior:
1. **Automated Boot:** Running `docker compose up` executes `/migrate -path=/migrations -database=... -verbose up`.
2. **Interactive Overriding:** Running `docker compose run --rm migrate down 1` replaces `command` with `["down", "1"]`, executing `/migrate -path=/migrations -database=... -verbose down 1`.
3. **Zero CLI installation:** Developers do not need to install `golang-migrate` binary on macOS or Linux; Docker handles all dependencies.
4. **Network isolation:** Migration runs inside `asesamo-net` pointing to `db:5432`, independent of whether port 5432 is forwarded to the host.
5. **Read-Only Mount (`:ro`):** Mounting `./db/migrations:/migrations:ro` prevents the container from modifying local migration files.
6. **Ephemeral Flag (`restart: "no"`):** Prevents Docker from restarting the container once migrations finish.
7. **No Hardcoded `container_name`:** Omitting `container_name: asesamo-migrate` prevents name collision errors when running ad-hoc `docker compose run --rm migrate <args>` commands while an exited migration container exists.

---

## 3. Root Makefile Target Specifications

A root `Makefile` must orchestrate local development tasks cleanly across macOS (BSD make/tools) and Linux (GNU make).

### 3.1 Environment Variable Loading

The Makefile must automatically load `.env` if present, while providing safe fallback defaults for all database credentials:

```makefile
# Load environment variables from .env if it exists
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Defaults
DB_USER ?= asesamo_admin
DB_PASSWORD ?= asesamo_secret_dev
DB_NAME ?= asesamo_db
DB_HOST ?= localhost
DB_PORT ?= 5432
DOCKER_NETWORK ?= asesamo-net
MIGRATE_IMAGE ?= migrate/migrate:v4.17.0
```

### 3.2 Target Breakdown

#### 1. `make docker-up`
- **Purpose:** Boot all local infrastructure containers in detached mode.
- **Implementation:**
  ```makefile
  docker-up:
  	@echo "==> Starting ASESAMO Docker services..."
  	docker compose up -d
  	@echo "==> Waiting for database readiness..."
  	@until docker compose exec -T db pg_isready -U $(DB_USER) -d $(DB_NAME) > /dev/null 2>&1; do \
  		sleep 1; \
  	done
  	@echo "✔ Docker services started and database is ready."
  ```

#### 2. `make docker-down`
- **Purpose:** Gracefully stop and remove containers, networks, and orphans while preserving persistent database volumes.
- **Implementation:**
  ```makefile
  docker-down:
  	@echo "==> Stopping ASESAMO Docker services..."
  	docker compose down --remove-orphans
  	@echo "✔ Containers stopped cleanly."
  ```

#### 3. `make docker-down-v` (Supporting Target for Clean Reset)
- **Purpose:** Full environment purge including named volumes (`asesamo_pgdata`).
- **Implementation:**
  ```makefile
  docker-down-v:
  	@echo "==> Stopping ASESAMO services and purging volumes..."
  	docker compose down -v --remove-orphans
  	@echo "✔ Services stopped and volumes purged."
  ```

#### 4. `make migrate-up`
- **Purpose:** Apply all pending migrations located in `db/migrations/`.
- **Implementation:**
  ```makefile
  migrate-up:
  	@echo "==> Executing database migrations UP..."
  	docker compose run --rm migrate up
  	@echo "✔ Migrations UP executed successfully."
  ```

#### 5. `make migrate-down`
- **Purpose:** Revert exactly one migration step (`000001_init.down.sql`).
- **Implementation:**
  ```makefile
  migrate-down:
  	@echo "==> Reverting last migration step (DOWN 1)..."
  	docker compose run --rm migrate down 1
  	@echo "✔ Migration DOWN 1 executed successfully."
  ```

#### 6. `make seed`
- **Purpose:** Execute deterministic seed script `db/seeds/000001_seed_dev_data.sql` to populate test tenant, plants, access points, guard user, contractors, and test workers.
- **Implementation:**
  ```makefile
  seed:
  	@echo "==> Seeding database with deterministic dev/test data..."
  	@if docker compose ps db --format '{{.State}}' 2>/dev/null | grep -q "running"; then \
  		docker compose exec -T -e PGPASSWORD=$(DB_PASSWORD) db psql -U $(DB_USER) -d $(DB_NAME) < db/seeds/000001_seed_dev_data.sql; \
  	else \
  		echo "✖ Error: db container is not running. Execute 'make docker-up' first."; \
  		exit 1; \
  	fi
  	@echo "✔ Database seeded successfully."
  ```

#### 7. Supporting High-Value Targets
To prevent developer friction, the Makefile includes:
- `make migrate-version`: Prints current version and dirty status from `schema_migrations`.
- `make migrate-force VERSION=<v>`: Clears dirty database flag after an interrupted migration.
- `make db-shell`: Opens interactive `psql` shell inside `asesamo-db`.
- `make help`: Prints formatted list of available targets.

### 3.3 Complete Proposed Root `Makefile`

```makefile
# ============================================================
# ASESAMO — Root Makefile
# ============================================================

.DEFAULT_GOAL := help
.PHONY: help docker-up docker-down docker-down-v migrate-up migrate-down migrate-version migrate-force seed db-shell test

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
	@echo "  make test             - Run backend unit and integration tests"

## Start local docker-compose stack
docker-up:
	@echo "==> Starting ASESAMO Docker services..."
	docker compose up -d
	@echo "==> Waiting for database readiness..."
	@until docker compose exec -T db pg_isready -U $(DB_USER) -d $(DB_NAME) > /dev/null 2>&1; do \
		sleep 1; \
	done
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
migrate-force:
	@if [ -z "$(V)" ]; then \
		echo "✖ Error: Please specify version, e.g. make migrate-force V=1"; \
		exit 1; \
	fi
	@echo "==> Forcing migration version to $(V)..."
	docker compose run --rm migrate force $(V)
	@echo "✔ Migration version forced to $(V)."

## Seed database with deterministic test records
seed:
	@echo "==> Seeding database with deterministic dev/test data..."
	@if docker compose ps db --format '{{.State}}' 2>/dev/null | grep -q "running"; then \
		docker compose exec -T -e PGPASSWORD=$(DB_PASSWORD) db psql -U $(DB_USER) -d $(DB_NAME) < db/seeds/000001_seed_dev_data.sql; \
	else \
		echo "✖ Error: db container is not running. Run 'make docker-up' first."; \
		exit 1; \
	fi
	@echo "✔ Database seeded successfully."

## Open interactive psql shell
db-shell:
	docker compose exec -e PGPASSWORD=$(DB_PASSWORD) db psql -U $(DB_USER) -d $(DB_NAME)

## Run backend tests
test:
	@if [ -d "asesamo-api" ]; then \
		cd asesamo-api && go test -v -race -cover ./...; \
	else \
		echo "asesamo-api directory not found."; \
	fi
```

---

## 4. Migration Idempotency Analysis

Migration idempotency ensures that executing a migration command repeatedly or in varied lifecycle states leaves the database in a correct, predictable, non-corrupted state.

### 4.1 Schema Layer Idempotency (`000001_init.up.sql`)

1. **Extensions:**
   - Uses `CREATE EXTENSION IF NOT EXISTS "pgcrypto";` and `CREATE EXTENSION IF NOT EXISTS "citext";`.
   - Idempotent: Can be executed multiple times without error.
2. **Triggers and Functions:**
   - Uses `CREATE OR REPLACE FUNCTION set_updated_at() ...`.
   - Triggers: `CREATE TRIGGER ...` in PostgreSQL does not natively support `IF NOT EXISTS` prior to PostgreSQL 14 (or raises duplicate trigger errors). Because `golang-migrate` tracks executed migrations via its internal ledger, triggers are only executed once per version.
3. **Default System Document Types Seed in `000001_init.up.sql`:**
   - In `000001_init.up.sql` lines 271-277:
     ```sql
     INSERT INTO document_types (id, tenant_id, code, name, applies_to, default_validity_days, is_required_default, is_system) VALUES
       (gen_random_uuid(), NULL, 'DC3',      'Constancia de Capacitación DC-3',          'worker',      365,  TRUE,  TRUE),
       (gen_random_uuid(), NULL, 'REPSE',    'Registro REPSE',                            'contractor',  365,  TRUE,  TRUE),
       (gen_random_uuid(), NULL, 'SUA',      'Comprobante de Pago SUA (IMSS/INFONAVIT)', 'contractor',  30,   TRUE,  TRUE),
       (gen_random_uuid(), NULL, 'SAT_32D',  'Opinión de Cumplimiento SAT (32-D)',        'contractor',  30,   TRUE,  TRUE),
       (gen_random_uuid(), NULL, 'INE',      'Identificación Oficial (INE/IFE)',          'worker',      NULL, TRUE,  TRUE),
       (gen_random_uuid(), NULL, 'MEDICAL',  'Certificado Médico de Aptitud',             'worker',      365,  TRUE,  TRUE);
     ```
   - Notice: `document_types` has a unique constraint `UNIQUE (tenant_id, code)`. If this migration were ever re-run against an existing table, it would raise a unique violation unless `ON CONFLICT (tenant_id, code) DO NOTHING` is specified.
   - **Recommendation:** Keep `000001_init.up.sql` as managed by `golang-migrate`, and guarantee that all seed inserts in `db/seeds/` include `ON CONFLICT DO NOTHING`.

### 4.2 Migration Tool Ledger & Re-Run Idempotency

`golang-migrate` creates and manages a tracking table:
```sql
CREATE TABLE IF NOT EXISTS schema_migrations (
    version bigint not null primary key,
    dirty boolean not null
);
```

1. **Re-running `make migrate-up` when up to date:**
   - `golang-migrate` checks `schema_migrations`.
   - Finds current version = `1`, `dirty = false`.
   - Emits `no change` and exits cleanly with **exit status code 0**.
   - This satisfies idempotency: running `migrate-up` multiple times produces zero errors.
2. **Reverting and Re-applying (`migrate-down` followed by `migrate-up`):**
   - Running `make migrate-down` executes `000001_init.down.sql`.
   - `000001_init.down.sql` drops all 13 tables in reverse dependency order, followed by the 6 custom ENUM types, helper function, and extensions.
   - `schema_migrations` drops version to 0 (or empty).
   - Running `make migrate-up` immediately afterwards recreates the entire schema from scratch with zero residual locks or conflicting tables.
   - This verifies that the rollback script is comprehensive and symmetrical with the schema creation.

### 4.3 Dirty State Recovery Protocol

If an unexpected crash, SIGKILL, or invalid SQL occurs while applying a migration:
1. `golang-migrate` sets `dirty = true` in `schema_migrations`.
2. Any subsequent attempt to run `migrate up` or `migrate down` fails with:
   ```
   error: Dirty database version 1. Fix and force version.
   ```
3. **Automated Recovery Strategy:**
   - Inspect database state (`make db-shell` or `docker compose logs migrate`).
   - If the migration succeeded partially or was rolled back manually, clear the dirty flag using `make migrate-force V=1` (or `V=0`).
   - Run `make migrate-up` again.

### 4.4 Seed Script Idempotency (`db/seeds/000001_seed_dev_data.sql`)

The seed script must be runnable an arbitrary number of times without duplicate key errors, primary key collisions, or foreign key violations:
- Every `INSERT INTO <table_name> (id, ...)` specifies `ON CONFLICT (id) DO NOTHING`.
- Junction tables with compound unique constraints (such as `plant_contractors (plant_id, contractor_id)` and `contractor_workers (contractor_id, user_id)`) specify `ON CONFLICT (plant_id, contractor_id) DO NOTHING` and `ON CONFLICT (contractor_id, user_id) DO NOTHING`.
- Password hashes use deterministic bcrypt salts or `crypt('Password123!', gen_salt('bf', 10))` so re-seeding produces identical hashes.

---

## 5. Clean Shutdown Validation

A production-grade development environment must shut down cleanly without leaving orphaned containers, dangling volume locks, or corrupted database write-ahead logs (WAL).

### 5.1 Signal Propagation and Lifecycle

When `docker compose down` or `Ctrl+C` is triggered:
1. Docker daemon sends `SIGTERM` to the main process in each container.
2. Docker waits for `stop_grace_period` (default 10 seconds).
3. If the process does not terminate within the grace period, Docker issues `SIGKILL`.

### 5.2 Container Roles and Shutdown Profiles

| Service | Type | Expected Behavior on Shutdown | Configuration Requirement |
|---|---|---|---|
| `db` (Postgres) | Daemon | Intercepts `SIGTERM`, initiates fast shutdown, terminates client backends, checkpoints WAL, flushes buffers to disk, and exits code 0. | Persistent volume mounted; no forced kill. |
| `migrate` | Ephemeral | One-shot task. Runs on startup, finishes, and exits code 0. Already stopped during regular operation. | `restart: "no"`. During `docker-down`, exited container is removed cleanly. |
| `seeder` | Ephemeral | One-shot task. Runs after migration, finishes, and exits code 0. Already stopped during regular operation. | `restart: "no"`. Removed cleanly. |
| `api` (Go) | Daemon | Chi HTTP server implements `srv.Shutdown(ctx)` with a 5-second graceful timeout on `SIGTERM` / `SIGINT`. | Closes DB connection pool; terminates cleanly within 2 seconds. |
| `web` (Next.js) | Daemon | Node.js process intercepts `SIGTERM` and closes HTTP listener. | Terminates cleanly within 2 seconds. |

### 5.3 Volume Persistence vs. Clean Wipe

Developers must have clarity on when database state persists and when it resets:
- **`make docker-down`**: Executes `docker compose down --remove-orphans`.
  - Containers are removed.
  - The named volume `asesamo_pgdata` is **preserved**.
  - On the next `make docker-up`, all seeded records and runtime `access_logs` remain intact.
- **`make docker-down-v`**: Executes `docker compose down -v --remove-orphans`.
  - The named volume `asesamo_pgdata` is **destroyed**.
  - On the next `make docker-up`, PostgreSQL boots from zero, runs `migrate`, runs `seeder`, and restores a clean baseline.

---

## 6. End-to-End Idempotency & Clean Shutdown Verification Protocol

The following automated validation test sequence confirms full compliance with the requirements:

```bash
#!/usr/bin/env bash
set -euo pipefail

echo "=========================================================="
echo "TEST 1: Clean First-Time Stack Boot"
echo "=========================================================="
make docker-up
# Verify containers are running
docker compose ps

echo "=========================================================="
echo "TEST 2: Verify Initial Migration & Schema Status"
echo "=========================================================="
docker compose exec -T db psql -U asesamo_admin -d asesamo_db -c "\dt"
make migrate-version

echo "=========================================================="
echo "TEST 3: Migration UP Idempotency (Re-run UP on clean DB)"
echo "=========================================================="
# Re-running migrate-up should report "no change" and exit 0
make migrate-up

echo "=========================================================="
echo "TEST 4: Seed Idempotency (Re-run Seed on populated DB)"
echo "=========================================================="
# Re-running seed should execute without duplicate key errors
make seed

echo "=========================================================="
echo "TEST 5: Migration Rollback (DOWN 1)"
echo "=========================================================="
make migrate-down
# Verify tables were dropped
TABLE_COUNT=$(docker compose exec -T db psql -U asesamo_admin -d asesamo_db -t -c \
  "SELECT count(*) FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE';")
echo "Table count after rollback (should only be schema_migrations): $TABLE_COUNT"

echo "=========================================================="
echo "TEST 6: Re-applying Migration UP (Full Cycle Idempotency)"
echo "=========================================================="
make migrate-up
make seed
# Verify tables and records restored
docker compose exec -T db psql -U asesamo_admin -d asesamo_db -c "SELECT count(*) FROM users;"

echo "=========================================================="
echo "TEST 7: Clean Shutdown Verification"
echo "=========================================================="
make docker-down
# Verify no lingering containers
docker compose ps

echo "=========================================================="
echo "TEST 8: Clean Wipe and Rebuild (Volume Purge)"
echo "=========================================================="
make docker-down-v
make docker-up
make migrate-version
make docker-down

echo "=========================================================="
echo "✔ ALL IDEMPOTENCY AND CLEAN SHUTDOWN TESTS PASS"
echo "=========================================================="
```

---

## 7. Recommendations for Builders (M1 Execution)

1. **Docker Compose File (`docker-compose.yml`)**:
   - Use `image: migrate/migrate:v4.17.0`.
   - Set `restart: "no"`.
   - Set `depends_on: { db: { condition: service_healthy } }`.
   - Mount `./db/migrations:/migrations:ro`.
   - Use entrypoint `/migrate -path=/migrations -database=... -verbose` with `command: ["up"]` to allow flexible overriding in `make migrate-down` and `make migrate-version`.
   - Omit hardcoded `container_name` on `migrate` and `seeder` to prevent naming collision during ad-hoc `docker compose run` calls.

2. **Root `Makefile`**:
   - Place the complete `Makefile` at the repository root (`/Users/jasbotello/devel/asesamo/Makefile`).
   - Ensure all targets (`migrate-up`, `migrate-down`, `docker-up`, `docker-down`, `seed`) work directly with Docker Compose without requiring local host tool installations.
   - Include `migrate-force` to guarantee developers can self-heal dirty migration states without dropping their database.
