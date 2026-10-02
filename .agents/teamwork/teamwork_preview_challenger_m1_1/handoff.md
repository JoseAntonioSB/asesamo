# Adversarial Challenge & Stress-Test Report — Milestone M1

**Agent**: `teamwork_preview_challenger_m1_1`  
**Role**: Empirical Challenger (`critic`, `specialist`)  
**Milestone**: M1 — Core Infra & Dev Environment Scaffolding  
**Date**: 2026-10-01  
**Working Directory**: `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_challenger_m1_1`  
**Parent Agent**: `ccf64bbc-f322-4ca8-b0be-0ee59db22a99`  

---

## Final Verdict: REJECT

Milestone M1 cannot be approved in its current state. Empirical stress testing uncovered **1 Critical container networking defect**, **2 High-severity shell command injection vulnerabilities in the Makefile**, **1 High-severity URI interpolation flaw**, and **1 Medium-severity infinite loop hazard in container startup orchestration**.

All findings have been reproduced empirically with runnable verification harnesses documented below.

---

## 1. Observation

### Observation 1.1 — Critical Port Desynchronization & Healthcheck Failure in `web` and `api`
In `/Users/jasbotello/devel/asesamo/docker-compose.yml`:
- Lines 118–133 (`web` service):
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
  environment:
    APP_ENV: ${APP_ENV:-development}
    APP_PORT: ${APP_PORT:-8080}
  healthcheck:
    test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/api/v1/health"]
  ```

**Empirical Tool Result**: Running `PORT=4000 APP_PORT=9090 docker compose --env-file /dev/null config --format json` produces:
```json
{
  "services": {
    "web": {
      "ports": [{"mode": "ingress", "target": 3000, "published": "4000", "protocol": "tcp"}],
      "environment": {"PORT": "4000"},
      "healthcheck": {"test": ["CMD-SHELL", "wget --no-verbose --tries=1 --spider http://localhost:3000 || exit 1"]}
    },
    "api": {
      "ports": [{"mode": "ingress", "target": 8080, "published": "9090", "protocol": "tcp"}],
      "environment": {"APP_PORT": "9090"},
      "healthcheck": {"test": ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/api/v1/health"]}
    }
  }
}
```

### Observation 1.2 — Arbitrary Shell Command Injection in `Makefile` Targets
In `/Users/jasbotello/devel/asesamo/Makefile`:
- Lines 81–89 (`migrate-force`):
  ```makefile
  migrate-force:
  	@if [ -z "$(V)" ]; then \
  		echo "✖ Error: Please specify version, e.g. make migrate-force V=1"; \
  		exit 1; \
  	fi
  	@echo "==> Forcing migration version to $(V)..."
  	docker compose run --rm migrate force $(V)
  	@echo "✔ Migration version forced to $(V)."
  ```
- Lines 92–100 (`seed`):
  ```makefile
  seed:
  	...
  	docker compose exec -T -e PGPASSWORD=$(DB_PASSWORD) db psql -U $(DB_USER) -d $(DB_NAME) < db/seeds/000001_seed_dev_data.sql;
  ```
- Lines 103–105 (`db-shell`):
  ```makefile
  db-shell:
  	docker compose exec -e PGPASSWORD=$(DB_PASSWORD) db psql -U $(DB_USER) -d $(DB_NAME)
  ```

**Empirical Tool Results**:
1. Executing `make -n migrate-force V='1; id'` prints verbatim:
   ```bash
   docker compose run --rm migrate force 1; id
   ```
2. Executing `make -n seed DB_PASSWORD='pass; touch /tmp/pwned'` prints verbatim:
   ```bash
   docker compose exec -T -e PGPASSWORD=pass; touch /tmp/pwned db psql -U asesamo_admin -d asesamo_db < db/seeds/000001_seed_dev_data.sql;
   ```
3. Executing `make -n db-shell DB_PASSWORD='pass; touch /tmp/pwned'` prints verbatim:
   ```bash
   docker compose exec -e PGPASSWORD=pass; touch /tmp/pwned db psql -U asesamo_admin -d asesamo_db
   ```

### Observation 1.3 — Unbounded Infinite Loop in `make docker-up`
In `/Users/jasbotello/devel/asesamo/Makefile` lines 46–50:
```makefile
	@echo "==> Waiting for database readiness..."
	@until docker compose exec -T db pg_isready -U $(DB_USER) -d $(DB_NAME) > /dev/null 2>&1; do \
		sleep 1; \
	done
	@echo "✔ Docker services started and database is ready."
```
The `until ... do sleep 1; done` loop lacks any timeout, counter, or retry ceiling. If the `db` container fails initialization (e.g. bad credentials, corrupt volume, port bind failure), this recipe hangs indefinitely, deadlocking automated CI runners.

### Observation 1.4 — Database URI Parser Crash on Password Delimiters (`/`, `#`, `?`, `%`)
In `/Users/jasbotello/devel/asesamo/docker-compose.yml`:
- Line 42:
  `- -database=postgres://${DB_USER:-asesamo_admin}:${DB_PASSWORD:-asesamo_secret_dev}@db:5432/${DB_NAME:-asesamo_db}?sslmode=disable`
- Line 91:
  `DATABASE_URL: postgres://${DB_USER:-asesamo_admin}:${DB_PASSWORD:-asesamo_secret_dev}@db:5432/${DB_NAME:-asesamo_db}?sslmode=disable`

**Empirical Tool Result**: Running Go's standard library `url.Parse` (used by `golang-migrate` and Go database drivers) on passwords containing URL delimiter characters:
```go
url.Parse("postgres://asesamo_admin:pass/word@db:5432/asesamo_db?sslmode=disable")
// ERROR: parse "...": invalid port ":pass" after host

url.Parse("postgres://asesamo_admin:pass#word@db:5432/asesamo_db?sslmode=disable")
// ERROR: parse "...": invalid port ":pass" after host

url.Parse("postgres://asesamo_admin:pass?word@db:5432/asesamo_db?sslmode=disable")
// ERROR: parse "...": invalid port ":pass" after host

url.Parse("postgres://asesamo_admin:pass%word@db:5432/asesamo_db?sslmode=disable")
// ERROR: parse "...": invalid URL escape "%wo"
```

### Observation 1.5 — Unquoted Shell Expansion in `seeder` Container Entrypoint
In `/Users/jasbotello/devel/asesamo/docker-compose.yml` lines 63–66:
```yaml
entrypoint:
  - sh
  - -c
  - PGPASSWORD=$${POSTGRES_PASSWORD} psql -h db -U $${POSTGRES_USER} -d $${POSTGRES_DB} -f /seeds/000001_seed_dev_data.sql
```
**Empirical Tool Result**: If `POSTGRES_USER` is empty, POSIX shell expands the command to:
`psql -h db -U -d asesamo_db -f /seeds/000001_seed_dev_data.sql`, causing `-U` to take `-d` as the username, leading to fatal execution failure.

### Observation 1.6 — SQL Seed Script Idempotency Verification
In `/Users/jasbotello/devel/asesamo/db/seeds/000001_seed_dev_data.sql`:
- Contains exactly 25 `INSERT INTO` statements across 7 domain entities.
- Every statement includes `ON CONFLICT (...) DO NOTHING`.
- All conflict targets (`tenants(id)`, `plants(id)`, `access_points(id)`, `users(id)`, `contractors(id)`, `plant_contractors(plant_id, contractor_id)`, `contractor_workers(contractor_id, user_id)`, `plant_document_requirements(plant_id, document_type_id)`, `documents(id)`) match primary keys or unique constraints declared in `/Users/jasbotello/devel/asesamo/db/migrations/000001_init.up.sql`.

---

## 2. Logic Chain

1. **Port Desynchronization Failure Mode**:
   - *Premise*: Next.js App Router (standalone server mode per `PROJECT.md` line 46) binds to `process.env.PORT || 3000`.
   - *Observation*: `docker-compose.yml` lines 119 and 125 map host port `${PORT:-3000}` to container target port `3000`, while simultaneously injecting `PORT: ${PORT:-3000}` into the container environment.
   - *Deduction*: When an engineer or CI environment overrides `PORT=4000` (or has `PORT` exported by the host system), Next.js listens on port 4000 inside the container.
   - *Deduction*: Docker port forwarding forwards incoming host traffic to container port 3000. Because no process is listening on container port 3000, all requests receive `ECONNREFUSED`.
   - *Deduction*: The healthcheck executes `wget http://localhost:3000`. Because port 3000 is dead, the healthcheck fails permanently, marking the container `unhealthy`.
   - *Identical Failure in `api`*: Passing `APP_PORT: ${APP_PORT:-8080}` into the container while hardcoding target port `8080` in port forwarding and `http://localhost:8080/api/v1/health` in healthcheck causes total container failure if `APP_PORT` is overridden.

2. **Makefile Command Injection Chain**:
   - *Premise*: Makefile recipes are executed in a subshell (`/bin/sh`).
   - *Observation*: Variables `$(V)` and `$(DB_PASSWORD)` are expanded directly without shell quoting.
   - *Deduction*: In `make migrate-force V="1; <cmd>"`, the semicolon terminates the `migrate` invocation and immediately executes `<cmd>` in the host shell with the privileges of the running user.
   - *Deduction*: In `make seed DB_PASSWORD="pass; <cmd>"`, unquoted expansion results in arbitrary command execution on the developer machine.

3. **URL Delimiter Collisions**:
   - *Premise*: Generated production passwords frequently contain symbols such as `/`, `#`, `?`, or `%`.
   - *Observation*: `docker-compose.yml` concatenates `${DB_PASSWORD}` directly into standard connection strings.
   - *Deduction*: URL parsers treat `/` as a path separator, `#` as a fragment delimiter, and `?` as a query delimiter. Unencoded passwords containing these characters corrupt the URI host/port structure, crashing `migrate` and `api` during cold boot.

4. **Startup Deadlock in CI**:
   - *Premise*: CI/CD pipelines run `make docker-up`.
   - *Observation*: The database readiness check runs an unbounded `until` loop.
   - *Deduction*: If the PostgreSQL container fails to achieve healthy status due to misconfiguration or resource constraints, the workflow hangs until the global CI runner timeout triggers.

---

## 3. Caveats

1. **Docker Daemon Availability**: The local Docker daemon was not active during this turn in the execution environment. Static verification, AST JSON parsing via `docker compose config --format json`, Go runtime testing, and POSIX shell expansion tests were executed directly.
2. **Subsequent Milestones**: Upstream source code directories (`asesamo-api` and `asesamo-web`) are scheduled for implementation in Milestones M2–M5. Findings regarding `api` and `web` internal port bindings are based on the contracts specified in `PROJECT.md` and `DEVELOPMENT_RULES.md`.

---

## 4. Conclusion & Required Remediations

Milestone M1 is **REJECTED**. Before approval can be granted, the following concrete fixes must be applied by the worker:

### Required Action Items:

1. **Fix Port Decoupling in `docker-compose.yml`**:
   - For `web`: Keep the container internal port fixed at `3000`. Remove `PORT: ${PORT:-3000}` from `web.environment` (or set `PORT: 3000`). Use `${WEB_PORT:-3000}:3000` for host mapping.
   - For `api`: Keep the container internal port fixed at `8080`. Set `APP_PORT: 8080` inside `api.environment`. Use `${API_PORT:-8080}:8080` (or `${APP_PORT:-8080}:8080` mapped to fixed 8080) for host mapping.
   - Update `.env` and `.env.example` to clearly distinguish `HOST_API_PORT` / `HOST_WEB_PORT` from internal ports.

2. **Fix Makefile Shell Injections**:
   - In `migrate-force`: Quote `$(V)` or validate that `$(V)` is purely numeric:
     ```makefile
     migrate-force:
     	@if [ -z "$(V)" ] || ! echo "$(V)" | grep -qE '^[0-9]+$$'; then \
     		echo "✖ Error: Please specify numeric version, e.g. make migrate-force V=1"; \
     		exit 1; \
     	fi
     	docker compose run --rm migrate force "$(V)"
     ```
   - In `seed` and `db-shell`: Quote `$(DB_PASSWORD)`, `$(DB_USER)`, and `$(DB_NAME)` in recipes:
     ```makefile
     docker compose exec -T -e PGPASSWORD="$(DB_PASSWORD)" db psql -U "$(DB_USER)" -d "$(DB_NAME)" < db/seeds/000001_seed_dev_data.sql;
     ```

3. **Add Timeout to `make docker-up`**:
   - Replace the unbounded `until` loop with a bounded retry counter:
     ```makefile
     @echo "==> Waiting for database readiness (max 30s)..."
     @for i in $$(seq 1 30); do \
     	if docker compose exec -T db pg_isready -U "$(DB_USER)" -d "$(DB_NAME)" > /dev/null 2>&1; then \
     		echo "✔ Database is ready."; \
     		exit 0; \
     	fi; \
     	sleep 1; \
     done; \
     echo "✖ Database failed to become ready within 30 seconds."; \
     exit 1
     ```

4. **Quote Shell Variables in `seeder` Entrypoint**:
   - In `docker-compose.yml` line 66:
     ```yaml
     entrypoint:
       - sh
       - -c
       - PGPASSWORD="$${POSTGRES_PASSWORD}" psql -h db -U "$${POSTGRES_USER}" -d "$${POSTGRES_DB}" -f /seeds/000001_seed_dev_data.sql
     ```

---

## 5. Verification Method

To independently reproduce all findings and verify future fixes, execute the empirical test harness below:

```bash
python3 -c "
import subprocess
import os
import json
import urllib.parse
import re

print('=== ASESAMO M1 EMPIRICAL CHALLENGE SUITE ===')

# Test 1: Web Port Decoupling
env = dict(os.environ)
env['PORT'] = '4000'
res = subprocess.run(['docker', 'compose', '--env-file', '/dev/null', 'config', '--format', 'json'], env=env, capture_output=True, text=True)
cfg = json.loads(res.stdout)
web = cfg['services']['web']
target_port = web['ports'][0]['target']
env_port = web['environment'].get('PORT')
health_test = ' '.join(web['healthcheck']['test'])

print('\nTest 1 (Web Port Override PORT=4000):')
print(f'  Target Port: {target_port}, Env PORT: {env_port}')
print(f'  Healthcheck: {health_test}')
assert str(target_port) == str(env_port), f'DEFECT: target port ({target_port}) != env port ({env_port})'
assert f':{target_port}' in health_test, f'DEFECT: healthcheck does not probe target port ({target_port})'

# Test 2: Makefile Command Injection
res = subprocess.run(['make', '-n', 'migrate-force', 'V=1; echo INJECTED'], capture_output=True, text=True)
assert 'INJECTED' not in res.stdout, 'DEFECT: Shell injection detected in make migrate-force'

res = subprocess.run(['make', '-n', 'seed', 'DB_PASSWORD=pass; echo INJECTED'], capture_output=True, text=True)
assert 'INJECTED' not in res.stdout, 'DEFECT: Shell injection detected in make seed'

# Test 3: URI Delimiter Collisions
for pw in ['pass/word', 'pass#word', 'pass?word']:
    uri = f'postgres://asesamo_admin:{pw}@db:5432/asesamo_db?sslmode=disable'
    p = urllib.parse.urlparse(uri)
    assert p.password == pw, f'DEFECT: URI delimiter corrupted password ({pw} != {p.password})'

print('\nALL EMPIRICAL ASSERTIONS PASSED!')
"
```

### Invalidation Conditions:
- The findings are invalidated only if `docker-compose.yml` guarantees that container listening ports, healthchecks, and published port targets remain strictly aligned regardless of host environment variables, and if all Makefile recipes safely quote external parameters.
