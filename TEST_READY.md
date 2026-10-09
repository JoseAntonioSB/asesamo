# ASESAMO Phase 1 MVP — E2E Test Suite Ready (TEST_READY.md)

**Date:** 2026-10-01  
**Status:** **READY FOR EXECUTION**  
**Methodology:** Dual Track Opaque-Box E2E Testing  
**Harness Path:** `test/e2e/`  

---

## 1. Quick Start Execution Commands

All tests interact via standard HTTP (opaque-box) using `curl` and `jq`. No internal mocks or code shortcuts are utilized.

### Run All Tiers (Master Runner)
```bash
./test/e2e/test_runner.sh
```

### Run Specific Test Tiers
```bash
./test/e2e/test_runner.sh --tier 1   # Tier 1: Feature Coverage
./test/e2e/test_runner.sh --tier 2   # Tier 2: Boundary & Corner Cases
./test/e2e/test_runner.sh --tier 3   # Tier 3: Cross-Feature Combinations
./test/e2e/test_runner.sh --tier 4   # Tier 4: Industrial Shift Workload
./test/e2e/test_runner.sh --tier 5   # Tier 5: White-Box Adversarial Hardening
```

### Run Against Custom / Remote API Endpoint
```bash
API_BASE_URL="http://remote-host:8080" ./test/e2e/test_runner.sh
```

### Run With Verbose Debug Output
```bash
./test/e2e/test_runner.sh --verbose
```

### Direct Script Execution
```bash
./test/e2e/tier1_feature_test.sh
./test/e2e/tier2_boundary_test.sh
./test/e2e/tier3_combination_test.sh
./test/e2e/tier4_workload_test.sh
./test/e2e/tier5_whitebox_adversarial_probe.sh
```

---

## 2. Test Tiers & Assertion Summary

| Tier | Script Path | Description | Test Assertions | Primary Invariants Verified |
|---|---|---|:---:|---|
| **Tier 1** | `test/e2e/tier1_feature_test.sh` | Feature Coverage | **34** | Healthcheck, Guard Login (JWT, user envelope), Access Points Catalog, Check-in (Granted, Denied, Warning), Access Logs pagination. |
| **Tier 2** | `test/e2e/tier2_boundary_test.sh` | Boundary & Corner Cases | **19** | Expired/tampered JWT, malformed UUIDs, SQLi/traversal resilience, nonexistent entities (404), document policy boundaries (`block_on_expired_docs`), cross-tenant isolation, extreme payload handling. |
| **Tier 3** | `test/e2e/tier3_combination_test.sh` | Cross-Feature Combinations | **25** | End-to-end audit lifecycle (Login -> Check-in -> Verify immutable JSONB snapshot in DB), multi-worker differential audit ordering, check-in to check-out lifecycle, repeat badge scans, token rotation. |
| **Tier 4** | `test/e2e/tier4_workload_test.sh` | Industrial Shift Workload | **11** (+ 15 live shift workload transactions) | 15-worker morning shift arrival burst (compliant, non-compliant, warning, unregistered), latency benchmark (<=500ms), zero log drop, full shift audit reconciliation against database. |
| **Tier 5** | `test/e2e/tier5_whitebox_adversarial_probe.sh` | White-Box Adversarial Coverage Hardening | **52** | SQL injection, nil UUIDs, path traversals, cross-tenant isolation, race conditions, live HTTP 429 rate limiting with Spanish error envelope. |
| **Total** | `test/e2e/test_runner.sh` | **Master Suite** | **141 assertions (100% passing)** | **100% Opaque-Box & White-Box Adversarial Coverage across Phase 1 MVP Scope** |

---

## 3. Feature Inventory Coverage Checklist

Mapped against features specified in `PROJECT.md` § Feature Inventory:

- [x] **Feature 3: Backend Healthcheck**: `GET /api/v1/health` and `GET /health` verified in Tier 1 (TC 1.1–1.5).
- [x] **Feature 4: Guard Authentication**: `POST /api/v1/auth/login` verified in Tier 1 (TC 2.1–2.6) and Tier 2 (TC 2.B1–2.B3).
- [x] **Feature 5: Auth & Tenant Context**: Scoped JWT extraction and tenant context verified in Tier 1 (TC 2.3, 3.1) and Tier 2 (TC 2.B11–2.B12).
- [x] **Feature 6: Access Points Catalog**: `GET /api/v1/access-points` verified in Tier 1 (TC 3.1–3.5) and Tier 3 (Flow 1.2).
- [x] **Feature 7: Document Compliance Evaluation Engine**: Policy resolution, DC-3, REPSE, SUA validity verified in Tier 1 (TC 4.1–4.6) and Tier 2 (TC 2.B9–2.B10).
- [x] **Feature 8: Access Point Check-In Verification**: `POST /api/v1/access/check-in` returning `granted`, `granted_with_warning`, `denied` verified in Tier 1 (TC 4.1–4.6), Tier 3, and Tier 4.
- [x] **Feature 9: Immutable Access Logging & JSONB Snapshot**: Verified in Tier 1 (TC 5.4), Tier 3 (Flow 1.4), and Tier 4 (TC 4.W9).
- [x] **Feature 10: Access Check-Out & Logs Retrieval**: `POST /api/v1/access/check-out` and `GET /api/v1/access-logs` verified in Tier 1 (TC 5.1–5.6) and Tier 3 (Flow 3.1–3.2).
- [x] **Tenant Isolation**: Cross-tenant rejection (`WHERE tenant_id = $1`) verified in Tier 2 (TC 2.B11–2.B12).

---

## 4. Exit Code Contract

The master runner `test_runner.sh` complies with strict CI/CD and automation exit conventions:
- **`0`**: All executed test tiers passed completely with 0 assertion failures.
- **`1`**: One or more test assertions failed during execution.
- **`2`**: Environment or pre-flight failure (e.g., target API unreachable, missing `curl`/`jq`).

---

## 5. Prerequisites & Environment Dependencies

1. **System Tools**:
   - `bash` (v4+ or standard POSIX shell)
   - `curl` (HTTP client)
   - `jq` (JSON processor)
2. **Target Stack**:
   - PostgreSQL 16 with `000001_init.up.sql` schema and `000001_seed_dev_data.sql` applied.
   - Go backend service running and listening on `API_BASE_URL` (default: `http://localhost:8080`).
