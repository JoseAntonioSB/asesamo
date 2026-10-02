# Handoff Report: Phase 2 Architecture & Codebase Survey
**Agent**: teamwork_preview_explorer_survey_1 (Explorer)  
**Recipient**: parent (Orchestrator, ID: `5af02f98-2cc9-400b-b4e1-67cc4f848f43`)  
**Timestamp**: 2026-09-14T18:59:00Z  
**Type**: Hard Handoff (Investigation Complete)

---

## 1. Observation

1. **Missing `googleapis` Dependency**:
   - Inspected `package.json` lines 16-27:
     ```json
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
     }
     ```
     `googleapis` is absent. `date-fns` and `dayjs` are also absent.
2. **Existing Google OAuth Fields in Prisma Schema**:
   - `prisma/schema.prisma` lines 90-91 in model `User`:
     ```prisma
     googleAccessToken  String?
     googleRefreshToken String?
     ```
   - `prisma/schema.prisma` line 214 in model `Appointment`:
     ```prisma
     googleEventId String?
     ```
   - `prisma/schema.prisma` lines 24-30:
     ```prisma
     enum AppointmentStatus {
       PENDING
       CONFIRMED
       COMPLETED
       CANCELLED
       NO_SHOW
     }
     ```
   - Currently absent from `User`: `googleTokenExpiry`, `googleCalendarId`, `googleEmail`.
   - `Appointment` records appointment start time as `dateTime DateTime` (line 211).
3. **Database Environment & Offline Test Harness**:
   - In `.env` line 1: `DATABASE_URL="postgresql://fisia_user:fisia_secret_2026@localhost:5432/appcitas?schema=public"`.
   - Port check `nc -z -v -w 2 localhost 5432` confirms PostgreSQL is offline in this sandbox environment.
   - `npx prisma validate` returns exit code 0 (`The schema at prisma/schema.prisma is valid 🚀`).
   - `npx prisma generate` compiles Prisma client in 95ms.
   - `npx tsx tests/verify-seed.ts` executed with exit code 0, verifying 10/10 checks across in-memory relational store.
   - `npx tsx tests/e2e/run-tests.ts` executed with exit code 0, validating 417 assertions across 53 test suites in 8ms.
   - `npm run build` completed with exit code 0, compiling 12 Next.js App Router routes with 0 TypeScript/packaging errors.
4. **Current Dashboard Navigation Structure**:
   - `src/components/dashboard/sidebar.tsx` lines 34-53 and `mobile-nav.tsx` lines 61-80 define 3 navigation links:
     - `/dashboard` (Resumen General)
     - `/dashboard/specialists` (Equipo de Especialistas)
     - `/dashboard/services` (Catálogo de Servicios)
   - `/dashboard/calendar` is not yet present in navigation.
5. **Route Interception in Middleware**:
   - `src/middleware.ts` lines 28-30:
     ```typescript
     export const config = {
       matcher: ['/dashboard/:path*', '/login', '/register'],
     };
     ```
     Public routes `/[slug]`, `/api/public/:path*`, and `/api/appointments/:path*` are not intercepted by auth middleware.

---

## 2. Logic Chain

1. **Dependency Installation**:
   - *From Observation 1*: `ORIGINAL_REQUEST.md` §R1 explicitly demands Google Calendar synchronization via `googleapis`.
   - *Inference*: `googleapis` must be added to `package.json` dependencies. Because `googleapis` ships with comprehensive internal typings, no `@types/googleapis` devDependency is needed.
2. **Schema Extension & Backward Compatibility**:
   - *From Observation 2*: `googleAccessToken`, `googleRefreshToken`, and `googleEventId` are already defined on `User` and `Appointment`.
   - *Inference*: Phase 2 requirements mandate token refresh management and UI sync status indication. Adding `googleTokenExpiry DateTime?`, `googleCalendarId String?`, and `googleEmail String?` to `User` directly fulfills this. Because all three additions are nullable, they are 100% non-breaking for existing seeds and queries.
3. **Simulation Mode Architecture for Resilient Offline Operation**:
   - *From Observation 3*: PostgreSQL and external network connections to Google Cloud are restricted/offline in local development and automated CI.
   - *Inference*: The availability and Google Calendar modules (`src/lib/google-calendar.ts` and `src/lib/availability.ts`) must support an integrated fallback/simulation mode when `GOOGLE_CLIENT_ID` or `GOOGLE_CLIENT_SECRET` are not configured. This ensures that offline tests, slot calculations, and public booking flows can execute without failure.
4. **Endpoint & UI Architecture**:
   - *From Observations 4 & 5*: Route handlers for public booking (`/api/public/[slug]/availability`, `/api/public/[slug]/book`) and appointment status updates (`/api/appointments/[id]`) will operate outside auth middleware, while `/dashboard/calendar` will be automatically protected. Adding `/dashboard/calendar` to `sidebar.tsx` and `mobile-nav.tsx` completes full dashboard discoverability.

---

## 3. Caveats

- **External Google Credentials**: Real OAuth against Google's live servers requires valid `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` issued in the Google Cloud Console with configured redirect URIs. The implementation must strictly adhere to the simulation fallback mode so local development and automated tests pass without credentials.
- **Timezones**: Physiotherapy appointments in Mexico typically operate under Mexican Central Time (America/Monterrey or America/Mexico_City). Native ISO strings (`toISOString()`) with explicit clinic operating hours (Mon-Fri 09:00-19:00, Sat 09:00-14:00) should be formatted consistently in UTC/local offsets.
- **No Other Caveats**: All models, seed data, and routes required for Phase 1 are fully validated and documented.

---

## 4. Conclusion

The APPCITAS codebase is exceptionally well architected and 100% prepared for Phase 2 implementation. The existing schema already has the core Google token fields, the test harness is extremely fast and robust, and the path forward requires:
1. Installing `googleapis`.
2. Augmenting `prisma/schema.prisma` with `googleTokenExpiry`, `googleCalendarId`, and `googleEmail`.
3. Implementing `src/lib/google-calendar.ts` (with simulation mode fallback) and `src/lib/availability.ts`.
4. Implementing the 6 Phase 2 API route handlers and `/dashboard/calendar` + `/[slug]` frontend pages.

A full technical report has been generated at:
`/Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_explorer_survey_1/survey_report.md`

---

## 5. Verification Method

To independently verify all observations in this report:

1. **Verify Prisma Schema**:
   ```bash
   npx prisma validate
   ```
   *Expected result*: Exit code 0, "The schema at prisma/schema.prisma is valid".
2. **Verify Phase 1 E2E Test Suite**:
   ```bash
   npx tsx tests/e2e/run-tests.ts
   ```
   *Expected result*: Exit code 0, 53 test suites passed, 417 assertions passed.
3. **Verify Seed Execution & Data Model Relationships**:
   ```bash
   npx tsx tests/verify-seed.ts
   ```
   *Expected result*: Exit code 0, all 10 verifications passed.
4. **Verify Clean Production Build**:
   ```bash
   npm run build
   ```
   *Expected result*: Exit code 0, Next.js compiles 12 routes with zero type or packaging errors.
