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


## Follow-up — 2026-10-08T14:58:25Z

notifica a tu padre lo que se hizo


## Follow-up — 2026-10-08T21:40:19Z

Build the complete Document Upload module for ASESAMO: an end-to-end capability allowing contractors and workers to upload compliance documents (DC-3, REPSE, SUA, INE, Medical) with metadata, persisted via an abstract storage interface (local disk volume in development, ready for Cloudflare R2/S3 in production), integrated into the Go backend API and Next.js web application.

Working directory: /Users/jasbotello/devel/asesamo
Integrity mode: demo

Reference materials:
- Architecture & standards: `DEVELOPMENT_RULES.md`
- Database schema: `db/migrations/000001_init.up.sql`
- Existing backend: `asesamo-api`
- Existing frontend: `asesamo-web`

## Requirements

### R1. Backend Storage Adapter & Clean Architecture Domain
Implement a file storage abstraction in `asesamo-api` following Clean Architecture:
- Define a domain `StorageService` interface (`Upload`, `GetURL`/`Download`, `Delete`).
- Provide a `LocalStorageService` implementation saving files to a designated directory/volume with sanitized file keys conforming to `{tenant_id_or_global}/{entity_type}/{id}/{filename_uuid}.{ext}`.
- Enforce file constraints: maximum size 10 MB, allowed MIME types (`application/pdf`, `image/jpeg`, `image/png`). Validate actual magic bytes, not just file extensions.

### R2. Document Upload & Management Endpoints
Implement REST endpoints in `asesamo-api` under authenticated routes:
- `POST /api/v1/documents/upload`: accepts `multipart/form-data` containing the file, `document_type_code`, `issued_at`, and `expires_at`. Associates the document with the authenticated worker (`user_id`) or contractor (`contractor_id`), saves the file via `StorageService`, and inserts a record into the `documents` table with initial status `pending_review`.
- `GET /api/v1/documents/my-documents`: returns list of uploaded documents and their review status for the authenticated user/contractor.
- `GET /api/v1/documents/:id/download`: serves or streams the file content securely.
- Ensure strict tenant isolation and error handling with Spanish user-facing messages.
- Provide comprehensive unit and integration tests for storage, upload validation, and failure conditions.

### R3. Frontend Document Upload UI
Implement the user-facing upload workflow in `asesamo-web`:
- Provide a responsive Document Management & Upload view allowing workers or contractors to view required documents and current status (`pending_review`, `approved`, `rejected`, `missing`).
- Provide an accessible upload dialog/form using shadcn/ui and Tailwind CSS with document type selection, date pickers for issued/expiration dates, drag-and-drop / file selector, and upload progress/status indicator with toast notifications (`ask-sonner`).
- Display clear client-side validation for file types and sizes before upload.

## Acceptance Criteria

### Storage & Backend Verification
- [ ] Uploading a valid PDF or JPG file via `POST /api/v1/documents/upload` succeeds with HTTP 201/200, saves the file to disk, and creates a row in `documents` with status `pending_review`.
- [ ] Attempting to upload a file exceeding 10 MB or with an unauthorized MIME type (e.g., `.exe`, `.txt`) is rejected with HTTP 400/422 and a Spanish error message without corrupting storage.
- [ ] `GET /api/v1/documents/my-documents` returns the uploaded documents with their status and metadata in the standard `{ "data": ... }` envelope.
- [ ] All Go tests across `asesamo-api` (`go test ./...`) pass with zero failures.

### Frontend Web Flow Verification
- [ ] An authenticated worker or contractor can navigate to the document management screen in `asesamo-web`.
- [ ] Submitting a file through the upload UI uploads the document, displays success feedback via toast, and refreshes the document list showing the item as "Pendiente de revisión".
