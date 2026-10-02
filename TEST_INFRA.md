# ASESAMO — Test Infrastructure Specification (TEST_INFRA.md)

**Document Version:** 1.0.0  
**Status:** Active / Ready for Execution  
**Project:** ASESAMO Phase 1 MVP (Industrial Access Control & Document Compliance)  
**Scope:** End-to-End Opaque-Box Testing Suite (Tiers 1–4)  

---

## 1. Test Philosophy

ASESAMO is an industrial safety access control platform operating at security checkpoints (*casetas*) in manufacturing, automotive, and heavy industry plants across Mexico. In this environment, system errors can result in unauthorized entry into hazardous areas, severe regulatory sanctions (STPS, IMSS, REPSE), or plant operational paralysis.

The testing philosophy is founded on five core tenets:

1. **Strict Opaque-Box Interaction**:
   All end-to-end tests interact exclusively with the system over HTTP through public REST API boundaries. No test accesses the database directly, mocks Go internal repositories, or leverages white-box code shortcuts. What the test verifies is identical to what a physical RFID/QR scanner, guard terminal, or automated turnstile observes.

2. **Progressive Testability & Decoupling**:
   Tests are written to the normative interface contracts defined in `PROJECT.md` and `DEVELOPMENT_RULES.md`. Each tier validates specific milestones and features, allowing the test suite to execute progressively as backend capabilities are brought online without waiting for frontend completion.

3. **Deterministic Expected Output Derivation**:
   Every test assertion derives from authoritative specifications:
   - Mexican labor and safety norms (STPS NOM-009, NOM-027, NOM-029, DC-3, IMSS/SUA, REPSE reform).
   - The standardized seed dataset defined in `db/seeds/000001_seed_dev_data.sql` and `PROJECT.md`.
   - Contract envelopes: `{ "data": ... }` and `{ "error": { "code", "message" } }`.

4. **Multi-Tenancy & Security Verification**:
   Testing strictly verifies tenant isolation (`WHERE tenant_id = $1`). Cross-tenant access attempts, token tampering, and role escalation are actively tested to ensure data boundaries are inviolable.

5. **Actionable Diagnostics & Zero Flakiness**:
   Tests output structured, color-coded summaries with exact HTTP statuses, timing benchmarks, payload snippets on failure, and deterministic exit codes (`0` for total pass, non-zero for any failure).

---

## 2. Feature Inventory Mapping

| # | Feature Name | API Endpoint / Method | Target Tier | Primary Test Script | Covered Scenarios |
|---|--------------|-----------------------|-------------|---------------------|-------------------|
| 1 | Health & Infrastructure Probe | `GET /api/v1/health` or `GET /health` | Tier 1 | `tier1_feature_test.sh` | Liveness check, database connection state, response format, method compliance, response timing. |
| 2 | Guard Authentication | `POST /api/v1/auth/login` | Tier 1, 2 | `tier1_feature_test.sh`<br>`tier2_boundary_test.sh` | Valid credentials, wrong password, inactive user, malformed request, missing fields, rate-limit resilience. |
| 3 | Access Points Catalog | `GET /api/v1/access-points` | Tier 1, 2 | `tier1_feature_test.sh`<br>`tier2_boundary_test.sh` | List active checkpoints, tenant scoping, unauthenticated denial, invalid token rejection. |
| 4 | Gate Check-In (Compliant) | `POST /api/v1/access/check-in` | Tier 1, 3, 4 | `tier1_feature_test.sh`<br>`tier3_combination_test.sh`<br>`tier4_workload_test.sh` | All documents approved and valid -> `result: "granted"`, denial reason null, JSONB snapshot generated. |
| 5 | Gate Check-In (Non-Compliant) | `POST /api/v1/access/check-in` | Tier 1, 2, 4 | `tier1_feature_test.sh`<br>`tier2_boundary_test.sh`<br>`tier4_workload_test.sh` | Expired DC-3, rejected medical exam -> `result: "denied"`, Spanish denial reason, audit log written. |
| 6 | Gate Check-In (Warning State) | `POST /api/v1/access/check-in` | Tier 1, 2 | `tier1_feature_test.sh`<br>`tier2_boundary_test.sh` | Expiring documents or plant policy `block_on_expired_docs = false` -> `result: "granted_with_warning"`. |
| 7 | Gate Check-Out | `POST /api/v1/access/check-out` | Tier 3 | `tier3_combination_test.sh` | Record exit timestamp for worker inside facility, close open access log. |
| 8 | Access Logs & Audit Trail | `GET /api/v1/access-logs` | Tier 1, 3, 4 | `tier1_feature_test.sh`<br>`tier3_combination_test.sh`<br>`tier4_workload_test.sh` | Paginated entry logs, immutable JSONB snapshot matching check-in decision, tenant isolation. |
| 9 | Boundary & Malformed Inputs | Multiple endpoints | Tier 2 | `tier2_boundary_test.sh` | Expired JWT, malformed UUID, nonexistent worker, nonexistent checkpoint, missing mandatory doc. |
| 10 | Cross-Feature Audit Flow | Login -> Check-in -> Logs | Tier 3 | `tier3_combination_test.sh` | Full end-to-end audit lifecycle verification ensuring database consistency. |
| 11 | Industrial Shift Workload | Batch check-in burst | Tier 4 | `tier4_workload_test.sh` | 15+ concurrent/sequential arrivals, mixed compliance, throughput (<500ms latency), zero log drop. |

---

## 3. Test Architecture

The E2E test harness is organized into four hierarchical tiers:

```
test/e2e/
├── test_runner.sh             # Master test harness orchestrator
├── tier1_feature_test.sh      # Core feature & API contract coverage (>=5 tests/feature)
├── tier2_boundary_test.sh     # Boundary conditions, negative tests, security invariants
├── tier3_combination_test.sh  # Multi-step stateful workflows & data integrity
└── tier4_workload_test.sh     # High-throughput realistic industrial shift simulation
```

### 3.1 Technology Stack & Tooling
- **POSIX Shell / Bash**: Universal portability across Linux, macOS, and CI/CD containers.
- **`curl`**: Native HTTP client with support for custom headers, HTTP response code extraction, timing measurements, and silent output.
- **`jq`**: JSON parsing and deterministic assertions against schema structures, arrays, and fields.
- **Exit Code Convention**:
  - `0`: All assertions passed successfully.
  - `1`: One or more assertion failures detected.
  - `2`: Configuration or pre-flight environment error (e.g., API unreachable, missing dependencies).

### 3.2 Environment Variables & Configurable Parameters

| Variable | Default Value | Description |
|---|---|---|
| `API_BASE_URL` | `http://localhost:8080` | Root URL of the Go REST API. |
| `GUARD_EMAIL` | `guard@asesamo.com` | Email of seeded test security guard. |
| `GUARD_PASSWORD` | `Password123!` | Password for test guard user. |
| `TEST_ACCESS_POINT_ID` | `33333333-3333-3333-3333-333333333333` | ID of seeded Caseta 1 checkpoint. |
| `WORKER_COMPLIANT_ID` | `66666666-6666-6666-6666-666666666661` | Worker Carlos Mendoza (All valid docs). |
| `WORKER_NON_COMPLIANT_ID` | `66666666-6666-6666-6666-666666666662` | Worker Roberto Morales (Expired DC-3, rejected medical). |
| `WORKER_WARNING_ID` | `66666666-6666-6666-6666-666666666663` | Worker Mateo Vargas (DC-3 expiring in 5 days). |
| `VERBOSE` | `0` | Set to `1` for full HTTP request/response payloads. |

---

## 4. Test Tiers Overview & Real-World Application Scenarios

### Tier 1: Feature Coverage (`tier1_feature_test.sh`)
Validates individual endpoint functionality against the contract envelopes:
- **Feature 1: System Health**: Liveness, DB connectivity, JSON envelope.
- **Feature 2: Authentication**: Guard login, JWT extraction, credential validation, error envelopes.
- **Feature 3: Access Points**: Checkpoint catalog retrieval, tenant filtering, auth protection.
- **Feature 4: Check-In Evaluation**: Granted, denied, and warning responses with document snapshots.
- **Feature 5: Access Logs Retrieval**: Audit listing, pagination parameters, sorting.

*Requirement: Minimum 5 distinct test cases per feature (Total: 25+ assertions).*

### Tier 2: Boundary & Corner Cases (`tier2_boundary_test.sh`)
Verifies system resilience against invalid inputs, adversarial payloads, and regulatory edge conditions:
- **Scenario A: Expired & Invalid JWT**: Tampered signature, expired tokens, empty Authorization header.
- **Scenario B: Malformed & Nonexistent Identifiers**: Non-UUID strings (`"abc-xyz"`), zero UUIDs (`00000000-...`), nonexistent worker records.
- **Scenario C: Missing & Expired Mandatory Documents**: Worker with no DC-3, expired DC-3 with `block_on_expired_docs = true` vs `block_on_expired_docs = false`.
- **Scenario D: Contractor Authorization Blockers**: Contractor in `suspended`, `blocked`, or `pending` status.
- **Scenario E: Deactivated Worker Profile**: Worker record flagged `is_active = false`.
- **Scenario F: Tenant Isolation Breach**: Guard from Tenant A attempting access point from Tenant B.

### Tier 3: Cross-Feature Combinations (`tier3_combination_test.sh`)
Validates stateful multi-step interactions and audit permanence:
- **Combination 1: End-to-End Entry & Audit Cycle**: Guard logs in -> Lists access points -> Checks in worker -> Queries access logs -> Confirms access log row contains matching worker UUID, access point, and immutable JSONB snapshot.
- **Combination 2: Check-In / Check-Out Lifecycle**: Guard checks in worker -> Worker completes job -> Guard calls check-out -> Confirms `checked_out_at` timestamp is written.
- **Combination 3: Rapid Double Entry**: Worker scans badge twice in under 5 seconds -> Verifies second attempt is logged without corrupting system state.
- **Combination 4: Multi-Worker Differential Audit**: Compliant worker enters, non-compliant worker rejected -> Verifies access logs reflect both outcomes accurately.

### Tier 4: Industrial Shift Workload (`tier4_workload_test.sh`)
Simulates the peak morning shift change at an industrial plant:
- **Scenario**: At 06:45 AM, a bus carrying contractor workers arrives at Caseta Principal.
- **Execution**: 15 rapid check-in requests are submitted in sequence/burst.
- **Composition**:
  - Compliant workers (Granted).
  - Non-compliant workers with expired certifications (Denied).
  - Workers with documents nearing expiration (Granted with warning).
  - Unknown badges / unregistered personnel (Rejected).
- **Invariants Verified**:
  - 100% of check-in decisions match expected compliance logic.
  - Zero dropped logs: Total check-in attempts == Total inserted `access_logs`.
  - Latency: Average check-in latency <= 250ms, 99th percentile <= 500ms.
  - Audit snapshot integrity: Document snapshots remain static and uncorrupted.

---

## 5. Coverage & Quality Thresholds

| Metric | Target Threshold | Method of Verification |
|---|---|---|
| **Tier 1 Feature Assertions** | >= 25 test cases (>= 5 per feature) | `test/e2e/tier1_feature_test.sh` counter |
| **Tier 2 Boundary Assertions** | >= 10 edge cases | `test/e2e/tier2_boundary_test.sh` counter |
| **Tier 3 Combination Assertions** | >= 4 multi-step flows | `test/e2e/tier3_combination_test.sh` counter |
| **Tier 4 Workload Assertions** | 15 simulated entries, 100% audit match | `test/e2e/tier4_workload_test.sh` reconciliation |
| **Overall E2E Pass Rate** | **100%** (0 failures tolerated) | `test/e2e/test_runner.sh` exit code |
| **Maximum Check-In Latency** | <= 500 ms per check-in | Timing checks in Tier 4 workload |
| **API Envelope Compliance** | 100% adherence to `{ "data" }` / `{ "error" }` | `jq` validation across all responses |

---

## 6. Execution Instructions

### Running the Entire Suite
```bash
./test/e2e/test_runner.sh
```

### Running Specific Tiers
```bash
./test/e2e/test_runner.sh --tier 1
./test/e2e/test_runner.sh --tier 2
./test/e2e/test_runner.sh --tier 3
./test/e2e/test_runner.sh --tier 4
```

### Running Against Custom Host
```bash
API_BASE_URL="http://remote-host:8080" ./test/e2e/test_runner.sh
```

### Direct Script Execution
```bash
./test/e2e/tier1_feature_test.sh
./test/e2e/tier2_boundary_test.sh
./test/e2e/tier3_combination_test.sh
./test/e2e/tier4_workload_test.sh
```
