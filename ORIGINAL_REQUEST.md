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


## Follow-up — 2026-10-15T16:07:31Z

Build the complete **Permisos de Trabajo de Alto Riesgo (PTAR Digital)** module for ASESAMO: an end-to-end capability enabling contractors and EHS supervisors to issue, validate, authorize, and monitor high-risk work permits in real-time, enforcing automated Mexican safety compliance locks (STPS: NOM-009, NOM-033, NOM-027, NOM-029).

Working directory: /Users/jasbotello/devel/asesamo
Integrity mode: development

Reference materials:
- Architecture & standards: `DEVELOPMENT_RULES.md`
- Existing database migrations: `db/migrations/`
- Existing backend: `asesamo-api` (Go Clean Architecture, PostgreSQL, Chi router)
- Existing frontend: `asesamo-web` (Next.js 14 App Router, Tailwind CSS, shadcn/ui)

## Requirements

### R1. Database Schema & Migration (`db/migrations/000003_create_work_permits.*.sql`)
Create the PostgreSQL migration for PTAR:
- `work_permits`: `id` (UUID), `tenant_id`, `permit_number` (e.g. `PTAR-2026-0001`), `permit_type` (`heights`, `confined_space`, `hot_work`, `electrical_loto`), `status` (`draft`, `pending_approval`, `approved_active`, `suspended`, `closed`, `cancelled`), `contractor_id`, `work_area`/`location`, `description`, `hazard_description`, `start_time`, `end_time`, `created_by_user_id`, `approved_by_user_id`, timestamps.
- `permit_workers`: junction table linking `permit_id` and `worker_id` with compliance status snapshot at assignment time.
- `permit_checklists`: JSONB or relational items detailing mandatory safety verifications per NOM (e.g., atmospheric test for confined spaces, fire extinguisher in place for hot work, harness inspection for heights, zero energy verification for LOTO).
- `permit_signatures`: digital sign-off record (role: `requester`, `contractor_supervisor`, `ehs_authorizer`, timestamp, signature metadata/notes).

### R2. Backend Service Domain & API (`asesamo-api`)
Implement following Clean Architecture:
- **Candado Inteligente (Compliance Engine):** When attaching workers or submitting permit for approval, verify:
  1. Worker has approved & active DC-3 corresponding to the hazard type (e.g., DC-3 Alturas for `heights`, DC-3 Espacios Confinados for `confined_space`).
  2. Worker has valid medical fitness certificate (`pending` or `expired` blocks issuance).
  3. Contractor company is active and not suspended.
- **Endpoints:**
  - `POST /api/v1/permits`: Create new PTAR draft with metadata, work area, hazard type, and schedule.
  - `GET /api/v1/permits`: List permits with filters (`status`, `permit_type`, `contractor_id`, date range).
  - `GET /api/v1/permits/:id`: Detailed permit view with assigned workers, compliance status check, checklist items, and signatures.
  - `POST /api/v1/permits/:id/workers`: Assign/remove workers (triggers Candado Inteligente verification).
  - `PATCH /api/v1/permits/:id/checklist`: Update safety verification checklist items.
  - `POST /api/v1/permits/:id/approve`: EHS digital authorization (validates all checklist items complete and worker compliance satisfied).
  - `POST /api/v1/permits/:id/status`: Transition status (`close`, `suspend`, `cancel`) with reason and audit timestamp.
- Strict tenant isolation and localized Spanish error messages. Unit tests for permit state transitions and compliance validation gate.

### R3. Frontend PTAR Management UI (`asesamo-web`)
Implement modern, responsive interface under `/permisos` (or `/ptar`):
- **Permit Dashboard & List:** Filterable grid/table by hazard badge (`Alturas NOM-009`, `Espacios Confinados NOM-033`, `Corte y Soldadura NOM-027`, `LOTO NOM-029`), status badge (Borrador, Pendiente EHS, Activo en Campo, Cerrado), contractor, and validity window.
- **PTAR Creation Wizard / Form:**
  - Step 1: Work details (type, area, schedule, description).
  - Step 2: Crew assignment with instant visual compliance flags (Green badge for validated DC-3/Medical, Red alert blocking assignment if non-compliant with reason).
  - Step 3: Interactive STPS checklist dynamically loaded based on permit type.
- **EHS Authorization & Execution Modal/Page:**
  - Fast review of checklist verification, atmospheric/safety measurements, and one-click EHS digital approval with status badge switch to "Activo en Campo".
  - Field status controls: Suspend (emergency/weather halt) and Close (work completed, area cleaned).

## Acceptance Criteria

### Database & Backend Verification
- [ ] Migration `000003_create_work_permits.up.sql` runs cleanly without schema conflicts.
- [ ] Attempting to approve a permit or assign a worker who lacks active DC-3 or valid medical cert fails with HTTP 422 and clear Spanish explanation ("El trabajador no cuenta con DC-3 vigente para trabajos en alturas", etc.).
- [ ] EHS authorization successfully transitions permit from `pending_approval` to `approved_active` and stores authorizer ID and timestamp.
- [ ] All Go tests (`go test ./...`) pass including new unit tests for the permit service and compliance checks.

### Frontend Flow Verification
- [ ] A contractor supervisor can create a new PTAR draft and select required workers.
- [ ] The crew selection screen visually flags compliant vs non-compliant workers with reasons (DC-3 missing/expired).
- [ ] An EHS user can review the checklist, verify compliance, and approve the permit.
- [ ] The permit appears in real-time as "Activo en Campo" with countdown / expiration time indicator.
- [ ] TypeScript and frontend build (`npm run build`) complete without errors.
