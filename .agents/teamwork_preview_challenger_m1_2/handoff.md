# Milestone 1 (M1) Handoff Report: Challenger 2

**Agent**: `teamwork_preview_challenger_m1_2` (Challenger 2)  
**Milestone**: Milestone 1 (Phase 2 — La Agenda y Google Calendar)  
**Assigned Directory**: `/Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_challenger_m1_2/`  
**Date**: 2026-09-14  
**Verdict**: **REQUEST_CHANGES**  

---

## 1. Observation

### 1.1 Automated Test Execution
1. **E2E Phase 2 Runner**:
   - Command: `npx tsx tests/e2e/phase2-runner.ts`
   - Result: Exit code `0`.
   - Output:
     ```
     TOTALS: 151 tests, 151 passed, 0 failed, 381 assertions verified in 35ms.
     Tier 1 (Feature Coverage)    : 75 tests  (Target: >= 40) -> [PASS]
     Tier 2 (Boundary & Edges)    : 60 tests  (Target: >= 35) -> [PASS]
     Tier 3 (Cross Combinations)  : 12 tests  (Target: >= 10) -> [PASS]
     Tier 4 (Clinical Scenarios)  : 4 tests  (Target: >= 4)  -> [PASS]
     Total Verified Tests         : 151 tests (Target: >= 40) -> [PASS]
     Total Verified Assertions    : 381 assertions (Target: >= 200) -> [PASS]
     ✔ All Phase 2 test tiers passed and met coverage thresholds successfully.
     ```

2. **Dedicated M1 Verification Suite**:
   - Command: `npx tsx tests/verify-m1-phase2.ts`
   - Result: Exit code `0`.
   - Output: `VERIFICATION RESULTS: 61 PASSED, 0 FAILED`.

3. **Static Typing & Build**:
   - Command: `npx tsc --noEmit`
   - Result: Exit code `0`, 0 errors.
   - Command: `npm run build`
   - Result: Exit code `0`, Next.js 14.2.24 compiled successfully, 19 route entry points generated.

4. **Adversarial Challenge Suite**:
   - Command: `npx tsx tests/adversarial-m1-phase2.test.ts`
   - Result: 18 PASSED, 1 FAILED.

### 1.2 Multi-Duration Service Collision Bug in `src/app/api/public/[slug]/book/route.ts`
- **Location**: `src/app/api/public/[slug]/book/route.ts`, lines 128–132 and lines 282–286:
  ```typescript
  128:           dateTime: {
  129:             lt: slotEnd,
  130:             gte: new Date(slotStart.getTime() - durationMin * 60 * 1000),
  131:           },
  ```
  and
  ```typescript
  282:           dateTime: {
  283:             lt: slotEnd,
  284:             gte: new Date(slotStart.getTime() - durationMin * 60 * 1000),
  285:           },
  ```
- **Observed Behavior**:
  When Patient 1 books a 90-minute session at 09:00 UTC (occupying Specialist A from 09:00 to 10:30 UTC), and Patient 2 books a 50-minute session at 10:00 UTC (10:00 to 10:50 UTC):
  `slotStart` = 10:00 UTC.
  `durationMin` = 50.
  `gte` filter evaluates to `10:00 - 50m = 09:10 UTC`.
  Patient 1's appointment starting at `09:00 UTC` fails the `gte: 09:10` filter (`09:00 < 09:10`).
  As a result, `tx.appointment.findFirst` returns `null`.
  The API route creates the overlapping appointment and returns HTTP `201 Created` instead of HTTP `409 Conflict`.
- **Verbatim Test Output from `tests/adversarial-m1-phase2.test.ts`**:
  ```
  ✖ FAIL [HIGH]    | MultiDuration   | Booking API Guard blocks booking overlapping prior multi-duration (90min) appointment
                   └─ Returned status: 201 (Expected: 409 Conflict)
  ```

### 1.3 High-Concurrency Race Condition Verification
- **Exact-Slot Race**:
  When 10 concurrent requests (`Promise.all`) target the exact same specialist and time slot:
  - 1 request received HTTP 201 Created.
  - 9 requests received HTTP 409 Conflict.
  - Database contains exactly 1 appointment for that slot.
- **Pooling ("any specialist") Race**:
  When 10 concurrent requests target `specialistId: 'any'` for the same slot where 2 specialists are free:
  - Workload balancing sort deterministically routes all 10 racing requests to the same specialist (the one with the lowest appointment count).
  - 1 request wins (201), 9 get 409.
  - A subsequent sequential request for `specialistId: 'any'` routes to the second specialist and succeeds (201).
  - A third sequential request receives 409 when all specialists are booked.

### 1.4 Google Calendar Integration & Fallback
- `src/lib/google-calendar.ts`:
  - When `GOOGLE_CLIENT_ID` is unset or contains 'mock'/'placeholder', `isGoogleConfigured` evaluates to `false`.
  - `generateAuthUrl` produces valid URLs with JSON-encoded state.
  - `handleOAuthCallback` provisions mock tokens and persists them on the `User` model, setting `googleCalendarId: 'primary'`.
  - `getSpecialistFreeBusy` returns `[]` without throwing exceptions when tokens are absent or when simulated network errors occur.
  - `createCalendarEvent` returns deterministic IDs prefixed with `mock_gcal_evt_` and Google Calendar event URLs.
  - `deleteCalendarEvent` deletes events idempotently and returns `true`.
  - `disconnectSpecialistGoogle` nullifies `googleAccessToken`, `googleRefreshToken`, `googleTokenExpiry`, `googleEmail`, and `googleCalendarId`.

### 1.5 Customer Deduplication
- Phone numbers with spaces, hyphens, and parentheses (`+52 81 9999 5555` vs `(81) 9999-5555`) normalize to clean numeric strings and link to the same customer record.
- Cross-tenant customer isolation: The same phone number booked in Clinic Alpha and Clinic Beta generates distinct Customer records scoped to each clinic's `clinicId`.
- Phone validation strictly returns HTTP 400 for numbers shorter than 7 digits or containing alphabetic characters.

---

## 2. Logic Chain

1. **Premise**: In clinical practice, physiotherapy clinics configure services of varying lengths (e.g., 30-min ultrasound, 50-min general evaluation, 90-min intensive therapy, 120-min post-op rehab).
2. **Observation (from 1.2)**: `src/app/api/public/[slug]/book/route.ts` line 130 and line 284 calculate the collision start boundary as `slotStart.getTime() - durationMin * 60 * 1000`, using the duration of the *new* booking request.
3. **Inference**: If an existing appointment has a duration $D_{prev} > D_{new}$, and starts at $T_{prev}$ where $T_{new} - D_{prev} < T_{prev} < T_{new} - D_{new}$, then $T_{prev}$ is strictly less than the `gte` filter, but the appointment still extends into the new appointment's time window ($T_{prev} + D_{prev} > T_{new}$).
4. **Empirical Confirmation**: When tested empirically in `tests/adversarial-m1-phase2.test.ts` (Challenge 5.2), a 50-minute booking at 10:00 UTC overlapping with a 90-minute session from 09:00 to 10:30 UTC returned HTTP 201 Created and persisted to the database.
5. **Impact**: Two patients are scheduled simultaneously with the same therapist, leading to calendar collisions in both the clinic database and Google Calendar.
6. **Contrasting Robustness in Availability Engine**: `src/lib/availability.ts` lines 268–278 correctly uses `(appt.service?.durationMin || 50)` from each individual appointment to calculate `apptEnd = apptStart + apptDuration` and checks `Math.max(slotStartTime, apptStart) < Math.min(slotEndTime, apptEnd)`.
7. **Conclusion**: While the UI correctly hides the slot, the booking API endpoint has a critical validation flaw that permits overlapping double-bookings. Therefore, the verdict must be **REQUEST_CHANGES**.

---

## 3. Caveats

- **No Caveats on Bug Reproduction**: The multi-duration collision defect was reproduced deterministically via automated test script `tests/adversarial-m1-phase2.test.ts`.
- **Out-of-Scope Production GCP Testing**: Real Google Cloud OAuth credentials and live GCP endpoints were not tested because real external credentials are not provisioned in the local environment; the offline simulation fallback was verified comprehensively.

---

## 4. Conclusion

**Verdict: REQUEST_CHANGES**

### Required Modifications for Approval
In `src/app/api/public/[slug]/book/route.ts`:
Replace the collision queries at lines 116–133, lines 168–187, and lines 270–287 with a duration-aware check.

**Recommended Fix**:
Instead of filtering only within `durationMin` of the requested service, query active appointments for the specialist on that date within a lookback window (e.g. up to 4 hours prior), including their service duration:
```typescript
const candidateAppts = await tx.appointment.findMany({
  where: {
    clinicId: clinic.id,
    specialistId: chosenSpecialist.id,
    status: {
      in: [
        AppointmentStatus.CONFIRMED,
        AppointmentStatus.PENDING,
        AppointmentStatus.COMPLETED,
        AppointmentStatus.NO_SHOW,
      ],
    },
    dateTime: {
      gte: new Date(slotStart.getTime() - 4 * 60 * 60 * 1000), // 4 hours lookback
      lt: slotEnd,
    },
  },
  include: {
    service: {
      select: { durationMin: true },
    },
  },
});

const hasConflict = candidateAppts.some((appt) => {
  const apptStart = new Date(appt.dateTime).getTime();
  const apptDur = (appt.service?.durationMin || 50) * 60 * 1000;
  const apptEnd = apptStart + apptDur;
  return Math.max(slotStart.getTime(), apptStart) < Math.min(slotEnd.getTime(), apptEnd);
});

if (hasConflict) {
  throw new Error('SLOT_CONFLICT');
}
```

Additionally:
- In `src/app/api/public/[slug]/book/route.ts` line 66, add a check rejecting bookings in the past (`if (slotStart.getTime() < Date.now()) return NextResponse.json({ error: 'No se pueden reservar citas en horarios pasados.' }, { status: 400 });`).

---

## 5. Verification Method

To independently verify all claims in this report:

1. **Execute Phase 2 Master Runner**:
   ```bash
   npx tsx tests/e2e/phase2-runner.ts
   ```
   *Expected*: 151 tests pass, exit code 0.

2. **Execute Dedicated M1 Verification Suite**:
   ```bash
   npx tsx tests/verify-m1-phase2.ts
   ```
   *Expected*: 61 assertions pass, exit code 0.

3. **Execute Adversarial Challenge Suite**:
   ```bash
   npx tsx tests/adversarial-m1-phase2.test.ts
   ```
   *Expected*: 18 passed, 1 failed (`MultiDuration | Booking API Guard blocks booking overlapping prior multi-duration (90min) appointment -> Returned status: 201 (Expected: 409 Conflict)`).

4. **Typecheck and Build**:
   ```bash
   npx tsc --noEmit
   npm run build
   ```
   *Expected*: 0 TypeScript errors, successful Next.js production build.

5. **Invalidation Condition**:
   If `tests/adversarial-m1-phase2.test.ts` passes with 19/19 passing (specifically with Challenge 5.2 returning HTTP 409 Conflict), the finding is resolved.
