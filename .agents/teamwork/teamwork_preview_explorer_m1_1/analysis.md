# Technical Analysis: Database & Environment Scaffolding (Milestone M1)

**Agent**: `teamwork_preview_explorer_m1_1`  
**Date**: 2026-10-01  
**Working Directory**: `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m1_1`  
**Related Documents**: `ORIGINAL_REQUEST.md`, `PROJECT.md`, `DEVELOPMENT_RULES.md`, `db/migrations/000001_init.up.sql`

---

## 1. Executive Summary

Milestone M1 establishes the foundational data and runtime environment for ASESAMO Phase 1 MVP. This analysis formulates the concrete, deterministic architecture for:
1. The **PostgreSQL 16 base container service** (`db`) in `docker-compose.yml`.
2. The **reliable container healthcheck** validating database readiness without race conditions.
3. The **clean, guaranteed enablement of PostgreSQL extensions** (`pgcrypto` and `citext`).
4. The **persistent storage and network architecture** (`asesamo_pgdata` volume and `asesamo-net` bridge).
5. The **environment variable management strategy** (`.env.example` and `.env`), supporting both local host execution (e.g., `go test`, `make`, psql CLI) and containerized multi-service execution.

---

## 2. PostgreSQL 16 Service Architecture (`db`)

### 2.1 Container Specifications

| Attribute | Specification | Rationale |
|---|---|---|
| **Image** | `postgres:16-alpine` | Official PostgreSQL 16 image based on Alpine Linux. Minimal footprint (~40 MB compressed vs ~150 MB Debian), reduced attack surface, rapid startup, and includes precompiled `contrib` extensions (`pgcrypto`, `citext`). |
| **Container Name** | `asesamo-db` | Deterministic, explicit naming simplifying CLI diagnostics (`docker logs asesamo-db`, `docker exec -it asesamo-db psql ...`). |
| **Restart Policy** | `unless-stopped` | Maintains database availability during developer reboots or transient crashes while allowing deliberate `docker compose down`. |
| **Port Mapping** | `5432:5432` (or `${DB_PORT:-5432}:5432`) | Exposes standard PostgreSQL port to the host machine. Essential for host-based development: `psql`, GUI clients (TablePlus/DBeaver), local Go test suites (`make test-integration`), and host migration scripts. |
| **Network** | `asesamo-net` | Dedicated bridge network providing isolated DNS service discovery for all stack containers (`db`, `migrate`, `seeder`, `api`, `web`). |
| **Volume Mount** | `asesamo_pgdata:/var/lib/postgresql/data` | Named volume ensuring database persistence across container teardowns and restarts, while supporting total resets via `docker compose down -v`. |

### 2.2 Container Environment Variables

The official PostgreSQL container requires the following environment variables:

```yaml
environment:
  POSTGRES_USER: ${DB_USER:-asesamo_admin}
  POSTGRES_PASSWORD: ${DB_PASSWORD:-asesamo_secret_dev}
  POSTGRES_DB: ${DB_NAME:-asesamo_db}
  PGDATA: /var/lib/postgresql/data/pgdata
```

- **`POSTGRES_USER`**: Defaults to `asesamo_admin`. Initialized with `SUPERUSER` privileges, permitting extension creation and schema alterations.
- **`POSTGRES_PASSWORD`**: Defaults to `asesamo_secret_dev`. Set via environment variable to avoid hardcoded credentials in production configurations.
- **`POSTGRES_DB`**: Defaults to `asesamo_db`. Automatically created upon container initialization.
- **`PGDATA`**: Configured to `/var/lib/postgresql/data/pgdata` (a subdirectory inside the volume mount). This prevents filesystem conflicts with volume mount point metadata or `lost+found` directories.

---

## 3. Healthcheck Specification & Race Condition Prevention

### 3.1 Healthcheck Definition

A robust healthcheck is critical because dependent containers (`migrate`, `seeder`, `api`) must not start until PostgreSQL has completed its internal initialization and is actively accepting SQL connections.

```yaml
healthcheck:
  test: ["CMD-SHELL", "pg_isready -U $$POSTGRES_USER -d $$POSTGRES_DB"]
  interval: 3s
  timeout: 3s
  retries: 5
  start_period: 10s
```

### 3.2 Deep Dive: Syntax & Variable Escaping (`$$POSTGRES_USER`)

1. **Docker Compose Escaping (`$$`)**:
   - In Docker Compose YAML, single `$` denotes host environment interpolation at compose parse time (e.g., `${DB_USER}`).
   - Double `$$` (`$$POSTGRES_USER` and `$$POSTGRES_DB`) escapes the symbol, passing a literal `$POSTGRES_USER` and `$POSTGRES_DB` into the container's shell (`CMD-SHELL`).
   - Inside the container, `/bin/sh` resolves `$POSTGRES_USER` to the value defined in the container's own `environment:` block (`asesamo_admin`).
   - This ensures the healthcheck adapts dynamically if the user overrides credentials in `.env`, preventing broken healthchecks due to mismatched parameters.

2. **Why `pg_isready`**:
   - `pg_isready` is the official client utility bundled with PostgreSQL.
   - It sends a connection status check to the PostgreSQL server without requiring authentication credentials or passwords, returning exit code `0` when the server is accepting connections.
   - When run inside the container without `-h`, `pg_isready` connects through the local Unix domain socket (`/var/run/postgresql/.s.PGSQL.5432`), bypassing TCP network latency and DNS lookup overhead.

3. **Timing Parameters**:
   - `interval: 3s`: Frequent checks to minimize startup latency for downstream services.
   - `timeout: 3s`: Prevents hung checks from stalling the lifecycle.
   - `retries: 5`: Allows up to 15 seconds of connection attempts before marking the container unhealthy.
   - `start_period: 10s`: Grace period during which failed health checks do not count against the retry limit. This accommodates initial database cluster creation (`initdb`) on cold starts.

### 3.3 Dependency Chain Integration

The healthcheck enables declarative startup sequencing across the entire stack:

```yaml
db (healthy)
   └──> migrate (completes 000001_init.up.sql)
           └──> seeder (completes 000001_seed_dev_data.sql)
                   └──> api (boots Go server on :8080 & healthy)
                           └──> web (boots Next.js on :3000)
```

This dependency topology completely eliminates startup race conditions during `docker compose up --build`.

---

## 4. Extension Enablement Strategy (`pgcrypto` & `citext`)

### 4.1 Required Extensions

1. **`pgcrypto`**:
   - Required by schema `000001_init.up.sql` line 12.
   - Generates UUID primary keys (`gen_random_uuid()`) for all 13 tables (`tenants`, `plants`, `access_points`, `users`, `contractors`, etc.).
   - Provides native cryptographic functions `crypt()` and `gen_salt('bf', 10)` used in seed scripts (`db/seeds/000001_seed_dev_data.sql`) to generate industry-standard bcrypt password hashes directly inside SQL.
2. **`citext`**:
   - Required by schema `000001_init.up.sql` line 13.
   - Powers the `CITEXT` data type for `users.email` (line 114), enforcing case-insensitive email comparisons and unique constraints at the database engine level (preventing duplicate accounts differing only in casing, e.g., `guard@asesamo.com` vs `Guard@asesamo.com`).

### 4.2 Multi-Layered Clean Activation

To guarantee clean, idempotent extension enablement under all operating conditions, we establish a two-tier activation architecture:

#### Tier 1: Canonical Migration Execution (Primary)
The migration file `db/migrations/000001_init.up.sql` opens with:
```sql
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "citext";
```
- Because the `migrate` container connects as `POSTGRES_USER` (`asesamo_admin`), it possesses superuser privileges in PostgreSQL 16.
- The `IF NOT EXISTS` clause guarantees complete idempotency: it executes cleanly on initial setup and does nothing if the extension is already loaded.
- In `000001_init.down.sql`, rollback executes:
  ```sql
  DROP EXTENSION IF EXISTS "citext";
  DROP EXTENSION IF EXISTS "pgcrypto";
  ```
  This cleanly dismantles extensions during rollbacks and allows clean re-application during subsequent `up` cycles.

#### Tier 2: Entrypoint Initialization Mount (Belt & Suspenders)
When developers spin up only the `db` service (e.g., `docker compose up db` for ad-hoc inspection or local testing without running containerized migrations), they should still have extensions ready.
By mounting an initialization script into `/docker-entrypoint-initdb.d/`, PostgreSQL's standard entrypoint script runs it during the initial cluster bootstrap:

`db/init/01_init_extensions.sql`:
```sql
-- Ensure core extensions are present on fresh database creation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "citext";
```

Mount in `docker-compose.yml`:
```yaml
volumes:
  - asesamo_pgdata:/var/lib/postgresql/data
  - ./db/init:/docker-entrypoint-initdb.d:ro
```

**Why this combination is optimal**:
1. When cold booting the cluster for the first time, PostgreSQL executes `/docker-entrypoint-initdb.d/01_init_extensions.sql` as the superuser.
2. When `migrate` executes `000001_init.up.sql`, `CREATE EXTENSION IF NOT EXISTS` sees the extensions are already loaded and proceeds without error.
3. If a developer runs `make migrate-down` and `make migrate-up` locally, `000001_init.up.sql` recreates the extensions even if the container was not restarted.
4. Neither extension requires `shared_preload_libraries` or server restarts in PostgreSQL 16.

---

## 5. Environment Variable Architecture (`.env` and `.env.example`)

### 5.1 Taxonomy & Portability

The environment variable configuration balances two requirements:
1. **Host-vs-Container Connection Parity**: Tools running directly on the host (e.g., `psql -h localhost`, `go test ./...`, local Next.js `npm run dev`) connect to `localhost:5432` or `localhost:8080`. Containers on the Docker network (`asesamo-net`) connect to `db:5432` or `api:8080`.
2. **Zero-Configuration Onboarding**: Every variable has a documented default fallback (`${VAR:-default}`) in `docker-compose.yml`, so running `docker compose up` succeeds immediately even before a developer customizes `.env`.

### 5.2 Variable Matrix

| Variable | Default / Dev Value | Purpose | Consumer |
|---|---|---|---|
| `DB_USER` | `asesamo_admin` | PostgreSQL username | `db`, `migrate`, `seeder`, `api` |
| `DB_PASSWORD` | `asesamo_secret_dev` | PostgreSQL user password | `db`, `migrate`, `seeder`, `api` |
| `DB_NAME` | `asesamo_db` | PostgreSQL database name | `db`, `migrate`, `seeder`, `api` |
| `DB_PORT` | `5432` | Exposed host port for PostgreSQL | `db`, host tools (`psql`, GUI) |
| `DATABASE_URL` | `postgres://asesamo_admin:asesamo_secret_dev@localhost:5432/asesamo_db?sslmode=disable` | Connection URL for host processes | Host tests, CLI migration tools |
| `DOCKER_DATABASE_URL` | `postgres://asesamo_admin:asesamo_secret_dev@db:5432/asesamo_db?sslmode=disable` | Connection URL for Docker container network | `migrate`, `seeder`, `api` container |
| `APP_ENV` | `development` | Application environment (`development` / `production`) | `api` (Go backend) |
| `APP_PORT` | `8080` | Port for Go REST API | `api`, host port mapping |
| `JWT_SECRET` | `super_secret_jwt_key_min_32_characters_long_12345` | HMAC-SHA256 secret key (≥ 32 chars per rules) | `api` (JWT issuance & validation) |
| `CORS_ALLOWED_ORIGINS`| `http://localhost:3000,http://127.0.0.1:3000,http://web:3000` | Whitelisted CORS origins | `api` chi middleware |
| `R2_ACCOUNT_ID` | `mock_account` | Cloudflare R2 Account ID (Dev mock) | `api` R2 storage adapter |
| `R2_ACCESS_KEY_ID` | `mock_key` | Cloudflare R2 Access Key | `api` R2 storage adapter |
| `R2_SECRET_ACCESS_KEY`| `mock_secret` | Cloudflare R2 Secret Access Key | `api` R2 storage adapter |
| `R2_BUCKET_NAME` | `asesamo-dev` | Cloudflare R2 Bucket name | `api` R2 storage adapter |
| `PORT` | `3000` | Port for Next.js web application | `web`, host port mapping |
| `NODE_ENV` | `development` | Node environment | `web` |
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080/api/v1` | Public client API URL accessed by browser | `web` client-side React code |
| `INTERNAL_API_URL` | `http://api:8080/api/v1` | Internal container API URL accessed by SSR | `web` Server Components / Server Actions |

---

## 6. Proposed File Contents

### 6.1 Target File: `.env.example`

```dotenv
# ============================================================
# ASESAMO — Environment Variables Reference
# Copy to .env for local development: cp .env.example .env
# ============================================================

# ------------------------------------------------------------
# 1. Database (PostgreSQL 16)
# ------------------------------------------------------------
DB_USER=asesamo_admin
DB_PASSWORD=asesamo_secret_dev
DB_NAME=asesamo_db
DB_PORT=5432

# Database Connection URLs
# - DATABASE_URL: for tools running directly on the host machine (e.g. host tests, psql)
# - DOCKER_DATABASE_URL: for services communicating inside docker-compose (e.g. api, migrate)
DATABASE_URL=postgres://asesamo_admin:asesamo_secret_dev@localhost:5432/asesamo_db?sslmode=disable
DOCKER_DATABASE_URL=postgres://asesamo_admin:asesamo_secret_dev@db:5432/asesamo_db?sslmode=disable

# ------------------------------------------------------------
# 2. Backend API (asesamo-api - Go)
# ------------------------------------------------------------
APP_ENV=development
APP_PORT=8080

# JWT signing secret (MUST be at least 32 characters long per DEVELOPMENT_RULES.md)
JWT_SECRET=super_secret_jwt_key_min_32_characters_long_12345

# Allowed CORS origins (comma-separated)
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://127.0.0.1:3000,http://web:3000

# ------------------------------------------------------------
# 3. Cloudflare R2 Storage (Mock / Dev S3-compatible Credentials)
# ------------------------------------------------------------
R2_ACCOUNT_ID=mock_account
R2_ACCESS_KEY_ID=mock_key
R2_SECRET_ACCESS_KEY=mock_secret
R2_BUCKET_NAME=asesamo-dev

# ------------------------------------------------------------
# 4. Frontend Web (asesamo-web - Next.js 14)
# ------------------------------------------------------------
PORT=3000
NODE_ENV=development

# Browser client-side calls route to host:8080
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1

# Server-side calls (SSR / Next.js Server Actions) route inside docker network to api:8080
INTERNAL_API_URL=http://api:8080/api/v1
```

### 6.2 Target File: `.env`

The `.env` file should be generated with identical development values so that any clone of the repository can run `docker compose up` out of the box without manual copy steps.

### 6.3 Target File: `db/init/01_init_extensions.sql`

```sql
-- ============================================================
-- ASESAMO — PostgreSQL 16 Initial Extensions Bootstrap
-- Mounted to /docker-entrypoint-initdb.d/ to ensure extensions
-- exist immediately upon database cluster creation.
-- ============================================================

CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "citext";
```

### 6.4 Target File: `docker-compose.yml` (Base Database Specification)

```yaml
version: "3.9"

services:
  # ------------------------------------------------------------
  # Database: PostgreSQL 16 with pgcrypto & citext
  # ------------------------------------------------------------
  db:
    image: postgres:16-alpine
    container_name: asesamo-db
    restart: unless-stopped
    ports:
      - "${DB_PORT:-5432}:5432"
    environment:
      POSTGRES_USER: ${DB_USER:-asesamo_admin}
      POSTGRES_PASSWORD: ${DB_PASSWORD:-asesamo_secret_dev}
      POSTGRES_DB: ${DB_NAME:-asesamo_db}
      PGDATA: /var/lib/postgresql/data/pgdata
    volumes:
      - asesamo_pgdata:/var/lib/postgresql/data
      - ./db/init:/docker-entrypoint-initdb.d:ro
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U $$POSTGRES_USER -d $$POSTGRES_DB"]
      interval: 3s
      timeout: 3s
      retries: 5
      start_period: 10s
    networks:
      - asesamo-net

  # ------------------------------------------------------------
  # Database Migration Runner (golang-migrate)
  # Coordinated with Explorer M1_3
  # ------------------------------------------------------------
  migrate:
    image: migrate/migrate:v4.17.0
    container_name: asesamo-migrate
    restart: "no"
    depends_on:
      db:
        condition: service_healthy
    volumes:
      - ./db/migrations:/migrations:ro
    command:
      [
        "-path=/migrations",
        "-database",
        "postgres://${DB_USER:-asesamo_admin}:${DB_PASSWORD:-asesamo_secret_dev}@db:5432/${DB_NAME:-asesamo_db}?sslmode=disable",
        "up"
      ]
    networks:
      - asesamo-net

  # ------------------------------------------------------------
  # Database Seeder (Deterministic Dev & Test Data)
  # Coordinated with Explorer M1_2
  # ------------------------------------------------------------
  seeder:
    image: postgres:16-alpine
    container_name: asesamo-seeder
    restart: "no"
    depends_on:
      migrate:
        condition: service_completed_successfully
    volumes:
      - ./db/seeds:/seeds:ro
    environment:
      POSTGRES_USER: ${DB_USER:-asesamo_admin}
      POSTGRES_PASSWORD: ${DB_PASSWORD:-asesamo_secret_dev}
      POSTGRES_DB: ${DB_NAME:-asesamo_db}
    entrypoint:
      [
        "sh",
        "-c",
        "PGPASSWORD=$${POSTGRES_PASSWORD} psql -h db -U $${POSTGRES_USER} -d $${POSTGRES_DB} -f /seeds/000001_seed_dev_data.sql"
      ]
    networks:
      - asesamo-net

  # ------------------------------------------------------------
  # Backend REST API (asesamo-api - Go Clean Architecture)
  # Activated in M2 / M5
  # ------------------------------------------------------------
  api:
    build:
      context: ./asesamo-api
      dockerfile: Dockerfile
    container_name: asesamo-api
    restart: unless-stopped
    ports:
      - "${APP_PORT:-8080}:8080"
    depends_on:
      db:
        condition: service_healthy
      migrate:
        condition: service_completed_successfully
      seeder:
        condition: service_completed_successfully
    environment:
      APP_ENV: ${APP_ENV:-development}
      APP_PORT: ${APP_PORT:-8080}
      DATABASE_URL: postgres://${DB_USER:-asesamo_admin}:${DB_PASSWORD:-asesamo_secret_dev}@db:5432/${DB_NAME:-asesamo_db}?sslmode=disable
      JWT_SECRET: ${JWT_SECRET:-super_secret_jwt_key_min_32_characters_long_12345}
      CORS_ALLOWED_ORIGINS: ${CORS_ALLOWED_ORIGINS:-http://localhost:3000,http://127.0.0.1:3000,http://web:3000}
      R2_ACCOUNT_ID: ${R2_ACCOUNT_ID:-mock_account}
      R2_ACCESS_KEY_ID: ${R2_ACCESS_KEY_ID:-mock_key}
      R2_SECRET_ACCESS_KEY: ${R2_SECRET_ACCESS_KEY:-mock_secret}
      R2_BUCKET_NAME: ${R2_BUCKET_NAME:-asesamo-dev}
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/api/v1/health"]
      interval: 5s
      timeout: 5s
      retries: 5
      start_period: 10s
    networks:
      - asesamo-net

  # ------------------------------------------------------------
  # Frontend Web Interface (asesamo-web - Next.js 14 App Router)
  # Activated in M4 / M5
  # ------------------------------------------------------------
  web:
    build:
      context: ./asesamo-web
      dockerfile: Dockerfile
      args:
        NEXT_PUBLIC_API_URL: ${NEXT_PUBLIC_API_URL:-http://localhost:8080/api/v1}
    container_name: asesamo-web
    restart: unless-stopped
    ports:
      - "${PORT:-3000}:3000"
    depends_on:
      api:
        condition: service_healthy
    environment:
      NODE_ENV: production
      PORT: ${PORT:-3000}
      NEXT_PUBLIC_API_URL: ${NEXT_PUBLIC_API_URL:-http://localhost:8080/api/v1}
      INTERNAL_API_URL: ${INTERNAL_API_URL:-http://api:8080/api/v1}
    healthcheck:
      test: ["CMD-SHELL", "wget --no-verbose --tries=1 --spider http://localhost:3000 || exit 1"]
      interval: 5s
      timeout: 5s
      retries: 5
      start_period: 15s
    networks:
      - asesamo-net

# ------------------------------------------------------------
# Volumes
# ------------------------------------------------------------
volumes:
  asesamo_pgdata:
    name: asesamo_pgdata

# ------------------------------------------------------------
# Networks
# ------------------------------------------------------------
networks:
  asesamo-net:
    name: asesamo-net
    driver: bridge
```

---

## 7. Security & Best Practices

1. **`.gitignore` Rules**:
   Since `.env` contains local secrets and tokens, it must be excluded from Git while preserving `.env.example`:
   ```gitignore
   # Environment files
   .env
   .env.local
   .env.*.local
   !.env.example
   ```
2. **Bcrypt Password Generation in SQL**:
   Using `crypt('Password123!', gen_salt('bf', 10))` in seed scripts relies on `pgcrypto`. This eliminates the need to compile external Go seed binaries or run manual hash pre-computation.
3. **Case-Insensitive Uniqueness via `citext`**:
   The `users.email` column uses `CITEXT`. This prevents case-manipulation authentication bypasses or duplicate records across authentication requests.
4. **Port Rebind Safety**:
   Exposing `${DB_PORT:-5432}` allows developers with an existing local PostgreSQL service running on port 5432 to change `DB_PORT=5433` in `.env` without modifying `docker-compose.yml`.

---

## 8. Coordination & Verification Plan

- **Coordination with `m1_2` (Seeds)**:
  `seeder` container definition uses `postgres:16-alpine` and mounts `./db/seeds:/seeds:ro`, executing `000001_seed_dev_data.sql` after `migrate` succeeds.
- **Coordination with `m1_3` (Migrations & Makefile)**:
  `migrate` container uses `migrate/migrate:v4.17.0` mounting `./db/migrations:/migrations:ro` and depends on `db: condition: service_healthy`.
- **Independent Verification**:
  1. Boot database only: `docker compose up -d db`
  2. Inspect healthcheck status: `docker compose ps` (should show `healthy` within 5-10s).
  3. Verify extensions: `docker compose exec db psql -U asesamo_admin -d asesamo_db -c "\dx"` (shows `pgcrypto` and `citext`).
  4. Test clean teardown and data volume wipe: `docker compose down -v`.
