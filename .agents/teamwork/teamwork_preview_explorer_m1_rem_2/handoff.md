# Handoff Report — Milestone M1 Iteration 2 (Makefile Remediation)

**Agent**: `teamwork_preview_explorer_m1_rem_2`  
**Role**: Teamwork Explorer (Read-Only Investigation & Remediation Architecture)  
**Parent Agent**: `ccf64bbc-f322-4ca8-b0be-0ee59db22a99`  
**Target File**: `/Users/jasbotello/devel/asesamo/Makefile`  
**Patch Artifact**: `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m1_rem_2/makefile_remediation.patch`  
**Analysis Artifact**: `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m1_rem_2/analysis.md`  
**Date**: 2026-10-01  

---

## 1. Observation

Direct observations from `/Users/jasbotello/devel/asesamo/Makefile` (lines 43–105) and empirical tool outputs:

### 1.1 Unvalidated and Unquoted `migrate-force`
Lines 81–89 in `Makefile`:
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
- Tool observation: Running `make -n migrate-force V='1; id'` produced:
  ```text
  docker compose run --rm migrate force 1; id
  ```
  Where `; id` is parsed by `/bin/sh` as an unquoted command delimiter.
- Tool observation: When testing naive recipe-level validation `@if [ -z "$(V)" ] || ! echo "$(V)" | grep -qE '^[0-9]+$$'; then ...`:
  Supplying `V='1" || id || "'` caused quote breakout on `[ -z "1" || id || "" ]`, executing `id` in the subshell:
  ```text
  uid=501(jasbotello) gid=20(staff) groups=20(staff)...
  ```
- Tool observation: Running Challenger's static test `assert 'INJECTED' not in res.stdout` on `make -n migrate-force V='1; echo INJECTED'` fails if `$(V)` is expanded directly in the recipe string, because Make textual macro expansion places `1; echo INJECTED` into the dry-run output.

### 1.2 Unquoted Database Credentials in `seed` and `db-shell`
Lines 92–105 in `Makefile`:
```makefile
seed:
	@echo "==> Seeding database with deterministic dev/test data..."
	@if docker compose ps db --format '{{.State}}' 2>/dev/null | grep -q "running"; then \
		docker compose exec -T -e PGPASSWORD=$(DB_PASSWORD) db psql -U $(DB_USER) -d $(DB_NAME) < db/seeds/000001_seed_dev_data.sql; \
	else \
		echo "✖ Error: db container is not running. Run 'make docker-up' first."; \
		exit 1; \
	fi
	@echo "✔ Database seeded successfully."

db-shell:
	docker compose exec -e PGPASSWORD=$(DB_PASSWORD) db psql -U $(DB_USER) -d $(DB_NAME)
```
- Tool observation: Running `make -n seed DB_PASSWORD='pass; touch /tmp/pwned'` expanded to:
  ```text
  docker compose exec -T -e PGPASSWORD=pass; touch /tmp/pwned db psql -U asesamo_admin -d asesamo_db < db/seeds/000001_seed_dev_data.sql;
  ```
  Where `; touch /tmp/pwned` is an unquoted shell command.
- Tool observation: If passwords contain double quotes (e.g. `pass" ; id ; echo "`), naive double-quoting in recipe text (`PGPASSWORD="$(DB_PASSWORD)"`) suffers quote breakout.

### 1.3 Unbounded Startup Deadlock in `docker-up`
Lines 43–51 in `Makefile`:
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
- Tool observation: The loop `until ... do sleep 1; done` has zero retry counters and zero timeout. If the database container fails or crashes, this recipe loops indefinitely.

---

## 2. Logic Chain

1. **Subshell Macro Expansion Vulnerability**:
   - *Premise*: GNU Make expands `$(VAR)` textually before constructing the shell command string passed to `/bin/sh -c`.
   - *Observation (1.1, 1.2)*: Without quoting and without format validation, metacharacters (`;`, `&`, `|`, `` ` ``, `$()`) become active shell control tokens.
   - *Observation (1.1)*: Even if quotes are added in the recipe text (`"$(VAR)"`), an attacker payload containing `"` breaks the quotation boundary.
   - *Deduction*: True defense requires decoupling variable transport from shell text interpolation. Passing parameters via the process environment array (`export VAR := $(VAR)`) and referencing them in the recipe as POSIX environment variables (`$$VAR`) guarantees that `/bin/sh` treats them strictly as data parameters, making quote breakout and command execution impossible.

2. **Regex Validation of Numeric Parameters**:
   - *Premise*: Migration version `V` in `migrate-force` must always be a non-negative integer.
   - *Observation (1.1)*: Checking `echo "$$V" | grep -qE '^[0-9]+$$'` strictly validates that `V` contains only digits `0-9`.
   - *Deduction*: Any injection payload or invalid format (`1; id`, `abc`, `1.0`, `-1`, empty string) fails regex validation, triggering `exit 1` before invoking `docker compose`.
   - *Observation (1.1)*: Because `make -n` outputs the recipe text containing `"$V"` rather than the expanded value, Challenger's check `assert 'INJECTED' not in res.stdout` passes.

3. **Global Subshell Export for Database Variables**:
   - *Premise*: `DB_USER`, `DB_PASSWORD`, and `DB_NAME` are used in `seed`, `db-shell`, and `docker-up`.
   - *Observation (1.2)*: Declaring `export DB_USER DB_PASSWORD DB_NAME` in `Makefile` causes Make to export them to the environment of all recipe subshells.
   - *Deduction*: Using `PGPASSWORD="$$DB_PASSWORD"` and `-U "$$DB_USER" -d "$$DB_NAME"` allows passwords with special characters (semicolons, dollar signs, quotes) to be passed safely without shell command parsing.

4. **Bounded Loop for Service Readiness**:
   - *Premise*: Readiness probes in developer tooling and CI workflows must terminate deterministically.
   - *Observation (1.3)*: Replacing `until ... done` with `for i in $$(seq 1 30); do ... done` provides an exact 30-second bounded ceiling with immediate break on readiness (`ready=1; break;`).
   - *Deduction*: If PostgreSQL does not become ready within 30 seconds, the recipe emits `✖ Error: Database failed to become ready within 30 seconds.` and exits with code 1, preventing CI runner hangs.

---

## 3. Caveats

1. **Local Docker Daemon**: The Docker daemon was not running in the current sandbox environment; all container execution checks were verified via Make dry-run (`make -n`), syntax tree parsing, POSIX shell simulation, and `git apply --check`.
2. **Scope**: This investigation and patch strictly covers `/Users/jasbotello/devel/asesamo/Makefile`. Related container configuration findings in `docker-compose.yml` (e.g. port decoupling and seeder entrypoint quoting) are addressed by peer remediation workers.

---

## 4. Conclusion

The exact, complete remediation for `/Users/jasbotello/devel/asesamo/Makefile` has been designed, validated against all adversarial injection vectors, and packaged as a clean patch:

### Proposed Patch Specification
```diff
--- a/Makefile
+++ b/Makefile
@@ -19,6 +19,8 @@ DB_HOST        ?= localhost
 DB_PORT        ?= 5432
 MIGRATE_IMAGE  ?= migrate/migrate:v4.17.0
 
+export DB_USER DB_PASSWORD DB_NAME
+
 ## Display help menu
 help:
 	@echo "ASESAMO Phase 1 MVP — Automation Targets"
@@ -43,10 +45,19 @@ help:
 docker-up:
 	@echo "==> Starting ASESAMO Docker services..."
 	docker compose up -d
-	@echo "==> Waiting for database readiness..."
-	@until docker compose exec -T db pg_isready -U $(DB_USER) -d $(DB_NAME) > /dev/null 2>&1; do \
+	@echo "==> Waiting for database readiness (max 30s)..."
+	@ready=0; \
+	for i in $$(seq 1 30); do \
+		if docker compose exec -T db pg_isready -U "$$DB_USER" -d "$$DB_NAME" > /dev/null 2>&1; then \
+			ready=1; \
+			break; \
+		fi; \
 		sleep 1; \
-	done
+	done; \
+	if [ $$ready -ne 1 ]; then \
+		echo "✖ Error: Database failed to become ready within 30 seconds."; \
+		exit 1; \
+	fi
 	@echo "✔ Docker services started and database is ready."
 
 ## Stop docker-compose stack (preserves volumes)
@@ -79,20 +90,21 @@ migrate-version:
 	docker compose run --rm migrate version
 
 ## Clear dirty state by forcing specific version (Usage: make migrate-force V=1)
+migrate-force: export V := $(V)
 migrate-force:
-	@if [ -z "$(V)" ]; then \
-		echo "✖ Error: Please specify version, e.g. make migrate-force V=1"; \
+	@if [ -z "$$V" ] || ! echo "$$V" | grep -qE '^[0-9]+$$'; then \
+		echo "✖ Error: Please specify a numeric version, e.g. make migrate-force V=1"; \
 		exit 1; \
 	fi
-	@echo "==> Forcing migration version to $(V)..."
-	docker compose run --rm migrate force $(V)
-	@echo "✔ Migration version forced to $(V)."
+	@echo "==> Forcing migration version to $$V..."
+	docker compose run --rm migrate force "$$V"
+	@echo "✔ Migration version forced to $$V."
 
 ## Seed database with deterministic test records
 seed:
 	@echo "==> Seeding database with deterministic dev/test data..."
 	@if docker compose ps db --format '{{.State}}' 2>/dev/null | grep -q "running"; then \
-		docker compose exec -T -e PGPASSWORD=$(DB_PASSWORD) db psql -U $(DB_USER) -d $(DB_NAME) < db/seeds/000001_seed_dev_data.sql; \
+		docker compose exec -T -e PGPASSWORD="$$DB_PASSWORD" db psql -U "$$DB_USER" -d "$$DB_NAME" < db/seeds/000001_seed_dev_data.sql; \
 	else \
 		echo "✖ Error: db container is not running. Run 'make docker-up' first."; \
 		exit 1; \
@@ -101,7 +113,7 @@ seed:
 
 ## Open interactive psql shell
 db-shell:
-	docker compose exec -e PGPASSWORD=$(DB_PASSWORD) db psql -U $(DB_USER) -d $(DB_NAME)
+	docker compose exec -e PGPASSWORD="$$DB_PASSWORD" db psql -U "$$DB_USER" -d "$$DB_NAME"
 
 ## Run backend unit tests
 test:
```

Patch file path: `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m1_rem_2/makefile_remediation.patch`

---

## 5. Verification Method

To independently verify this remediation, run the following automated test harness:

```bash
python3 -c "
import subprocess

patch_file = '/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m1_rem_2/makefile_remediation.patch'

# 1. Verify patch validity
res = subprocess.run(['git', 'apply', '--check', patch_file], capture_output=True, text=True)
assert res.returncode == 0, f'Patch check failed: {res.stderr}'
print('✔ git apply --check passed')

# 2. Verify Challenger Test 2 assertions against patched version
# (Applying patch in a temporary test file)
with open('/Users/jasbotello/devel/asesamo/Makefile', 'r') as f:
    orig = f.read()

# Generate patched content
import tempfile
with tempfile.NamedTemporaryFile('w', delete=False) as tmp:
    tmp_path = tmp.name

subprocess.run(['patch', '-o', tmp_path, '/Users/jasbotello/devel/asesamo/Makefile', patch_file], check=True)

# Test 2.1: migrate-force dry-run
res1 = subprocess.run(['make', '-n', '-f', tmp_path, 'migrate-force', 'V=1; echo INJECTED'], capture_output=True, text=True)
assert 'INJECTED' not in res1.stdout, 'DEFECT: Shell injection detected in make migrate-force'
print('✔ migrate-force dry-run injection check passed')

# Test 2.2: seed dry-run
res2 = subprocess.run(['make', '-n', '-f', tmp_path, 'seed', 'DB_PASSWORD=pass; echo INJECTED'], capture_output=True, text=True)
assert 'INJECTED' not in res2.stdout, 'DEFECT: Shell injection detected in make seed'
print('✔ seed dry-run injection check passed')

# Test 2.3: db-shell dry-run
res3 = subprocess.run(['make', '-n', '-f', tmp_path, 'db-shell', 'DB_PASSWORD=pass; echo INJECTED'], capture_output=True, text=True)
assert 'INJECTED' not in res3.stdout, 'DEFECT: Shell injection detected in make db-shell'
print('✔ db-shell dry-run injection check passed')

# Test 2.4: Injection payload rejection
for bad_v in ['', 'abc', '1; id', '1\" || id || \"', '\$(id)', '\`id\`']:
    r = subprocess.run(['make', '-f', tmp_path, 'migrate-force', f'V={bad_v}'], capture_output=True, text=True)
    assert r.returncode != 0, f'Bad V={bad_v} was not rejected'
print('✔ All malicious migration versions rejected')

# Test 2.5: docker-up dry-run
res_up = subprocess.run(['make', '-n', '-f', tmp_path, 'docker-up'], capture_output=True, text=True)
assert 'seq 1 30' in res_up.stdout, 'DEFECT: Bounded loop missing in docker-up'
print('✔ docker-up bounded loop confirmed')

import os
os.unlink(tmp_path)
print('\nALL INDEPENDENT VERIFICATIONS PASSED!')
"
```

### Invalidation Conditions:
- If `make -n migrate-force V='1; echo INJECTED'` outputs the literal string `INJECTED` rather than `$V`.
- If `make migrate-force V='1; id'` or `make migrate-force V='1" || id || "'` executes `id` or succeeds with exit code 0.
- If `make docker-up` runs an unbounded `until` loop without a retry count limit.
