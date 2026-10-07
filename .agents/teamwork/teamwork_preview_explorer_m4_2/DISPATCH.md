# Dispatch Assignment: M4 Explorer 2 (Auth Flow, API Client & State)
Timestamp: 2026-10-07T16:25:00Z
Agent: teamwork_preview_explorer_m4_2
Working Directory: /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m4_2

## Objective
Analyze and formulate the authentication architecture, API client layer, and route protection for `asesamo-web/`.

## Authoritative Inputs
- /Users/jasbotello/devel/asesamo/.agents/teamwork/ORIGINAL_REQUEST.md
- /Users/jasbotello/devel/asesamo/PROJECT.md (Milestone 4, Authentication & HTTP envelopes)
- /Users/jasbotello/devel/asesamo/DEVELOPMENT_RULES.md
- `asesamo-api` contracts: `POST /api/v1/auth/login`, `GET /api/v1/access-points`, `POST /api/v1/access/check-in`, `POST /api/v1/access/check-out`

## Scope of Analysis
1. API Client Layer (`lib/api.ts` or `lib/api-client.ts`):
   - Base URL configuration via `NEXT_PUBLIC_API_URL` (defaulting to `http://localhost:8080`).
   - Standard envelope unwrapping: `{ data: T, meta?: M }` and typed error handling: `{ error: { code: string, message: string } }`.
   - Handling Authorization headers (`Bearer <token>`).
   - Methods: `login(email, password)`, `getAccessPoints()`, `checkIn(workerId, accessPointId)`, `checkOut(workerId, accessPointId)`.
2. Guard Authentication (`app/(auth)/login/page.tsx` or `app/login/page.tsx`):
   - Form for `email` and `password` with default helper/placeholder `guard@asesamo.com` / `Password123!`.
   - Client-side error banner showing backend Spanish messages or network errors.
   - Storage of JWT token (HTTP-only cookie or client cookie / localStorage) with redirection to `/caseta`.
3. Route Protection & Auth State (`middleware.ts` or auth provider context):
   - Protect `/caseta` route (redirect unauthenticated guards to `/login`).
   - Redirect authenticated guards from `/login` to `/caseta`.
   - Logout functionality (clearing token and redirecting to `/login`).
4. Deliver concrete code blueprints and error handling contracts.

## Output Requirements
Write your detailed analysis to `analysis.md` and executive summary to `handoff.md` in your working directory. Call send_message when done.
