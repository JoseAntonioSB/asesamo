# Dispatch: Reviewer 1 (M1 — Google Calendar & Availability Engine)

## Context & Mission
You are Reviewer 1 for Milestone 1 (M1) of APPCITAS Phase 2.
Your mission is to perform an objective, rigorous code review of the backend availability engine (`src/lib/availability.ts`) and Google Calendar module (`src/lib/google-calendar.ts`).

## Inputs
- Authoritative requirements: /Users/jasbotello/devel/APPCITAS/.agents/ORIGINAL_REQUEST.md
- Project architecture: /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_orchestrator_phase2/PROJECT.md
- Milestone 1 Handoff: /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_worker_m1/handoff.md
- Source files:
  - `src/lib/availability.ts`
  - `src/lib/google-calendar.ts`
  - `prisma/schema.prisma`

## Tasks
1. Verify operating hours rules: Mon-Fri 09:00-19:00, Sat 09:00-14:00, Sunday closed.
2. Verify slot generation math: 60-min increments, ~50 min duration, clipping at clinic close time.
3. Verify collision logic: DB appointments (excluding CANCELLED) and Google Calendar busy spans.
4. Verify "Cualquier especialista disponible" aggregation and workload balancing.
5. Verify Google Calendar OAuth handling and resilient offline fallback mode.
6. Run `npx tsx tests/verify-m1-phase2.ts` and `npm run build` to confirm.

## Deliverable
Write your review report to `/Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_reviewer_m1_1/review_report.md` and your `handoff.md` with an explicit verdict: **APPROVE** or **REQUEST_CHANGES**. Send a completion message back.

## 2026-09-14T19:36:30Z
You are Reviewer 1 for Milestone 1 (M1) of APPCITAS Phase 2.
Your assigned working directory is: /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_reviewer_m1_1/
Read your dispatch instructions at /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_reviewer_m1_1/DISPATCH.md, the authoritative request at /Users/jasbotello/devel/APPCITAS/.agents/ORIGINAL_REQUEST.md, and M1 handoff at /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_worker_m1/handoff.md.
Review src/lib/availability.ts, src/lib/google-calendar.ts, and prisma/schema.prisma.
Run test and build commands, write your review_report.md and handoff.md with an explicit verdict (APPROVE or REQUEST_CHANGES), and send a completion message back.
