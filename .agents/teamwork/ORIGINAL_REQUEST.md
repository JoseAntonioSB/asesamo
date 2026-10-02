# Original User Request

## Initial Request — 2026-10-01T22:13:53Z

Build the foundational Phase 1 MVP of ASESAMO: an end-to-end industrial access control system with a Go REST API backend and a Next.js web application orchestrating the guard check-in flow with document compliance verification.

Working directory: /Users/jasbotello/devel/asesamo
Integrity mode: demo

Reference materials:
- Architecture & standards: `DEVELOPMENT_RULES.md`
- Database schema: `db/migrations/000001_init.up.sql`

## Requirements

### R1. Backend Service (`asesamo-api`)
Implement a Go backend following Clean Architecture and the rules defined in `DEVELOPMENT_RULES.md`. It must:
- Connect to PostgreSQL and apply the schema in `db/migrations/000001_init.up.sql`.
- Provide authentication (`POST /api/v1/auth/login`) returning a JWT for users.
- Provide access point check-in verification (`POST /api/v1/access/check-in`) that receives a worker UUID and access point ID, validates whether the worker and their documents are compliant according to the plant's rules, and records an immutable log in `access_logs` with a JSONB `document_snapshot`.
- Include unit and integration tests verifying both permitted and denied access conditions with tenant isolation.

### R2. Frontend Web Interface (`asesamo-web`)
Implement a Next.js (App Router) web application styled with Tailwind CSS and shadcn/ui. It must:
- Provide an authentication screen for security personnel (guard login).
- Provide a Caseta Check-in interface where the guard can submit or scan a worker's UUID.
- Display an immediate, clear visual status indicator (Green/Granted, Yellow/Warning, Red/Denied) showing the worker's name, contractor company, and document validation status summary.

### R3. Local Orchestration & Environment
Provide a root `docker-compose.yml` and environment configuration allowing the entire stack (PostgreSQL database, Go backend API, and Next.js web app) to be built and run locally with a single command (`docker compose up`).

## Acceptance Criteria

### Environment & Compilation
- [ ] Running `docker compose up --build` boots PostgreSQL, the Go backend API, and the Next.js frontend without fatal startup errors.
- [ ] The database migration `000001_init.up.sql` applies successfully on container initialization.

### Backend API Verification
- [ ] Authentication endpoint generates a valid JWT when provided valid guard credentials.
- [ ] Calling the check-in endpoint with a worker whose required documents are valid returns `granted` and inserts a corresponding record in `access_logs`.
- [ ] Calling the check-in endpoint with a worker having missing or expired mandatory documents evaluates against plant policy, returning `denied` (or warning according to policy) with the denial reason stored in `access_logs`.
- [ ] Automated tests run via `go test ./...` pass with zero failures.

### End-to-End User Flow
- [ ] A guard can sign in through the Next.js UI using test credentials.
- [ ] From the Caseta UI, submitting a test worker UUID executes the check-in and renders the granted/denied compliance result card in the browser.
