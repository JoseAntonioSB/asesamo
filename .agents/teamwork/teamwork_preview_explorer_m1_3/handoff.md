# Handoff Report: Database Migration Container Execution, Makefile Targets & Idempotency Strategy

**Milestone:** M1 — Database & Environment Scaffolding  
**Agent:** teamwork_preview_explorer_m1_3  
**Date:** 2026-10-01  
**Working Directory:** `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m1_3`

---

## 1. Observation

Direct observations from the repository codebase, architecture documentation, and tooling definitions:

1. **Database Schema & Migrations Location:**
   - Existing migration files are located at:
     - `/Users/jasbotello/devel/asesamo/db/migrations/000001_init.up.sql` (441 lines)
     - `/Users/jasbotello/devel/asesamo/db/migrations/000001_init.down.sql` (32 lines)
   - `000001_init.up.sql` declares extensions (`pgcrypto`, `citext`), trigger function `set_updated_at()`, 13 tables (`tenants`, `plants`, `access_points`, `users`, `refresh_tokens`, `contractors`, `plant_contractors`, `contractor_workers`, `document_types`, `plant_document_requirements`, `documents`, `access_logs`, `notifications`), 6 custom ENUM types, and inserts default system document types (`DC3`, `REPSE`, `SUA`, `SAT_32D`, `INE`, `MEDICAL`) at lines 271-277.
   - `000001_init.down.sql` contains `DROP TABLE IF EXISTS` in reverse dependency order (lines 7-19), drops all 6 custom ENUM types (lines 21-26), drops `set_updated_at()` (line 28), and drops extensions `citext` and `pgcrypto` (lines 30-31).

2. **Architectural Guidelines (`DEVELOPMENT_RULES.md`):**
   - Line 18: `Migraciones: golang-migrate (archivos .sql numerados 000001_description.up.sql)`.
   - Line 42: `Entorno local: Docker Compose (backend + DB + herramientas auxiliares en contenedores)`.
   - Lines 163-169: `Cada migración es idempotente en el .up.sql cuando sea posible... Nunca se modifica una migración ya aplicada en producción; siempre se crea una nueva.`
   - Lines 301-314: Defines expected Makefile targets:
     ```makefile
     make run              # Levanta el servidor local
     make build            # Compila el binario
     make test             # Tests unitarios
     make test-integration # Tests de integración (requiere Docker)
     make migrate-up       # Aplica migraciones pendientes
     make migrate-down     # Revierte la última migración
     make sqlc             # Regenera el código sqlc desde los .sql
     make lint             # Corre golangci-lint
     make docker-up        # Levanta docker-compose (DB + app)
     make docker-down      # Baja docker-compose
     ```
   - Lines 331-344: `docker-compose.yml` must include `api`, `web`, `db` (PostgreSQL 16), and `migrate` (container that runs golang-migrate on startup).

3. **Docker Compose & Tooling Constraints:**
   - On Docker Hub, the official image repository for `golang-migrate` is `migrate/migrate:v4.17.0` (`https://hub.docker.com/r/migrate/migrate`), whereas `golang-migrate/migrate` is the upstream GitHub project name.
   - `db` healthcheck is defined in `survey_report.md` (lines 81-86) as `test: ["CMD-SHELL", "pg_isready -U $${POSTGRES_USER} -d $${POSTGRES_DB}"]`.
   - If `container_name: asesamo-migrate` is hardcoded on the migration service, executing ad-hoc CLI commands via `docker compose run --rm migrate <args>` can trigger container name collision errors against an existing stopped container.
   - If `command` in `docker-compose.yml` holds all flags (`-path`, `-database`, `-verbose`, `up`), running `docker compose run --rm migrate down 1` replaces the entire command, causing `migrate` to fail because `-path` and `-database` are omitted. Setting `entrypoint` to `/migrate -path=/migrations -database=... -verbose` and `command: ["up"]` enables both automatic boot execution and seamless CLI subcommand arguments (`down 1`, `version`, `force <v>`).

---

## 2. Logic Chain

1. **From Observation 1 and 2 (Schema and Project Requirements):**
   - The system requires automatic migration execution upon stack boot, manual migration rollback capabilities, and deterministic database seeding.
   - Because `000001_init.up.sql` establishes 13 tables with intricate foreign keys and default catalog rows, migrations must run strictly after PostgreSQL is ready, and before the backend API attempts connection pool initialization.

2. **From Observation 3 (Docker Hub Registry Naming & Portability):**
   - The official Docker Hub image is `migrate/migrate:v4.17.0`. Using this image tag guarantees that any environment running standard Docker can pull the image without namespace errors.

3. **From Observation 2 & 3 (Readiness Synchronization):**
   - Relying on `depends_on: [db]` alone causes race conditions because PostgreSQL takes several seconds after container launch to initialize shared buffers and accept connections.
   - Gating `migrate` behind `depends_on: { db: { condition: service_healthy } }` guarantees that `migrate` executes only when `pg_isready` returns exit code 0.

4. **From Observation 3 (Command Overriding & CLI Portability):**
   - Configuring `entrypoint` in `docker-compose.yml` with the persistent flags (`-path=/migrations`, `-database=...`, `-verbose`) and setting `command: ["up"]` allows `docker compose run --rm migrate down 1` to append `["down", "1"]` to the entrypoint.
   - This eliminates the need for host machines to install `golang-migrate`, Go, or `psql` locally, enabling developers on macOS, Linux, or CI to run `make migrate-up`, `make migrate-down`, and `make seed` out-of-the-box.

5. **From Observation 1 & 2 (Idempotency and Rollback Symmetry):**
   - `golang-migrate` records applied migrations in `schema_migrations`. Subsequent runs of `migrate up` detect version 1 and exit 0 with `no change`, verifying forward idempotency.
   - Because `000001_init.down.sql` drops tables in strict reverse foreign-key order, custom ENUMs, triggers, and extensions, executing `make migrate-down` followed by `make migrate-up` restores a clean schema from zero without residual artifacts or locks.
   - In `000001_seed_dev_data.sql`, every row insertion must use `ON CONFLICT DO NOTHING` to guarantee that executing `make seed` multiple times is idempotent and safe.

6. **From Observation 2 & 3 (Clean Shutdown):**
   - Ephemeral tasks (`migrate`, `seeder`) must declare `restart: "no"` to avoid infinite restart loops.
   - `make docker-down` issues `docker compose down --remove-orphans`, giving daemon containers (`db`, `api`, `web`) 10 seconds to process `SIGTERM`, checkpoint databases, and exit cleanly without data corruption, while preserving persistent named volume `asesamo_pgdata`.
   - `make docker-down-v` purges `asesamo_pgdata` when a total zero-state rebuild is required.

---

## 3. Caveats

1. **Docker Daemon Availability in Sandbox:** During this exploration turn, the local Docker daemon socket was not running in the sandbox environment. This investigation is strictly read-only and verified through static code and architecture analysis. When builders implement M1, the Docker daemon must be running.
2. **Host vs Docker Database Port Conflict:** If a developer runs a local PostgreSQL server on port 5432 outside Docker, port mapping `5432:5432` will conflict. The Makefile is designed so all commands (`migrate-up`, `migrate-down`, `seed`, `db-shell`) execute via Docker container networking (`asesamo-net` -> `db:5432`), ensuring independence from host port collisions.
3. **Dirty Migration State Recovery:** If an unexpected power loss or killed process occurs during migration execution, `golang-migrate` flags `dirty = true` in `schema_migrations`. To avoid developer blockage, `make migrate-force V=1` has been included in the Makefile specification.

---

## 4. Conclusion

The strategy for database migration execution, Makefile automation, and idempotency validation is finalized and fully specified for builders:

1. **Docker Compose `migrate` Service:**
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
2. **Root `Makefile`:** Fully defined with required targets (`docker-up`, `docker-down`, `migrate-up`, `migrate-down`, `seed`) plus resilience targets (`docker-down-v`, `migrate-version`, `migrate-force`, `db-shell`, `help`).
3. **Idempotency & Clean Shutdown:** Fully documented test workflow covering UP idempotency (`no change` -> exit 0), seed idempotency (`ON CONFLICT DO NOTHING`), rollback symmetry (DOWN 1 -> UP), and graceful SIGTERM handling.

---

## 5. Verification Method

To independently verify the implementation once builders assemble the M1 files:

### Step 1: File Existence Verification
Inspect presence of key configuration files:
- `docker-compose.yml`
- `Makefile`
- `db/migrations/000001_init.up.sql`
- `db/migrations/000001_init.down.sql`
- `db/seeds/000001_seed_dev_data.sql`

### Step 2: Stack Boot & Health Verification
```bash
make docker-up
docker compose ps
# Expected: db is healthy; migrate and seeder exited with status 0.
```

### Step 3: Migration Idempotency Test
```bash
make migrate-up
# Expected: Outputs "no change", returns exit code 0.
```

### Step 4: Seed Idempotency Test
```bash
make seed
# Expected: Executes successfully without primary key or unique constraint errors.
```

### Step 5: Rollback and Full Re-apply Cycle Test
```bash
make migrate-down
# Expected: Reverts migration 000001 cleanly.
make migrate-up
make seed
# Expected: Re-applies migration 000001 and seeds database cleanly.
```

### Step 6: Clean Shutdown Test
```bash
make docker-down
# Expected: All containers stop and are removed without SIGKILL timeouts.
```

### Invalidation Conditions:
- `migrate` container restarts continuously instead of exiting code 0.
- `make migrate-down` leaves orphaned tables or types in PostgreSQL.
- `make seed` fails with unique key violation when executed twice.
- `docker compose run --rm migrate down 1` fails due to missing `-path` or `-database` parameters.
