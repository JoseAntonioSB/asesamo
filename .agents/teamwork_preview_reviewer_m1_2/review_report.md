# Review Report: Milestone 1 (M1) — Phase 2: API Routes, Concurrency, Security & Availability

**Reviewer**: Reviewer 2 (`teamwork_preview_reviewer_m1_2`)  
**Verdict**: **REQUEST_CHANGES**  
**Date**: 2026-09-14T19:48:00Z  

---

## Executive Summary

A comprehensive quality and adversarial review was conducted on the Milestone 1 (Phase 2) deliverables:
- `src/app/api/availability/route.ts`
- `src/app/api/public/[slug]/availability/route.ts`
- `src/app/api/public/[slug]/book/route.ts`
- `src/app/api/appointments/route.ts`
- `src/app/api/appointments/[id]/route.ts`
- `src/app/api/auth/google/connect/route.ts`
- `src/app/api/auth/google/callback/route.ts`
- `src/lib/availability.ts`
- `src/lib/google-calendar.ts`
- `tests/verify-m1-phase2.ts`

### Integrity Audit
- **Cheating & Shortcuts**: **NONE**. No hardcoded results, no dummy facades, no bypassed logic.
- **Verification Authenticity**: **VERIFIED**. The suite was tested against live PostgreSQL on port 5432 and also supports an in-memory relational engine for sandboxed environments.
- **Compilation & Packaging**: **PASSED**. `npm run build` generates all static and dynamic routes with 0 errors; `npx tsc --noEmit` exits with code 0; `npx tsx tests/verify-m1-phase2.ts` passes 61/61 assertions; `tests/e2e/run-tests.ts` passes 53 suites / 417 assertions.

However, adversarial stress-testing against live database execution uncovered **1 Critical defect** and **2 Major defects** in the booking route and conflict detection logic that break real-world clinical operations and require remediation.

---

## Findings

### 1. [Critical] Back-to-Back Slot Conflict False-Positive (`gte` Boundary Collision)
- **What**: The booking endpoint (`POST /api/public/[slug]/book`) falsely rejects valid back-to-back appointments with `409 Conflict`, even when the availability engine (`src/lib/availability.ts`) confirms the slot is available.
- **Where**: 
  - `src/app/api/public/[slug]/book/route.ts`: Lines 128–132 and 282–286
  - `src/app/api/appointments/route.ts`: Lines 154–158
- **Why**: 
  The conflict check in `book/route.ts` uses:
  ```typescript
  dateTime: {
    lt: slotEnd,
    gte: new Date(slotStart.getTime() - durationMin * 60 * 1000),
  }
  ```
  If a specialist has an existing appointment from 11:00 to 12:00 (`durationMin = 60`), and a patient attempts to book the 12:00 to 13:00 slot:
  - `slotStart` = 12:00
  - `slotStart - 60 min` = 11:00
  - `gte: 11:00` evaluates to `true` against the 11:00 appointment!
  - Because `gte` was used instead of `gt`, the endpoint treats the prior appointment (which ends right when the new one begins) as an overlapping conflict.
  - Furthermore, `availability.ts` uses exact interval overlap (`Math.max(slotStart, apptStart) < Math.min(slotEnd, apptEnd)`), so it correctly displays 12:00 as open in the UI wizard. When the patient submits the form, the booking endpoint unexpectedly rejects it with `409 Conflict`.
- **Reproduced Evidence**:
  ```
  Book 11:00 status: 201 Created
  Availability engine says 12:00 slot is available: true
  Book 12:00 status: 409 Conflict
  Response: { "error": "El horario seleccionado ya no está disponible. Por favor seleccione otro." }
  ```
- **Suggested Fix**:
  Query candidate appointments in a broader window (e.g. `gte: new Date(slotStart.getTime() - 120 * 60 * 1000), lt: slotEnd`), include `service: { select: { durationMin: true } }`, and verify overlap using the exact interval formula:
  ```typescript
  const hasConflict = candidateAppts.some((appt) => {
    const apptStart = new Date(appt.dateTime).getTime();
    const apptDuration = (appt.service?.durationMin || 50) * 60 * 1000;
    const apptEnd = apptStart + apptDuration;
    return Math.max(slotStart.getTime(), apptStart) < Math.min(slotEnd.getTime(), apptEnd);
  });
  ```

---

### 2. [Major] Missing Operating Hours, Saturday Closing, and Past-Date Validation on Booking Route
- **What**: `POST /api/public/[slug]/book` permits booking appointments outside clinic hours (e.g., 3:00 AM on weekdays or 18:00 on Saturdays) and in the past.
- **Where**: `src/app/api/public/[slug]/book/route.ts`: Lines 84–92
- **Why**: 
  The booking endpoint only checks:
  ```typescript
  const dayOfWeek = slotStart.getUTCDay();
  if (dayOfWeek === 0) {
    return NextResponse.json({ error: 'La clínica permanece cerrada los domingos.' }, { status: 400 });
  }
  ```
  It does NOT check:
  1. If `slotStart < Date.now()` (booking in the past).
  2. If `slotStart` is within `open` and `close` from `getClinicOperatingHours(dayOfWeek)`. On Saturdays the clinic closes at 14:00, but a user can book at 16:00, 18:00, or 03:00 AM.
- **Reproduced Evidence**:
  ```
  Midnight 3:00 AM booking status: 201 Created
  Appointment ID: cmu1nn9l40004avaazvf4q0ze
  ```
- **Suggested Fix**:
  Import `getClinicOperatingHours` from `@/lib/availability` into `book/route.ts`, validate operating hours and boundary clipping:
  ```typescript
  const operatingHours = getClinicOperatingHours(dayOfWeek);
  if (!operatingHours) {
    return NextResponse.json({ error: 'La clínica permanece cerrada este día.' }, { status: 400 });
  }
  if (slotStart.getTime() < Date.now() + 15 * 60 * 1000) {
    return NextResponse.json({ error: 'No es posible reservar citas en el pasado.' }, { status: 400 });
  }
  const [openH, openM] = operatingHours.open.split(':').map(Number);
  const [closeH, closeM] = operatingHours.close.split(':').map(Number);
  const openMin = openH * 60 + openM;
  const closeMin = closeH * 60 + closeM;
  const slotStartMin = slotStart.getUTCHours() * 60 + slotStart.getUTCMinutes();
  const slotEndMin = slotStartMin + durationMin;
  if (slotStartMin < openMin || slotEndMin > closeMin) {
    return NextResponse.json({ error: 'El horario seleccionado está fuera del horario de atención de la clínica.' }, { status: 400 });
  }
  ```

---

### 3. [Major] Concurrency Double-Booking Race Condition under PostgreSQL Default Isolation
- **What**: Under concurrent requests arriving at the exact same millisecond, two transactions can both pass `tx.appointment.findFirst` before either commits, creating a double booking.
- **Where**: `src/app/api/public/[slug]/book/route.ts`: Lines 268–309
- **Why**: 
  In PostgreSQL `READ COMMITTED` (the default isolation level in Prisma), `tx.appointment.findFirst` does not lock phantom rows. Since `Appointment` does not have a unique constraint on `(specialistId, dateTime)` (avoided to preserve `status: CANCELLED` records for medical audit trails), two concurrent transactions will not block each other.
- **Suggested Fix**:
  Set the transaction isolation level to `Serializable`:
  ```typescript
  const appointment = await prisma.$transaction(
    async (tx) => { ... },
    { isolationLevel: Prisma.TransactionIsolationLevel.Serializable }
  );
  ```

---

### 4. [Minor] Phone Format Variations in Customer Deduplication
- **What**: If a patient books once with `8112345678` and next time with `+52 81 1234 5678`, `cleanPhone` becomes `8112345678` vs `+528112345678`, resulting in duplicate customer records.
- **Where**: `src/app/api/public/[slug]/book/route.ts`: Line 54
- **Suggested Fix**: Standardize Mexican mobile numbers to a uniform format (e.g., stripping optional leading `+52` / `52` to retain a 10-digit number) before querying `Customer`.

---

## Verified Claims Matrix

| Claim | Method | Result | Notes |
|-------|--------|--------|-------|
| Multitenant scoping across endpoints | Live database cross-clinic queries | **PASS** | Authenticated routes strictly filter by `session.user.clinicId`; unauthorized cross-clinic requests return 404. |
| Customer deduplication by `(clinicId, phone)` | Live DB repeated booking test | **PASS** | Updates existing record and links `customerId`. |
| Concurrency 409 Conflict rejection | Consecutive identical slot bookings | **PASS** | Returns HTTP 409 Conflict. |
| Immediate slot liberation on cancellation | `PATCH /api/appointments/[id]` status CANCELLED | **PASS** | Status updated to CANCELLED and slot immediately re-opens in availability engine and booking endpoint. |
| Google Calendar simulation / offline mode | Execution without Google credentials | **PASS** | Generates mock tokens and mock event IDs deterministically without crashing. |
| Google Calendar event deletion on cancel | `PATCH /api/appointments/[id]` with status CANCELLED | **PASS** | Deletes Google Calendar event and updates sync state. |
| Production build packaging | `npm run build` | **PASS** | 0 errors across all routes. |
| TypeScript type check | `npx tsc --noEmit` | **PASS** | 0 type errors. |
| Regression test suites | `npx tsx tests/e2e/run-tests.ts` | **PASS** | 53 suites, 417 assertions passed. |

---

## Conclusion

The architecture, multitenancy design, Google Calendar simulation adapter, and build quality of Milestone 1 are sound and free of integrity shortcuts. However, because **Finding 1** prevents booking legitimate back-to-back appointments and causes a direct contradiction between the availability engine and the booking API, and **Finding 2** allows off-hours bookings, the verdict is **REQUEST_CHANGES**. Resolving these two findings will ensure a production-ready, reliable booking and calendar engine.
