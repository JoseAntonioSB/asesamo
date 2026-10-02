# Handoff Report: APPCITAS Phase 2 Specification Mining

- **Agent**: Spec Miner (`teamwork_preview_spec_miner_survey`)
- **Recipient**: Orchestrator (`5af02f98-2cc9-400b-b4e1-67cc4f848f43`)
- **Date**: 2026-09-14T19:05:00Z
- **Type**: Hard Handoff (Task Complete)

---

## 1. Observation

1. **Authoritative Request File**:
   `/Users/jasbotello/devel/APPCITAS/.agents/ORIGINAL_REQUEST.md` (lines 50–93, header `## 2026-09-14T18:47:27Z`):
   - R1: "Implement Google Calendar synchronization for specialists using `googleapis` with OAuth2... Intelligent availability engine (`src/lib/availability.ts`) that computes bookable slots (respecting ~50 min typical service duration and clinic operating hours: Mon-Fri 9:00-19:00, Sat 9:00-14:00) by subtracting local DB appointments and Google Calendar busy intervals (`freeBusy`)... Resilient fallback / simulation mode when Google credentials are not yet configured in local development..."
   - R2: "Create `/dashboard/calendar` providing: Weekly and daily agenda views displaying appointments color-coded by specialist and status (`CONFIRMED`, `COMPLETED`, `CANCELLED`, `NO_SHOW`)... Specialist filtering and date navigation... Appointment detail drawer / modal to review patient details, contact, service, and update appointment status... Google Calendar sync status indicator per specialist with a one-click connection prompt."
   - R3: "Build a responsive, mobile-first public booking flow at `/[slug]` (e.g., `/clinica-fisia`): Step 1: Select Service... Step 2: Select Specialist (or "Cualquier especialista disponible")... Step 3: Date and time slot picker dynamically rendered from the availability engine... Step 4: Patient information form (Name, WhatsApp phone number, optional email/notes)... Step 5: Booking confirmation screen with summary and calendar event feedback."
   - Acceptance Criteria: "Availability engine accurately calculates open slots and blocks conflicting appointments or busy Google Calendar times. Public booking endpoint (`/api/public/[slug]/book`) creates or links `Customer`, creates `Appointment`, and triggers calendar synchronization. Status updates on appointments (`/api/appointments/[id]`) correctly update state and free up slots upon cancellation. `/dashboard/calendar` is accessible and functional... `npm run build` completes with 0 errors... End-to-end verification test suite passes..."

2. **Database Schema & Model Readiness**:
   File: `/Users/jasbotello/devel/APPCITAS/prisma/schema.prisma`
   - Model `User` (lines 80–104) contains `googleAccessToken String?` (line 90) and `googleRefreshToken String?` (line 91).
   - Model `Appointment` (lines 205–229) contains `googleEventId String?` (line 214), `status AppointmentStatus @default(CONFIRMED)` (line 212), `dateTime DateTime` (line 211), and relations to `Clinic`, `User` (specialist), `Service`, `Customer`.
   - Model `Customer` (lines 181–200) contains `clinicId`, `name`, `email?`, `phone`, with composite index `@@index([clinicId, phone])` (line 198) enabling patient deduplication.
   - Command `npx prisma validate` exited with code 0:
     "The schema at prisma/schema.prisma is valid 🚀".

3. **Current Build & Test Baseline**:
   - Running `npm run build` exited with code 0: 9 static/dynamic pages compiled cleanly, with Next.js 14.2.24 and TypeScript 5.6.3.
   - Running `npx tsx tests/e2e/run-tests.ts` exited with code 0: 53 tests passed, 0 failed, 417 assertions verified in 17ms.
   - Package dependencies (`package.json`): `googleapis` is not yet installed; it will need to be installed (`npm install googleapis`) during implementation.

4. **Detailed Specification Report Generated**:
   File: `/Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_spec_miner_survey/spec_report.md`
   - Contains 41 discovered features in structured table.
   - Contains 20 edge cases in structured table.
   - Contains exact mathematical availability rules, operating hours matrix, boundary clipping logic, Google Calendar OAuth contract, UI/UX specs for `/dashboard/calendar` and `/[slug]`, and full API endpoint schemas.

---

## 2. Logic Chain

1. **Operating Hours & Slot Math**:
   From Observation 1, clinic hours are Mon–Fri 09:00–19:00, Sat 09:00–14:00, and Sunday closed.
   For services with ~50 min duration, standard 60-minute grid increments (`09:00`, `10:00`, ..., `18:00` weekdays; `09:00`, ..., `13:00` Saturday) leave a natural 10-minute sanitation/turnover buffer between patient sessions.
   Any slot where `start + duration > close` (e.g. 18:30 on weekdays or 13:30 on Saturday) must be excluded to prevent appointments overrunning clinic closing hours.

2. **Collision Detection & Slot Liberation**:
   From Observation 1 and 2, appointments block time slots only when their status is NOT `CANCELLED`.
   When an appointment status is updated to `CANCELLED` (via `PATCH /api/appointments/[id]`), the slot is excluded from the collision check, instantly reopening it for booking on `/[slug]`.
   Similarly, live events from Google Calendar `freeBusy` query are checked for interval overlap `max(T_start, B_start) < min(T_end, B_end)` and subtracted from the availability pool.

3. **Resilient Offline Fallback**:
   From Observation 1, local development must run seamlessly even without Google API credentials.
   When `GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET` are not provided in `.env`, the system executes in simulation mode: mock OAuth tokens are assigned upon connect, `freeBusy` queries return mock or empty busy spans without crashing, and appointment creation outputs mock event IDs (`mock_evt_...`).

4. **"Cualquier Especialista" & Workload Balancing**:
   From Observation 1 and 2, when patients select "Cualquier especialista disponible", the engine checks all active specialists assigned to the selected service via `SpecialistService`.
   For each open time slot, if at least one therapist is free, the slot is presented as bookable. Upon booking submission, the engine assigns the appointment to the free specialist with the lowest appointment load on that date, ensuring optimal clinic resource distribution.

5. **Patient Deduplication & Concurrency Guard**:
   From Observation 2, `Customer` has `@@index([clinicId, phone])`. The booking API endpoint (`POST /api/public/[slug]/book`) performs an upsert/match on `(clinicId, phone)`, preventing duplicate patient records.
   Double bookings are prevented by re-verifying slot availability within the transaction immediately before inserting the `Appointment`, rejecting collisions with `409 Conflict`.

---

## 3. Caveats

1. **Google Credentials**:
   Live Google Calendar synchronization requires real OAuth client credentials from the Google Cloud Console. Until provided by the user in `.env`, the implementation must rely on the offline simulation fallback mode specified in `spec_report.md`.
2. **Timezone Assumptions**:
   The pilot clinic ("Clínica Fisia") is located in San Pedro Garza García, NL (`America/Monterrey`, UTC-6). Timestamps should be stored in UTC in the database, with availability calculations performed against the clinic's local operating hours.
3. **External WhatsApp API**:
   Phase 2 specifies a client-side click-to-chat action (`https://wa.me/...`) for WhatsApp confirmation. Automated server-side Meta Cloud API messaging is part of Phase 3 (as noted in `Roadmap.md`).

---

## 4. Conclusion

All requirements for APPCITAS Phase 2 have been thoroughly mined, analyzed, and synthesized into an authoritative specification report (`spec_report.md`).
- Schema prerequisites are already satisfied in `prisma/schema.prisma`.
- Exact availability formulas, clinic hours, boundary conditions, and collision rules are documented.
- Public booking flow (5 steps) and dashboard calendar UI specifications are detailed and aligned with design guidelines (`ui-ux-pro-max`, `apple-design`, `minimalist-ui`).
- API endpoint contracts and test verification strategies are complete and ready for downstream architects, test writers, and implementers.

---

## 5. Verification Method

1. **Inspect Specification Report**:
   ```bash
   cat /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_spec_miner_survey/spec_report.md
   ```
   Verify sections 3 (Features Discovered: 41 items) and 4 (Edge Cases: 20 items), plus sections 5–11.

2. **Verify Database Schema**:
   ```bash
   npx prisma validate
   ```
   Ensures schema syntax and relations remain 100% valid.

3. **Verify Baseline Test Suite**:
   ```bash
   npx tsx tests/e2e/run-tests.ts
   ```
   Ensures zero regressions across the 53 existing Phase 1 tests.

4. **Verify Application Build**:
   ```bash
   npm run build
   ```
   Ensures clean Next.js compilation without TypeScript errors.
