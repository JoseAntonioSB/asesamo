# Scope: Phase 2 — La Agenda y Google Calendar

## Current Objective
Deliver the complete, production-ready implementation of Phase 2 for APPCITAS:
1. Google Calendar Integration & Availability Engine (`src/lib/google-calendar.ts`, `src/lib/availability.ts`, API endpoints).
2. Interactive Clinic Calendar & Agenda Dashboard (`/dashboard/calendar` with daily/weekly views, filters, appointment drawer, sync badges).
3. Public Booking Portal (`/[slug]` 5-step responsive flow, slot booking, customer deduplication, WhatsApp confirmation).
4. End-to-end Verification & Build passing 100% tests with 0 errors.

## Milestones
| # | Name | Scope | Dependencies | Status |
|---|------|-------|-------------|--------|
| M1 | Google Calendar & Availability Engine | `googleapis` installation, schema extension, `src/lib/google-calendar.ts`, `src/lib/availability.ts`, `/api/availability`, `/api/public/[slug]/book`, `/api/appointments/[id]`, `/api/auth/google/*` | none | IN_PROGRESS |
| M2 | Interactive Clinic Calendar Dashboard | `/dashboard/calendar`, weekly/daily views, specialist filters, appointment drawer, status updater, Google sync badges, navigation links | M1 | PLANNED |
| M3 | Public Booking Portal (`/[slug]`) | Responsive 5-step booking flow at `src/app/[slug]/page.tsx`, booking wizard, customer form, confirmation card | M1 | PLANNED |
| M4 | Comprehensive E2E Verification & Build | Full E2E test execution (Tiers 1-4), adversarial test hardening (Tier 5), zero build errors (`npm run build`), forensic audit | M1, M2, M3 | PLANNED |

## Parallel Dual Track: E2E Testing Track
- `E2E Testing Track` runs concurrently: test suite design, requirement-based opaque-box tests across 4 Tiers, culminating in `TEST_READY.md`.
