# APPCITAS Phase 2: Comprehensive Specification Report
**The Agenda, Availability Engine, Google Calendar Integration & Public Booking Portal**

- **Document Version**: 2.0.0
- **Author**: Specification Miner
- **Status**: Authoritative Reference Specification
- **Target Working Directory**: `/Users/jasbotello/devel/APPCITAS`
- **Scope**: Phase 2 Requirements (R1, R2, R3) and System Integrations

---

## 1. Executive Summary & Architecture Overview

APPCITAS Phase 2 expands the multi-tenant physiotherapy foundation (Phase 1) with an appointment scheduling, availability, and external calendar synchronization system.

The core architecture connects three distinct layers:
1. **Intelligent Availability Engine (`src/lib/availability.ts`) & Google Calendar Adapter (`src/lib/google-calendar.ts`)**:
   Computes conflict-free appointment slots by synthesizing clinic operating hours, service durations, specialist assignments, active local database bookings (filtering out `CANCELLED`), and live Google Calendar busy intervals queried via Google Calendar API `freeBusy`. It incorporates a deterministic offline simulation mode when OAuth credentials are absent.
2. **Interactive Clinic Calendar & Agenda Dashboard (`/dashboard/calendar`)**:
   A desktop and mobile-responsive agenda for clinic owners and specialists featuring weekly and daily timeline grids, specialist and status filtering, appointment detail slide-over drawer with quick WhatsApp links, status update transitions (confirmed, completed, cancelled, no-show), and specialist Google Calendar sync badges.
3. **Public Multi-Tenant Booking Portal (`/[slug]`)**:
   A high-converting, mobile-first 5-step booking flow following minimalist healthcare design tokens (`ui-ux-pro-max`, `apple-design`, `minimalist-ui`). Patients can choose services, select specific therapists or auto-assign ("Cualquier especialista"), pick dynamic slots, enter WhatsApp contact info, and receive immediate booking confirmation with Google Calendar event generation.

---

## 2. Authoritative Specification Source Matrix

| Component / Requirement | Primary Authoritative Source | Code / Schema Reference | Priority |
|-------------------------|------------------------------|-------------------------|----------|
| **Phase 2 Scope & Acceptance Criteria** | `ORIGINAL_REQUEST.md` (2026-09-14T18:47:27Z) | Lines 50–93 | Mandatory (P0) |
| **Multitenant Data Models** | `prisma/schema.prisma` | Models: `Clinic`, `User`, `Service`, `SpecialistService`, `Customer`, `Appointment`, `Order` | Ground Truth (P0) |
| **Seeded Baseline Data** | `prisma/seed.ts` | 5 Specialists, 6 Services, "Clínica Fisia" | Reference (P0) |
| **Authentication & Tenant Context** | `src/lib/auth.ts`, `src/middleware.ts` | NextAuth JWT & Session, Route Matcher | Ground Truth (P0) |
| **UI Aesthetics & Design System** | `ui-ux-pro-max`, `apple-design`, `minimalist-ui` | Healthcare teal accents, calm neutrals, responsive drawers | Mandatory (P1) |
| **E2E Testing Standard** | `tests/e2e/harness.ts` | Standalone zero-dependency test runner | Mandatory (P0) |

---

## 3. Features Discovered

| # | Category | Feature | Description | Inputs | Outputs | Error Behavior | Discovered Via |
|---|----------|---------|-------------|--------|---------|----------------|----------------|
| 1 | Availability Engine | Clinic Operating Hours Filter | Filters available slots based on clinic hours: Mon-Fri 09:00-19:00, Sat 09:00-14:00, Sun closed. | Date string (`YYYY-MM-DD`), Clinic ID | Open operating window `{ open: string, close: string }` or `closed: true` | Returns 0 slots and `isOpen: false` for closed days | `ORIGINAL_REQUEST.md` R1 |
| 2 | Availability Engine | Dynamic Service Duration Slicing | Slices daily working window into discrete slots respecting service duration (~50 min typical, or 45/60 min per service). | `serviceId`, Date | List of time slots with `startTime` and `endTime` | 404 if service does not exist or is inactive | `ORIGINAL_REQUEST.md` R1 |
| 3 | Availability Engine | Boundary End-Time Clipping | Excludes slots whose `startTime + durationMin` exceeds clinic closing time (e.g. 18:30 on weekday or 13:30 on Sat). | Slot start time, service duration, closing time | Boolean slot eligibility | Excluded from generated slots | Codebase math probe |
| 4 | Availability Engine | Past Time Slot Pruning | Disqualifies any slots earlier than current time (`Date.now()`) when querying availability for today. | Requested date, current server timestamp | Filtered array of future slots | Excluded from results | Robustness probe |
| 5 | Availability Engine | Local DB Conflict Detection | Blocks slots that collide with existing non-cancelled DB appointments for the specialist. | `specialistId`, `dateTime`, `durationMin` | Boolean collision status | Marks slot unavailable | `ORIGINAL_REQUEST.md` R1 |
| 6 | Availability Engine | Cancellation Slot Liberation | Ignores `CANCELLED` appointments during collision detection, instantly reopening the slot for booking. | Appointment records with status `CANCELLED` | Available slot reinstated | N/A (liberation logic) | `ORIGINAL_REQUEST.md` Criteria |
| 7 | Availability Engine | Google Calendar `freeBusy` Collision | Queries specialist's Google Calendar `freebusy.query` to block slots overlapping personal events. | Specialist OAuth tokens, time window | Google busy intervals array `[{ start, end }]` | Falls back to local DB if API error occurs | `ORIGINAL_REQUEST.md` R1 |
| 8 | Availability Engine | Google Offline Fallback / Simulation Mode | Enables 100% offline development and testing with deterministic mock availability when Google keys are missing. | Environment absence or test flag | Simulated slots and mock tokens | Clean fallback, no crash | `ORIGINAL_REQUEST.md` R1 |
| 9 | Availability Engine | "Cualquier Especialista" Slot Aggregation | Aggregates all bookable slots across all active specialists assigned to the service. | `serviceId`, date, `specialistId='any'` | Merged list of available slots with candidate specialist | Empty list if no specialist is free | `ORIGINAL_REQUEST.md` R3 |
| 10 | Availability Engine | Workload-Balanced Specialist Assignment | Auto-selects the free specialist with the lowest appointment count for that day when booking "any specialist". | Candidate free specialists, date | Selected `specialistId` | 409 Conflict if none available | Dispatch & Architecture probe |
| 11 | Availability Engine | Specialist Qualification Verification | Ensures only specialists assigned to the service via `SpecialistService` and `isActive: true` can be booked. | `serviceId`, `specialistId` | Qualified specialist record | 400 Bad Request if specialist unqualified | `prisma/schema.prisma` |
| 12 | Google Calendar | Specialist OAuth Connection Flow | Generates Google OAuth consent URL with calendar scopes and handles callback to store tokens. | `specialistId`, OAuth `code` | Stored `googleAccessToken`, `googleRefreshToken` | Redirect with error toast if rejected | `ORIGINAL_REQUEST.md` R1 |
| 13 | Google Calendar | Specialist OAuth Disconnect | Clears Google tokens from the specialist's `User` record to unlink the calendar. | `specialistId`, authenticated session | Cleared tokens in database | 401 Unauthorized if non-owner / wrong specialist | Security probe |
| 14 | Google Calendar | Automatic Event Creation | Creates an event with patient and service details on specialist's Google Calendar upon booking. | Appointment details, patient contact, start/end time | Created Google Event ID stored on `Appointment` | In offline mode, generates deterministic mock event ID | `ORIGINAL_REQUEST.md` R1 |
| 15 | Google Calendar | Two-Way Event Deletion on Cancellation | Deletes or cancels the Google Calendar event when appointment status changes to `CANCELLED`. | `googleEventId`, specialist tokens | Event deleted from Google Calendar | Logs warning and completes DB cancellation if Google fails | `ORIGINAL_REQUEST.md` Criteria |
| 16 | Google Calendar | Google Sync Status Indicator | Exposes Google Calendar connection state for specialists in UI and API. | Specialist record tokens | Sync state: `CONNECTED`, `DISCONNECTED`, `SIMULATED` | Renders connection CTA button if disconnected | `ORIGINAL_REQUEST.md` R2 |
| 17 | Public Booking | Tenant Slug Resolution | Resolves public clinic profile by slug (e.g. `/clinica-fisia`) for the booking wizard. | URL slug param | Clinic name, address, phone, branding | Custom 404 page if slug not found | `ORIGINAL_REQUEST.md` R3 |
| 18 | Public Booking | Step 1: Service Catalog Selection | Displays active clinic services with name, description, duration chip, and MXN price. | Clinic services query | Selected `serviceId` | Disables "Continuar" until selection made | `ORIGINAL_REQUEST.md` R3 |
| 19 | Public Booking | Step 2: Specialist Selection | Renders "Cualquier especialista disponible" plus list of specialists assigned to the service. | Selected `serviceId`, clinic specialists | Selected `specialistId` or `'any'` | Shows warning if no specialists assigned | `ORIGINAL_REQUEST.md` R3 |
| 20 | Public Booking | Step 3: Date & Slot Picker | Interactive date selector (next 14-30 days, Sundays disabled) and reactive time slot grid. | Selected date, service, specialist | Selected `dateTime` slot | Shows friendly empty state if day is full | `ORIGINAL_REQUEST.md` R3 |
| 21 | Public Booking | Step 4: Patient Info Form | Form collecting patient name, WhatsApp phone number, optional email, and appointment notes. | Patient input data | Validated patient payload | Inline validation errors for missing/invalid fields | `ORIGINAL_REQUEST.md` R3 |
| 22 | Public Booking | Step 5: Booking Confirmation Screen | Renders confirmation card with booking folio, summary details, calendar status, and action links. | Created appointment payload | Visual confirmation UI | Fallback error screen if creation failed | `ORIGINAL_REQUEST.md` R3 |
| 23 | Public Booking | Public Booking API (`/api/public/[slug]/book`) | Public transaction endpoint creating/linking `Customer`, creating `Appointment`, and triggering calendar sync. | Slug, `serviceId`, `specialistId`, `dateTime`, `patient` | 201 Created with appointment data | 400 (validation), 404 (not found), 409 (conflict) | `ORIGINAL_REQUEST.md` Criteria |
| 24 | Public Booking | Customer Deduplication & Upsert | Links appointment to existing `Customer` by `clinicId` + `phone`, or creates a new customer. | `clinicId`, phone, name, email, notes | `Customer` record ID | Returns validation error if phone format invalid | `prisma/schema.prisma` |
| 25 | Public Booking | Concurrency Race Condition Guard | Atomic verification immediately before appointment creation to prevent double-booking identical slots. | Specialist, dateTime, duration | Successful booking or conflict rejection | 409 Conflict: "El horario ya no está disponible" | Technical probe |
| 26 | Public Booking | WhatsApp Follow-Up Link | Action button on confirmation screen opening `https://wa.me/{phone}` with prefilled confirmation text. | Clinic phone, patient name, service, date/time | WhatsApp web/native chat launch | Graceful fallback if clinic phone is missing | Roadmap & UX probe |
| 27 | Public Booking | Add to Google Calendar Action | Provides calendar URL / `.ics` download link on confirmation screen for the patient's personal calendar. | Appointment start/end, service name, address | Google Calendar web event URL | Browser handles URL | UX probe |
| 28 | Agenda Dashboard | Clinic Calendar View (`/dashboard/calendar`) | Full interactive calendar page within protected clinic dashboard layout. | Authenticated session with `clinicId` | Weekly/daily interactive agenda view | Redirects to `/login` if unauthenticated | `ORIGINAL_REQUEST.md` R2 |
| 29 | Agenda Dashboard | Weekly Agenda View | 6-column grid (Monday to Saturday) spanning 09:00 to 19:00 with positioned appointment cards. | Week date range, appointments | Color-coded appointment blocks on grid | Empty grid columns on non-working hours | `ORIGINAL_REQUEST.md` R2 |
| 30 | Agenda Dashboard | Daily Agenda View | Focused single-day schedule view showing exact time slots and therapist allocations. | Selected single date | Detailed timeline list of appointments | Empty state if no appointments for the day | `ORIGINAL_REQUEST.md` R2 |
| 31 | Agenda Dashboard | Date Range Navigation | Header controls: "Hoy", "← Anterior", "Siguiente →", with formatted localized date header. | Current date state | Updated date range triggering appointment refetch | N/A | `ORIGINAL_REQUEST.md` R2 |
| 32 | Agenda Dashboard | Specialist Filter Dropdown | Dropdown/chip filter allowing staff to view all specialists or filter to a single therapist. | Selected `specialistId` | Filtered calendar appointments | N/A | `ORIGINAL_REQUEST.md` R2 |
| 33 | Agenda Dashboard | Status Color Coding | Visual color badges for appointments: `CONFIRMED` (teal), `COMPLETED` (emerald), `CANCELLED` (rose/strikethrough), `NO_SHOW` (amber). | `appointment.status` | Applied CSS classes and badges | Fallback neutral badge for unknown status | `ORIGINAL_REQUEST.md` R2 |
| 34 | Agenda Dashboard | Specialist Palette Accents | Distinct therapist accent color/dot to differentiate staff members across the agenda view. | `specialistId` index | Therapist color indicator | Consistent palette cycling (5 distinct colors) | UX probe |
| 35 | Agenda Dashboard | Appointment Detail Drawer | Slide-over drawer opening upon clicking an appointment card, displaying patient, service, and sync details. | Selected appointment object | Interactive detail sheet | Dismisses smoothly on click-outside or close button | `ORIGINAL_REQUEST.md` R2 |
| 36 | Agenda Dashboard | Status Update API (`PATCH /api/appointments/[id]`) | Protected endpoint allowing clinic staff to update status (`CONFIRMED`, `COMPLETED`, `CANCELLED`, `NO_SHOW`). | `appointmentId`, `status`, optional `notes` | 200 OK with updated appointment | 401 (unauthorized), 404 (not found), 400 (invalid) | `ORIGINAL_REQUEST.md` Criteria |
| 37 | Agenda Dashboard | Staff Quick Booking Modal | Allows receptionist/specialist to create an appointment directly from the dashboard calendar. | Staff form payload | 201 Created appointment, updates agenda grid | Form validation errors displayed via toasts | Clinical workflow probe |
| 38 | Agenda Dashboard | Google Calendar Connect Banner | Header banner/card showing Google sync status per specialist with one-click OAuth button. | Current specialist Google tokens | Status badge & connect/disconnect action | Disabled in simulation mode with informative tooltip | `ORIGINAL_REQUEST.md` R2 |
| 39 | Navigation | Calendar Sidebar & Mobile Nav Link | Adds "Agenda y Citas" to dashboard sidebar (`sidebar.tsx`) and mobile drawer (`mobile-nav.tsx`). | Current pathname | Highlighted active navigation item | N/A | Layout compliance probe |
| 40 | Multitenancy | Tenant Isolation on Appointments | Strictly enforces `where: { clinicId }` across all appointment, customer, and availability queries. | Session `clinicId` | Scoped data records | 403 / 404 if cross-tenant access attempted | Architectural rule |
| 41 | Toast Feedback | Sonner Toast Notifications | User notifications for appointment booking, status transitions, Google sync connection, and errors. | Toast trigger event | Styled Sonner toast popups | Auto-dismisses after 3-4s | Design system probe |

---

## 4. Edge Cases

| # | Feature | Input / Scenario | Observed / Specified Behavior |
|---|---------|------------------|-------------------------------|
| 1 | Operating Hours | Date requested is a Sunday (e.g. 2026-09-20). | Availability engine returns `isOpen: false`, `slots: []`. In UI, Sundays are disabled and grayed out. |
| 2 | Operating Hours | Slot calculation on Saturday at 13:30 for a 50 min service. | Slot would end at 14:20, exceeding Saturday 14:00 closing time. Engine clips and omits the slot. |
| 3 | Operating Hours | Slot calculation on Weekday at 18:30 for a 50 min service. | Slot would end at 19:20, exceeding 19:00 closing time. Engine clips and omits the slot. |
| 4 | Past Time Filter | User loads booking page today at 15:00. | All slots from 09:00 to 15:00 for today are omitted or marked unavailable. Future slots (16:00, 17:00, 18:00) remain available. |
| 5 | Collision Detection | Existing DB appointment at 10:00-10:50 with status `CONFIRMED`. | 10:00 slot is blocked for that specialist. Other free specialists can still offer 10:00 if booking "any". |
| 6 | Collision Detection | Existing DB appointment at 10:00-10:50 updated to `CANCELLED`. | Engine excludes `CANCELLED` appointments; 10:00 slot is immediately reopened and bookable. |
| 7 | Google freeBusy | Specialist has personal Google Calendar event marked "Busy" at 11:00-12:00. | Engine receives busy interval from Google `freeBusy` API and excludes 11:00 slot for that specialist. |
| 8 | Google freeBusy | Specialist has personal Google Calendar event marked "Free" (Show me as available). | Google Calendar does not include transparent events in `freeBusy` response; slot remains bookable. |
| 9 | Offline Fallback | `GOOGLE_CLIENT_ID` or `GOOGLE_CLIENT_SECRET` not set in `.env`. | System runs in offline simulation mode without throwing unhandled exceptions. Connect generates mock tokens, freeBusy returns empty or simulated busy spans, and event insertion yields mock ID. |
| 10 | Google API Error | Google OAuth token refresh fails or network error occurs during `freeBusy` query. | Engine catches error gracefully, logs diagnostic warning, falls back to local DB appointments, and avoids 500 error on public booking page. |
| 11 | Concurrent Booking | Two patients simultaneously attempt to book the exact same specialist and slot. | Handled via DB atomic check inside transaction: first request succeeds (201), second is rejected with 409 Conflict ("Horario ya reservado"). |
| 12 | Workload Balancing | Booking with `specialistId = 'any'`, Specialist A has 3 appointments today, Specialist B has 0. | Engine assigns appointment to Specialist B, distributing clinic patient load evenly. |
| 13 | Service Qualification | Patient attempts to book Specialist C for "Punción Seca", but Specialist C is not assigned to this service in `SpecialistService`. | API validates assignment matrix and returns 400 Bad Request ("Especialista no asignado a este servicio"). |
| 14 | Inactive Specialist | Specialist is deactivated (`isActive: false`) in clinic roster. | Specialist is excluded from public specialist list, "any specialist" pool, and slot generation. |
| 15 | Patient Deduplication | Returning patient books appointment with existing phone number but slightly different name casing or new notes. | System matches existing `Customer` by `(clinicId, phone)`, updates fields if provided, and links appointment without creating duplicate records. |
| 16 | Invalid Phone | Patient enters invalid phone number (e.g. "12345" or "abc"). | UI blocks submission; API validates regex/length (10-15 digits) and responds with 400 Bad Request. |
| 17 | Unknown Tenant Slug | User navigates to `/clinica-inexistente`. | Next.js dynamic route renders healthcare 404 page ("Clínica no encontrada") with link back to main site. |
| 18 | Timezone Conversion | Patient in California (UTC-7) books appointment in Monterrey clinic (America/Monterrey, UTC-6). | Operating hours are evaluated in clinic's timezone. Appointment start time is saved as normalized UTC ISO-8601 string (`YYYY-MM-DDTHH:mm:ss.sssZ`). |
| 19 | Event Deletion Failure | Appointment cancelled in dashboard, but Google API call to delete event fails (e.g. network timeout). | Database status is still updated to `CANCELLED` to ensure patient slot is freed; error is logged without failing the API response. |
| 20 | Service With Zero Staff | Service created in clinic but no specialists have been assigned yet. | Public booking wizard shows warning badge: "No hay especialistas asignados a este servicio actualmente" and disables slot step. |

---

## 5. Availability Engine Math & Calculation Rules (`src/lib/availability.ts`)

### 5.1 Clinic Operating Hours Matrix
Clinic operations are strictly bound by the following schedule:

```
┌───────────────┬─────────────────┬───────────────────┬──────────────┐
│ Day of Week   │ Operating Hours │ Opening Time (T0) │ Closing (T1) │
├───────────────┼─────────────────┼───────────────────┼──────────────┤
│ Monday        │ 09:00 – 19:00   │ 09:00:00          │ 19:00:00     │
│ Tuesday       │ 09:00 – 19:00   │ 09:00:00          │ 19:00:00     │
│ Wednesday     │ 09:00 – 19:00   │ 09:00:00          │ 19:00:00     │
│ Thursday      │ 09:00 – 19:00   │ 09:00:00          │ 19:00:00     │
│ Friday        │ 09:00 – 19:00   │ 09:00:00          │ 19:00:00     │
│ Saturday      │ 09:00 – 14:00   │ 09:00:00          │ 14:00:00     │
│ Sunday        │ CLOSED          │ N/A               │ N/A          │
└───────────────┴─────────────────┴───────────────────┴──────────────┘
```

### 5.2 Slot Generation & Duration Slicing
1. **Time Base**: Times are computed in 24-hour format (`HH:mm`).
2. **Service Duration ($D$)**: Retrieved from `Service.durationMin`. Default is ~50 minutes (typical physiotherapy session), but may be 45 min, 50 min, or 60 min.
3. **Slot Stepping Interval ($S$)**: 
   - Standard clinical grid increment is 60 minutes (on the hour: `09:00`, `10:00`, `11:00`, `12:00`, `13:00`, `14:00`, `15:00`, `16:00`, `17:00`, `18:00` for weekdays; `09:00`, `10:00`, `11:00`, `12:00`, `13:00` for Saturdays).
   - This natural 60-minute grid creates a built-in **10-minute turnover and sanitation buffer** between consecutive 50-minute clinical sessions (or 15 min for 45-minute sessions).
4. **Boundary Condition**:
   A slot beginning at $T_{start}$ with duration $D$ is valid if and only if:
   $$T_{start} \ge T_{open} \quad \text{and} \quad T_{start} + D \le T_{close}$$
   - On weekdays ($T_{close} = 19:00$): The last valid hourly slot starts at `18:00` (ends at `18:50` $\le 19:00$).
   - On Saturdays ($T_{close} = 14:00$): The last valid hourly slot starts at `13:00` (ends at `13:50` $\le 14:00$).

### 5.3 Collision Detection Pipeline
A slot $[T_{start}, T_{end})$ is deemed **UNAVAILABLE** for specialist $S_i$ if:
1. **Past Time Collision**: Requested date is today and $T_{start} < \text{CurrentTime} + \text{LeadBuffer}$ (where LeadBuffer $\ge 15$ min).
2. **Local Database Collision**:
   There exists an `Appointment` record where:
   - `specialistId` $= S_i$
   - `status` $\ne \text{'CANCELLED'}$ (i.e. status is `PENDING`, `CONFIRMED`, `COMPLETED`, or `NO_SHOW`)
   - The time intervals overlap:
     $$\max(T_{start}, A_{start}) < \min(T_{end}, A_{end})$$
     Where $A_{start} = \text{dateTime}$ and $A_{end} = A_{start} + \text{service.durationMin}$.
3. **Google Calendar Collision**:
   If specialist $S_i$ has connected Google Calendar:
   There exists a busy span $[B_{start}, B_{end})$ in the `freeBusy` query result where:
   $$\max(T_{start}, B_{start}) < \min(T_{end}, B_{end})$$

### 5.4 Specialist Assignment Resolution
- **Case 1: Specific Specialist Selected (`specialistId = 'user_xxx'`)**:
  1. Validate specialist exists in clinic, has `role = 'SPECIALIST'`, `isActive = true`.
  2. Validate specialist is linked to `serviceId` via `SpecialistService`.
  3. Compute free slots for this specialist only.
- **Case 2: "Cualquier Especialista Disponible" (`specialistId = 'any'`)**:
  1. Find all active specialists assigned to `serviceId`.
  2. For each candidate slot in the day:
     - Check each assigned specialist for availability.
     - If at least one specialist is free, mark the slot as `available = true`.
     - Assign candidate to the free specialist who has the **lowest count of confirmed appointments** on that date (workload balancing).

### 5.5 Offline Simulation & Fallback Mode
- If `process.env.GOOGLE_CLIENT_ID` or `process.env.GOOGLE_CLIENT_SECRET` is unset:
  - System operates in `GOOGLE_OFFLINE_MODE = true`.
  - `freeBusy` queries return empty arrays (or mock busy spans if seeded for testing).
  - Appointment creation generates a mock Google event ID: `mock_evt_<timestamp>_<rand>`.
  - Dashboard UI displays: "Google Calendar: Modo Simulación (Offline)" badge with an active green/blue indicator.

---

## 6. Google Calendar Integration Contract (`src/lib/google-calendar.ts`)

### 6.1 Authentication & Scopes
- **Library**: `googleapis` (OAuth2 client).
- **OAuth Scopes**:
  - `https://www.googleapis.com/auth/calendar.events` (read/write calendar events).
  - `https://www.googleapis.com/auth/calendar.readonly` (read free/busy metadata).
- **OAuth Parameters**:
  - `access_type: 'offline'`
  - `prompt: 'consent'` (forces issuance of `refresh_token`).
  - `state`: Serialized and signed/encoded `{ specialistId, clinicId }`.

### 6.2 Calendar Event Payload
When a booking is confirmed, the engine calls Google Calendar API:
```typescript
const eventPayload = {
  summary: `[APPCITAS] ${service.name} — ${customer.name}`,
  description: [
    `Cita de Fisioterapia — ${clinic.name}`,
    `Paciente: ${customer.name}`,
    `Teléfono / WhatsApp: ${customer.phone}`,
    customer.email ? `Email: ${customer.email}` : null,
    `Servicio: ${service.name} (${service.durationMin} min)`,
    `Especialista: ${specialist.name}`,
    notes ? `Notas: ${notes}` : null,
  ].filter(Boolean).join('\n'),
  start: {
    dateTime: slotStart.toISOString(),
    timeZone: clinicTimezone, // e.g. 'America/Monterrey'
  },
  end: {
    dateTime: slotEnd.toISOString(),
    timeZone: clinicTimezone,
  },
  attendees: customer.email ? [{ email: customer.email, displayName: customer.name }] : [],
};
```

### 6.3 Event Lifecycle Synchronization
- **Booking Creation**: `calendar.events.insert` $\rightarrow$ store returned ID in `Appointment.googleEventId`.
- **Booking Cancellation**: If `appointment.googleEventId` exists $\rightarrow$ `calendar.events.delete` $\rightarrow$ remove event from specialist's calendar.

---

## 7. Public Booking Portal UI/UX Specification (`/[slug]`)

### 7.1 Architecture & Theme
- **Path**: `src/app/[slug]/page.tsx`
- **Design Language**: Minimalist healthcare (`ui-ux-pro-max`, `apple-design`, `minimalist-ui`). Calm slate neutrals (`#f8fafc`, `#e2e8f0`, `#0f172a`), medical teal primary accents (`teal-600`, `teal-700`, `teal-50`), soft rounded cards (`rounded-2xl`), crisp typography.
- **Device Support**: Mobile-first responsive layout (375px to 1440px).

### 7.2 5-Step Flow Specification

```
┌─────────────────────────────────────────────────────────────────────────┐
│ [1. Servicio] ──> [2. Especialista] ──> [3. Horario] ──> [4. Datos] ──> [5. Confirmación] │
└─────────────────────────────────────────────────────────────────────────┘
```

1. **Step 1: Seleccionar Servicio**
   - Renders service cards with name, description, duration pill (`50 min`), and bold price (`$750.00 MXN`).
   - Radio card selection with teal border and checkmark indicator.
   - Sticky/fixed bottom "Continuar" button (disabled until service selected).
2. **Step 2: Seleccionar Especialista**
   - Primary option: "Cualquier especialista disponible" with sparkles/shuffle icon and description: "Te asignamos automáticamente al primer terapeuta disponible".
   - Followed by cards of specialists assigned to the service, showing therapist name, specialty badge, and avatar initials.
   - "Atrás" and "Continuar" buttons.
3. **Step 3: Fecha y Horario**
   - Horizontal date picker strip (scrollable on mobile) displaying the next 14 to 30 days. Sundays visually disabled with "Cerrado" tag.
   - Reactive time slot grid (fetching `/api/availability`).
   - Skeleton shimmer cards while loading.
   - Slots displayed as clean pills (`09:00`, `10:00`, `11:00`, etc.). Selected slot highlighted in solid teal.
   - "Atrás" and "Continuar" buttons.
4. **Step 4: Datos del Paciente**
   - Inputs:
     - Nombre Completo (required, text)
     - Teléfono WhatsApp (required, phone format)
     - Correo Electrónico (optional, email format)
     - Motivo de Consulta / Notas (optional, textarea)
   - Trust badge: "Tus datos están protegidos y solo se utilizarán para coordinar tu cita clínica."
   - "Confirmar Reserva" CTA button with loading spinner.
5. **Step 5: Confirmación de Cita**
   - Celebratory checkmark header: "¡Cita Confirmada con Éxito!".
   - Summary card:
     - Folio ID (e.g. `CITA-FISIA-7821`)
     - Clínica: Name, Address, Phone
     - Servicio: Name, Duration, Price
     - Especialista: Name, Specialty
     - Fecha y Hora: Formatted in Spanish (e.g. "Martes, 15 de Septiembre 2026 a las 10:00 hrs")
     - Google Calendar sync status badge
   - Actions:
     - "Abrir en WhatsApp" (opens `https://wa.me/{clinicPhone}?text=...`)
     - "Añadir a mi Calendario"
     - "Agendar otra cita" (resets wizard)

---

## 8. Interactive Clinic Calendar & Agenda Dashboard UI/UX Specification (`/dashboard/calendar`)

### 8.1 Layout & Navigation
- **Path**: `src/app/dashboard/calendar/page.tsx`
- **Integration**: Linked in sidebar (`src/components/dashboard/sidebar.tsx`) and mobile drawer (`src/components/dashboard/mobile-nav.tsx`) with Lucide `CalendarDays` icon and active state highlighting.

### 8.2 Header Controls
- **View Switcher**: Segmented pill control: "Semana" | "Día".
- **Date Navigation**:
  - `[← Anterior]` `[Hoy]` `[Siguiente →]`
  - Localized date header (e.g. "14 al 19 de Septiembre, 2026").
- **Specialist Filter**:
  - Dropdown/pill selector: "Todos los especialistas" or individual specialist cards with colored avatars.
- **Google Calendar Sync Status**:
  - Connection indicator badge per specialist.
  - "Conectar Google Calendar" CTA button if disconnected.
- **Quick Action**:
  - `[+ Nueva Cita]` button opening manual appointment modal for walk-ins/phone bookings.

### 8.3 Weekly & Daily Agenda Grids
- **Weekly View**:
  - Horizontal headers: 6 day columns (Lunes through Sábado; Domingo omitted).
  - Vertical axis: Hourly ticks from 09:00 to 19:00.
  - Appointment cards positioned dynamically by time:
    - Patient Name (bold)
    - Service name & duration chip
    - Specialist color badge
    - Status pill (`CONFIRMED`, `COMPLETED`, `CANCELLED`, `NO_SHOW`)
- **Daily View**:
  - Expanded single-day timeline with full patient notes and quick action buttons.

### 8.4 Appointment Detail Drawer (Slide-Over)
Clicking any appointment opens a slide-over sheet displaying:
- **Patient Profile**: Name, WhatsApp phone number (with one-click "Enviar WhatsApp" action link), email.
- **Clinical Booking Details**: Service, price, duration, assigned specialist, date and time.
- **Google Calendar Sync**: Sync badge and Google Event ID.
- **Status Action Bar**:
  - `[Marcar Confirmada]` (`CONFIRMED`)
  - `[Marcar Completada]` (`COMPLETED`)
  - `[Marcar No Asistió]` (`NO_SHOW`)
  - `[Cancelar Cita]` (`CANCELLED` — triggers slot liberation and Google event deletion with Sonner toast feedback).

---

## 9. Comprehensive API Endpoint Contracts

### 9.1 `GET /api/availability`
Public endpoint returning bookable slots for a service and date.

- **Query Parameters**:
  - `clinicSlug` (string) OR `clinicId` (string) — Required.
  - `serviceId` (string) — Required.
  - `date` (string, `YYYY-MM-DD`) — Required.
  - `specialistId` (string, optional) — Specific specialist ID or `'any'`. Default: `'any'`.

- **Response (200 OK — Open Day)**:
  ```json
  {
    "date": "2026-09-15",
    "dayOfWeek": 2,
    "isOpen": true,
    "operatingHours": {
      "open": "09:00",
      "close": "19:00"
    },
    "service": {
      "id": "srv_eval",
      "name": "Evaluación Inicial",
      "durationMin": 60,
      "price": 850.00
    },
    "slots": [
      {
        "time": "09:00",
        "endTime": "10:00",
        "available": true,
        "specialistId": "usr_sofia",
        "specialistName": "Lic. Sofia Garza"
      },
      {
        "time": "10:00",
        "endTime": "11:00",
        "available": true,
        "specialistId": "usr_mateo",
        "specialistName": "Lic. Mateo Rivas"
      }
    ],
    "totalAvailable": 8
  }
  ```

- **Response (200 OK — Sunday / Closed)**:
  ```json
  {
    "date": "2026-09-20",
    "dayOfWeek": 0,
    "isOpen": false,
    "message": "La clínica permanece cerrada los domingos.",
    "slots": [],
    "totalAvailable": 0
  }
  ```

- **Error Codes**:
  - `400 Bad Request`: Missing `serviceId` or invalid `date` format.
  - `404 Not Found`: Clinic or Service not found.

---

### 9.2 `POST /api/public/[slug]/book`
Public endpoint to reserve an appointment.

- **URL Parameter**: `slug` (clinic tenant slug, e.g. `clinica-fisia`).
- **Request Body**:
  ```json
  {
    "serviceId": "srv_eval",
    "specialistId": "usr_sofia", // or "any"
    "dateTime": "2026-09-15T10:00:00.000Z",
    "patient": {
      "name": "Carlos Benítez",
      "phone": "+528119998877",
      "email": "carlos.benitez@gmail.com",
      "notes": "Dolor en hombro derecho"
    }
  }
  ```

- **Response (201 Created)**:
  ```json
  {
    "success": true,
    "appointment": {
      "id": "apt_cuid123",
      "dateTime": "2026-09-15T10:00:00.000Z",
      "status": "CONFIRMED",
      "googleEventId": "gevt_mock_8291",
      "googleSynced": true,
      "clinic": {
        "name": "Clínica Fisia",
        "slug": "clinica-fisia",
        "phone": "+52 81 8000 1122",
        "address": "Av. Humberto Lobo 555, San Pedro Garza García, NL"
      },
      "service": {
        "id": "srv_eval",
        "name": "Evaluación Inicial",
        "durationMin": 60,
        "price": 850.00
      },
      "specialist": {
        "id": "usr_sofia",
        "name": "Lic. Sofia Garza",
        "specialty": "Fisioterapia Deportiva"
      },
      "customer": {
        "id": "cust_123",
        "name": "Carlos Benítez",
        "phone": "+528119998877",
        "email": "carlos.benitez@gmail.com"
      }
    }
  }
  ```

- **Error Codes**:
  - `400 Bad Request`: Missing patient name/phone, invalid phone format, or missing dateTime.
  - `404 Not Found`: Clinic or Service not found.
  - `409 Conflict`: Slot is no longer available ("El horario seleccionado ya no está disponible. Por favor seleccione otro.").

---

### 9.3 `GET /api/appointments`
Protected clinic staff endpoint returning appointments for calendar views.

- **Authentication**: Required (`getAuthSession()`). Scoped to `session.user.clinicId`.
- **Query Parameters**:
  - `startDate`: ISO date string (e.g. `2026-09-14T00:00:00.000Z`).
  - `endDate`: ISO date string (e.g. `2026-09-20T23:59:59.999Z`).
  - `specialistId` (optional): Filter by specialist.
  - `status` (optional): Filter by status (`CONFIRMED`, `COMPLETED`, `CANCELLED`, `NO_SHOW`).

- **Response (200 OK)**:
  ```json
  {
    "appointments": [
      {
        "id": "apt_cuid123",
        "dateTime": "2026-09-15T10:00:00.000Z",
        "status": "CONFIRMED",
        "notes": "Dolor en hombro derecho",
        "googleEventId": "gevt_mock_8291",
        "googleSynced": true,
        "service": {
          "id": "srv_eval",
          "name": "Evaluación Inicial",
          "durationMin": 60,
          "price": 850.00
        },
        "specialist": {
          "id": "usr_sofia",
          "name": "Lic. Sofia Garza",
          "specialty": "Fisioterapia Deportiva"
        },
        "customer": {
          "id": "cust_123",
          "name": "Carlos Benítez",
          "phone": "+528119998877",
          "email": "carlos.benitez@gmail.com"
        }
      }
    ]
  }
  ```

---

### 9.4 `PATCH /api/appointments/[id]`
Protected endpoint to update appointment status.

- **Authentication**: Required (`getAuthSession()`). Scoped to `session.user.clinicId`.
- **URL Parameter**: `id` (Appointment CUID).
- **Request Body**:
  ```json
  {
    "status": "CANCELLED", // or "COMPLETED", "NO_SHOW", "CONFIRMED"
    "notes": "Cancelado por el paciente vía telefónica" // optional
  }
  ```

- **Response (200 OK)**:
  ```json
  {
    "success": true,
    "appointment": {
      "id": "apt_cuid123",
      "status": "CANCELLED",
      "googleSynced": false // or deleted
    }
  }
  ```

- **Error Codes**:
  - `401 Unauthorized`: No active session.
  - `404 Not Found`: Appointment not found or does not belong to clinic.
  - `400 Bad Request`: Invalid status value.

---

### 9.5 Google OAuth Endpoints
- `GET /api/auth/google/connect?specialistId=...`: Initiates Google OAuth consent flow or mock simulation connection.
- `GET /api/auth/google/callback?code=...&state=...`: Receives authorization code, exchanges for tokens, stores on `User`, and redirects to `/dashboard/calendar?google_connected=true`.
- `POST /api/auth/google/disconnect`: Disconnects specialist Google Calendar, clearing tokens.
- `GET /api/auth/google/status`: Returns current sync status across specialists.

---

## 10. Database Models & Schema Validation

The Prisma schema already possesses all required fields for Phase 2:
1. `User.googleAccessToken` (`String?`) & `User.googleRefreshToken` (`String?`): Used for storing OAuth tokens.
2. `Appointment.googleEventId` (`String?`): Used for two-way synchronization.
3. `Appointment.status` (`AppointmentStatus` enum: `PENDING`, `CONFIRMED`, `COMPLETED`, `CANCELLED`, `NO_SHOW`).
4. `Appointment.dateTime` (`DateTime`): Stores exact appointment start timestamp.
5. `Customer.phone` & `Customer.clinicId`: Unique identity composite index `[clinicId, phone]` ensures seamless patient deduplication.

Multi-tenancy verification rule:
All queries on `Appointment`, `Customer`, `Service`, and `User` **MUST** include `where: { clinicId }`.

---

## 11. Test Coverage Strategy & Acceptance Verification

To guarantee zero regression and 100% acceptance compliance, Phase 2 tests must cover:

1. **Tier 1: Feature Coverage Tests**:
   - Availability Engine operating hours boundary verification (Mon-Fri 9-19, Sat 9-14, Sun closed).
   - Dynamic service duration calculations (~50 min, 45 min, 60 min).
   - Collision detection for local DB bookings.
   - Slot liberation when appointment is marked `CANCELLED`.
   - "Cualquier especialista disponible" slot aggregation and workload balancing.
   - Public booking endpoint (`/api/public/[slug]/book`) appointment & customer creation.
   - Google Calendar offline simulation fallback mode.
   - Appointment status updates (`PATCH /api/appointments/[id]`).

2. **Tier 2: Boundary & Edge Case Tests**:
   - Boundary clip: slot start time + duration exceeding closing time (e.g. 18:30 on weekday, 13:30 on Sat).
   - Past time pruning: past slots on current day are excluded.
   - Unassigned specialist rejection (specialist not linked to service in `SpecialistService`).
   - Inactive specialist rejection (`isActive: false`).
   - Concurrent race condition: 409 Conflict on double booking.
   - Non-existent tenant slug returns 404.

3. **Tier 3 & Tier 4: Clinical Workflows**:
   - Complete booking journey: Patient opens `/[slug]` -> selects service -> picks slot -> submits info -> receives confirmation -> calendar updates.
   - Receptionist calendar workflow: Staff views week agenda in `/dashboard/calendar` -> filters by therapist -> inspects drawer -> marks appointment as cancelled -> verifies slot is immediately bookable again.

4. **Production Build Integrity**:
   - `npm run build` exits 0 with no TypeScript errors across dynamic route `/[slug]` and `/dashboard/calendar`.
