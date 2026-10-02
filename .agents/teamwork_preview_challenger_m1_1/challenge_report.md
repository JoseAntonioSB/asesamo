# Adversarial Challenge Report: APPCITAS Phase 2 Milestone 1 (M1)
**Target**: Availability Engine (`src/lib/availability.ts`) & Public Booking Route (`src/app/api/public/[slug]/book/route.ts`)  
**Investigator**: Challenger 1 (`teamwork_preview_challenger_m1_1`)  
**Date**: 2026-09-14T19:48:00Z  
**Verdict**: **REQUEST_CHANGES**

---

## Executive Challenge Summary

**Overall Risk Assessment**: **CRITICAL**

An empirical adversarial review was conducted on the Milestone 1 availability engine, operating schedule boundaries, interval overlap algebra, cancelled appointment slot liberation, multi-specialist pooling, and public booking collision guards.

While core operating hours (Mon-Fri 09:00-19:00, Sat 09:00-14:00, Sun closed) and standard single-duration availability computations pass verification, empirical adversarial stress-testing revealed **1 CRITICAL vulnerability**, **1 HIGH vulnerability**, and **1 MEDIUM operational flaw** that compromise schedule integrity and allow overlapping double-bookings:

1. **[CRITICAL] Asymmetric Service Duration Collision in Public Booking (`src/app/api/public/[slug]/book/route.ts`)**:
   When an appointment exists for a longer service (e.g. 90m or 120m), the booking endpoint's collision query only subtracts the duration of the *new* service being booked (e.g. 50m). An overlapping booking during the latter portion of the existing appointment passes validation, returns `HTTP 201 Created`, and successfully persists an overlapping double-booking.
2. **[HIGH] Concurrency Double-Booking Race Condition under Parallel Bursts**:
   The booking transaction in `POST /api/public/[slug]/book` relies on `tx.appointment.findFirst` under default PostgreSQL Read Committed isolation without row-level locks (`SELECT ... FOR UPDATE`), advisory locks, or database unique exclusion constraints. Under simultaneous parallel requests for the exact same slot and specialist, multiple requests commit duplicate bookings.
3. **[MEDIUM] Premature Current-Day Afternoon Slot Pruning via UTC `Z` Normalization**:
   In `src/lib/availability.ts`, `buildDateTime` hardcodes UTC string generation (`T${timeStr}:00.000Z`). In local clinic timezones like Monterrey/CDMX (UTC-6), afternoon slots (14:00-18:00) are evaluated as 14:00Z-18:00Z. When a patient queries today's availability in the afternoon (e.g. 13:36 local / 19:36 UTC), ALL remaining slots for the day are pruned as past time, locking patients out of same-day appointments.
4. **[INFORMATIONAL] E2E Availability Test Suite Decoupled from Production Code**:
   `tests/e2e/phase2-availability.test.ts` executes all assertions against an in-file `ReferenceAvailabilityEngine` class rather than importing `src/lib/availability.ts`. Production availability code was tested only via `tests/verify-m1-phase2.ts`.

---

## Challenges & Empirical Findings

### [CRITICAL] Challenge 1: Asymmetric Service Duration Overlap in Public Booking Route

- **Assumption Challenged**: The booking endpoint (`src/app/api/public/[slug]/book/route.ts`) will reject any appointment that collides with an existing appointment.
- **Flawed Code**: `src/app/api/public/[slug]/book/route.ts`, lines 130-131 and lines 283-285:
  ```typescript
  dateTime: {
    lt: slotEnd,
    gte: new Date(slotStart.getTime() - durationMin * 60 * 1000),
  }
  ```
- **Attack Scenario**:
  1. Specialist Bruno Diaz is booked for a 90-minute service (`Terapia Intensiva 90m`) starting at 13:00 on 2026-11-04 (occupying 13:00 to 14:30).
  2. An external patient or bot submits `POST /api/public/[slug]/book` requesting a 50-minute service (`Terapia Estándar 50m`) with Specialist Bruno Diaz starting at 14:00.
  3. In `book/route.ts`, `durationMin = 50`.
  4. `slotStart` = 14:00.
  5. The query checks: `dateTime >= 14:00 - 50m = 13:10` and `dateTime < 14:50`.
  6. The existing appointment started at 13:00. Because `13:00 < 13:10`, it does NOT match `gte: 13:10`!
  7. Both the preliminary collision check and the atomic transaction check return `null`.
  8. `publicBook` returns `201 Created` and creates appointment `apt_adv_52...` at 14:00.
- **Empirical Evidence**:
  Executed via `tests/adversarial-phase2-m1.test.ts`:
  ```
  ✖ FAIL [CRITICAL]: Booking route rejects booking overlapping with longer prior appointment (13:00 90m vs 14:00 50m)
  Details: VULNERABILITY DETECTED: Returned status 201 with body {
    "success": true,
    "appointment": {
      "id": "apt_adv_52_...",
      "dateTime": "2026-11-04T14:00:00.000Z",
      "status": "CONFIRMED",
      "service": { "durationMin": 50, "name": "Terapia Estándar 50m" },
      "specialist": { "name": "Lic. Bruno Diaz" }
    }
  }
  ```
- **Blast Radius**: Double-booked specialists when clinics offer services of varying durations (50m, 60m, 90m, 120m).
- **Suggested Mitigation**:
  In `src/app/api/public/[slug]/book/route.ts`, use a wide lookback window (or full day range) and fetch existing appointments with `service: { select: { durationMin: true } }`, applying the exact interval overlap formula already used in `src/lib/availability.ts`:
  ```typescript
  const candidateAppointments = await prisma.appointment.findMany({
    where: {
      clinicId: clinic.id,
      specialistId: chosenSpecialist.id,
      status: { in: [AppointmentStatus.CONFIRMED, AppointmentStatus.PENDING, AppointmentStatus.COMPLETED, AppointmentStatus.NO_SHOW] },
      dateTime: {
        lt: slotEnd,
        gte: new Date(slotStart.getTime() - 240 * 60 * 1000), // 4h maximum service lookback
      },
    },
    include: { service: { select: { durationMin: true } } },
  });

  const hasConflict = candidateAppointments.some((appt) => {
    const apptStart = new Date(appt.dateTime).getTime();
    const apptEnd = apptStart + (appt.service?.durationMin || 50) * 60 * 1000;
    return Math.max(slotStart.getTime(), apptStart) < Math.min(slotEnd.getTime(), apptEnd);
  });

  if (hasConflict) {
    return NextResponse.json({ error: 'El horario seleccionado ya no está disponible.' }, { status: 409 });
  }
  ```

---

### [HIGH] Challenge 2: Concurrency Double-Booking Race Condition under Parallel Bursts

- **Assumption Challenged**: The atomic transaction in `book/route.ts` prevents concurrent race double-bookings.
- **Flawed Code**: `src/app/api/public/[slug]/book/route.ts`, lines 268-309:
  ```typescript
  const appointment = await prisma.$transaction(async (tx) => {
    const finalConflict = await tx.appointment.findFirst({ ... });
    if (finalConflict) throw new Error('SLOT_CONFLICT');
    return await tx.appointment.create({ ... });
  });
  ```
- **Attack Scenario**:
  5 concurrent HTTP requests arrive simultaneously for the exact same specialist and time slot.
  Under PostgreSQL's default Read Committed transaction isolation level, all 5 transactions execute their `findFirst` query before any of the transactions commit their `create`. All 5 transactions observe 0 rows, bypass the conflict check, and create 5 overlapping appointments.
- **Empirical Evidence**:
  Executed via `tests/adversarial-phase2-m1.test.ts`:
  ```
  ✖ FAIL [HIGH]: Concurrency Race Guard under parallel burst: Exactly 1 succeeds with 201 and others with 409 (got 5 successes, 0 conflicts)
  Details: RACE CONDITION: Application allows 5 simultaneous bookings for exact same slot
  ```
- **Blast Radius**: High-demand slots or duplicate clicks by users create multiple confirmed appointments for a single therapist.
- **Suggested Mitigation**:
  Implement PostgreSQL advisory locking (`await tx.$executeRaw\`SELECT pg_advisory_xact_lock(hashtext(${chosenSpecialist.id || 'any'}))\``) or serializable transaction isolation, or an application-level mutex around slot booking.

---

### [MEDIUM] Challenge 3: Premature Current-Day Afternoon Slot Pruning via UTC `Z` Mismatch

- **Assumption Challenged**: Time slot past pruning correctly prunes past times while keeping remaining business hours bookable.
- **Flawed Code**: `src/lib/availability.ts`, lines 77-79 and line 306:
  ```typescript
  export function buildDateTime(dateStr: string, timeStr: string): Date {
    return new Date(`${dateStr}T${timeStr}:00.000Z`);
  }
  ...
  const isPastTime = slotStartDate.getTime() < now + leadBufferMs;
  ```
- **Attack Scenario**:
  Local time in Mexico City / Monterrey is `2026-09-14T13:36:30-06:00` (1:36 PM).
  Current UTC time is `2026-09-14T19:36:30Z`.
  A customer requests availability for today (`date = "2026-09-14"`).
  The slot grid generates 14:00, 15:00, 16:00, 17:00, 18:00 (which are 2:00 PM - 6:00 PM local time).
  `buildDateTime` creates `2026-09-14T14:00:00.000Z`.
  `slotStartDate.getTime()` is 14:00 UTC.
  Because 14:00 UTC < 19:36 UTC, all afternoon slots are marked `isPastTime = true`.
- **Empirical Evidence**:
  Executed via `tests/adversarial-phase2-m1.test.ts`:
  ```
  Slots evaluated for today (2026-09-14): 09:00:BLOCKED, 10:00:BLOCKED, 11:00:BLOCKED, 12:00:BLOCKED, 13:00:BLOCKED, 14:00:BLOCKED, 15:00:BLOCKED, 16:00:BLOCKED, 17:00:BLOCKED, 18:00:BLOCKED
  ✖ FAIL [MEDIUM]: Afternoon slots on current day (14:00-18:00) remain bookable during afternoon business hours (13:36 local)
  Details: TIMEZONE GAP: All slots for today (including 14:00-18:00) were clipped as past time because UTC comparison treated them as 14:00Z-18:00Z (before current UTC 19:36Z). Available count: 0
  ```
- **Blast Radius**: Same-day bookings become impossible after 12:45 PM local clinic time.
- **Suggested Mitigation**:
  Normalize timezone offset when comparing against `now` or allow clinic timezone configuration (`America/Monterrey` / UTC-6) in `buildDateTime`.

---

## Stress Test Results Summary

| Test Case / Scenario | Expected Behavior | Actual Behavior | Result |
|---|---|---|---|
| Sunday operating hours across months | 0 slots, isOpen: false | 0 slots, isOpen: false | **PASS** |
| Saturday 50m operating hours (09:00-14:00) | Exactly 5 slots (09:00-13:00), 14:00 clipped | 5 slots, 14:00 clipped | **PASS** |
| Weekday 50m operating hours (09:00-19:00) | Exactly 10 slots (09:00-18:00), 19:00 clipped | 10 slots, 19:00 clipped | **PASS** |
| Leap Year Day (2028-02-29 Tuesday) | 10 slots, 09:00 to 18:00 | 10 slots, 09:00 to 18:00 | **PASS** |
| Saturday 90m service clipping | Clips 13:00 slot, returns 4 slots (09:00-12:00) | 4 slots, last slot 12:00 | **PASS** |
| Saturday 120m service clipping | Includes 12:00 (ends 14:00), clips 13:00 | 4 slots, last slot 12:00 | **PASS** |
| Weekday 90m service clipping | Clips 18:00 slot, returns 9 slots (09:00-17:00) | 9 slots, last slot 17:00 | **PASS** |
| Weekday 120m service clipping | Includes 17:00 (ends 19:00), clips 18:00 | 9 slots, last slot 17:00 | **PASS** |
| Service exceeding daily hours (660m) | 0 slots, isOpen: true, no crash | 0 slots, no crash | **PASS** |
| Touching boundary (10:00-11:00 vs 11:00-12:00) | 11:00 slot is FREE (0ms overlap) | 11:00 slot is FREE | **PASS** |
| Touching boundary (09:00-10:00 vs 10:00-11:00) | 09:00 slot is FREE | 09:00 slot is FREE | **PASS** |
| Overlapping appointment tail (13:00-14:30) | 13:00 & 14:00 BLOCKED, 15:00 FREE | 13:00 & 14:00 BLOCKED, 15:00 FREE | **PASS** |
| Multi-status collision (NO_SHOW, PENDING, COMPLETED) | All 3 statuses block slot | All 3 statuses block slot | **PASS** |
| Slot liberation on CANCELLED | Slot immediately liberated | Slot immediately liberated | **PASS** |
| Intermingled cancelled + active on same slot | Remained BLOCKED | Remained BLOCKED | **PASS** |
| All active specialists booked | Pooled slot unavailable, no ID leaked | Unavailable, specialistId undefined | **PASS** |
| Inactive specialist assigned to service | Inactive specialist ignored, slot unavailable | Inactive specialist ignored | **PASS** |
| Workload balancing across therapists | Picks specialist with fewest bookings today | Correctly picked Spec B | **PASS** |
| Sequential double-booking attempt | First succeeds (201), second rejected (409) | 201, then 409 | **PASS** |
| Parallel burst double-booking attempt | 1 succeeds (201), 4 rejected (409) | **5 succeeded (201), 0 rejected (409)** | **FAIL** |
| Asymmetric service duration collision (90m vs 50m) | Rejected with 409 Conflict | **Allowed with 201 Created (Double Booking)** | **FAIL** |
| Post-cancellation slot re-booking | Slot immediately re-bookable (201) | Succeeded (201) | **PASS** |
| Current-day afternoon slot availability (13:36 local) | 14:00-18:00 slots bookable | **All 09:00-18:00 slots blocked (UTC Z mismatch)** | **FAIL** |

---

## Verdict & Recommended Action

**Verdict**: **REQUEST_CHANGES**

**Required Changes for Approval**:
1. Fix `src/app/api/public/[slug]/book/route.ts` collision query to check true interval overlap against existing appointment durations, preventing double-bookings when services have asymmetric durations (e.g. 90m vs 50m).
2. Harden the booking transaction in `src/app/api/public/[slug]/book/route.ts` with mutual exclusion or advisory lock against parallel concurrent double-booking bursts.
3. Fix timezone handling in `src/lib/availability.ts` to prevent premature same-day afternoon slot pruning.
