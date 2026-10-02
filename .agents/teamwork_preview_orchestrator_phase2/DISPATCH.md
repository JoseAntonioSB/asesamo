## 2026-09-14T18:54:00Z
You are the Project Orchestrator for Phase 2 (La Agenda y Google Calendar) of APPCITAS.

Your working directory is: /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_orchestrator_phase2/
The project root workspace is: /Users/jasbotello/devel/APPCITAS
The authoritative user requirements are recorded in: /Users/jasbotello/devel/APPCITAS/.agents/ORIGINAL_REQUEST.md under header "## 2026-09-14T18:47:27Z".

## Phase 2 Scope & Requirements:
### R1. Google Calendar Integration & Availability Engine
- Implement Google Calendar synchronization for specialists using `googleapis` with OAuth2.
- OAuth flow to connect specialist Google accounts, storing access/refresh tokens securely.
- Intelligent availability engine (`src/lib/availability.ts`) computing bookable slots (respecting ~50 min typical service duration and clinic operating hours: Mon-Fri 9:00-19:00, Sat 9:00-14:00) by subtracting local DB appointments and Google Calendar busy intervals (`freeBusy`).
- Automatic creation of calendar events with appointment details when a booking is confirmed.
- Resilient fallback / simulation mode when Google credentials are not yet configured in local development, ensuring seamless offline testing.

### R2. Interactive Clinic Calendar & Agenda Dashboard
- Create `/dashboard/calendar` providing:
  - Weekly and daily agenda views displaying appointments color-coded by specialist and status (`CONFIRMED`, `COMPLETED`, `CANCELLED`, `NO_SHOW`).
  - Specialist filtering and date navigation.
  - Appointment detail drawer / modal to review patient details, contact, service, and update appointment status.
  - Google Calendar sync status indicator per specialist with a one-click connection prompt.

### R3. Public Booking Portal (`/[slug]`)
- Build a responsive, mobile-first public booking flow at `/[slug]` (e.g., `/clinica-fisia`):
  - Step 1: Select Service (displaying duration ~50 min and price).
  - Step 2: Select Specialist (or "Cualquier especialista disponible").
  - Step 3: Date and time slot picker dynamically rendered from the availability engine.
  - Step 4: Patient information form (Name, WhatsApp phone number, optional email/notes).
  - Step 5: Booking confirmation screen with summary and calendar event feedback.

### Acceptance Criteria:
- Availability engine accurately calculates open slots and blocks conflicting appointments or busy Google Calendar times.
- Public booking endpoint (`/api/public/[slug]/book`) creates or links `Customer`, creates `Appointment`, and triggers calendar synchronization.
- Status updates on appointments (`/api/appointments/[id]`) correctly update state and free up slots upon cancellation.
- `/dashboard/calendar` is accessible and functional, displaying current bookings.
- Public booking page at `/[slug]` renders smoothly on both desktop and mobile viewports following the minimalist healthcare design aesthetic.
- `npm run build` completes with 0 errors across all routes including dynamic `[slug]` and calendar views.
- End-to-end verification test suite passes covering availability calculations and booking creation.
