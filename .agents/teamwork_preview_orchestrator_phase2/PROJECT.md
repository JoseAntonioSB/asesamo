# Project: APPCITAS Phase 2 — La Agenda y Google Calendar

## Architecture
- **Next.js App Router (14.2.24)** + TypeScript + Tailwind CSS.
- **Database**: Prisma ORM with PostgreSQL schema. Core models: `Clinic`, `User` (specialist), `Service`, `SpecialistService`, `Customer`, `Appointment`.
- **Google Calendar Integration (`src/lib/google-calendar.ts`)**:
  - `googleapis` OAuth2 client for specialist Google account linking.
  - Token refresh management.
  - `freeBusy` query integration for real-time external calendar collisions.
  - Automated event creation (`events.insert`) with patient and appointment metadata.
  - Resilient offline simulation fallback mode when Google credentials are not configured in local environment.
- **Availability Engine (`src/lib/availability.ts`)**:
  - Operating hours: Monday–Friday 09:00–19:00, Saturday 09:00–14:00, Sunday closed.
  - Time slot increment: 60-minute grid matching standard ~50 min service duration with 10-minute turnaround buffer.
  - Collision subtraction: Local database appointments (`CONFIRMED`, `PENDING`, `COMPLETED`, `NO_SHOW` block time; `CANCELLED` frees time) + Google Calendar `busy` intervals.
  - "Cualquier especialista disponible" mode: returns union of all active specialists' open slots, load-balancing assignment by selecting the therapist with least bookings on that day.
- **Dashboard Calendar (`/dashboard/calendar`)**:
  - Weekly and Daily interactive agenda views.
  - Color-coding by specialist and appointment status.
  - Specialist filter dropdown and date navigation controls (Prev, Next, Today).
  - Appointment details drawer/modal with patient contact info, service details, and status update actions (`CONFIRMED`, `COMPLETED`, `CANCELLED`, `NO_SHOW`).
  - Google Calendar sync status badges per specialist with connection flow trigger.
- **Public Booking Flow (`/[slug]`)**:
  - Step 1: Select Service (~50 min duration, price, description).
  - Step 2: Select Specialist or "Cualquier especialista disponible".
  - Step 3: Date & time slot picker dynamically rendered from the availability engine.
  - Step 4: Patient information form (Name, WhatsApp phone, optional email/notes).
  - Step 5: Confirmation screen with appointment summary, calendar sync feedback, and WhatsApp click-to-chat action.

---

## Feature Inventory
| # | Feature | Description | Milestone | Source |
|---|---------|-------------|-----------|--------|
| 1 | Google OAuth Integration | Specialist OAuth2 authentication using `googleapis`, token persistence | M1 | ORIGINAL_REQUEST §R1 |
| 2 | Token Refresh Management | Automatic renewal of expired Google OAuth access tokens | M1 | spec_report §3.1 |
| 3 | Resilient Fallback Simulation | Mock OAuth and calendar sync when credentials are not configured | M1 | ORIGINAL_REQUEST §R1 |
| 4 | Google FreeBusy Sync | Query external Google Calendar busy spans to prevent collisions | M1 | ORIGINAL_REQUEST §R1 |
| 5 | Calendar Event Creation | Automatically create Google Calendar event on appointment confirmation | M1 | ORIGINAL_REQUEST §R1 |
| 6 | Clinic Operating Hours Engine | Mon-Fri 09:00-19:00, Sat 09:00-14:00, Sunday closed validation | M1 | ORIGINAL_REQUEST §R1 |
| 7 | Slot Grid Generation | 60-min slots respecting ~50 min service duration and 10 min turnover | M1 | ORIGINAL_REQUEST §R1 |
| 8 | Local DB Collision Detection | Exclude occupied slots for non-cancelled appointments | M1 | ORIGINAL_REQUEST §R1 |
| 9 | Slot Liberation on Cancel | Instantly reopen slot when appointment is updated to `CANCELLED` | M1 | ORIGINAL_REQUEST §AC |
| 10 | "Cualquier Especialista" Mode | Pool availability across qualified specialists and load-balance assignment | M1 | ORIGINAL_REQUEST §R3 |
| 11 | Availability API Endpoint | `GET /api/availability` returning bookable time slots | M1 | spec_report §8.1 |
| 12 | Public Booking API Endpoint | `POST /api/public/[slug]/book` creating customer and appointment | M1 | ORIGINAL_REQUEST §AC |
| 13 | Appointment Status Update API | `PATCH /api/appointments/[id]` updating status (`CONFIRMED`, `COMPLETED`, `CANCELLED`, `NO_SHOW`) | M1 | ORIGINAL_REQUEST §AC |
| 14 | Patient Deduplication | Upsert/match customer by `(clinicId, phone)` | M1 | spec_report §3.4 |
| 15 | Dashboard Navigation Update | Add `/dashboard/calendar` to sidebar, mobile nav, and header | M2 | ORIGINAL_REQUEST §R2 |
| 16 | Calendar Weekly View | 7-day grid displaying booked appointments with specialist color tags | M2 | ORIGINAL_REQUEST §R2 |
| 17 | Calendar Daily View | Detailed time-column daily agenda view with status badges | M2 | ORIGINAL_REQUEST §R2 |
| 18 | Specialist Filter Control | Filter calendar appointments by individual specialist or all specialists | M2 | ORIGINAL_REQUEST §R2 |
| 19 | Date Navigation Controls | Previous, Next, Today quick date navigation buttons | M2 | ORIGINAL_REQUEST §R2 |
| 20 | Appointment Detail Drawer | Modal/slide-over drawer with patient data, notes, and status actions | M2 | ORIGINAL_REQUEST §R2 |
| 21 | Appointment Status Changer | Quick action dropdown to change status with instant UI update | M2 | ORIGINAL_REQUEST §R2 |
| 22 | Google Sync Status Badges | Specialist sync indicator (Connected vs Not Connected) with one-click connect | M2 | ORIGINAL_REQUEST §R2 |
| 23 | Public Portal Slug Resolution | Dynamic `/[slug]` resolving clinic by unique slug | M3 | ORIGINAL_REQUEST §R3 |
| 24 | Step 1: Service Selection | Interactive service cards displaying title, duration, price, description | M3 | ORIGINAL_REQUEST §R3 |
| 25 | Step 2: Specialist Selection | Select therapist card or "Cualquier especialista disponible" | M3 | ORIGINAL_REQUEST §R3 |
| 26 | Step 3: Date & Slot Picker | Calendar date selector + dynamic grid of available time slots | M3 | ORIGINAL_REQUEST §R3 |
| 27 | Step 4: Patient Info Form | Form with Name, WhatsApp phone number, optional email and clinical notes | M3 | ORIGINAL_REQUEST §R3 |
| 28 | Step 5: Booking Confirmation | Success card with summary, calendar event confirmation, WhatsApp action | M3 | ORIGINAL_REQUEST §R3 |
| 29 | Mobile-First Responsive UX | Fluid layout and touch targets for mobile and desktop screens | M3 | ORIGINAL_REQUEST §R3 |
| 30 | Healthcare Minimalist Aesthetic | Clean medical-teal palette, Apple-style feedback, accessible typography | M3 | ORIGINAL_REQUEST §AC |
| 31 | Concurrency Double-Booking Guard | Atomic validation in booking transaction to prevent race collisions | M1 | spec_report §4.7 |
| 32 | E2E Test Suite (Tiers 1-4) | Comprehensive opaque-box test suite verifying availability and booking | M4 | ORIGINAL_REQUEST §AC |
| 33 | Production Build Verification | `npm run build` passes with 0 errors across all routes | M4 | ORIGINAL_REQUEST §AC |

---

## Milestones
| # | Name | Scope | Dependencies | Status |
|---|------|-------|-------------|--------|
| M1 | Google Calendar & Availability Engine | `googleapis` installation, schema extension (`googleTokenExpiry`, `googleCalendarId`, `googleEmail`), `src/lib/google-calendar.ts` (with simulation mode), `src/lib/availability.ts`, `/api/availability`, `/api/public/[slug]/book`, `/api/appointments/[id]`, `/api/auth/google/*` | none | PLANNED |
| M2 | Interactive Clinic Calendar Dashboard | `/dashboard/calendar` page, weekly & daily agenda views, specialist filters, appointment detail drawer, status updater, Google sync badges, navigation links | M1 | PLANNED |
| M3 | Public Booking Portal (`/[slug]`) | Responsive 5-step booking flow at `src/app/[slug]/page.tsx`, components in `src/components/booking/`, mobile-first responsive design, slot selection, customer submission, confirmation screen | M1 | PLANNED |
| M4 | Comprehensive E2E Verification & Build | Full E2E test execution (Tiers 1-4), adversarial test hardening (Tier 5), zero build errors (`npm run build`), forensic audit | M1, M2, M3 | PLANNED |

---

## Interface Contracts
### 1. `src/lib/google-calendar.ts`
```typescript
export interface GoogleAuthTokens {
  accessToken: string;
  refreshToken?: string;
  expiryDate?: number;
}

export interface FreeBusyInterval {
  start: Date;
  end: Date;
}

export interface CreateEventParams {
  specialistId: string;
  summary: string;
  description: string;
  startTime: Date;
  endTime: Date;
  attendeeEmail?: string;
}

export function getGoogleOAuthClient(): any;
export function generateAuthUrl(specialistId: string): string;
export function handleOAuthCallback(code: string, specialistId: string): Promise<GoogleAuthTokens>;
export function getSpecialistFreeBusy(specialistId: string, timeMin: Date, timeMax: Date): Promise<FreeBusyInterval[]>;
export function createCalendarEvent(params: CreateEventParams): Promise<{ eventId: string; htmlLink?: string }>;
```

### 2. `src/lib/availability.ts`
```typescript
export interface TimeSlot {
  time: string; // "HH:MM" e.g. "09:00"
  dateTime: string; // ISO-8601 string
  specialistId?: string;
  specialistName?: string;
  available: boolean;
}

export interface GetAvailabilityParams {
  clinicId: string;
  serviceId: string;
  date: string; // "YYYY-MM-DD"
  specialistId?: string; // specific specialist ID or undefined for "Cualquier especialista"
}

export function getBookableSlots(params: GetAvailabilityParams): Promise<TimeSlot[]>;
```

### 3. Public Booking Endpoint: `POST /api/public/[slug]/book`
```typescript
// Request Body:
{
  serviceId: string;
  specialistId?: string; // specific ID or omitted/"any"
  dateTime: string; // ISO-8601 string e.g. "2026-09-15T10:00:00.000Z"
  customerName: string;
  customerPhone: string;
  customerEmail?: string;
  notes?: string;
}

// Response (201 Created):
{
  success: true,
  appointment: {
    id: string;
    dateTime: string;
    status: "CONFIRMED";
    specialist: { id: string; name: string };
    service: { id: string; name: string; duration: number; price: number };
    customer: { id: string; name: string; phone: string };
    googleEventId?: string;
  }
}
```

---

## Code Layout
```
src/
├── app/
│   ├── [slug]/
│   │   └── page.tsx                         # Public mobile-first booking portal (M3)
│   ├── api/
│   │   ├── appointments/
│   │   │   ├── route.ts                     # List & query appointments (M1)
│   │   │   └── [id]/
│   │   │       └── route.ts                 # Update status (PATCH) / Get details (M1)
│   │   ├── auth/
│   │   │   └── google/
│   │   │       ├── connect/route.ts         # Google OAuth redirect (M1)
│   │   │       └── callback/route.ts        # Google OAuth callback handler (M1)
│   │   ├── availability/
│   │   │   └── route.ts                     # Availability query endpoint (M1)
│   │   └── public/
│   │       └── [slug]/
│   │           ├── route.ts                 # Clinic info by slug (M3)
│   │           ├── availability/route.ts    # Public availability for slug (M1/M3)
│   │           └── book/route.ts            # Public booking creation (M1/M3)
│   └── dashboard/
│       └── calendar/
│           └── page.tsx                     # Interactive Clinic Agenda (M2)
├── components/
│   ├── booking/
│   │   ├── booking-wizard.tsx               # 5-step booking container (M3)
│   │   ├── step-service.tsx                 # Step 1: Select Service (M3)
│   │   ├── step-specialist.tsx              # Step 2: Select Specialist (M3)
│   │   ├── step-datetime.tsx                # Step 3: Date & Slot Picker (M3)
│   │   ├── step-patient-info.tsx            # Step 4: Patient Info Form (M3)
│   │   └── step-confirmation.tsx            # Step 5: Confirmation Card (M3)
│   └── dashboard/
│       ├── calendar/
│       │   ├── agenda-view.tsx              # Week & Day view toggle/renderer (M2)
│       │   ├── calendar-header.tsx          # Specialist filters & date nav (M2)
│       │   ├── appointment-drawer.tsx       # Detail drawer & status actions (M2)
│       │   └── google-sync-modal.tsx        # Sync status & connect action (M2)
│       ├── sidebar.tsx                      # Add /dashboard/calendar (M2)
│       └── mobile-nav.tsx                   # Add /dashboard/calendar (M2)
└── lib/
    ├── google-calendar.ts                   # Googleapis OAuth & sync + simulation (M1)
    └── availability.ts                      # Intelligent slot availability engine (M1)
tests/
└── e2e/
    └── phase2-runner.ts                     # Comprehensive Phase 2 test harness (M4)
```
