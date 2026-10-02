# Remediation Analysis: E2E Synchronization & Safe URI Handling

**Agent**: `teamwork_preview_explorer_m1_rem_3`  
**Milestone**: M1 Iteration 2 (Remediation)  
**Date**: 2026-10-01  
**Working Directory**: `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m1_rem_3`  
**Scope**: E2E Access Point ID synchronization and safe URI / password encoding guidance across `.env.example`, `.env`, and container connection strings.

---

## 1. Executive Summary

Milestone M1 review and challenger stress-testing surfaced two critical configuration and interface resilience issues:
1. **E2E Access Point ID Desynchronization**: In `test/e2e/common.sh` line 24, the default environment variable `TEST_ACCESS_POINT_ID` was set to `33333333-3333-3333-3333-333333333333` (ending in `...3333`). However, the authoritative database seed script `db/seeds/000001_seed_dev_data.sql` (lines 57, 68) and interface contracts in `PROJECT.md` (lines 119, 150) declare Caseta Principal Norte as `33333333-3333-3333-3333-333333333331` (ending in `...3331`). Executing the test suite without explicitly passing `TEST_ACCESS_POINT_ID` caused E2E test cases TC 3.3 (catalog verification) and TC 4.1 (check-in evaluation) to fail against an unseeded entity ID.
2. **Database URI Parsing & Password Delimiter Resilience**: Challenger 1 (Observation 1.4) identified that passwords containing RFC 3986 reserved delimiter characters (`#`, `?`, `/`, `@`, `:`, `%`, `&`, `+`) cause standard URL parsers (including Go's `net/url` used by `golang-migrate` and `pgx`) to fail fatally during container boot. Furthermore, naive string concatenation in `docker-compose.yml` (`postgres://${DB_USER}:${DB_PASSWORD}@...`) lacks percent-encoding, preventing deployment with complex passwords.

This document formulates the exact, verified remediation for both issues, including patch specifications, environment guidance, and architectural guidelines for Milestone M2.

---

## 2. Part 1: E2E Access Point UUID Desynchronization Analysis

### 2.1 File & Contract Inspection

A comprehensive pattern search for `33333333` across the entire workspace identified the discrepancy:

| File | Line | Content | Status |
|---|---|---|---|
| `db/seeds/000001_seed_dev_data.sql` | 57 | `-- Deterministic UUID: 33333333-3333-3333-3333-333333333331` | Authoritative Source (Seed) |
| `db/seeds/000001_seed_dev_data.sql` | 68 | `'33333333-3333-3333-3333-333333333331',` | Authoritative Source (Seed) |
| `PROJECT.md` | 119 | `"access_point_id": "33333333-3333-3333-3333-333333333331"` | Authoritative Contract (Check-in) |
| `PROJECT.md` | 150 | `"id": "33333333-3333-3333-3333-333333333331"` | Authoritative Contract (Snapshot) |
| `test/e2e/common.sh` | 24 | `export TEST_ACCESS_POINT_ID="${TEST_ACCESS_POINT_ID:-33333333-3333-3333-3333-333333333333}"` | **Defect (Ends in ...3333)** |
| `TEST_INFRA.md` | 83 | `\| TEST_ACCESS_POINT_ID \| 33333333-3333-3333-3333-333333333333 \|` | **Defect (Documentation table)** |

### 2.2 Impact and Blast Radius in Test Suite

The test suite relies heavily on `TEST_ACCESS_POINT_ID`:
- **`test/e2e/tier1_feature_test.sh` (Line 149)**:
  `assert_json_expr "Catalog contains active Caseta Principal" "$BODY" '(.data[] | select(.id == "'"${TEST_ACCESS_POINT_ID}"'")) != null'`
  Because the seed only inserts `...3331`, this assertion evaluates to `false` and fails.
- **`test/e2e/tier1_feature_test.sh` (Lines 172, 196, 216)**:
  Submits check-ins with `access_point_id: ${TEST_ACCESS_POINT_ID}`. Because `...3333` does not exist in the database, the backend returns HTTP 404 or foreign key validation error, causing check-in assertions for compliant, non-compliant, and warning workers to fail.
- **`test/e2e/tier2_boundary_test.sh` (Lines 69, 84, 97, 117, 159, 177, 235)**:
  Uses `${TEST_ACCESS_POINT_ID}` as the valid baseline access point. Interestingly, line 136 uses `00000000-0000-0000-0000-000000000099` as the nonexistent access point test; having `${TEST_ACCESS_POINT_ID}` also nonexistent created semantic confusion.
- **`test/e2e/tier3_combination_test.sh` (Line 59) and `tier4_workload_test.sh` (Line 43)**:
  Falls back to `${TEST_ACCESS_POINT_ID}` if catalog query is empty or returns null.

### 2.3 Proposed Patch for `test/e2e/common.sh`
File: `/Users/jasbotello/devel/asesamo/test/e2e/common.sh` (Line 24)

```diff
--- a/test/e2e/common.sh
+++ b/test/e2e/common.sh
@@ -21,7 +21,7 @@
 export GUARD_PASSWORD="${GUARD_PASSWORD:-Password123!}"
 export TEST_TENANT_ID="${TEST_TENANT_ID:-11111111-1111-1111-1111-111111111111}"
 export TEST_PLANT_ID="${TEST_PLANT_ID:-22222222-2222-2222-2222-222222222222}"
-export TEST_ACCESS_POINT_ID="${TEST_ACCESS_POINT_ID:-33333333-3333-3333-3333-333333333333}"
+export TEST_ACCESS_POINT_ID="${TEST_ACCESS_POINT_ID:-33333333-3333-3333-3333-333333333331}"
 export WORKER_COMPLIANT_ID="${WORKER_COMPLIANT_ID:-66666666-6666-6666-6666-666666666661}"
 export WORKER_NON_COMPLIANT_ID="${WORKER_NON_COMPLIANT_ID:-66666666-6666-6666-6666-666666666662}"
 export WORKER_WARNING_ID="${WORKER_WARNING_ID:-66666666-6666-6666-6666-666666666663}"
```

### 2.4 Accompanying Patch for `TEST_INFRA.md`
File: `/Users/jasbotello/devel/asesamo/TEST_INFRA.md` (Line 83)

```diff
--- a/TEST_INFRA.md
+++ b/TEST_INFRA.md
@@ -80,7 +80,7 @@
 | `API_BASE_URL` | `http://localhost:8080` | Root URL of the Go REST API. |
 | `GUARD_EMAIL` | `guard@asesamo.com` | Email of seeded test security guard. |
 | `GUARD_PASSWORD` | `Password123!` | Password for test guard user. |
-| `TEST_ACCESS_POINT_ID` | `33333333-3333-3333-3333-333333333333` | ID of seeded Caseta 1 checkpoint. |
+| `TEST_ACCESS_POINT_ID` | `33333333-3333-3333-3333-333333333331` | ID of seeded Caseta 1 checkpoint. |
 | `WORKER_COMPLIANT_ID` | `66666666-6666-6666-6666-666666666661` | Worker Carlos Mendoza (All valid docs). |
 | `WORKER_NON_COMPLIANT_ID` | `66666666-6666-6666-6666-666666666662` | Worker Roberto Morales (Expired DC-3, rejected medical). |
 | `WORKER_WARNING_ID` | `66666666-6666-6666-6666-666666666663` | Worker Mateo Vargas (DC-3 expiring in 5 days). |
```

---

## 3. Part 2: Password Encoding Guidance & Safe URI Handling

### 3.1 Empirical Root Cause Investigation

PostgreSQL connection strings formatted as standard RFC 3986 URIs follow the schema:
`postgres://[user[:password]@]host[:port][/database][?parameter_list]`

When passwords contain RFC 3986 reserved delimiter characters, URI parsers cannot distinguish between delimiter syntax and credential payload:
- `/` is parsed as the end of authority / beginning of database path.
- `#` is parsed as the beginning of the URI fragment.
- `?` is parsed as the beginning of query parameters.
- `:` is parsed as username/password or host/port separator.
- `%` is parsed as the prefix of an octet escape sequence (fails if followed by non-hex chars).

#### Empirical Test in Go Runtime (`net/url`):
Testing Go's standard library `url.Parse` (the exact parser invoked by `golang-migrate` and PostgreSQL drivers):

| Password String | Raw URI Parsing Result | Percent-Encoded URI Result |
|---|---|---|
| `asesamo_secret_dev` | OK (parsed: `asesamo_secret_dev`) | OK (parsed: `asesamo_secret_dev`) |
| `pass/word` | **ERROR**: `invalid port ":pass" after host` | OK (encoded: `pass%2Fword`, parsed: `pass/word`) |
| `pass#word` | **ERROR**: `invalid port ":pass" after host` | OK (encoded: `pass%23word`, parsed: `pass#word`) |
| `pass?word` | **ERROR**: `invalid port ":pass" after host` | OK (encoded: `pass%3Fword`, parsed: `pass?word`) |
| `pass%word` | **ERROR**: `invalid URL escape "%wo"` | OK (encoded: `pass%25word`, parsed: `pass%word`) |
| `p@ss#w0rd!%^&*()` | **ERROR**: `invalid URL escape "%^&"` | OK (encoded: `p%40ss%23w0rd%21%25%5E%26%2A%28%29`, parsed: `p@ss#w0rd!%^&*()`) |

**Finding**: URL percent-encoding guarantees 100% preservation of all passwords without parser crashes.

### 3.2 Current Implementation State in Project
1. **Default dev configuration**: `.env` and `.env.example` currently use `asesamo_secret_dev`, which consists solely of alphanumeric characters and underscores. In local development with default values, no parsing errors occur.
2. **Missing developer guidance**: Neither `.env.example` nor `.env` provides documentation or warning regarding URI reserved characters in passwords. When developers or deployers change `DB_PASSWORD` to a generated secret containing `#`, `?`, `/`, or `%`, `docker compose up` crashes in the `migrate` container.
3. **`docker-compose.yml` interpolation gap**:
   In `docker-compose.yml`:
   ```yaml
   migrate:
     entrypoint:
       - /migrate
       - -path=/migrations
       - -database=postgres://${DB_USER:-asesamo_admin}:${DB_PASSWORD:-asesamo_secret_dev}@db:5432/${DB_NAME:-asesamo_db}?sslmode=disable
   api:
     environment:
       DATABASE_URL: postgres://${DB_USER:-asesamo_admin}:${DB_PASSWORD:-asesamo_secret_dev}@db:5432/${DB_NAME:-asesamo_db}?sslmode=disable
   ```
   Even if a user sets `DOCKER_DATABASE_URL` in `.env` with a percent-encoded password, `docker-compose.yml` ignores `DOCKER_DATABASE_URL` and blindly expands `${DB_PASSWORD}`!

### 3.3 Remediation Actions

#### Action 1: Add Clear Password Encoding Guidance to `.env.example` and `.env`
Add an explicit documentation block directly above the connection URLs in both `.env.example` and `.env`.

**Unified Diff Patch for `.env.example`**:
```diff
--- a/.env.example
+++ b/.env.example
@@ -14,6 +14,16 @@
 # Database Connection URLs
 # - DATABASE_URL: for tools running directly on the host machine (e.g. host tests, psql)
 # - DOCKER_DATABASE_URL: for services communicating inside docker-compose (e.g. api, migrate)
+#
+# IMPORTANT URI ENCODING GUIDANCE:
+# Standard PostgreSQL connection URIs require RFC 3986 reserved delimiter characters
+# (such as #, ?, /, @, :, %, &, +, and spaces) in usernames or passwords to be
+# URL percent-encoded to prevent connection parser syntax errors in Go / golang-migrate.
+# - Default dev password ("asesamo_secret_dev") contains only alphanumeric characters (no encoding required).
+# - If customizing DB_PASSWORD with special symbols:
+#     Raw Character:  #    ?    /    @    :    %    &    +   [space]
+#     Percent-Encode: %23  %3F  %2F  %40  %3A  %25  %26  %2B  %20
+#   Example: password "pass#123?" -> "postgres://asesamo_admin:pass%23123%3F@..."
 DATABASE_URL=postgres://asesamo_admin:asesamo_secret_dev@localhost:5432/asesamo_db?sslmode=disable
 DOCKER_DATABASE_URL=postgres://asesamo_admin:asesamo_secret_dev@db:5432/asesamo_db?sslmode=disable
```

**Unified Diff Patch for `.env`**:
Identical hunk applied to `/Users/jasbotello/devel/asesamo/.env`.

#### Action 2: Allow `DOCKER_DATABASE_URL` Override in `docker-compose.yml` (Coordination with `rem_1`)
Update `docker-compose.yml` lines 42 and 91 so that if `DOCKER_DATABASE_URL` is defined in `.env`, it takes precedence over the template interpolation:

```diff
--- a/docker-compose.yml
+++ b/docker-compose.yml
@@ -39,7 +39,7 @@
     entrypoint:
       - /migrate
       - -path=/migrations
-      - -database=postgres://${DB_USER:-asesamo_admin}:${DB_PASSWORD:-asesamo_secret_dev}@db:5432/${DB_NAME:-asesamo_db}?sslmode=disable
+      - -database=${DOCKER_DATABASE_URL:-postgres://${DB_USER:-asesamo_admin}:${DB_PASSWORD:-asesamo_secret_dev}@db:5432/${DB_NAME:-asesamo_db}?sslmode=disable}
       - -verbose
     command: ["up"]
     networks:
@@ -88,7 +88,7 @@
     environment:
       APP_ENV: ${APP_ENV:-development}
       APP_PORT: ${APP_PORT:-8080}
-      DATABASE_URL: postgres://${DB_USER:-asesamo_admin}:${DB_PASSWORD:-asesamo_secret_dev}@db:5432/${DB_NAME:-asesamo_db}?sslmode=disable
+      DATABASE_URL: ${DOCKER_DATABASE_URL:-postgres://${DB_USER:-asesamo_admin}:${DB_PASSWORD:-asesamo_secret_dev}@db:5432/${DB_NAME:-asesamo_db}?sslmode=disable}
       JWT_SECRET: ${JWT_SECRET:-super_secret_jwt_key_min_32_characters_long_12345}
       CORS_ALLOWED_ORIGINS: ${CORS_ALLOWED_ORIGINS:-http://localhost:3000,http://127.0.0.1:3000,http://web:3000}
```

This guarantees:
1. Backwards compatibility: if `DOCKER_DATABASE_URL` is not provided, defaults work cleanly.
2. Production safety: when a complex password is used, defining `DOCKER_DATABASE_URL` in `.env` or production secrets manager guarantees proper URI parsing.

#### Action 3: Architectural Guidelines for Milestone M2 (Backend Config Loader)
For the M2 backend team implementing `asesamo-api/internal/infra/config`:
1. **Config Source Precedence**:
   Allow either `DATABASE_URL` or discrete environment variables: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`.
2. **Safe URI Construction in Go**:
   If discrete parameters are provided, construct the connection string via Go's `net/url` standard builder:
   ```go
   u := &url.URL{
       Scheme: "postgres",
       User:   url.UserPassword(cfg.DBUser, cfg.DBPassword),
       Host:   net.JoinHostPort(cfg.DBHost, cfg.DBPort),
       Path:   cfg.DBName,
   }
   q := u.Query()
   q.Set("sslmode", cfg.DBSSLMode)
   u.RawQuery = q.Encode()
   connStr := u.String()
   ```
   `url.UserPassword` automatically handles percent-encoding for arbitrary password characters without developer intervention.
3. **Keyword-Value Driver Strings**:
   Alternatively, PostgreSQL drivers like `pgx` support key-value format (`host=... port=... user=... password='...' dbname=... sslmode=...`), where quotes escape special characters without URL encoding.

---

## 4. Verification & Testing Evidence

All proposed patches were verified using `patch --dry-run` and empirical test scripts:
1. `patch --dry-run -p1 < e2e_common.patch`: Cleanly applies with exit code 0.
2. `patch --dry-run -p1 < test_infra.patch`: Cleanly applies with exit code 0.
3. `patch --dry-run -p1 < env_guidance.patch`: Cleanly applies with exit code 0.
4. `patch --dry-run -p1 < env_local.patch`: Cleanly applies with exit code 0.
5. `patch --dry-run -p1 < docker_compose_uri.patch`: Cleanly applies with exit code 0.
6. Empirical URL encoding test across 8 delimiter categories passed with 100% fidelity.
