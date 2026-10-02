# Dispatch: E2E Test Writer (Phase 2 Test Suite)

## Context & Mission
You are the E2E Test Writer for APPCITAS Phase 2 (Google Calendar, Availability Engine, Agenda & Public Booking).
Your mission is to build a comprehensive, requirement-driven, opaque-box test harness and test suite covering all 4 tiers of testing.

## Inputs
- Authoritative requirements: /Users/jasbotello/devel/APPCITAS/.agents/ORIGINAL_REQUEST.md
- Specification report: /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_spec_miner_survey/spec_report.md
- Project architecture & contracts: /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_orchestrator_phase2/PROJECT.md
- Existing Phase 1 test harness: /Users/jasbotello/devel/APPCITAS/tests/e2e/run-tests.ts

## Responsibilities & Files Owned
You own exclusively:
- `tests/e2e/phase2-runner.ts` (or integration into `tests/e2e/` test suite)
- `tests/e2e/phase2-availability.test.ts`
- `tests/e2e/phase2-booking.test.ts`
- `tests/e2e/phase2-calendar-sync.test.ts`
- `/Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_orchestrator_phase2/TEST_READY.md`
- `/Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_orchestrator_phase2/TEST_INFRA.md`

## Testing Methodology (4 Tiers)
1. **Tier 1: Feature Coverage (>=5 tests per feature)**:
   - Operating hours filtering (Mon-Fri 09:00-19:00, Sat 09:00-14:00, Sun closed).
   - Time slot generation (~50 min duration in 60 min intervals).
   - Collision subtraction with existing database appointments.
   - Google Calendar freeBusy busy intervals subtraction.
   - Booking creation: Customer deduplication by phone + Appointment creation.
   - Status update (`PATCH /api/appointments/[id]`): marking `CANCELLED` liberates slot.
   - "Cualquier especialista disponible" pooling and least-busy therapist assignment.
   - Offline fallback simulation when Google credentials are not set.

2. **Tier 2: Boundary & Corner Cases (>=5 tests per feature)**:
   - Earliest slot boundary (09:00) and latest slot boundary (18:00 weekday, 13:00 Saturday).
   - Slot overrun clipping: appointment starting at 18:30 must not be offered if closing is 19:00.
   - Exact boundary overlap (appointment ends at 10:00, next starts at 10:00 -> not a collision).
   - Partial overlap (Google busy from 10:15 to 10:45 -> blocks 10:00 slot).
   - Expired Google tokens triggering refresh logic or simulation mode.
   - Invalid phone number format or missing required fields.

3. **Tier 3: Cross-Feature Interactions**:
   - Booking an appointment immediately removes that slot from subsequent availability queries.
   - Cancelling an appointment immediately restores that slot in subsequent availability queries.
   - Concurrent booking attempt for same slot (race condition) -> second booking returns conflict / error.
   - Customer booking two different services with different specialists on different days.

4. **Tier 4: Real-World Scenarios**:
   - Full-day simulation for a busy physiotherapy clinic with 3 specialists.
   - Public booking flow execution: Customer visits `/[slug]`, picks service, picks "Cualquier especialista", chooses slot, enters details, confirms appointment, verifies Google Calendar event ID returned and status is `CONFIRMED`.

## Verification & Deliverable
- Run your test suite using `npx tsx tests/e2e/phase2-runner.ts` (or `npx tsx tests/e2e/run-tests.ts`).
- Ensure all tests are self-contained and mock external network calls cleanly.
- Create `/Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_orchestrator_phase2/TEST_READY.md` detailing test counts, commands, and tier coverage.
- Write `handoff.md` and send a completion message to the orchestrator.

## 2026-09-14T19:11:22Z
You are the E2E Test Writer for APPCITAS Phase 2 (Google Calendar, Availability Engine, Agenda & Public Booking).
Your assigned working directory is: /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_test_writer_phase2/
Read your dispatch instructions at /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_test_writer_phase2/DISPATCH.md, the authoritative request at /Users/jasbotello/devel/APPCITAS/.agents/ORIGINAL_REQUEST.md, and spec_report.md at /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_spec_miner_survey/spec_report.md.
Build a comprehensive opaque-box test suite for Phase 2 covering Tiers 1-4 in tests/e2e/ (minimum 40+ tests across availability, operating hours, collisions, Google sync fallback, booking, and cancellation). Run the test harness, create TEST_READY.md at /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_orchestrator_phase2/TEST_READY.md, write your handoff.md, and send a completion message back.

