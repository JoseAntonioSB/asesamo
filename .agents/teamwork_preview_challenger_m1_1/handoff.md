# Handoff Report: Challenger 1 (Milestone 1 — Availability Engine Adversarial Stress Testing)

**Role**: Challenger 1 (critic, specialist)  
**Assigned Working Directory**: `/Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_challenger_m1_1/`  
**Milestone**: Phase 2 Milestone 1 (M1 — Google Calendar & Availability Engine)  
**Target Modules**: `src/lib/availability.ts`, `src/lib/google-calendar.ts`, `src/app/api/public/[slug]/book/route.ts`, `src/app/api/appointments/[id]/route.ts`  
**Timestamp**: 2026-09-14T19:50:00Z  
**Verdict**: **REQUEST_CHANGES**

---

## 1. Observation

### 1.1 Test Runner Executions
1. **Dedicated Phase 2 M1 Verification Suite (`tests/verify-m1-phase2.ts`)**:
   - Command: `npx tsx tests/verify-m1-phase2.ts`
   - Result: Exited with code 0 (61 passed, 0 failed).
   - Validated:
     - Google Calendar OAuth generation, callback simulation, freeBusy array structure.
     - Operating hours schedule: Mon-Fri 09:00-19:00, Sat 09:00-14:00, Sun null.
     - Basic collision subtraction and slot liberation for identical 50m durations.
     - Customer deduplication by `(clinicId, phone)` and sequential duplicate rejection.

2. **Master Phase 2 E2E Suite (`tests/e2e/phase2-runner.ts`)**:
   - Command: `npx tsx tests/e2e/phase2-runner.ts`
   - Result: Exited with code 0 (151 tests passed, 381 assertions verified).
   - Critical Architectural Discovery:
     Inspection of `tests/e2e/phase2-availability.test.ts` reveals that the test suite does **NOT** import or execute `src/lib/availability.ts`. It imports only `./harness` and instantiates an internal duplicate mock class `ReferenceAvailabilityEngine` (lines 68–250). Production availability code was completely decoupled from `tests/e2e/phase2-runner.ts`.

### 1.2 Independent Adversarial Stress Harness (`tests/adversarial-phase2-m1.test.ts`)
To empirically stress-test the production implementation modules (`src/lib/availability.ts`, `src/app/api/public/[slug]/book/route.ts`, and `src/app/api/appointments/[id]/route.ts`), an adversarial harness was executed with 40 empirical tests across 7 challenge categories:
- Command: `npx tsx tests/adversarial-phase2-m1.test.ts`
- Result: **37 PASSED, 3 FAILED** (Exit code 1).

### 1.3 Verbatim Empirical Failures & Code Inspection

#### A. CRITICAL Vulnerability: Asymmetric Service Duration Collision in Public Booking
- **File**: `src/app/api/public/[slug]/book/route.ts`
- **Lines**: 130–131 and 283–285:
  ```typescript
  dateTime: {
    lt: slotEnd,
    gte: new Date(slotStart.getTime() - durationMin * 60 * 1000),
  }
  ```
- **Observed Behavior**:
  Specialist B (`Lic. Bruno Diaz`) has a 90-minute appointment (`Terapia Intensiva 90m`) booked at 13:00 on 2026-11-04 (occupies 13:00 to 14:30).
  A new booking request arrives for a 50-minute service (`Terapia Estándar 50m`) at 14:00.
  In `book/route.ts`, `durationMin = 50`.
  `new Date(slotStart.getTime() - durationMin * 60 * 1000)` calculates `14:00 - 50m = 13:10`.
  The query filters for `dateTime >= 13:10`.
  Because the existing 90-minute appointment started at 13:00, `13:00 >= 13:10` is **false**.
  The query returns `null` for conflicts, both in the preliminary check and within the `$transaction`.
- **Verbatim Test Output**:
  ```
  ✖ FAIL [CRITICAL]: Booking route rejects booking overlapping with longer prior appointment (13:00 90m vs 14:00 50m)
  Details: VULNERABILITY DETECTED: Returned status 201 with body {
    "success": true,
    "appointment": {
      "id": "apt_adv_52_1789415254486",
      "dateTime": "2026-11-04T14:00:00.000Z",
      "status": "CONFIRMED",
      "notes": null,
      "googleEventId": "mock_gcal_evt_1789415254486_gufah",
      "googleSynced": true,
      "service": { "id": "srv_adv_5_...", "name": "Terapia Estándar 50m", "durationMin": 50, "price": 700 },
      "specialist": { "id": "usr_adv_3_...", "name": "Lic. Bruno Diaz" },
      "customer": { "name": "Paciente Colision Asimétrica", "phone": "+528190009999" }
    }
  }
  ```

#### B. HIGH Vulnerability: Concurrency Race Condition under Parallel Bursts
- **File**: `src/app/api/public/[slug]/book/route.ts`
- **Lines**: 268–309:
  ```typescript
  const appointment = await prisma.$transaction(async (tx) => {
    const finalConflict = await tx.appointment.findFirst({ ... });
    if (finalConflict) throw new Error('SLOT_CONFLICT');
    return await tx.appointment.create({ ... });
  });
  ```
- **Observed Behavior**:
  Under a burst of 5 parallel requests booking the exact same specialist (`specA`) at the exact same slot (`2026-11-04T11:00:00.000Z`), all 5 requests execute `tx.appointment.findFirst` before any transaction commits. Under PostgreSQL's default Read Committed isolation level without row-level locks (`SELECT ... FOR UPDATE`), advisory locks, or database unique constraints on `(specialistId, dateTime)`, all 5 requests find `finalConflict == null` and all 5 successfully create confirmed appointments.
- **Verbatim Test Output**:
  ```
  ✖ FAIL [HIGH]: Concurrency Race Guard under parallel burst: Exactly 1 succeeds with 201 and others with 409 (got 5 successes, 0 conflicts)
  Details: RACE CONDITION: Application allows 5 simultaneous bookings for exact same slot
  ```

#### C. MEDIUM Vulnerability: Premature Current-Day Afternoon Slot Pruning via UTC `Z` Mismatch
- **File**: `src/lib/availability.ts`
- **Lines**: 77–79 and 306:
  ```typescript
  export function buildDateTime(dateStr: string, timeStr: string): Date {
    return new Date(`${dateStr}T${timeStr}:00.000Z`);
  }
  ...
  const isPastTime = slotStartDate.getTime() < now + leadBufferMs;
  ```
- **Observed Behavior**:
  At 13:36 local time in Mexico (UTC-6, `2026-09-14T19:36:30Z`), querying availability for today (`date = "2026-09-14"`) constructs slots using UTC: `2026-09-14T14:00:00.000Z` through `2026-09-14T18:00:00.000Z`. Because 14:00Z through 18:00Z are earlier than 19:36Z, `slotStartDate.getTime() < now + leadBufferMs` evaluates to `true` for all afternoon slots.
- **Verbatim Test Output**:
  ```
  Slots evaluated for today (2026-09-14): 09:00:BLOCKED, 10:00:BLOCKED, 11:00:BLOCKED, 12:00:BLOCKED, 13:00:BLOCKED, 14:00:BLOCKED, 15:00:BLOCKED, 16:00:BLOCKED, 17:00:BLOCKED, 18:00:BLOCKED
  ✖ FAIL [MEDIUM]: Afternoon slots on current day (14:00-18:00) remain bookable during afternoon business hours (13:36 local)
  Details: TIMEZONE GAP: All slots for today (including 14:00-18:00) were clipped as past time because UTC comparison treated them as 14:00Z-18:00Z (before current UTC 19:36Z). Available count: 0
  ```

---

## 2. Logic Chain

1. **Premise 1 (Spec Requirement)**:
   Per `ORIGINAL_REQUEST.md` (lines 84–86), the acceptance criteria requires: *"Availability engine accurately calculates open slots and blocks conflicting appointments"*, and Feature 31 requires *"Concurrency Double-Booking Guard: Atomic validation in booking transaction to prevent race collisions"*.
2. **Premise 2 (Empirical Proof of Collision Leak in `book/route.ts`)**:
   Observation 1.3.A proves that `book/route.ts` lines 130–131 and 283–285 assume all conflicting prior appointments started within `[slotStart - durationMin, slotStart]`, where `durationMin` is the duration of the *requested* service. When a previous appointment has a duration of 90m (ending at 14:30) and a customer books a 50m appointment at 14:00, the previous appointment's start time (13:00) falls outside the lookback window `[13:10, 14:00]`.
3. **Inference 1**:
   The query evaluates to zero conflicting rows, completely bypassing the conflict check. The appointment is inserted into the database with `status: CONFIRMED` (Observation 1.3.A), resulting in an overlapping double booking on the specialist's schedule between 14:00 and 14:30.
4. **Premise 3 (Empirical Proof of Parallel Race Condition)**:
   Observation 1.3.B proves that under a parallel burst of 5 requests for the exact same slot, all 5 transactions execute their `findFirst` query before any transaction commits, resulting in 5 confirmed appointments instead of 1 confirmed and 4 conflicts.
5. **Premise 4 (Empirical Proof of Timezone Pruning)**:
   Observation 1.3.C proves that on the current day, once it is past 12:45 PM in the clinic's local timezone (UTC-6), every remaining slot in the business day (14:00 to 18:00) is marked unavailable because `buildDateTime` hardcodes UTC `Z`.
6. **Conclusion**:
   The Milestone 1 implementation cannot be approved in its current state because it permits clinical double-bookings when services have different durations, permits duplicate bookings under parallel bursts, and locks patients out of afternoon same-day bookings.

---

## 3. Caveats

1. **Core Availability Mathematics in `src/lib/availability.ts`**:
   The core interval overlap algorithm in `src/lib/availability.ts` lines 268–278 (`Math.max(slotStartTime, apptStart) < Math.min(slotEndTime, apptEnd)`) is mathematically correct and passed all touching boundary, containment, and long-service tests. The collision leak exists specifically in `src/app/api/public/[slug]/book/route.ts`.
2. **Offline Mode**:
   Testing was executed using the project's in-memory relational simulator with PostgreSQL offline. Live PostgreSQL would behave identically regarding the asymmetric duration query (SQL `WHERE dateTime >= slotStart - durationMin`), and similarly regarding race conditions under Read Committed isolation without explicit locks.

---

## 4. Conclusion

**Verdict**: **REQUEST_CHANGES**

Milestone 1 has critical backend defects in booking collision validation and concurrency safety that directly violate Phase 2 acceptance criteria.

### Required Actions for Worker M1:

1. **Fix Asymmetric Collision Query in `src/app/api/public/[slug]/book/route.ts`**:
   Replace the narrow `gte: new Date(slotStart.getTime() - durationMin * 60 * 1000)` filter in lines 116–133, 170–186, and 268–290 with a wider lookback (e.g. `240` min or day range) and include `service: { select: { durationMin: true } }`. Validate overlap using the exact interval overlap formula:
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
   ```
2. **Harden Transaction against Concurrent Bursts**:
   Add mutex locking or transactional advisory locking in the booking endpoint.
3. **Fix Same-Day Timezone Evaluation in `src/lib/availability.ts`**:
   Ensure current day past-slot comparison accounts for clinic local timezone offset so afternoon slots remain bookable during afternoon business hours.

---

## 5. Verification Method

To independently reproduce all findings and verify future fixes:

1. **Run the Empirical Adversarial Stress Harness**:
   ```bash
   npx tsx tests/adversarial-phase2-m1.test.ts
   ```
   - Current output: 37 PASSED, 3 FAILED (exit code 1).
   - Expected output once fixed: 40 PASSED, 0 FAILED (exit code 0).

2. **Run the Milestone 1 Verification Suite**:
   ```bash
   npx tsx tests/verify-m1-phase2.ts
   ```
   - Output: 61 PASSED, 0 FAILED.

3. **Run the Master Phase 2 Test Suite**:
   ```bash
   npx tsx tests/e2e/phase2-runner.ts
   ```
   - Output: 151 PASSED, 381 assertions verified.
