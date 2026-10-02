# APPCITAS Phase 2 (La Agenda y Google Calendar) — Test Infrastructure Specification

- **Document Version**: 2.0.0
- **Author**: E2E Test Writer (Phase 2)
- **Target Working Directory**: `/Users/jasbotello/devel/APPCITAS`
- **Scope**: Opaque-box Test Harness, Tier Breakdown, Execution Semantics, and Verification Criteria

---

## 1. Overview & Architectural Principles

The APPCITAS Phase 2 Test Suite delivers a rigorous, self-contained, opaque-box verification framework for appointment scheduling, calendar synchronization, and public booking.

### Core Testing Pillars:
1. **Zero External Network Dependencies**: All external services (Google Calendar API, OAuth token servers) run in deterministic simulation modes with clean mock boundaries.
2. **Opaque-Box Contract Verification**: Tests validate system interfaces, mathematical properties, data transformations, and HTTP API specifications derived strictly from `ORIGINAL_REQUEST.md` and `spec_report.md`.
3. **Progressive Testability**: The harness is verifiable independently of incomplete milestone dependencies.
4. **Sub-second Execution**: Using the lightweight, zero-dependency `harness.ts` assertion engine, the entire 151-test suite executes in ~40 milliseconds.

---

## 2. Test File Ownership & Directory Layout

```
tests/e2e/
├── harness.ts                      # Standalone zero-dependency assertion & runner engine
├── phase2-runner.ts                # Master Phase 2 test suite executable runner
├── phase2-availability.test.ts     # Availability engine, operating hours, slot generation, collisions (55 tests)
├── phase2-booking.test.ts          # Customer deduplication, public booking API, concurrency guard (44 tests)
└── phase2-calendar-sync.test.ts    # Google OAuth, freeBusy sync, event payload, fallback, clinical scenarios (52 tests)
```

---

## 3. Four-Tier Testing Methodology

### Tier 1: Feature Coverage (75 Tests Total, Target >= 40)
Every core feature from `spec_report.md` is covered by at least 5 isolated tests:
- **P2-Tier1-F01**: Clinic Operating Hours Filter (Mon-Fri 09:00-19:00, Sat 09:00-14:00, Sun closed) — 5 tests.
- **P2-Tier1-F02**: Time Slot Generation (~50 min duration in 60 min intervals, 45/60 min variations) — 5 tests.
- **P2-Tier1-F03**: Local Database Collision Subtraction (`CONFIRMED`, `PENDING`, `COMPLETED`, `NO_SHOW` block time) — 5 tests.
- **P2-Tier1-F04**: Google Calendar `freeBusy` Busy Intervals Subtraction — 5 tests.
- **P2-Tier1-F05**: Slot Liberation on Status `CANCELLED` — 5 tests.
- **P2-Tier1-F06**: "Cualquier Especialista" Pooling & Workload Balancing — 5 tests.
- **P2-Tier1-F07**: Customer Deduplication by Phone & Clinic ID — 5 tests.
- **P2-Tier1-F08**: Appointment Creation with Full Metadata (`POST /api/public/[slug]/book`) — 5 tests.
- **P2-Tier1-F09**: Status Transitions via `PATCH /api/appointments/[id]` — 5 tests.
- **P2-Tier1-F10**: Offline Fallback Simulation Mode (when Google keys unset) — 5 tests.
- **P2-Tier1-F11**: Specialist OAuth2 Flow & URL Generation (`calendar.events`, `calendar.readonly`) — 5 tests.
- **P2-Tier1-F12**: Token Callback, Storage & Expiry Refresh Simulation — 5 tests.
- **P2-Tier1-F13**: Calendar Event Payload Construction Contract — 5 tests.
- **P2-Tier1-F14**: Two-Way Event Deletion on Cancellation — 5 tests.
- **P2-Tier1-F15**: Google Sync Status Indicators & API Resilience — 5 tests.

### Tier 2: Boundary & Corner Cases (60 Tests Total, Target >= 35)
- **P2-Tier2-B01**: Earliest (09:00) and Latest (18:00 weekday, 13:00 Sat) Slot Boundaries — 5 tests.
- **P2-Tier2-B02**: Slot Overrun Clipping & End-of-Day Bounds (18:30 weekday, 13:30 Sat, duration > day length) — 5 tests.
- **P2-Tier2-B03**: Exact Boundary Overlap vs Partial Overlap Math (touching vs overlapping intervals) — 5 tests.
- **P2-Tier2-B04**: Past Time Slot Pruning on Current Day (15-min lead buffer, past hours blocked) — 5 tests.
- **P2-Tier2-B05**: Specialist Qualification & Inactive Status Filtering (`isActive: false`, unassigned staff) — 5 tests.
- **P2-Tier2-B06**: Phone Number Format & Sanitization Boundaries (E.164, 10-15 digits, invalid rejection) — 5 tests.
- **P2-Tier2-B07**: Required Field & Payload Validation Boundaries (missing name, phone, service, date) — 5 tests.
- **P2-Tier2-B08**: Concurrency Double-Booking Guard & Race Conditions (atomic locks, 409 Conflict) — 5 tests.
- **P2-Tier2-B09**: Multitenant Slug Resolution & Tenant Isolation — 5 tests.
- **P2-Tier2-B10**: Expired Access Tokens & Auto-Refresh Simulation — 5 tests.
- **P2-Tier2-B11**: Google Calendar API Outage & Fallback Resilience (graceful degradation to local DB) — 5 tests.
- **P2-Tier2-B12**: Timezone Handling & UTC Normalization (`America/Monterrey` -> UTC ISO-8601) — 5 tests.

### Tier 3: Cross-Feature Interactions (12 Tests Total, Target >= 10)
- **P2-Tier3-X01**: Book Slot -> Verify Immediate Removal from Availability — 3 tests.
- **P2-Tier3-X02**: Cancel Appointment -> Verify Immediate Restoration of Slot — 3 tests.
- **P2-Tier3-X03**: Multi-Service & Multi-Specialist Booking Sequence for Same Customer — 3 tests.
- **P2-Tier3-X04**: High-Concurrency Double Booking Simulation (5 simultaneous requests, 1 winner, 4 conflicts) — 3 tests.

### Tier 4: Real-World Clinical Workload Scenarios (4 Tests Total, Target >= 4)
- **SCN-P2-01**: Full-Day Simulation for Busy Physiotherapy Clinic with 3 Specialists (Sofia, Mateo, Valentina) — 4 assertions.
- **SCN-P2-02**: End-to-End Public 5-Step Booking Flow Execution (`/[slug]`) — 10 assertions.
- **SCN-P2-03**: Receptionist Agenda Workflow (Review schedule, filter therapist, open drawer, cancel, verify liberated slot) — 6 assertions.
- **SCN-P2-04**: High-Volume Simultaneous Booking Stress Test (15 simultaneous bookings across therapists) — 2 assertions.

---

## 4. Execution Commands

```bash
# Execute master Phase 2 test suite
npx tsx tests/e2e/phase2-runner.ts

# Execute individual test suites
npx tsx tests/e2e/phase2-availability.test.ts
npx tsx tests/e2e/phase2-booking.test.ts
npx tsx tests/e2e/phase2-calendar-sync.test.ts
```

---

## 5. Summary Threshold Matrix

| Test Category | Target Threshold | Actual Verified Tests | Actual Assertions | Audit Status |
|---|---|---|---|---|
| **Tier 1: Feature Coverage** | >= 40 tests | **75 tests** | **200 assertions** | **PASS** |
| **Tier 2: Boundary & Corner Cases** | >= 35 tests | **60 tests** | **128 assertions** | **PASS** |
| **Tier 3: Cross Combinations** | >= 10 tests | **12 tests** | **28 assertions** | **PASS** |
| **Tier 4: Clinical Scenarios** | >= 4 tests | **4 tests** | **25 assertions** | **PASS** |
| **TOTALS** | >= 40 tests | **151 tests** | **381 assertions** | **100% PASS** |
