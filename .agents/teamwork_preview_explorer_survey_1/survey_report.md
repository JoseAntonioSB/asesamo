# Survey Report: APPCITAS Phase 2 Architecture & Foundation
**Surveyor**: teamwork_preview_explorer_survey_1 (Explorer)  
**Date**: 2026-09-14  
**Target Project**: APPCITAS (Physiotherapy Clinic Multitenant SaaS)  
**Scope**: Schema, Database Setup, Google Calendar Integration, Availability Engine, API Routes & Migration Requirements

---

## 1. Executive Summary

APPCITAS Phase 1 (*El Cimiento*) established a robust, multitenant healthcare foundation with Next.js 14 (App Router), Prisma ORM 5.22, PostgreSQL, NextAuth.js, and a bespoke in-memory test harness verifying 417 assertions across 53 test suites in under 10ms.

For Phase 2 (*La Agenda y Google Calendar*), the database schema already anticipated Google integration by incorporating `googleAccessToken` and `googleRefreshToken` on the `User` model, and `googleEventId` on the `Appointment` model. However, key dependencies (notably `googleapis`), Google OAuth lifecycle management, the intelligent availability engine (`src/lib/availability.ts`), appointment management endpoints, the interactive dashboard calendar (`/dashboard/calendar`), and the public booking portal (`/[slug]`) must now be architected and implemented.

---

## 2. Dependencies Audit (`package.json`)

### 2.1 Current Dependencies
```json
{
  "dependencies": {
    "@prisma/client": "5.22.0",
    "bcryptjs": "2.4.3",
    "clsx": "2.1.1",
    "lucide-react": "0.460.0",
    "next": "14.2.24",
    "next-auth": "4.24.8",
    "react": "18.3.1",
    "react-dom": "18.3.1",
    "sonner": "1.7.1",
    "tailwind-merge": "2.5.4"
  },
  "devDependencies": {
    "@types/bcryptjs": "2.4.6",
    "@types/node": "20.17.0",
    "@types/react": "18.3.12",
    "@types/react-dom": "18.3.1",
    "autoprefixer": "10.4.20",
    "postcss": "8.4.49",
    "prisma": "5.22.0",
    "tailwindcss": "3.4.15",
    "tsx": "4.19.2",
    "typescript": "5.6.3"
  }
}
```

### 2.2 Missing Dependencies for Phase 2
1. **`googleapis`**:
   - **Requirement**: ORIGINAL_REQUEST §R1 specifies: *"Implement Google Calendar synchronization for specialists using `googleapis` with OAuth2"*.
   - **Status**: Currently **NOT** installed.
   - **Recommendation**: Install `googleapis` (`npm install googleapis`). Note: `googleapis` bundles its own first-class TypeScript types, so `@types/googleapis` is not needed.
2. **Date Libraries (`date-fns` / `dayjs`)**:
   - **Status**: Neither is currently installed.
   - **Assessment**: The project currently uses native JavaScript `Date` and Intl formatters (e.g. in `src/lib/utils.ts` and `src/lib/serialize.ts`). Native date arithmetic (converting minutes to milliseconds, extracting day-of-week 0-6, formatted date strings `YYYY-MM-DD`) is clean and avoids external dependency bloat. However, if timezone-aware formatting or heavy calendar interval math is desired, `date-fns` can be added without conflict, or lightweight native date helpers can be implemented directly inside `src/lib/availability.ts`.
3. **Calendar UI Libraries**:
   - In `Roadmap.md`, external libraries like `FullCalendar` were mentioned. However, FullCalendar in Next.js 14 App Router frequently encounters SSR hydration mismatches, complex CSS bundle overrides, and heavy footprint. Building a responsive weekly/daily agenda view using Tailwind CSS + Lucide icons (already installed and thoroughly styled in Phase 1) provides seamless UI coherence, 0 extra bundle overhead, and instant mobile responsiveness.

---

## 3. Database Schema & Data Models (`prisma/schema.prisma`)

### 3.1 Model Analysis

| Model | Role in Multitenancy & Phase 2 | Existing Fields Relevant to Phase 2 | Proposed Phase 2 Adjustments |
|---|---|---|---|
| `Clinic` | Root tenant entity. Foreign key `clinicId` enforced on all child models. | `id`, `slug` (@unique), `name`, `phone`, `address` | None required. Operating hours (Mon-Fri 9-19, Sat 9-14) can be defined as clinic defaults. |
| `User` | Owners & Specialists (`role`: `OWNER`, `SPECIALIST`). Max 5 specialists quota. | `googleAccessToken String?`, `googleRefreshToken String?` | Add `googleTokenExpiry DateTime?`, `googleCalendarId String?`, `googleEmail String?`. |
| `Service` | Clinic services catalog. Duration in minutes and pricing. | `id`, `clinicId`, `name`, `durationMin`, `price` | None required. Already provides `durationMin` (e.g. 45, 50, 60 min) for slot calculations. |
| `SpecialistService` | Join table for M:N specialist assignments to services. | `specialistId`, `serviceId`, `clinicId` | None required. Accurately identifies which specialists can perform which services. |
| `Customer` | Clinic patients directory. Unique index on `[clinicId, phone]`. | `id`, `clinicId`, `name`, `phone`, `email?`, `notes?` | None required. Captures patient WhatsApp phone, name, and notes. |
| `Appointment` | Clinical session bookings. | `dateTime DateTime`, `status AppointmentStatus`, `googleEventId String?` | Optional: Add `endTime DateTime?` for explicit range indexing. `dateTime` serves as start time. |
| `AppointmentStatus` | Enum: `PENDING`, `CONFIRMED`, `COMPLETED`, `CANCELLED`, `NO_SHOW` | Already has all required values. | Fully matches Phase 2 specifications. |
| `Order` / `Waitlist` | Cash closing & cancellation queues. | Structured for Phase 1 & 3. | Ready for integration with appointment cancellation. |

### 3.2 Proposed Schema Extensions (100% Backwards-Compatible)

```prisma
model User {
  // ... existing fields ...
  googleAccessToken  String?
  googleRefreshToken String?
  googleTokenExpiry  DateTime?   // NEW: Token expiration timestamp for proactive refresh
  googleCalendarId   String?     // NEW: Designated Google Calendar ID (defaults to 'primary')
  googleEmail        String?     // NEW: Connected Google Account email for UI badge display
  // ... relations ...
}

model Appointment {
  // ... existing fields ...
  dateTime      DateTime        // Start timestamp
  endTime       DateTime?       // NEW: Explicit end timestamp (dateTime + service.durationMin)
  googleEventId String?         // Google Calendar event ID
  // ... relations ...
}
```

*Note: All additions are nullable (`?`), guaranteeing zero regressions for existing seeds, queries, and tests.*

---

## 4. Existing Backend Utilities & API Architecture

### 4.1 Utility Modules
- **`src/lib/prisma.ts`**: Global PrismaClient singleton. Handles BigInt serialization and log verbosity.
- **`src/lib/auth.ts`**: NextAuth credentials configuration with JWT session strategy. Includes test bypass hook `globalThis.__TEST_SESSION__` used by the test runner for hermetic execution.
- **`src/lib/serialize.ts`**: Deep serialization utility converting Prisma `Decimal` instances into plain JavaScript numbers, and safe handling of dates.
- **`src/lib/utils.ts`**: Contains `cn()` (clsx + tailwind-merge), `formatCurrency()` (Intl MXN), and `formatDuration()`.

### 4.2 Route Protection & Middleware (`src/middleware.ts`)
- Matches `/dashboard/:path*`, `/login`, and `/register`.
- Automatically redirects unauthenticated requests targeting `/dashboard/*` to `/login?callbackUrl=...`.
- Public routes (such as `/[slug]`, `/api/public/*`, `/api/appointments/*`) are intentionally **not** intercepted by middleware, allowing public booking flows to execute freely.

### 4.3 Existing API Endpoints
- `POST /api/clinic/register`: Clinic tenant creation + Owner user creation.
- `GET /api/specialists`: Fetches active specialists and service assignment counts.
  - *Phase 2 note*: Should be updated or augmented to report `hasGoogleCalendarConnected: boolean` and `googleEmail: string | null` without exposing sensitive access/refresh tokens.
- `POST /api/specialists`: Enforces 5-specialist active quota, creates credentials.
- `GET /api/services`: Lists clinic services and specialist associations.
- `POST /api/services`: Creates clinical services with duration and price validation.
- `PUT /api/services/[id]/assign`: Atomically replaces specialist assignments for a service.

---

## 5. Database Setup, Connectivity & Test Simulation Engine

### 5.1 Environment Configuration (`.env`)
```bash
DATABASE_URL="postgresql://fisia_user:fisia_secret_2026@localhost:5432/appcitas?schema=public"
NEXTAUTH_URL="http://localhost:3000"
NEXTAUTH_SECRET="fisia_super_secret_jwt_key_2026_production_grade"
NODE_ENV="development"
```

### 5.2 Offline Resiliency & Test Harness Architecture
- In the local development sandbox, direct socket connection to external PostgreSQL instances is restricted or offline.
- Phase 1 solved this elegantly: `tests/verify-seed.ts`, `tests/adversarial-m2.test.ts`, and `tests/e2e/harness.ts` implement an in-memory relational simulation engine with genuine constraint validation, multi-tenant isolation checks, and relationship resolution.
- Running `npx tsx tests/e2e/run-tests.ts` verifies 417 assertions across all 4 tiers with 100% success in 8ms.
- Running `npx prisma validate` and `npx prisma generate` execute with zero errors.
- Running `npm run build` compiles 12 routes with 0 TypeScript or packaging errors.

---

## 6. Phase 2 Architectural Blueprint & Implementation Plan

### 6.1 Google Calendar Integration Module (`src/lib/google-calendar.ts`)
A dedicated module encapsulating Google Calendar interactions with OAuth2:
1. **OAuth2 Setup**:
   ```typescript
   import { google } from 'googleapis';

   export function getOAuth2Client() {
     return new google.auth.OAuth2(
       process.env.GOOGLE_CLIENT_ID,
       process.env.GOOGLE_CLIENT_SECRET,
       process.env.GOOGLE_REDIRECT_URI || `${process.env.NEXTAUTH_URL}/api/google/callback`
     );
   }
   ```
2. **Simulation / Fallback Mode (Mandatory for R1 & Offline Testing)**:
   - When `process.env.GOOGLE_CLIENT_ID` or `process.env.GOOGLE_CLIENT_SECRET` are not set:
     - `isGoogleConfigured()` returns `false`.
     - `getFreeBusy()` returns an empty busy interval array `[]` (or simulated busy slots if testing conflicts).
     - `createCalendarEvent()` generates a mock event ID (`mock_gevt_${Date.now()}`) and logs event creation.
     - Specialists can trigger a simulated connection in the UI (`/api/google/auth?mock=true`) for instant testing without requiring real Google Cloud Console credentials.
3. **Token Refresh Strategy**:
   - Checks `user.googleTokenExpiry`. If expired (or expiring within 5 minutes) and `googleRefreshToken` is present:
     `oauth2Client.setCredentials({ refresh_token: user.googleRefreshToken })`
     `const { credentials } = await oauth2Client.refreshAccessToken()`
     Updates `User` in database with new access token and expiry.
4. **Calendar Operations**:
   - `fetchFreeBusy(tokens, timeMin, timeMax, calendarId = 'primary')`: Returns `{ start: string, end: string }[]`.
   - `createCalendarEvent(tokens, details)`: Inserts event with summary `[APPCITAS] {serviceName} - {patientName}`, description with phone and notes, and start/end times.
   - `deleteCalendarEvent(tokens, eventId)`: Removes event upon appointment cancellation.

### 6.2 Intelligent Availability Engine (`src/lib/availability.ts`)
1. **Operating Hours Rules**:
   - Monday – Friday: `09:00` to `19:00`
   - Saturday: `09:00` to `14:00`
   - Sunday: `CLOSED` (0 slots)
2. **Slot Generation Algorithm**:
   - Inputs: `clinicId: string`, `date: string (YYYY-MM-DD)`, `serviceId: string`, `specialistId?: string` ('any' or specific ID).
   - Fetch service duration (`service.durationMin`, e.g. 45-60 min).
   - Generate candidate slot intervals during clinic operating hours.
   - Query local DB for existing appointments on that date where `status !== 'CANCELLED'`.
   - For eligible specialist(s), query Google Calendar `freeBusy` (or simulation).
   - Filter candidate slots: a slot `[start, end]` is open if:
     - It does not overlap any local DB appointment (`aptStart < slotEnd && aptEnd > slotStart`).
     - It does not overlap any Google Calendar busy span (`busyStart < slotEnd && busyEnd > slotStart`).
     - If the date is today, slot start is strictly in the future (`slotStart > Date.now()`).
   - For `specialistId === 'any'`: a slot is bookable if **at least one** qualified specialist is free.

### 6.3 Required New API Endpoints

| Method | Route | Description | Auth Required |
|---|---|---|:---:|
| `GET` | `/api/google/auth` | Initiates Google OAuth consent flow or simulation | Yes (Owner/Specialist) |
| `GET` | `/api/google/callback` | OAuth code exchange callback; persists tokens on User | No (Google redirect) |
| `POST` | `/api/google/disconnect` | Clears Google tokens from specialist User record | Yes (Owner/Specialist) |
| `GET` | `/api/appointments` | Lists appointments for clinic with date range & specialist filters | Yes (Dashboard) |
| `GET` | `/api/appointments/[id]` | Fetches appointment detail (patient, service, specialist, order) | Yes (Dashboard) |
| `PATCH` | `/api/appointments/[id]` | Updates status (`CONFIRMED`, `COMPLETED`, `CANCELLED`, `NO_SHOW`) | Yes (Dashboard) |
| `GET` | `/api/public/[slug]/config` | Returns clinic name, address, active services & specialist roster | No (Public) |
| `GET` | `/api/public/[slug]/availability` | Returns open bookable slots for date + service + specialist | No (Public) |
| `POST` | `/api/public/[slug]/book` | Atomic booking: creates Customer, Appointment, triggers Google sync | No (Public) |

### 6.4 UI Views & Frontend Components
1. **Interactive Clinic Calendar (`/dashboard/calendar`)**:
   - Weekly and Daily view toggles.
   - Date navigation (Today, Prev Week/Day, Next Week/Day, Date picker).
   - Specialist filter selector.
   - Color-coded appointment chips (by specialist and status: `CONFIRMED` emerald, `COMPLETED` blue, `CANCELLED` rose, `NO_SHOW` amber).
   - Appointment detail modal / drawer: Patient name, WhatsApp phone (clickable link `https://wa.me/...`), service, time, notes, and status update selector.
   - Specialist Google Sync Status Bar: Shows badge per specialist (Connected / Disconnected) with 1-click connect button.
2. **Dashboard Navigation Integration**:
   - Update `src/components/dashboard/sidebar.tsx` and `mobile-nav.tsx` to include `/dashboard/calendar` (`Calendar` icon, badge `"Agenda"`).
3. **Public Booking Portal (`src/app/[slug]/page.tsx`)**:
   - Clean, minimalist healthcare aesthetic following `ui-ux-pro-max` and `minimalist-ui`.
   - **Step 1**: Service Selection (name, description, duration ~50 min, price).
   - **Step 2**: Specialist Selection (specialist card with specialty or "Cualquier especialista disponible").
   - **Step 3**: Interactive Date Picker & Available Time Slot Chips (fetched dynamically from availability engine).
   - **Step 4**: Patient Details Form (Name, WhatsApp Phone Number, optional Email and Notes).
   - **Step 5**: Confirmation Summary Card with appointment details, clinic location, and WhatsApp contact prompt.

---

## 7. Migration & Compatibility Assessment

1. **Schema Backwards Compatibility**:
   Adding optional fields `googleTokenExpiry`, `googleCalendarId`, `googleEmail` on `User` and `endTime` on `Appointment` introduces **no breaking changes**. Existing queries using Prisma select/include will continue functioning identically.
2. **Seed Script Idempotency**:
   `prisma/seed.ts` already provisions sample appointments and specialists without specifying these optional fields, preserving 100% compatibility.
3. **E2E Test Suite Integrity**:
   All 417 existing assertions across Tiers 1 through 4 remain unaffected. A dedicated Phase 2 test suite (`tests/e2e/phase2-calendar-booking.test.ts`) can be integrated into the test runner to verify availability logic, slot exclusions, simulation mode, and booking creation.

---

## 8. Summary Checklist for Implementation Team

- [ ] Install `googleapis` in `package.json`.
- [ ] Add optional Google token expiry and email fields to `prisma/schema.prisma` and run `npx prisma generate`.
- [ ] Implement `src/lib/google-calendar.ts` with OAuth2 handling, Google Calendar API calls, and local simulation mode.
- [ ] Implement `src/lib/availability.ts` with operating hours, DB conflict subtraction, and Google freeBusy integration.
- [ ] Create Google OAuth endpoints: `/api/google/auth`, `/api/google/callback`, `/api/google/disconnect`.
- [ ] Create appointment management endpoints: `/api/appointments`, `/api/appointments/[id]`.
- [ ] Create public booking endpoints: `/api/public/[slug]/config`, `/api/public/[slug]/availability`, `/api/public/[slug]/book`.
- [ ] Build `/dashboard/calendar` interactive agenda with weekly/daily views and specialist sync indicators.
- [ ] Add calendar link to sidebar and mobile navigation.
- [ ] Build high-converting public booking portal at `src/app/[slug]/page.tsx`.
- [ ] Create Phase 2 automated test suite validating availability calculations and booking transactions.
- [ ] Verify `npm run build` completes with 0 errors across all routes.
