# Dispatch: Spec Miner Survey (Phase 2 Detailed Requirements Mining)

## Context & Mission
You are a Spec Miner analyzing the authoritative requirements and specification source of truth for APPCITAS Phase 2.
Your goal is to extract exhaustive requirements, constraints, edge cases, data flows, and acceptance criteria for:
- R1: Google Calendar Integration & Availability Engine (`src/lib/availability.ts`, OAuth, tokens, freeBusy, offline fallback).
- R2: Interactive Clinic Calendar & Agenda Dashboard (`/dashboard/calendar`, week/day views, color codes, filters, appointment drawer, sync indicator).
- R3: Public Booking Portal (`/[slug]`, 5-step responsive flow, mobile-first, /api/public/[slug]/book, cancellation release).

## Inputs
- Authoritative requirements: /Users/jasbotello/devel/APPCITAS/.agents/ORIGINAL_REQUEST.md (specifically under `## 2026-09-14T18:47:27Z` and previous headers)
- Project root: /Users/jasbotello/devel/APPCITAS

## Tasks
1. Extract all functional requirements for R1, R2, and R3.
2. Formulate explicit availability rules:
   - Operating hours: Mon-Fri 09:00-19:00, Sat 09:00-14:00, Sunday closed.
   - Slot duration: default service duration (~50 min, configurable per service).
   - Buffer times / step intervals (e.g. 30 or 60 min intervals).
   - Collision detection: local DB appointments (excluding CANCELLED) + Google Calendar `busy` spans.
   - Specialist assignment: specific specialist vs "Cualquier especialista disponible" (least-busy or first available).
   - Google Calendar offline / fallback mode: deterministic mock / simulation when GOOGLE_CLIENT_ID / secret is not set.
3. Formulate UI/UX requirements for R2 (Agenda Dashboard) and R3 (Public Booking Portal).
4. Formulate API endpoints specifications (`/api/auth/google/...`, `/api/availability`, `/api/public/[slug]/book`, `/api/appointments/[id]`, etc.).
5. Document edge cases: timezone handling, conflicting bookings (race conditions / transactions), expired OAuth tokens, non-existent slugs.

## Deliverable
Write your comprehensive specification report to:
`/Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_spec_miner_survey/spec_report.md`
And write your `handoff.md` and `progress.md` in:
`/Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_spec_miner_survey/`
Then send a completion message back to your orchestrator.

## 2026-09-14T18:56:13Z
You are a Spec Miner analyzing the authoritative requirements and specification source of truth for APPCITAS Phase 2.
Your assigned working directory is: /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_spec_miner_survey/
Read your dispatch file at /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_spec_miner_survey/DISPATCH.md and the authoritative request at /Users/jasbotello/devel/APPCITAS/.agents/ORIGINAL_REQUEST.md.
Extract all requirements, availability calculation rules (clinic hours Mon-Fri 9-19 Sat 9-14, slot duration ~50min, freeBusy collision, offline fallback), UI/UX specs for /dashboard/calendar and /[slug], and API endpoint contracts.
Write your findings to /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_spec_miner_survey/spec_report.md and your handoff.md, then send a completion message back to the orchestrator.
