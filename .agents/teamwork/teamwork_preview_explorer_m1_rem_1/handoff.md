# Handoff Report — Milestone M1 Iteration 2 (Remediation)

**Agent**: `teamwork_preview_explorer_m1_rem_1`  
**Role**: Teamwork Explorer (`investigation`, `synthesis`)  
**Milestone**: M1 Iteration 2 — Remediation Formulation  
**Working Directory**: `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m1_rem_1`  
**Target Recipient**: `ccf64bbc-f322-4ca8-b0be-0ee59db22a99` (teamwork_preview_orchestrator)  
**Date**: 2026-10-01  

---

## 1. Observation

### 1.1 `docker-compose.yml` Web and API Port Desynchronization
- In `/Users/jasbotello/devel/asesamo/docker-compose.yml`:
  - Lines 118–134 (`web` service):
    ```yaml
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
    ```
  - Lines 79–103 (`api` service):
    ```yaml
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
    ```
- Tool command executed: `PORT=4000 APP_PORT=9090 docker compose --env-file /dev/null config`
  - Output observed:
    ```yaml
    api:
      environment:
        APP_PORT: "9090"
      ports:
        - mode: ingress
          published: "9090"
          target: 8080
      healthcheck:
        test: [CMD, wget, --no-verbose, --tries=1, --spider, http://localhost:8080/api/v1/health]
    web:
      environment:
        PORT: "4000"
      ports:
        - mode: ingress
          published: "4000"
          target: 3000
      healthcheck:
        test: [CMD-SHELL, wget --no-verbose --tries=1 --spider http://localhost:3000 || exit 1]
    ```

### 1.2 `docker-compose.yml` Seeder Unquoted Shell Expansion
- In `/Users/jasbotello/devel/asesamo/docker-compose.yml` lines 63–66 (`seeder` service):
  ```yaml
  entrypoint:
    - sh
    - -c
    - PGPASSWORD=$${POSTGRES_PASSWORD} psql -h db -U $${POSTGRES_USER} -d $${POSTGRES_DB} -f /seeds/000001_seed_dev_data.sql
  ```
  The variables `$${POSTGRES_PASSWORD}`, `$${POSTGRES_USER}`, and `$${POSTGRES_DB}` lack double quotes.

### 1.3 `.env` and `.env.example` Variable Definitions
- In `/Users/jasbotello/devel/asesamo/.env` and `/Users/jasbotello/devel/asesamo/.env.example`:
  - Lines 23–24:
    ```ini
    APP_ENV=development
    APP_PORT=8080
    ```
  - Lines 43–44:
    ```ini
    PORT=3000
    NODE_ENV=development
    ```
  Neither file defines `API_PORT` or `WEB_PORT` to distinguish host port forwarding from container application listeners.

### 1.4 Verification Tool Executions
- `git apply --check .agents/teamwork/teamwork_preview_explorer_m1_rem_1/remediation_docker_compose.patch .agents/teamwork/teamwork_preview_explorer_m1_rem_1/remediation_env.patch .agents/teamwork/teamwork_preview_explorer_m1_rem_1/remediation_env_example.patch`
  - Result: Returned exit code 0 (all 3 patches apply cleanly).
- Python test script `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m1_rem_1/test_patch.py` executing `docker compose -f proposed_docker-compose.yml config --format json` under `PORT=4000 APP_PORT=9090 WEB_PORT=4000 API_PORT=9090`:
  - Result: Returned exit code 0; `target_port == env_port` confirmed for both services, healthcheck targets confirmed.

---

## 2. Logic Chain

1. **Web Port Coupling Failure**:
   - *Observation*: `docker-compose.yml` lines 119 and 125 use `${PORT:-3000}` simultaneously for the host published port and container `PORT`.
   - *Deduction*: Next.js reads `process.env.PORT` inside the container. If `PORT=4000` is present in the host environment, Next.js listens on container port 4000.
   - *Deduction*: The Docker port mapping forwards host traffic to container target port 3000. Because Next.js listens on 4000, port 3000 is unopened, returning `ECONNREFUSED`.
   - *Deduction*: The healthcheck `wget http://localhost:3000` fails, marking `web` permanently `unhealthy`.
   - *Remediation*: Mapping host port via `${WEB_PORT:-3000}:3000` and locking container environment `PORT: 3000` decouples host publishing from container runtime. Next.js always binds to 3000, aligning with target port 3000 and the healthcheck at `http://localhost:3000`.

2. **API Port Coupling Failure**:
   - *Observation*: `docker-compose.yml` lines 80 and 90 use `${APP_PORT:-8080}` for both host mapping and container `APP_PORT`.
   - *Deduction*: If `APP_PORT=9090` is set on the host, the Go binary binds to 9090 inside the container.
   - *Deduction*: Docker routes incoming host traffic to container port 8080, which is unopened.
   - *Deduction*: The healthcheck `wget http://localhost:8080/api/v1/health` and inter-service call `INTERNAL_API_URL: http://api:8080/api/v1` both fail because container port 8080 is dead.
   - *Remediation*: Mapping host port via `${API_PORT:-8080}:8080` and locking container environment `APP_PORT: 8080` guarantees the Go binary always listens on 8080. Host port remapping via `API_PORT` functions seamlessly without breaking container healthchecks or internal network communication.

3. **Seeder Argument Splitting**:
   - *Observation*: `docker-compose.yml` line 66 expands `$${POSTGRES_USER}` without quotes in `sh -c`.
   - *Deduction*: If `POSTGRES_USER` contains whitespace or evaluates to empty, POSIX shell shifts arguments, interpreting `-d` as the username.
   - *Remediation*: Quoting variables as `PGPASSWORD="$${POSTGRES_PASSWORD}" psql -h db -U "$${POSTGRES_USER}" -d "$${POSTGRES_DB}"` prevents word splitting and handles special characters safely.

4. **Configuration Clarity**:
   - *Observation*: `.env` and `.env.example` only defined `APP_PORT` and `PORT`.
   - *Deduction*: Adding `API_PORT=8080` and `WEB_PORT=3000` with clear comments establishes an unambiguous contract between Docker Compose host mappings and host binary development.

---

## 3. Caveats

1. **Docker Daemon Status**: The Docker daemon was not running during this turn. Static analysis, schema AST parsing via `docker compose config --format json`, and POSIX shell expansion validation were performed.
2. **Makefile Hardening Scope**: The prompt specifically requested remediation of `docker-compose.yml`, `.env`, and `.env.example`. Challenger M1_1 also flagged Makefile shell injection and unbounded retry loops; these are included as advisory artifacts (`advisory_makefile_remediation.patch` and `proposed_Makefile`) for the orchestrator and builder to adopt at their discretion.

---

## 4. Conclusion

The root causes of Challenger M1_1's rejection are fully resolved by four targeted remediations. All patch specifications have been generated, dry-run tested with `git apply --check`, and validated against container configuration schemas.

### Remediation Diff Summary:

#### A. `docker-compose.yml`
```diff
--- a/docker-compose.yml
+++ b/docker-compose.yml
@@ -63,7 +63,7 @@
     entrypoint:
       - sh
       - -c
-      - PGPASSWORD=$${POSTGRES_PASSWORD} psql -h db -U $${POSTGRES_USER} -d $${POSTGRES_DB} -f /seeds/000001_seed_dev_data.sql
+      - PGPASSWORD="$${POSTGRES_PASSWORD}" psql -h db -U "$${POSTGRES_USER}" -d "$${POSTGRES_DB}" -f /seeds/000001_seed_dev_data.sql
     networks:
       - asesamo-net
 
@@ -79,7 +79,7 @@
     restart: unless-stopped
     ports:
-      - "${APP_PORT:-8080}:8080"
+      - "${API_PORT:-8080}:8080"
     depends_on:
       db:
         condition: service_healthy
@@ -88,3 +88,3 @@
     environment:
       APP_ENV: ${APP_ENV:-development}
-      APP_PORT: ${APP_PORT:-8080}
+      APP_PORT: 8080
       DATABASE_URL: postgres://${DB_USER:-asesamo_admin}:${DB_PASSWORD:-asesamo_secret_dev}@db:5432/${DB_NAME:-asesamo_db}?sslmode=disable
@@ -118,3 +118,3 @@
     ports:
-      - "${PORT:-3000}:3000"
+      - "${WEB_PORT:-3000}:3000"
     depends_on:
@@ -124,3 +124,3 @@
     environment:
       NODE_ENV: production
-      PORT: ${PORT:-3000}
+      PORT: 3000
       NEXT_PUBLIC_API_URL: ${NEXT_PUBLIC_API_URL:-http://localhost:8080/api/v1}
```

#### B. `.env` and `.env.example`
```diff
--- a/.env
+++ b/.env
@@ -23,2 +23,4 @@
 APP_ENV=development
+API_PORT=8080
 APP_PORT=8080
@@ -43,2 +45,4 @@
+WEB_PORT=3000
 PORT=3000
 NODE_ENV=development
```

---

## 5. Verification Method

To independently verify the remediations, run the following commands:

1. **Verify Patches Apply Cleanly**:
   ```bash
   git apply --check .agents/teamwork/teamwork_preview_explorer_m1_rem_1/remediation_docker_compose.patch
   git apply --check .agents/teamwork/teamwork_preview_explorer_m1_rem_1/remediation_env.patch
   git apply --check .agents/teamwork/teamwork_preview_explorer_m1_rem_1/remediation_env_example.patch
   ```

2. **Verify Challenger M1_1 Test 1 Compliance**:
   ```bash
   python3 /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m1_rem_1/test_patch.py
   ```

3. **Verify Host Port Override Resistance**:
   ```bash
   PORT=4000 APP_PORT=9090 WEB_PORT=4000 API_PORT=9090 docker compose -f /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m1_rem_1/proposed_docker-compose.yml config --format json | jq '.services.web.environment.PORT, .services.api.environment.APP_PORT'
   # Expected output:
   # "3000"
   # "8080"
   ```

### Invalidation Conditions:
- If Next.js or Go services bind to variable host ports inside container environments rather than fixed ports `3000` and `8080`.
- If container healthchecks test ports other than the fixed container listening ports `3000` and `8080`.
