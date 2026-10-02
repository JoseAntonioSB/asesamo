# Dispatch: Milestone 1 Worker (Google Calendar & Availability Engine)

## Context & Mission
You are the Implementation Worker for Milestone 1 (M1) of APPCITAS Phase 2.
Your mission is to implement Google Calendar integration, the intelligent availability engine, and core Phase 2 API route handlers.

## Mandatory Integrity Warning
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. A teamwork_preview_auditor will independently verify your work. Integrity violations WILL be detected and your work WILL be rejected.

## Inputs
- Authoritative requirements: /Users/jasbotello/devel/APPCITAS/.agents/ORIGINAL_REQUEST.md
- Specification report: /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_spec_miner_survey/spec_report.md
- Project architecture: /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_orchestrator_phase2/PROJECT.md
- Backend survey report: /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_explorer_survey_1/survey_report.md

## Exclusive Write Ownership
You own and may modify or create:
- `package.json` (add `googleapis`)
- `prisma/schema.prisma` (add `googleTokenExpiry DateTime?`, `googleCalendarId String?`, `googleEmail String?` to `model User`, and any required indices)
- `src/lib/google-calendar.ts`
- `src/lib/availability.ts`
- `src/app/api/availability/route.ts`
- `src/app/api/public/[slug]/book/route.ts`
- `src/app/api/public/[slug]/availability/route.ts`
- `src/app/api/appointments/route.ts`
- `src/app/api/appointments/[id]/route.ts`
- `src/app/api/auth/google/connect/route.ts`
- `src/app/api/auth/google/callback/route.ts`

DO NOT touch UI files owned by M2 (`src/app/dashboard/calendar/`) or M3 (`src/app/[slug]/page.tsx`).

## Technical Requirements
1. **Google Calendar Integration (`src/lib/google-calendar.ts`)**:
   - Install `googleapis` (`npm install googleapis`).
   - Use `google.auth.OAuth2` with credentials from `process.env.GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URI`.
   - Implement `generateAuthUrl(specialistId: string): string` with scopes `['https://www.googleapis.com/auth/calendar.events', 'https://www.googleapis.com/auth/calendar.freebusy']`.
   - Implement `handleOAuthCallback(code: string, specialistId: string)`: exchange code for tokens, persist `googleAccessToken`, `googleRefreshToken`, `googleTokenExpiry`, `googleEmail` on `User`.
   - Implement `getSpecialistFreeBusy(specialistId: string, timeMin: Date, timeMax: Date)`: return `{ start: Date, end: Date }[]`.
   - Implement `createCalendarEvent(params)`: creates Google Calendar event with patient summary, description, start and end times, returns eventId.
   - **Resilient Fallback Mode**: When Google credentials are empty/mock, operate in simulation mode. Do not throw unhandled errors; return deterministic mock busy periods or mock event IDs (`mock_gcal_evt_...`).

2. **Intelligent Availability Engine (`src/lib/availability.ts`)**:
   - Operating hours: Mon-Fri 09:00-19:00, Sat 09:00-14:00, Sun closed.
   - Slot grid: 60-minute intervals (09:00, 10:00, ..., 18:00 on weekdays; 09:00-13:00 on Saturday).
   - Collision subtraction:
     - Local DB appointments matching specialist + date with status in `['CONFIRMED', 'PENDING', 'COMPLETED', 'NO_SHOW']`.
     - Subtraction: if `status === 'CANCELLED'`, the appointment does NOT block availability.
     - Google Calendar `busy` intervals: subtract any slot where `max(slotStart, busyStart) < min(slotEnd, busyEnd)`.
   - "Cualquier especialista disponible" mode:
     - If `specialistId` is omitted or `"any"`, query all active specialists associated with the service via `SpecialistService`.
     - Slot is available if AT LEAST ONE qualified specialist is free.
     - Select least-busy specialist on that day for assignment upon booking.

3. **API Endpoints**:
   - `GET /api/availability` and `GET /api/public/[slug]/availability`: Query parameters `serviceId`, `date` (YYYY-MM-DD), optional `specialistId`. Returns array of bookable slots.
   - `POST /api/public/[slug]/book`:
     - Resolves clinic by `slug`.
     - Validates payload with Zod or thorough validation: `serviceId`, `dateTime`, `customerName`, `customerPhone`, optional `specialistId`, `customerEmail`, `notes`.
     - Finds or creates `Customer` by `(clinicId, phone)`.
     - Re-verifies slot availability to prevent double-booking.
     - Creates `Appointment` with status `CONFIRMED`.
     - Triggers Google Calendar event creation (real or simulation).
     - Returns 201 Created with appointment details.
   - `GET /api/appointments`: List appointments with filters (`specialistId`, `dateMin`, `dateMax`, `status`).
   - `PATCH /api/appointments/[id]`: Update appointment `status` (`CONFIRMED`, `COMPLETED`, `CANCELLED`, `NO_SHOW`). When updated to `CANCELLED`, slot is liberated.
   - `GET /api/auth/google/connect` and `GET /api/auth/google/callback`.

4. **Build & Test Verification**:
   - Run `npx prisma generate`.
   - Run `npx tsx tests/e2e/run-tests.ts` to ensure 0 regressions.
   - Run `npm run build` to ensure 0 compile/type errors.

## Deliverables
- Working directory: `/Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_worker_m1/`
- Write `progress.md` and `handoff.md` in your directory.
- Include exact build and test command outputs in `handoff.md`.
- Send completion message back to the orchestrator.

## 2026-09-14T13:11:22Z
You are the Implementation Worker for Milestone 1 (M1) of APPCITAS Phase 2.
Your assigned working directory is: /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_worker_m1/
Read your dispatch instructions at /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_worker_m1/DISPATCH.md and the authoritative request at /Users/jasbotello/devel/APPCITAS/.agents/ORIGINAL_REQUEST.md.

MANDATORY INTEGRITY WARNING:
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. A teamwork_preview_auditor will independently verify your work. Integrity violations WILL be detected and your work WILL be rejected.

Implement Google Calendar integration (googleapis, OAuth, freeBusy, event creation, fallback simulation), intelligent availability engine (src/lib/availability.ts), and API endpoints (/api/availability, /api/public/[slug]/book, /api/appointments/[id], /api/auth/google/*).
Run prisma generate, test verification, and npm run build. Write your handoff.md with verified command outputs and send a completion message back.

