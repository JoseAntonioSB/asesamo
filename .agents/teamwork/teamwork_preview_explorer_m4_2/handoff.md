# Handoff Report: Milestone M4 — Auth Flow, API Client Layer & Route Protection

**Agent:** `teamwork_preview_explorer_m4_2`  
**Milestone:** M4  
**Date:** 2026-10-07  
**Status:** Hard Handoff (Investigation & Specification Complete)

---

## 1. Observation

1. **Backend Endpoint Contracts in `asesamo-api/internal/adapter/http/router.go`**:
   - `POST /api/v1/auth/login` (Line 55): Public endpoint guarded by 10 req/min IP rate limiter (`httprate.LimitBy(10, 1*time.Minute)`). Handled by `AuthHandler.Login`.
   - `GET /api/v1/access-points` (Line 62): Protected by `customMiddleware.Auth(cfg.JWTSecret)`. Handled by `AccessPointHandler.List`.
   - `POST /api/v1/access/check-in` (Line 66): Protected by JWT auth. Handled by `AccessHandler.CheckIn`.
   - `POST /api/v1/access/check-out` (Line 67): Protected by JWT auth. Handled by `AccessHandler.CheckOut`.
   - `GET /api/v1/access-logs` (Line 68): Protected by JWT auth. Handled by `AccessHandler.ListLogs`.

2. **Backend Envelopes in `asesamo-api/internal/adapter/http/dto/envelope.go`**:
   - Success envelope (Line 13–16):
     ```json
     { "data": <payload>, "meta": <meta_object_optional> }
     ```
   - Error envelope (Line 24–27):
     ```json
     { "error": { "code": "UPPER_SNAKE_CASE", "message": "Mensaje en español" } }
     ```

3. **Backend DTOs and Payloads**:
   - `auth_dto.go` (Line 5–24): `LoginRequest{ email, password }`, returning `LoginResponseData{ token, user: UserProfileResponse{ id, tenant_id, email, full_name, role } }`.
   - `access_dto.go` (Line 12–17): `CheckInRequest{ worker_id, worker_user_id, access_point_id, notes }`. Accepts either `worker_id` or `worker_user_id`.
   - `access_point_dto.go` (Line 6–13): `AccessPointResponse{ id, plant_id, tenant_id, name, description, is_active }`.
   - `access/check_in.go` (Line 27–38): `CheckInOutput` returns `access_log_id`, `result` (`granted` | `granted_with_warning` | `denied`), `denial_reason`, `checked_in_at`, `worker`, `contractor`, `plant`, `access_point`, `document_snapshot`, `documents`.

4. **Seeded Test Credentials in `db/seeds/000001_seed_dev_data.sql` (Line 83–101)**:
   - Email: `guard@asesamo.com`
   - Password: `Password123!`
   - User ID: `44444444-4444-4444-4444-444444444444`
   - Tenant ID: `11111111-1111-1111-1111-111111111111`
   - Role: `guard`
   - Full Name: `Oficial Roberto Caseta`
   - Access Point ID: `33333333-3333-3333-3333-333333333331` (`Caseta Principal Norte`)

5. **Environment Configuration in `docker-compose.yml` (Line 115, 126)**:
   - `NEXT_PUBLIC_API_URL: ${NEXT_PUBLIC_API_URL:-http://localhost:8080/api/v1}`
   - Fallback specified in dispatch: `http://localhost:8080`.

---

## 2. Logic Chain

1. **Base URL Normalization**:
   - *Observation*: Some environments set `NEXT_PUBLIC_API_URL` to `http://localhost:8080`, while `docker-compose.yml` sets `http://localhost:8080/api/v1`.
   - *Inference*: A hardcoded concatenation of `/api/v1` would cause double prefixing (`.../api/v1/api/v1`) in Docker, while omitting it would break local dev if only `http://localhost:8080` is passed.
   - *Solution*: `normalizeApiUrl` strips trailing slashes and appends `/api/v1` only if it does not already end with `/api/v1`.

2. **Envelope Unwrapping & Error Handling**:
   - *Observation*: Every backend response is either `{ data: T, meta?: M }` or `{ error: { code, message } }`.
   - *Inference*: The frontend HTTP client must intercept responses, unwrapping `body.data` on `response.ok`, and constructing a typed `ApiError` with `code` and Spanish `message` on `!response.ok`.
   - *Solution*: `ApiClient.request<T>()` inspects `body.error`, maps status codes, and throws `ApiError`. On 401 Unauthorized for protected endpoints, it triggers session cleanup and redirects to `/login?error=session_expired`.

3. **Route Protection via Edge Middleware + Cookie Storage**:
   - *Observation*: Next.js App Router Edge Middleware (`middleware.ts`) runs on the server before client hydration and cannot access `localStorage`.
   - *Inference*: Storing the JWT purely in `localStorage` makes server-side / Edge route protection impossible without layout flash or hydration lag.
   - *Solution*: Dual storage strategy in `lib/auth.ts`: `asesamo_token` is written to both a browser cookie (`SameSite=Lax`, `path=/`) and `localStorage`. `middleware.ts` reads `request.cookies.get('asesamo_token')` to redirect unauthenticated users from `/caseta` to `/login`, and authenticated users from `/login` to `/caseta`.

4. **Guard Authentication Screen & Demo UX**:
   - *Observation*: Acceptance criteria require test guards to sign in easily with `guard@asesamo.com` / `Password123!`.
   - *Inference*: Security guards on industrial tablets or testing evaluators benefit from zero-friction demo credential autofill.
   - *Solution*: `app/login/page.tsx` features an interactive "Autocompletar" button populating both fields, client validation preventing malformed email requests, and a prominent Spanish alert banner displaying backend errors (`INVALID_CREDENTIALS`, `TOO_MANY_REQUESTS`, or connection failures).

5. **Cross-Explorer Integration**:
   - *Observation*: Explorer M4_1 is designing UI components (`button`, `card`, `input`, etc.) and Explorer M4_3 is designing `app/caseta/page.tsx` and check-in scanner inputs.
   - *Inference*: Explorer M4_3 requires typed methods `getAccessPoints()`, `checkIn()`, and `checkOut()`, while the header requires `useAuth()` to show the active guard name (`Oficial Roberto Caseta`) and allow logout.
   - *Solution*: `lib/api/client.ts` and `lib/api.ts` export clean async functions matching M4_3's exact needs, and `contexts/auth-context.tsx` provides the reactive `useAuth()` hook.

---

## 3. Caveats

1. **HTTPS in Production**: The cookie helper sets the `Secure` attribute only when `window.location.protocol === 'https:'`. In local development (`http://localhost:3000`), it omits `Secure` so cookies persist properly over plain HTTP.
2. **Token Refresh Rotation**: Phase 1 MVP backend uses 1-hour access tokens without an active `/auth/refresh` endpoint in Chi router. Session expiration gracefully logs the guard out and displays a Spanish warning banner. Refresh token rotation is scheduled for Phase 2.
3. **No Code Modification Permitted**: As an Explorer agent, no source code files in `asesamo-web/` were modified. Complete specifications and drop-in code blueprints are provided in `analysis.md`.

---

## 4. Conclusion

The authentication architecture, API client layer, and route protection for `asesamo-web/` are fully resolved, typed, and specified. The specifications provide:
- `lib/api/types.ts`: Exhaustive TypeScript definitions matching `asesamo-api` contracts.
- `lib/api/client.ts`: Resilient `ApiClient` class with URL normalization, envelope unwrapping, bearer injection, and error translation.
- `lib/api.ts`: Convenient façade allowing `@/lib/api` imports.
- `lib/auth.ts`: Dual cookie/localStorage token persistence helper.
- `contexts/auth-context.tsx`: React Context and `useAuth()` hook for user state and logout.
- `middleware.ts`: Next.js Edge route guard protecting `/caseta` and handling unauthenticated/authenticated redirects.
- `app/login/page.tsx`: Industrial guard login page with demo auto-complete, validation, and Spanish error alerts.
- `components/auth/ProtectedRoute.tsx`: Client fallback route guard wrapper.

All specifications are documented in `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m4_2/analysis.md`.

---

## 5. Verification Method

1. **Contract Invariant Check**:
   - Inspect `asesamo-api/internal/adapter/http/router.go` and confirm endpoints `/api/v1/auth/login`, `/api/v1/access-points`, `/api/v1/access/check-in`, `/api/v1/access/check-out`, `/api/v1/access-logs` match the routes called in `lib/api/client.ts`.
2. **DTO & Error Alignment**:
   - Inspect `asesamo-api/internal/core/domain/errors/errors.go` and ensure `INVALID_CREDENTIALS`, `TOO_MANY_REQUESTS`, and `USER_INACTIVE` correspond to the error handling in `app/login/page.tsx`.
3. **Deterministic Credentials Verification**:
   - Inspect `db/seeds/000001_seed_dev_data.sql` lines 83–101 and verify email `guard@asesamo.com` and password `Password123!` match the Quick-Fill helper in `app/login/page.tsx`.
4. **End-to-End Implementation Readiness**:
   - Review `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m4_2/analysis.md`. All code blocks are self-contained, typed, and ready for immediate implementation by Builder agents.
