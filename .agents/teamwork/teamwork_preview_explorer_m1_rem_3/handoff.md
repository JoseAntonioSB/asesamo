# Handoff Report: E2E and URI Resilience Remediation

**Agent**: `teamwork_preview_explorer_m1_rem_3`  
**Role**: Explorer (`investigator`, `analyzer`)  
**Milestone**: M1 Iteration 2 (Remediation)  
**Date**: 2026-10-01  
**Working Directory**: `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m1_rem_3`  
**Parent Agent**: `ccf64bbc-f322-4ca8-b0be-0ee59db22a99`  
**Verdict**: **REMEDIATION SPECIFICATION COMPLETE**

---

## 1. Observation

### Observation 1.1 — Default Access Point UUID Desynchronization in `test/e2e/common.sh`
- In `/Users/jasbotello/devel/asesamo/test/e2e/common.sh` line 24:
  ```bash
  export TEST_ACCESS_POINT_ID="${TEST_ACCESS_POINT_ID:-33333333-3333-3333-3333-333333333333}"
  ```
- In `/Users/jasbotello/devel/asesamo/TEST_INFRA.md` line 83:
  ```markdown
  | `TEST_ACCESS_POINT_ID` | `33333333-3333-3333-3333-333333333333` | ID of seeded Caseta 1 checkpoint. |
  ```
- In `/Users/jasbotello/devel/asesamo/db/seeds/000001_seed_dev_data.sql`:
  - Line 57: `-- Deterministic UUID: 33333333-3333-3333-3333-333333333331`
  - Line 68: `'33333333-3333-3333-3333-333333333331',`
- In `/Users/jasbotello/devel/asesamo/PROJECT.md`:
  - Line 119: `"access_point_id": "33333333-3333-3333-3333-333333333331"`
  - Line 150: `"id": "33333333-3333-3333-3333-333333333331",`

**Empirical Result**: Running `grep_search` across the repository confirmed that `33333333-3333-3333-3333-333333333333` appears in exactly two locations (`test/e2e/common.sh:24` and `TEST_INFRA.md:83`). All database seed statements and interface specifications declare the primary checkpoint as `33333333-3333-3333-3333-333333333331`.

### Observation 1.2 — E2E Test Failures Caused by UUID Mismatch
- In `test/e2e/tier1_feature_test.sh`:
  - Line 149 (TC 3.3):
    ```bash
    assert_json_expr "Catalog contains active Caseta Principal (${TEST_ACCESS_POINT_ID})" "$BODY" \
      '(.data[] | select(.id == "'"${TEST_ACCESS_POINT_ID}"'")) != null'
    ```
  - Lines 172, 196, 216 (TC 4.1, TC 4.2, TC 4.3): Check-in payloads send `"access_point_id": "${TEST_ACCESS_POINT_ID}"`.
- When `TEST_ACCESS_POINT_ID` defaults to `...3333`, TC 3.3 fails because the seeded access point is `...3331`. Furthermore, check-in requests with `...3333` fail because the foreign key/checkpoint record does not exist in the database.

### Observation 1.3 — Database Connection URI Parser Crash on Password Delimiters
- In Go `net/url` and Python `urllib.parse`, parsing database URIs where the password contains unescaped delimiter characters (`#`, `?`, `/`, `%`) produces fatal errors:
  ```go
  url.Parse("postgres://asesamo_admin:pass/word@db:5432/asesamo_db?sslmode=disable")
  // ERROR: parse "...": invalid port ":pass" after host
  url.Parse("postgres://asesamo_admin:pass#word@db:5432/asesamo_db?sslmode=disable")
  // ERROR: parse "...": invalid port ":pass" after host
  url.Parse("postgres://asesamo_admin:pass?word@db:5432/asesamo_db?sslmode=disable")
  // ERROR: parse "...": invalid port ":pass" after host
  url.Parse("postgres://asesamo_admin:pass%word@db:5432/asesamo_db?sslmode=disable")
  // ERROR: parse "...": invalid URL escape "%wo"
  url.Parse("postgres://asesamo_admin:p@ss#w0rd!%^&*()@db:5432/asesamo_db?sslmode=disable")
  // ERROR: parse "...": invalid URL escape "%^&"
  ```
- Conversely, when special characters are URL percent-encoded (RFC 3986):
  ```go
  u, _ := url.Parse("postgres://asesamo_admin:pass%23word@db:5432/asesamo_db?sslmode=disable")
  pw, _ := u.User.Password() // returns "pass#word"
  ```
  100% of tested password cases parse without errors and return the original plaintext password.

### Observation 1.4 — Configuration and Docker Compose State
- In `.env.example` and `.env` (lines 14–19):
  `DATABASE_URL` and `DOCKER_DATABASE_URL` are provided with default `asesamo_secret_dev`. No documentation, warning, or encoding guidance exists for developers or operators using complex passwords.
- In `docker-compose.yml`:
  - Line 42: `- -database=postgres://${DB_USER:-asesamo_admin}:${DB_PASSWORD:-asesamo_secret_dev}@db:5432/${DB_NAME:-asesamo_db}?sslmode=disable`
  - Line 91: `DATABASE_URL: postgres://${DB_USER:-asesamo_admin}:${DB_PASSWORD:-asesamo_secret_dev}@db:5432/${DB_NAME:-asesamo_db}?sslmode=disable`
  Docker Compose performs simple string substitution on `${DB_PASSWORD}`. Even if a user provides an encoded `DOCKER_DATABASE_URL` in `.env`, `docker-compose.yml` ignores it and concatenates raw `${DB_PASSWORD}`.

---

## 2. Logic Chain

1. **Test Failure Mechanism**:
   - *Observation 1.1*: `db/seeds/000001_seed_dev_data.sql` inserts Caseta Principal Norte with UUID `33333333-3333-3333-3333-333333333331`.
   - *Observation 1.2*: `test/e2e/common.sh` line 24 defaults `TEST_ACCESS_POINT_ID` to `...3333`.
   - *Deduction*: When `test/e2e/test_runner.sh` runs without explicit environment overrides, tests querying `/api/v1/access-points` search for `...3333` and find null, failing TC 3.3. Check-ins for `...3333` fail with HTTP 404 or FK validation errors.
   - *Conclusion*: Changing default `TEST_ACCESS_POINT_ID` in `test/e2e/common.sh` from `...3333` to `...3331` restores alignment with the database seeds and interface contracts.

2. **URI Parser Fragility & Password Delimiters**:
   - *Observation 1.3*: Standard RFC 3986 URL parsers treat `/`, `#`, `?`, and `%` as syntactic delimiters. Passwords containing these characters corrupt the host/port authority or trigger invalid escape sequence errors in `golang-migrate` and database drivers.
   - *Observation 1.4*: Neither `.env.example` nor `.env` instructs operators on URL percent-encoding.
   - *Deduction*: Deployers generating passwords containing standard punctuation will encounter immediate boot failures in `migrate` and `api`.
   - *Deduction*: Naive `${DB_PASSWORD}` concatenation in `docker-compose.yml` prevents overriding the connection string with a pre-encoded URL.
   - *Conclusion*: Adding prominent encoding guidance to `.env.example`/`.env` and updating `docker-compose.yml` to allow `${DOCKER_DATABASE_URL:-...}` override ensures complete URI resilience across environments.

---

## 3. Caveats

1. **Read-Only Explorer Scope**: In accordance with the Explorer protocol, no direct file writes were performed on root codebase files (`test/e2e/common.sh`, `.env`, `.env.example`, `docker-compose.yml`). Exact `.patch` files have been authored and verified with `patch --dry-run` in this agent's directory.
2. **Coordination with `rem_1`**: Agent `teamwork_preview_explorer_m1_rem_1` is analyzing port decoupling and seeder variable quoting in `docker-compose.yml`. The advisory patch `docker_compose_uri.patch` provided here should be merged with `rem_1`'s edits to `docker-compose.yml` lines 42 and 91.
3. **M2 Backend Implementation**: Full discrete connection parameter support (`DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`) will be implemented in Milestone M2 when `asesamo-api/internal/infra/config` is developed.

---

## 4. Conclusion

The exact remediations for E2E and URI resilience are formulated and verified.

### Actionable Remediation Artifacts in Working Directory:

1. **`e2e_common.patch`**:
   - **Target**: `test/e2e/common.sh` line 24.
   - **Change**: Changes default `TEST_ACCESS_POINT_ID` from `33333333-3333-3333-3333-333333333333` to `33333333-3333-3333-3333-333333333331`.
2. **`test_infra.patch`**:
   - **Target**: `TEST_INFRA.md` line 83.
   - **Change**: Updates documentation table reference for `TEST_ACCESS_POINT_ID` to `...3331`.
3. **`env_guidance.patch`**:
   - **Target**: `.env.example` lines 14–19.
   - **Change**: Adds explicit RFC 3986 URI encoding guidance, character escape reference table, and example.
4. **`env_local.patch`**:
   - **Target**: `.env` lines 14–19.
   - **Change**: Applies identical URI encoding guidance to the active local `.env`.
5. **`docker_compose_uri.patch`**:
   - **Target**: `docker-compose.yml` lines 42 and 91.
   - **Change**: Replaces raw `${DB_PASSWORD}` string concatenation with `${DOCKER_DATABASE_URL:-postgres://${DB_USER:-...}...}` fallback.

---

## 5. Verification Method

To independently verify the proposed remediations:

### 1. Dry-Run All Patches
Execute from repository root:
```bash
patch --dry-run -p1 < .agents/teamwork/teamwork_preview_explorer_m1_rem_3/e2e_common.patch
patch --dry-run -p1 < .agents/teamwork/teamwork_preview_explorer_m1_rem_3/test_infra.patch
patch --dry-run -p1 < .agents/teamwork/teamwork_preview_explorer_m1_rem_3/env_guidance.patch
patch --dry-run -p1 < .agents/teamwork/teamwork_preview_explorer_m1_rem_3/env_local.patch
patch --dry-run -p1 < .agents/teamwork/teamwork_preview_explorer_m1_rem_3/docker_compose_uri.patch
```
*Expected Result*: All 5 patches report `patching file ...` with exit code `0` and zero rejections.

### 2. Verify UUID Synchronization
```bash
grep -n "33333333" test/e2e/common.sh db/seeds/000001_seed_dev_data.sql PROJECT.md
```
*Expected Result*: All access point UUID declarations end in `333333333331`.

### 3. Verify Password Delimiter Resilience
Run the verification test harness:
```bash
python3 -c "
import urllib.parse
test_passwords = ['asesamo_secret_dev', 'pass/word', 'pass#word', 'pass?word', 'pass@word', 'pass%word', 'p@ss#w0rd!%^&*()']
for pw in test_passwords:
    encoded = urllib.parse.quote(pw, safe='')
    uri = f'postgres://asesamo_admin:{encoded}@db:5432/asesamo_db?sslmode=disable'
    parsed = urllib.parse.urlparse(uri)
    recovered = urllib.parse.unquote(parsed.password)
    assert recovered == pw, f'Failed for {pw}'
print('All password encoding assertions passed.')
"
```
*Expected Result*: Prints `All password encoding assertions passed.` with exit code `0`.

### Invalidation Conditions
- Any change to `db/seeds/000001_seed_dev_data.sql` that modifies the primary access point UUID away from `...3331`.
- Any Go driver change that departs from RFC 3986 percent-encoding in database URIs.
