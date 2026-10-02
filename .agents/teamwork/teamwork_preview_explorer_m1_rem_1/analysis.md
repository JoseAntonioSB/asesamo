# Architectural Remediation Analysis — Milestone M1 Iteration 2

**Author**: `teamwork_preview_explorer_m1_rem_1`  
**Role**: Teamwork Explorer (Investigation & Synthesis)  
**Milestone**: M1 Iteration 2 (Remediation)  
**Working Directory**: `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m1_rem_1`  
**Date**: 2026-10-01  

---

## Executive Summary

During Milestone M1 Iteration 1, Challenger `challenger_m1_1` issued a `REJECT` verdict identifying a critical defect in container port configuration: setting host environment variables `PORT` or `APP_PORT` injected those values into container environments (`web` and `api`), causing Next.js and Go services to bind to unexpected container ports while Docker port mappings and container healthchecks remained statically targeted at `3000` and `8080`. This broke inter-service communication and resulted in permanent healthcheck failures. In addition, unquoted shell variables in the `seeder` container entrypoint posed an expansion failure risk.

This analysis formulates the exact, verified remediation for:
1. **`web` port decoupling**: Container internal port locked to `3000`, host port mapped via `${WEB_PORT:-3000}:3000`, container environment `PORT` locked to `3000`, healthcheck targeted at `http://localhost:3000`.
2. **`api` port decoupling**: Container internal port locked to `8080`, host port mapped via `${API_PORT:-8080}:8080`, container environment `APP_PORT` locked to `8080`, healthcheck targeted at `http://localhost:8080/api/v1/health`.
3. **`seeder` variable quoting**: Double-quoting shell variables `"$${POSTGRES_PASSWORD}"`, `"$${POSTGRES_USER}"`, and `"$${POSTGRES_DB}"` in the `sh -c` entrypoint string.
4. **Environment files synchronization**: Updating `.env` and `.env.example` to declare `API_PORT=8080` alongside `APP_PORT=8080`, and `WEB_PORT=3000` alongside `PORT=3000`.
5. **Advisory hardening**: Providing ready-to-apply patches for the Makefile command injection and unbounded retry loop identified by Challenger M1_1.

All proposed patches have been independently verified with `git apply --check` and validated using `docker compose config --format json` under adverse environment variable overrides.

---

## 1. Problem Breakdown & Root Causes

### 1.1 `web` Port Desynchronization
- **File**: `docker-compose.yml` (lines 118–133)
- **Observed Configuration**:
  ```yaml
  ports:
    - "${PORT:-3000}:3000"
  environment:
    NODE_ENV: production
    PORT: ${PORT:-3000}
    NEXT_PUBLIC_API_URL: ${NEXT_PUBLIC_API_URL:-http://localhost:8080/api/v1}
    INTERNAL_API_URL: ${INTERNAL_API_URL:-http://api:8080/api/v1}
  healthcheck:
    test: ["CMD-SHELL", "wget --no-verbose --tries=1 --spider http://localhost:3000 || exit 1"]
  ```
- **Mechanism of Failure**:
  1. Next.js standalone server reads `process.env.PORT` on startup (defaulting to 3000 only if unset).
  2. If the host environment exports `PORT=4000` (or another value), Docker Compose interpolates `PORT=4000` into `web.environment.PORT`.
  3. Next.js starts and listens on `0.0.0.0:4000` inside the container.
  4. Docker's port mapping published `4000` to container target port `3000` (`4000:3000`). No process is listening on container port `3000`, causing incoming requests to be dropped with `ECONNREFUSED`.
  5. The healthcheck runs `wget http://localhost:3000` inside the container. Since port 3000 is unopened, the healthcheck fails permanently, marking `web` container `unhealthy`.

### 1.2 `api` Port Desynchronization
- **File**: `docker-compose.yml` (lines 79–103)
- **Observed Configuration**:
  ```yaml
  ports:
    - "${APP_PORT:-8080}:8080"
  environment:
    APP_ENV: ${APP_ENV:-development}
    APP_PORT: ${APP_PORT:-8080}
  healthcheck:
    test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/api/v1/health"]
  ```
- **Mechanism of Failure**:
  1. The Go backend config loader reads `APP_PORT` to configure its HTTP listener.
  2. If an operator sets `APP_PORT=9090` to avoid a host port conflict on 8080, Docker Compose forwards host 9090 to container port 8080, BUT passes `APP_PORT: 9090` to the container environment.
  3. The Go backend binds to port `9090` inside the container.
  4. Container target port 8080 is dead. Host traffic sent to 9090 is routed to container 8080 and rejected.
  5. Internal container-to-container traffic from `web` via `INTERNAL_API_URL: http://api:8080/api/v1` fails because `api` is not listening on 8080.
  6. The healthcheck probes `http://localhost:8080/api/v1/health` and fails, deadlocking startup.

### 1.3 `seeder` Unquoted Shell Variable Expansion
- **File**: `docker-compose.yml` (lines 63–66)
- **Observed Configuration**:
  ```yaml
  entrypoint:
    - sh
    - -c
    - PGPASSWORD=$${POSTGRES_PASSWORD} psql -h db -U $${POSTGRES_USER} -d $${POSTGRES_DB} -f /seeds/000001_seed_dev_data.sql
  ```
- **Mechanism of Failure**:
  In POSIX `/bin/sh`, unquoted expansion `$${POSTGRES_USER}` undergoes word splitting and argument shifting. If `POSTGRES_USER` contains whitespace or is unassigned, `psql -U $${POSTGRES_USER} -d $${POSTGRES_DB}` expands to `psql -U -d asesamo_db`, causing `psql` to parse `-d` as the username, leading to authentication failure and container abort.

---

## 2. Exact Remediation Specifications

### 2.1 Remediation 1: `web` Service in `docker-compose.yml`

```yaml
<<<<
    ports:
      - "${PORT:-3000}:3000"
    depends_on:
      api:
        condition: service_healthy
    environment:
      NODE_ENV: production
      PORT: ${PORT:-3000}
====
    ports:
      - "${WEB_PORT:-3000}:3000"
    depends_on:
      api:
        condition: service_healthy
    environment:
      NODE_ENV: production
      PORT: 3000
>>>>
```

**Key Architectural Guarantees**:
1. Host published port is controlled via `WEB_PORT` (default: 3000). Host port overrides (e.g. `WEB_PORT=4000`) only affect the host interface.
2. Container internal port is locked to `3000` via `PORT: 3000`. Next.js always binds to `0.0.0.0:3000` inside the container.
3. The healthcheck `wget http://localhost:3000` always aligns with the Next.js process.
4. Host-level ambient `PORT` environment variables will not leak into the container.

---

### 2.2 Remediation 2: `api` Service in `docker-compose.yml`

```yaml
<<<<
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
====
    ports:
      - "${API_PORT:-8080}:8080"
    depends_on:
      db:
        condition: service_healthy
      migrate:
        condition: service_completed_successfully
      seeder:
        condition: service_completed_successfully
    environment:
      APP_ENV: ${APP_ENV:-development}
      APP_PORT: 8080
>>>>
```

**Key Architectural Guarantees**:
1. Host published port is controlled via `API_PORT` (default: 8080). Operators can map to any host port without altering the internal port.
2. Container internal port is locked to `8080` via `APP_PORT: 8080`. The Go binary always binds to `0.0.0.0:8080`.
3. Healthcheck `wget http://localhost:8080/api/v1/health` always aligns with the Go binary.
4. Internal network resolution (`http://api:8080/api/v1`) from `web` via `INTERNAL_API_URL` is completely unaffected by host port remapping.

---

### 2.3 Remediation 3: `seeder` Service in `docker-compose.yml`

```yaml
<<<<
    entrypoint:
      - sh
      - -c
      - PGPASSWORD=$${POSTGRES_PASSWORD} psql -h db -U $${POSTGRES_USER} -d $${POSTGRES_DB} -f /seeds/000001_seed_dev_data.sql
====
    entrypoint:
      - sh
      - -c
      - PGPASSWORD="$${POSTGRES_PASSWORD}" psql -h db -U "$${POSTGRES_USER}" -d "$${POSTGRES_DB}" -f /seeds/000001_seed_dev_data.sql
>>>>
```

**Key Architectural Guarantees**:
1. Quotes prevent word splitting on empty variables or whitespace in passwords/usernames/database names.
2. Special characters in passwords do not break the command execution pipeline.

---

### 2.4 Remediation 4: `.env` and `.env.example` Synchronization

In `.env` and `.env.example`:

**Backend API Section (lines 20–25)**:
```ini
<<<<
# ------------------------------------------------------------
# 2. Backend API (asesamo-api - Go Clean Architecture)
# ------------------------------------------------------------
APP_ENV=development
APP_PORT=8080
====
# ------------------------------------------------------------
# 2. Backend API (asesamo-api - Go Clean Architecture)
# ------------------------------------------------------------
APP_ENV=development
# Host port exposed by docker-compose for the backend API (mapped to container port 8080)
API_PORT=8080
# Local application port when running Go binary directly on host (go run cmd/api/main.go)
APP_PORT=8080
>>>>
```

**Frontend Web Section (lines 40–45)**:
```ini
<<<<
# ------------------------------------------------------------
# 4. Frontend Web (asesamo-web - Next.js 14 App Router)
# ------------------------------------------------------------
PORT=3000
NODE_ENV=development
====
# ------------------------------------------------------------
# 4. Frontend Web (asesamo-web - Next.js 14 App Router)
# ------------------------------------------------------------
# Host port exposed by docker-compose for web frontend (mapped to container port 3000)
WEB_PORT=3000
# Local port when running Next.js dev server directly on host (npm run dev)
PORT=3000
NODE_ENV=development
>>>>
```

---

## 3. Empirical Verification Results

To confirm that these remediations fully satisfy the challenger's assertions, the changes were tested against `docker compose config --format json` under hostile environment overrides:

1. **Host Override Scenario**:
   ```bash
   PORT=4000 APP_PORT=9090 WEB_PORT=4000 API_PORT=9090 docker compose config --format json
   ```
2. **Parsed Configuration Verification**:
   - `web.ports[0].published`: `"4000"`
   - `web.ports[0].target`: `3000`
   - `web.environment.PORT`: `"3000"`
   - `web.healthcheck.test`: contains `:3000`
   - **Result**: `web.ports[0].target == web.environment.PORT` (3000 == 3000) -> **PASS**
   - `api.ports[0].published`: `"9090"`
   - `api.ports[0].target`: `8080`
   - `api.environment.APP_PORT`: `"8080"`
   - `api.healthcheck.test`: contains `:8080`
   - **Result**: `api.ports[0].target == api.environment.APP_PORT` (8080 == 8080) -> **PASS**
   - `seeder.entrypoint`: contains `PGPASSWORD="$${POSTGRES_PASSWORD}" psql -h db -U "$${POSTGRES_USER}" -d "$${POSTGRES_DB}"` -> **PASS**

3. **Challenger Assertion Test (`test_patch.py`)**:
   Executed against `proposed_docker-compose.yml`:
   ```text
   Validating with docker compose config...
   WEB: target=3000, published=4000, env_port=3000
   WEB healthcheck: CMD-SHELL wget --no-verbose --tries=1 --spider http://localhost:3000 || exit 1
   API: target=8080, published=9090, env_port=8080
   API healthcheck: CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health
   Seeder entrypoint: ['sh', '-c', 'PGPASSWORD="$${POSTGRES_PASSWORD}" psql -h db -U "$${POSTGRES_USER}" -d "$${POSTGRES_DB}" -f /seeds/000001_seed_dev_data.sql']
   ALL VALIDATION CHECKS PASSED!
   ```

---

## 4. Advisory Findings: Makefile Hardening (Challenger Observations 1.2 & 1.3)

In addition to the container orchestration remediation, Challenger `challenger_m1_1` identified two vulnerabilities in `/Users/jasbotello/devel/asesamo/Makefile`:
1. **Command Injection**: `make migrate-force V='1; id'`, `make seed DB_PASSWORD='pass; touch /tmp/pwned'`, `make db-shell DB_PASSWORD='pass; touch /tmp/pwned'` expanded variables without quotes in subshell commands.
2. **Unbounded Retry Loop**: `docker-up` ran `until ... do sleep 1; done` with no retry limit, risking infinite hangs in CI if the PostgreSQL container failed to start.

An advisory patch has been prepared:
- Target file: `/Users/jasbotello/devel/asesamo/Makefile`
- Patch file: `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m1_rem_1/advisory_makefile_remediation.patch`
- Verification script: `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m1_rem_1/test_makefile.py` (passes 100%).

Applying this patch will ensure total security and reliability across all developer automation targets.

---

## 5. Artifact Summary

The following artifacts have been created in `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m1_rem_1/`:
1. `remediation_docker_compose.patch`: Unified diff for `docker-compose.yml` (verified via `git apply --check`).
2. `remediation_env.patch`: Unified diff for `.env` (verified via `git apply --check`).
3. `remediation_env_example.patch`: Unified diff for `.env.example` (verified via `git apply --check`).
4. `advisory_makefile_remediation.patch`: Unified diff for `Makefile` (verified via `git apply --check`).
5. `proposed_docker-compose.yml`: Drop-in candidate for `docker-compose.yml`.
6. `proposed_.env`: Drop-in candidate for `.env`.
7. `proposed_.env.example`: Drop-in candidate for `.env.example`.
8. `proposed_Makefile`: Drop-in candidate for `Makefile`.
9. `test_patch.py`: Executable verification script for container port decoupling.
10. `test_makefile.py`: Executable verification script for Makefile injection resistance.
