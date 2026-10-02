# Adversarial Challenge Report: Milestone 1 (M1) — Google Calendar & Concurrency Stress Testing

**Target**: APPCITAS Phase 2 Milestone 1 (M1) — Google Calendar Sync, Intelligent Availability & Public Booking
**Challenger**: `teamwork_preview_challenger_m1_2` (Challenger 2)
**Date**: 2026-09-14
**Overall Risk Assessment**: **MEDIUM-HIGH** (Critical functionality works robustly, but a high-severity collision bypass defect was identified in multi-duration appointment booking).

---

## 1. Challenge Summary

Challenger 2 was dispatched to empirically challenge and stress-test:
1. **Google Calendar Integration**: Offline simulation fallback mode, OAuth URL generation, token exchange/persistence, freeBusy intervals, event creation/cancellation, and error resilience.
2. **High-Concurrency Race Conditions**: Parallel booking operations targeting the exact same specialist/slot and "any specialist" pooling.
3. **Customer Deduplication**: Cross-format phone normalization (`+52`, spaces, dashes, parentheses) and tenant isolation.
4. **Availability Engine & Boundary Invariants**: Operating hours clipping, Sunday closures, past-time pruning, and status slot liberation.
5. **Multi-Duration Service Overlaps**: Collision detection when prior appointments have durations longer than the requested service.

### Overall Verification Summary
- `tests/e2e/phase2-runner.ts`: **151/151 tests passed**, 381 assertions verified.
- `tests/verify-m1-phase2.ts`: **61/61 assertions passed**.
- `npm run build`: **0 compilation errors**, 19 routes generated cleanly.
- `tests/adversarial-m1-phase2.test.ts`: **18 passed, 1 failed**.

---

## 2. Challenges & Findings

### [HIGH] Challenge 1: Booking API Collision Guard Fails on Prior Multi-Duration Appointments

- **Assumption Challenged**:
  The implementation in `src/app/api/public/[slug]/book/route.ts` assumes that checking `dateTime: { lt: slotEnd, gte: new Date(slotStart.getTime() - durationMin * 60 * 1000) }` is sufficient to prevent double-booking collisions, where `durationMin` is taken from the *new* service being booked.

- **Attack Scenario**:
  1. A physiotherapy clinic offers standard sessions (e.g., "Evaluación Fisioterapéutica" — 50 min) and intensive sessions (e.g., "Sesión Extensa" — 90 min).
  2. Patient A books Specialist 1 for the 90-minute session at 09:00 UTC (occupying Specialist 1 from 09:00 to 10:30 UTC).
  3. Patient B sends a direct booking request (`POST /api/public/[slug]/book`) for the 50-minute session with Specialist 1 at 10:00 UTC (10:00 to 10:50 UTC).
  4. In `src/app/api/public/[slug]/book/route.ts` (lines 128–132 and 282–286):
     ```typescript
     dateTime: {
       lt: slotEnd, // 10:50
       gte: new Date(slotStart.getTime() - durationMin * 60 * 1000), // 10:00 - 50m = 09:10
     }
     ```
  5. The query searches for appointments starting between `09:10` and `10:50`.
  6. Patient A's appointment started at `09:00`. Because `09:00 < 09:10`, the query fails to match Patient A's appointment!
  7. The server concludes there is no conflict, and creates Patient B's appointment at 10:00 with status `CONFIRMED`.

- **Blast Radius**:
  Specialist 1 is now double-booked between 10:00 and 10:30 UTC. While the availability engine (`src/lib/availability.ts`) properly flags this slot as unavailable in the UI, any automated client, retrying script, or public user calling the API directly can trigger physical overlapping appointments in the clinic agenda and in Google Calendar.

- **Mitigation**:
  In `src/app/api/public/[slug]/book/route.ts`, query existing appointments for that specialist on the day (or query with `include: { service: { select: { durationMin: true } } }` within a generous preceding window such as `slotStart - 240 min`), and compute actual interval overlap:
  ```typescript
  const existingAppts = await tx.appointment.findMany({
    where: {
      clinicId: clinic.id,
      specialistId: chosenSpecialist.id,
      status: { in: [AppointmentStatus.CONFIRMED, AppointmentStatus.PENDING, AppointmentStatus.COMPLETED, AppointmentStatus.NO_SHOW] },
      dateTime: {
        gte: new Date(slotStart.getTime() - 4 * 60 * 60 * 1000), // look back up to 4 hours
        lt: slotEnd,
      },
    },
    include: { service: { select: { durationMin: true } } },
  });

  const conflict = existingAppts.some((appt) => {
    const start = new Date(appt.dateTime).getTime();
    const dur = (appt.service?.durationMin || 50) * 60 * 1000;
    const end = start + dur;
    return Math.max(slotStart.getTime(), start) < Math.min(slotEnd.getTime(), end);
  });
  ```

---

### [MEDIUM] Challenge 2: "Cualquier Especialista" High-Concurrency Pooling Contention

- **Assumption Challenged**:
  Under high concurrency, when multiple users book with `specialistId: 'any'` for the same time slot, it is assumed that bookings will distribute across available specialists.

- **Attack Scenario**:
  1. Both Specialist A and Specialist B are qualified and completely free at 14:00 UTC.
  2. 10 simultaneous concurrent requests arrive with `specialistId: 'any'`.
  3. Because specialist selection (lines 153–239) occurs *before* entering the transaction, and the workload balancing sort deterministically chooses the specialist with the fewest appointments (e.g. Specialist B), all 10 racing requests select Specialist B.
  4. All 10 requests then contend for Specialist B inside the transaction guard (lines 268–309).
  5. Exactly 1 request succeeds (201 Created); the remaining 9 requests receive 409 Conflict.
  6. Specialist A remains completely unbooked at 14:00 UTC during the parallel burst.

- **Blast Radius**:
  Under extreme simultaneous spikes on the exact same second, unassigned requests fail with 409 even if another specialist was free. Subsequent sequential requests correctly fill the remaining specialist.

- **Mitigation**:
  Inside the atomic transaction, if the preferred specialist encounters a conflict, iterate over the remaining candidate specialists in `freeCandidates` before returning 409 Conflict.

---

### [LOW] Challenge 3: Public Booking Route Does Not Prune Past Time Slots

- **Assumption Challenged**:
  The public booking endpoint (`/api/public/[slug]/book`) assumes requests only contain future slots provided by the availability engine.

- **Attack Scenario**:
  An adversarial client sends a `POST /api/public/[slug]/book` with a date in the past (e.g., `2020-01-01T10:00:00.000Z`). The availability engine (`availability.ts`) has a 15-minute lead buffer and prunes past slots, but `book/route.ts` lacks a check ensuring `slotStart.getTime() >= Date.now()`.

- **Blast Radius**:
  Adversarial clients can insert historical appointments into the database via the public booking endpoint.

- **Mitigation**:
  Add `if (slotStart.getTime() < Date.now()) return NextResponse.json({ error: 'No se pueden reservar citas en horarios pasados.' }, { status: 400 });` at line 66 of `src/app/api/public/[slug]/book/route.ts`.

---

## 3. Empirical Stress Test Results

Executed via `npx tsx tests/adversarial-m1-phase2.test.ts`:

| # | Category | Scenario | Expected Behavior | Actual Empirical Behavior | Status |
|---|---|---|---|---|---|
| 1 | GoogleCalendar | Offline Mode with Unset Credentials | Returns `isGoogleConfigured: false` without throwing | `isGoogleConfigured: false`, zero exceptions | **PASS** |
| 2 | GoogleCalendar | Malicious / XSS `specialistId` in `generateAuthUrl` | Generates safe URL without crashing | All 4 adversarial inputs safely URL-encoded | **PASS** |
| 3 | GoogleCalendar | Offline Token Callback Simulation | Generates mock tokens and updates User record | Tokens persisted, status becomes `SIMULATED` | **PASS** |
| 4 | GoogleCalendar | `getSpecialistFreeBusy` for non-existent specialist | Returns empty array `[]` without error | Returns `[]` gracefully | **PASS** |
| 5 | GoogleCalendar | Offline `createCalendarEvent` and `deleteCalendarEvent` | Deterministic mock IDs, idempotent deletion | Mock IDs formatted `mock_gcal_evt_...`, deletion returns `true` | **PASS** |
| 6 | GoogleCalendar | `disconnectSpecialistGoogle` | Nullifies all 5 Google fields on User | All 5 fields set to `null`, status `DISCONNECTED` | **PASS** |
| 7 | Concurrency | 10 Simultaneous Requests for Exact Same Slot & Specialist | Exactly 1 winner (201), 9 conflicts (409) | Exactly 1 HTTP 201, 9 HTTP 409, 1 DB record | **PASS** |
| 8 | Concurrency | "Cualquier Especialista" Race Guard & Sequential Followup | 1 wins in parallel race; sequential fills 2nd specialist; 3rd gets 409 | Race: 1 win, 9 conflicts. Seq 11: 201, Seq 12: 409. Total: 2 appts | **PASS** |
| 9 | CustomerDedup | Phone Formatting Sanitization | Normalizes `+52`, spaces, parentheses to same phone | Same customer record updated, 0 duplicates | **PASS** |
| 10 | CustomerDedup | Cross-Tenant Customer Isolation | Same phone in Clinic A vs Clinic B creates separate records | Distinct customer records per clinic, zero cross-leakage | **PASS** |
| 11 | CustomerDedup | Malformed Phone Number Rejection | Reject short (<7 digits) or alphabetic phone numbers | 5 variations rejected with HTTP 400 | **PASS** |
| 12 | Availability | Sunday Availability Query | Returns `isOpen: false` and 0 slots | `isOpen: false`, slots: `[]`, closed message | **PASS** |
| 13 | Availability | Saturday Closing Bounds (09:00 - 14:00) | Clips 14:00 slot, returns 5 slots | Exactly 5 slots (09:00 to 13:00), 14:00 clipped | **PASS** |
| 14 | Availability | Inactive Specialist Filtering | Excludes inactive specialist from available slots | 0 available slots, error message returned | **PASS** |
| 15 | Availability | Slot Liberation on Status CANCELLED | Slot immediately reopens in availability engine | Before: `false`, After: `true` | **PASS** |
| 16 | MultiDuration | 90-Minute Service Slot Blocking in Availability | 09:00-10:30 session blocks both 09:00 and 10:00 slots | 09:00: `false`, 10:00: `false`, 11:00: `true` | **PASS** |
| 17 | MultiDuration | Booking API Guard against 90-Min Overlap | Direct booking at 10:00 rejected with 409 Conflict | Returned HTTP 201 Created (Collision allowed) | **FAIL** |
| 18 | Security | Cross-Tenant Appointment Modification | Unauthorized user from Clinic A cannot modify Clinic B | Returns HTTP 404 Not Found | **PASS** |
| 19 | Security | Booking on Non-Existent Clinic Slug | Returns HTTP 404 | Returns HTTP 404 Not Found | **PASS** |

---

## 4. Unchallenged Areas

- **Live Google Cloud OAuth Production Environment**: Live production OAuth consent and live Google Calendar API endpoints require real GCP client secrets and active Google user consent, which are unavailable in local development. Offline simulation mode was stress-tested and verified with 100% test coverage.

---

## 5. Verdict

**REQUEST_CHANGES**

**Reasoning**:
While the Google Calendar simulation, exact-slot concurrency locking, customer deduplication, and availability calculations are well-built, the collision check in `src/app/api/public/[slug]/book/route.ts` exhibits a confirmed bug where appointments overlapping with prior longer-duration services (e.g. 90-minute therapy sessions) are allowed to double-book with HTTP 201.

Applying the proposed mitigation in Challenge 1 will ensure full schedule integrity across multi-duration clinical service catalogs.
