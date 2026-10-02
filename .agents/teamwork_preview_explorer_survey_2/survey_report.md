# APPCITAS Frontend, App Router, UI Components & Test Suite Survey Report

**Survey Date:** 2026-09-14  
**Investigator:** Teamwork Explorer Survey Agent (`teamwork_preview_explorer_survey_2`)  
**Workspace Root:** `/Users/jasbotello/devel/APPCITAS`  
**Target Milestone:** Phase 2 (La Agenda y Google Calendar) Foundation  

---

## Executive Summary

APPCITAS Phase 1 (*El Cimiento*) has established a clean, production-ready foundation using **Next.js 14.2.24 (App Router)**, **React 18.3.1**, **TypeScript 5.6.3**, **Tailwind CSS 3.4.15**, **NextAuth 4.24.8**, and **Prisma ORM 5.22.0**.

The frontend employs an Apple-inspired minimalist healthcare aesthetic (calm neutrals, soft medical teal `#0D9488` / `#14B8A6`, crisp typography, and fluid `active:scale-[0.98]` tactile feedback) without external component framework bloat (no Radix UI, no heavy form libraries).

The test suite is powered by a high-performance, standalone TypeScript assertion engine (`tests/e2e/harness.ts`) executed via `npx tsx`, passing **417/417 assertions in 8ms** with zero dependencies. Production build (`next build`) compiles cleanly with **0 TypeScript or packaging errors**.

Key preparation findings for Phase 2:
1. **Google Calendar fields** (`googleAccessToken`, `googleRefreshToken`, `googleEventId`) and `AppointmentStatus` enums already exist in `prisma/schema.prisma`.
2. **Dashboard Calendar plug-in points** are cleanly identified in `sidebar.tsx`, `mobile-nav.tsx`, `header.tsx`, and `dashboard/page.tsx`.
3. **Public routing `/[slug]`** is not yet implemented, but the tenant data model (`Clinic.slug`), normalization routines, and header sharing buttons are already configured for it. Middleware specifically leaves `/[slug]` unintercepted for public patient access.
4. **Missing npm script:** `package.json` currently lacks `"test"`. Running `npm test` fails with `Missing script: "test"`. Tests must currently be invoked via `npx tsx tests/e2e/run-tests.ts`.

---

## 1. Next.js App Router Map & Layout Hierarchy

```
src/app/
├── layout.tsx                     # Global Root Layout (Sonner Toaster, Tailwind, force-dynamic)
├── page.tsx                       # Landing Page (public marketing & auth redirection)
├── (auth)/                        # Route Group for Authentication
│   ├── login/
│   │   └── page.tsx               # Credentials Login Page (NextAuth signIn, test quick-fills)
│   └── register/
│       └── page.tsx               # Clinic Registration Page (calls /api/clinic/register)
├── dashboard/                     # Authenticated Dashboard Shell
│   ├── layout.tsx                 # Dashboard Layout (SessionProvider, Sidebar, MobileNav, Header)
│   ├── page.tsx                   # Overview Page (KPI bento cards, clinic roster & service previews)
│   ├── specialists/
│   │   └── page.tsx               # Specialists Roster (5-quota indicator, search, OnboardModal)
│   └── services/
│       └── page.tsx               # Services Catalog (pricing, duration, assignment matrix modal)
│   └── [calendar/ (PHASE 2)]      # Target plug-in for weekly/daily interactive agenda
├── [slug]/ (PHASE 2)              # Target dynamic route for public booking portal
└── api/                           # Route Handlers
    ├── auth/[...nextauth]/
    │   └── route.ts               # NextAuth Handler (Credentials Provider, JWT callbacks)
    ├── clinic/
    │   └── register/
    │       └── route.ts           # Clinic & Owner Provisioning (slug generation, bcrypt)
    ├── services/
    │   ├── route.ts               # Services GET, POST, PATCH (with Decimal serialization)
    │   └── [id]/
    │       └── assign/
    │           └── route.ts       # Bulk Specialist Assignment PUT
    ├── specialists/
    │   └── route.ts               # Specialists GET, POST (5-specialist tenant quota)
    ├── [public/[slug]/book (P2)]  # Target endpoint for public patient appointment booking
    └── [appointments/[id] (P2)]   # Target endpoint for status updates and calendar sync
```

### Route Descriptions & Direct Observations

| Route | Type | File Path | Direct Observations & Implementation Details |
|---|---|---|---|
| `/` | Server Component | `src/app/page.tsx` | Renders public landing page. Invokes `getAuthSession()`: if session with `clinicId` exists, immediately redirects to `/dashboard`. |
| `/layout.tsx` | Server Component | `src/app/layout.tsx` | Root layout. Sets `export const dynamic = 'force-dynamic'`. Injects `<Toaster position="top-right" richColors closeButton duration={3500} offset={20} />` from Sonner. Imports `'tailwindcss/tailwind.css'`. Sets `lang="es"`. |
| `/login` | Client Component | `src/app/(auth)/login/page.tsx` | Uses NextAuth `signIn('credentials')`. Handles errors, displays quick-fill chips for pilot accounts (`Dr. Alejandro Morales`, `Lic. Sofia Garza`). |
| `/register` | Client Component | `src/app/(auth)/register/page.tsx` | Multi-field clinic & owner sign-up form. Submits to `/api/clinic/register`, automatically triggers `signIn` upon success. |
| `/dashboard` | Client Component | `src/app/dashboard/page.tsx` | Welcome banner with clinic live pill, quick-action link copier for public booking, 4 KPI cards (Specialists 5/5, Services count, Today's appointments placeholder, Cash closing foundation), and roster/services preview tables. |
| `/dashboard/layout.tsx` | Client Component | `src/app/dashboard/layout.tsx` | Wraps children in `<SessionProvider>`. Houses desktop `<Sidebar />` (fixed left, hidden `<1024px`), mobile `<MobileNav />` drawer, and sticky `<Header />` with menu trigger. |
| `/dashboard/specialists` | Client Component | `src/app/dashboard/specialists/page.tsx` | Roster management. Fetches `/api/specialists`. Displays 5-specialist quota progress bar (`quotaUsed / 5`), search filter, specialist cards, and opens `<OnboardSpecialistModal />`. |
| `/dashboard/services` | Client Component | `src/app/dashboard/services/page.tsx` | Service catalog. Fetches `/api/services` and `/api/specialists`. Displays service cards with duration/price, inline assignment checkboxes, new service creation modal, and bulk `<SpecialistAssignmentModal />`. |
| `/api/auth/[...nextauth]` | Route Handler | `src/app/api/auth/[...nextauth]/route.ts` | NextAuth route handler importing `authOptions` from `@/lib/auth`. |
| `/api/clinic/register` | Route Handler | `src/app/api/clinic/register/route.ts` | Creates clinic with auto-slug generation (kebab-case + duplicate collision suffix), hashes owner password with `bcryptjs`, executes in Prisma `$transaction`. |
| `/api/specialists` | Route Handler | `src/app/api/specialists/route.ts` | `GET`: lists tenant specialists. `POST`: enforces 5 active specialists quota per clinic, checks duplicate email, hashes temporary password. |
| `/api/services` | Route Handler | `src/app/api/services/route.ts` | `GET`: lists tenant services with assigned specialists. `POST`: creates service. `PATCH`: toggles single specialist assignment. Uses `serializeDecimal`. |
| `/api/services/[id]/assign` | Route Handler | `src/app/api/services/[id]/assign/route.ts` | `PUT`: atomic transaction replacing all specialist assignments for a service. Verifies all specialist IDs belong to caller's clinic. |

---

## 2. UI Styling, Design System Tokens & Component Library

### Design Tokens & Tailwind Configuration (`tailwind.config.ts`)

- **Palette:** Custom `medical` teal palette extending default Tailwind colors:
  - `medical-50`: `#F0FDFA` (subtle tint backgrounds)
  - `medical-100`: `#CCFBF1` (light accent pills)
  - `medical-200`: `#99F6E4`
  - `medical-500`: `#14B8A6` (bright teal accents)
  - `medical-600`: `#0D9488` (primary brand color for buttons, active links, logos)
  - `medical-700`: `#0F766E` (hover & focus states)
  - `medical-800`: `#115E59`
  - `medical-900`: `#134E4A` (deep contrast headers)
  - `medical-950`: `#042F2E`
- **Neutrals:** Standard Tailwind `slate` scale (`slate-50` page background, `slate-100`/`slate-200` borders and input surfaces, `slate-600`/`slate-900` text).
- **Status Colors:**
  - Success/Active: `emerald-500`, `emerald-50`, `emerald-700`
  - Warning/Pending: `amber-500`, `amber-50`, `amber-700`
  - Danger/Inactive: `rose-500`, `rose-50`, `rose-700`
  - Info/Services: `sky-500`, `sky-50`, `sky-700`
- **Typography:** Inter, system-ui, -apple-system, BlinkMacSystemFont, sans-serif.
- **Physical Interaction (Apple Design):**
  - Buttons and interactive cards implement `active:scale-[0.98]` tactile press states.
  - Transitions configured with `transition-all duration-150 ease-out`.
  - Focus rings use `focus-visible:ring-2 focus-visible:ring-teal-500 focus-visible:ring-offset-2`.
- **Iconography:** `lucide-react` (v0.460.0). Strict rule against emojis: all visual indicators use Lucide SVG icons.
- **Toast Notifications:** `sonner` (v1.7.1) configured in `src/app/layout.tsx` with customized toast pill styling (`!rounded-2xl !border !border-slate-200 !shadow-lg !font-sans`).

### Component Library Inventory (`src/components/ui/`)

| Component | File Path | Props & Capabilities |
|---|---|---|
| `Button` | `src/components/ui/button.tsx` | Variants: `primary`, `secondary`, `outline`, `ghost`, `danger`. Sizes: `sm` (h-8), `md` (h-10), `lg` (h-12). Supports `isLoading` (renders spinning `Loader2`), `leftIcon`, `rightIcon`, `disabled`, and `active:scale-[0.98]`. |
| `Badge` | `src/components/ui/badge.tsx` | Variants: `success`, `warning`, `danger`, `info`, `neutral`. Optional `withDot` prop rendering pulsing status indicator. |
| `Card` | `src/components/ui/card.tsx` | Composed family: `Card`, `CardHeader`, `CardTitle`, `CardDescription`, `CardContent`, `CardFooter`. Optional `hoverEffect` with subtle lift and scale. |
| `Input` | `src/components/ui/input.tsx` | Text/password input with `label`, `error` message display, `helperText`, `leftIcon`, `rightIcon`. Focus states styled with teal ring. |
| `Modal` | `src/components/ui/modal.tsx` | Accessible dialog with backdrop blur (`backdrop-blur-sm`), escape key dismiss, document body scroll-locking, and width configurations (`sm` to `2xl`). |

### Component & Form Library Findings
- **Radix UI:** Not installed in `package.json`. Modals and interactive elements are pure React components built on Tailwind.
- **Form Libraries:** Neither `react-hook-form` nor `zod` are installed in `package.json`. Forms in `login`, `register`, `onboard-modal`, and `services` use React `useState`, controlled `<input>` elements, and native event handlers with Sonner feedback.

---

## 3. Existing `/dashboard` Navigation & `/dashboard/calendar` Plug-In Plan

### Current Shell Components

1. **Desktop Sidebar (`src/components/dashboard/sidebar.tsx`):**
   - Fixed width (`w-64 fixed inset-y-0 left-0 z-30`).
   - Hidden on mobile (`hidden lg:flex`), persistent on large viewports (`>=1024px`).
   - Navigation array (lines 34–53):
     ```typescript
     const navItems = [
       { label: 'Resumen General', href: '/dashboard', icon: LayoutDashboard, exact: true },
       { label: 'Equipo de Especialistas', href: '/dashboard/specialists', icon: Users, badge: '5' },
       { label: 'Catálogo de Servicios', href: '/dashboard/services', icon: Stethoscope, badge: '6' },
     ];
     ```
   - Renders clinic brand header at top, Phase 1 badge banner, active route highlighting (`bg-teal-50 text-teal-900 font-semibold`), user credentials footer, and `signOut` button.

2. **Mobile Drawer (`src/components/dashboard/mobile-nav.tsx`):**
   - Drawer container (`fixed inset-y-0 left-0 w-72 z-50 lg:hidden`).
   - Contains identical `navItems` array, closing on navigation link clicks or escape key.

3. **Sticky Header (`src/components/dashboard/header.tsx`):**
   - Sticky top bar (`sticky top-0 z-20 h-16 bg-white/80 backdrop-blur-md`).
   - Computes title via `getPageTitle(pathname)`:
     ```typescript
     const getPageTitle = (path: string) => {
       if (path === '/dashboard') return 'Resumen General';
       if (path.startsWith('/dashboard/specialists')) return 'Equipo de Especialistas';
       if (path.startsWith('/dashboard/services')) return 'Catálogo de Servicios';
       return 'Panel de Control';
     };
     ```
   - Features clinic live pill and public booking URL copy button (`https://appcitas.mx/${displaySlug}`).

4. **Dashboard Overview (`src/app/dashboard/page.tsx`):**
   - Metric 3 in KPI grid (lines 188–210) is currently an unlinked placeholder titled `"Citas de Hoy"` with subtitle `"Fase 2: Conexión Google Calendar"`.

### Exact Plug-In Specifications for `/dashboard/calendar`

To cleanly integrate Phase 2 calendar without breaking existing layouts:

1. **Add Nav Item to `sidebar.tsx` and `mobile-nav.tsx`:**
   ```typescript
   import { Calendar } from 'lucide-react';
   // In navItems:
   {
     label: 'Agenda y Calendario',
     href: '/dashboard/calendar',
     icon: Calendar,
   }
   ```
2. **Update Title Resolution in `header.tsx`:**
   ```typescript
   if (path.startsWith('/dashboard/calendar')) return 'Agenda y Calendario';
   ```
3. **Connect Metric 3 in `dashboard/page.tsx`:**
   Convert Metric 3 card into an interactive `<Link href="/dashboard/calendar">` with hover effects matching Metrics 1 and 2.
4. **Create Calendar Route (`src/app/dashboard/calendar/page.tsx`):**
   - Implement weekly/daily toggle.
   - Specialist filter bar with Google Calendar sync connection status badges.
   - Color-coded appointment blocks (`CONFIRMED`, `COMPLETED`, `CANCELLED`, `NO_SHOW`).
   - Slide-over drawer or modal for appointment detail review & status transitions.

---

## 4. Public Routing Capability & Dynamic `[slug]` Resolution

### Current State
- **Path `src/app/[slug]`:** Does not yet exist.
- **Middleware configuration (`src/middleware.ts`):**
  ```typescript
  export const config = {
    matcher: ['/dashboard/:path*', '/login', '/register'],
  };
  ```
  **Crucial architectural insight:** Middleware strictly intercepts only `/dashboard/*`, `/login`, and `/register`. Public routes matching `/[slug]` will bypass authentication checks without requiring middleware alterations.
- **Clinic Slug Resolution:**
  - `model Clinic` in `prisma/schema.prisma` enforces `@unique` on `slug`.
  - Registration API (`/api/clinic/register`) normalizes clinic names into clean kebab-case slugs and appends incremental numeric counters on collisions.
  - Pilot clinic seed sets `slug: "clinica-fisia"`.
  - NextAuth JWT token and session attach `clinicSlug` to the session user.
  - The dashboard header and overview page already construct the public URL format `https://appcitas.mx/${clinicSlug}`.

### Architecture for Phase 2 Public Booking Portal

```
src/app/[slug]/
└── page.tsx                      # Public Booking Flow at /[slug] (e.g. /clinica-fisia)

src/app/api/public/[slug]/
└── book/
    └── route.ts                  # Public Booking Endpoint (POST)
```

1. **`src/app/[slug]/page.tsx`:**
   - Dynamic App Router page receiving `{ params: { slug: string } }`.
   - Fetches tenant clinic data:
     ```typescript
     const clinic = await prisma.clinic.findUnique({
       where: { slug: params.slug },
       include: {
         services: { where: { isActive: true } },
         users: { where: { role: 'SPECIALIST', isActive: true } },
       },
     });
     if (!clinic) notFound();
     ```
   - 5-step booking flow:
     - Step 1: Select Service (shows price & duration).
     - Step 2: Select Specialist (or "Cualquier especialista disponible").
     - Step 3: Date & time slot picker powered by `src/lib/availability.ts`.
     - Step 4: Patient details (name, WhatsApp phone, notes).
     - Step 5: Confirmation screen with booking ID and calendar event summary.
2. **`src/app/api/public/[slug]/book/route.ts`:**
   - Public POST endpoint.
   - Resolves clinic by slug.
   - Upserts `Customer` record by clinicId and phone.
   - Creates `Appointment` with `status: CONFIRMED`.
   - Calls Google Calendar sync helper (or simulation mode fallback).

---

## 5. Existing Database Schema Alignment with Phase 2

Review of `prisma/schema.prisma` reveals that **Phase 2 schema requirements were proactively modeled in Phase 1**:

1. **Specialist Google Calendar Tokens:**
   `User` model contains lines 90–91:
   ```prisma
   googleAccessToken  String?
   googleRefreshToken String?
   ```
2. **Appointment Google Calendar Event ID:**
   `Appointment` model contains line 214:
   ```prisma
   googleEventId String?
   ```
3. **Appointment Status Enum:**
   `AppointmentStatus` enum (lines 24–30) contains:
   `PENDING`, `CONFIRMED`, `COMPLETED`, `CANCELLED`, `NO_SHOW`.
4. **Appointment Relationships:**
   `Appointment` belongs to `Clinic`, `User` (specialist), `Service`, `Customer`, and optional `Order`.
5. **No Migration Needed:**
   The Prisma schema already has all columns required for Phase 2. `npx prisma validate` passes cleanly.

---

## 6. Test Suite & Build Pipeline Audit

### Framework & Harness Configuration
- **No Vitest, Jest, or Playwright:** None of these testing frameworks are installed in `package.json`.
- **Standalone TypeScript Test Architecture:**
  - Test runner: `tests/e2e/run-tests.ts`.
  - Test harness: `tests/e2e/harness.ts` (typed assertion engine with colorized output, high-resolution timing, suite reports).
  - Tiers covered:
    - **Tier 1 (Feature Coverage):** 20 features (F01–F20), 158 assertions.
    - **Tier 2 (Boundary & Corner Cases):** 7 boundary suites (B01–B07), 138 assertions.
    - **Tier 3 (Cross Combinations):** 20 cross-feature clinical workflows, 64 assertions.
    - **Tier 4 (Clinical Workload Scenarios):** 6 end-to-end scenarios, 57 assertions.
- **In-Memory Mock Fallback:** When PostgreSQL is offline, test suites activate an in-memory relational simulation engine, allowing ultra-fast local validation.
- **Milestone Verification Scripts:**
  - `tests/verify-seed.ts`: Validates database seed idempotency and entity integrity (10/10 pass).
  - `tests/verify-m2.ts`: Validates NextAuth, credentials, route protection middleware, APIs, and quota limits (99/99 pass).
  - `tests/verify-m3.ts`: Validates UI components, dashboard shell, cards, modals, and design guidelines (59/59 pass).

### Execution Commands & Live Test Verification Results

| Command | Status | Output / Assertions | Execution Time |
|---|---|---|---|
| `npx tsx tests/e2e/run-tests.ts` | **PASS (Code 0)** | 53 suites, 417 assertions verified | 8 ms |
| `npx tsx tests/verify-seed.ts` | **PASS (Code 0)** | 10/10 seed & data assertions | 2 ms |
| `npx tsx tests/verify-m2.ts` | **PASS (Code 0)** | 99/99 auth, middleware, API assertions | 3 ms |
| `npx tsx tests/verify-m3.ts` | **PASS (Code 0)** | 59/59 UI and dashboard assertions | 2 ms |
| `npm run build` (`next build`) | **PASS (Code 0)** | 0 TS errors, 9 static pages compiled | ~9 s |
| `npm test` | **FAIL (Code 1)** | `npm error Missing script: "test"` | N/A |

### Critical Finding on `npm test`
In `package.json`, the `"scripts"` object is currently:
```json
"scripts": {
  "dev": "next dev",
  "build": "next build",
  "start": "next start",
  "prisma:validate": "prisma validate",
  "prisma:generate": "prisma generate",
  "prisma:seed": "tsx prisma/seed.ts"
}
```
There is no `"test"` script. Running `npm test` results in `npm error Missing script: "test"`.

**Recommendation for implementation phase:** Add the following script to `package.json`:
```json
"test": "tsx tests/e2e/run-tests.ts"
```

---

## 7. Recommended Implementation Blueprint for Phase 2

1. **Availability Engine (`src/lib/availability.ts`):**
   - Create slot computation engine taking `clinicId`, `specialistId` (or all), `serviceDurationMin` (~50 min), and date range.
   - Operating hours: Mon–Fri 09:00–19:00, Sat 09:00–14:00, Sun closed.
   - Subtract local DB `Appointment` records where `status != 'CANCELLED'`.
   - Subtract Google Calendar busy intervals (`freeBusy`).
   - Provide offline / simulated Google Calendar fallback mode when OAuth credentials are unset.

2. **Google Calendar Client (`src/lib/google-calendar.ts`):**
   - Wrapper using `googleapis` with OAuth2 client.
   - Methods: `getAuthUrl(specialistId)`, `handleCallback(code, specialistId)`, `getFreeBusy(specialistId, start, end)`, `createEvent(specialistId, appointment)`.
   - Graceful fallback simulation when `GOOGLE_CLIENT_ID` is missing in development.

3. **Dashboard Calendar Route (`src/app/dashboard/calendar/page.tsx`):**
   - Add navigation links in `sidebar.tsx` and `mobile-nav.tsx`.
   - Update title in `header.tsx`.
   - Build calendar view with weekly/daily toggle, specialist filter pills, Google sync indicators, and appointment detail drawer.

4. **Public Booking Route (`src/app/[slug]/page.tsx`):**
   - Dynamic tenant booking portal at `/[slug]`.
   - Service selection → Specialist selection → Date/time slot picker → Patient form → Confirmation.

5. **API Endpoints:**
   - `POST /api/public/[slug]/book`: Patient public booking.
   - `GET /api/public/[slug]/availability`: Bookable slot query.
   - `PATCH /api/appointments/[id]`: Status changes (`CONFIRMED`, `COMPLETED`, `CANCELLED`, `NO_SHOW`).
   - `GET /api/specialists/[id]/google/auth`: Initiates OAuth.
   - `GET /api/specialists/[id]/google/callback`: Saves access/refresh tokens.

6. **Test Suite Expansion:**
   - Add Phase 2 tests (`tests/e2e/phase2-calendar.test.ts`) covering slot calculation, Google Calendar mock fallback, public booking creation, and cancellation slot reclamation.
   - Register Phase 2 tests in `tests/e2e/run-tests.ts`.
   - Add `"test": "tsx tests/e2e/run-tests.ts"` to `package.json`.
