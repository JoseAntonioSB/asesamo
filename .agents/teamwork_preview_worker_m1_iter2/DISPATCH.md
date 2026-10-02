# Dispatch: Milestone 1 Worker — Iteration 2 (Remediation & Hardening)

## Context & Mission
You are the fresh Implementation Worker for Milestone 1 Iteration 2.
Your mission is to resolve the critical and major defects identified by Reviewer 2, Challenger 1, and Challenger 2 in `src/app/api/public/[slug]/book/route.ts`, `src/app/api/appointments/route.ts`, and `src/lib/availability.ts`.

## Mandatory Integrity Warning
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. A teamwork_preview_auditor will independently verify your work. Integrity violations WILL be detected and your work WILL be rejected.

## Inputs
- Authoritative requirements: /Users/jasbotello/devel/APPCITAS/.agents/ORIGINAL_REQUEST.md
- Gate status: /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_orchestrator_phase2/GATE_STATUS.md
- Reviewer 2 report: /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_reviewer_m1_2/review_report.md
- Challenger 1 report & test: /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_challenger_m1_1/challenge_report.md & `tests/adversarial-phase2-m1.test.ts`
- Challenger 2 report & test: /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_challenger_m1_2/challenge_report.md & `tests/adversarial-m1-phase2.test.ts`

## Required Fixes

### 1. Canonical Interval Collision Guard (`src/app/api/public/[slug]/book/route.ts` & `src/app/api/appointments/route.ts`)
The previous collision check queried `gte: slotStart - durationMin` using the *new* service's duration.
This created two bugs:
- **Asymmetric Duration Bug**: A prior 90-minute appointment (e.g. 09:00 - 10:30) was missed when booking a 50-minute slot at 10:00 because `09:00 < (10:00 - 50m)`.
- **Back-to-Back False Collision**: A prior 60-minute appointment ending exactly at 10:00 was flagged as a collision for a 10:00 slot.
**Fix**:
Query all non-cancelled appointments for that specialist on the target calendar day (or `gte: slotStart - 6 hours, lt: slotEnd + 6 hours`), including `service: { select: { durationMin: true } }`.
Then test each appointment for canonical overlap:
```typescript
const apptStart = appt.dateTime.getTime();
const apptEnd = apptStart + (appt.service?.durationMin || 50) * 60 * 1000;
const hasOverlap = Math.max(slotStart.getTime(), apptStart) < Math.min(slotEnd.getTime(), apptEnd);
```
If `hasOverlap` is true, reject with 409 Conflict.
Apply this to both specific specialist bookings and "any specialist" candidate evaluation.

### 2. Strict Operating Hours Validation (`src/app/api/public/[slug]/book/route.ts`)
Directly validate in the booking route:
- Sunday (day 0): Clinic closed -> reject with 400 Bad Request.
- Saturday (day 6): 09:00 - 14:00. If `slotStart < 09:00` or `slotEnd > 14:00` -> reject with 400 Bad Request.
- Mon-Fri (days 1-5): 09:00 - 19:00. If `slotStart < 09:00` or `slotEnd > 19:00` -> reject with 400 Bad Request.

### 3. Timezone / Local Time Parity (`src/lib/availability.ts`)
In `isPastTime`, ensure slots are checked accurately against current time respecting local clinic time (UTC-6) so afternoon slots for "today" are not prematurely marked as past when it is still morning/early afternoon in Mexico.

## Verification Requirements
Run and verify all tests:
1. `npx tsx tests/verify-m1-phase2.ts`
2. `npx tsx tests/adversarial-m1-phase2.test.ts` (all 19 tests must pass)
3. `npx tsx tests/adversarial-phase2-m1.test.ts` (all 40 tests must pass)
4. `npx tsx tests/e2e/run-tests.ts` (Phase 1 regression suite)
5. `npm run build` (0 compile errors)

Write your `handoff.md` and send completion message back to orchestrator.

## 2026-09-14T19:50:00Z
You are the Implementation Worker for Milestone 1 Iteration 2 (Remediation & Hardening).
Your assigned working directory is: /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_worker_m1_iter2/
Read your dispatch instructions at /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_worker_m1_iter2/DISPATCH.md and the authoritative request at /Users/jasbotello/devel/APPCITAS/.agents/ORIGINAL_REQUEST.md.

MANDATORY INTEGRITY WARNING:
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. A teamwork_preview_auditor will independently verify your work. Integrity violations WILL be detected and your work WILL be rejected.

Apply the canonical interval collision fixes in src/app/api/public/[slug]/book/route.ts and src/app/api/appointments/route.ts, strict operating hours validation, and timezone parity in src/lib/availability.ts.
Verify that tests/verify-m1-phase2.ts, tests/adversarial-m1-phase2.test.ts, tests/adversarial-phase2-m1.test.ts, and npm run build all succeed.
Write your handoff.md and send completion message back.
