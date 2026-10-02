# Handoff Report — Frontend Specification Survey (`teamwork_preview_spec_miner_survey_2`)

## 1. Observation
1. **Authoritative Project Constraints (`DEVELOPMENT_RULES.md`):**
   - Line 27: Framework is **Next.js 14+** with **App Router** (Server Components by default).
   - Line 29-30: Styling with **Tailwind CSS**; UI components using **shadcn/ui** (on Radix UI primitives).
   - Line 31: Async/state management with **TanStack Query** (React Query v5).
   - Line 41: Repositories separated: `asesamo-api` (Go) and `asesamo-web` (Next.js).
   - Line 137: TypeScript strict: `strict: true` in `tsconfig.json`. No explicit `any`.
   - Line 139: Component naming: one component per file, filename matches component (`ContractorCard.tsx`).
   - Line 140: Server vs Client: default Server Component, add `"use client"` only when browser events, state, or hooks are needed.
   - Line 141: Fetching: all API calls pass through `lib/api/` encapsulating base URL and JWT token. Components never do raw `fetch`.
   - Line 147: User-facing error messages in the UI must be in **Spanish**.
   - Line 177-184: Standard envelope structure: `{ "data": ... }` for success, `{ "error": { "code": "...", "message": "..." } }` for error.
   - Line 194-195: JWT claims include `sub`, `tenant_id`, `role`, `exp` (1 hour expiration).
   - Line 202: Role `guard` is the gate guard (only access module).
   - Line 356-359: Mandatory UI skills: `ui-styling` (shadcn/ui + Tailwind + a11y), `ask-sonner` (Sonner toast notifications).

2. **Database Schema & Data Model (`db/migrations/000001_init.up.sql`):**
   - Lines 77-87: `access_points` table with `id`, `plant_id`, `tenant_id`, `name`, `is_active`.
   - Lines 101-108: `user_role` enum includes `'guard'`, `'plant_admin'`, `'super_admin'`, `'worker'`.
   - Lines 111-127: `users` table where QR payload is directly the user UUID (line 120: "QR code: the UUID itself is used as the QR payload").
   - Lines 158-183: `contractors` table with `name`, `rfc`, `repse_number`, `contractor_type`.
   - Lines 188-202: `plant_contractors` table with `status`: `'pending' | 'active' | 'suspended' | 'blocked'`.
   - Lines 216-229: `contractor_workers` table with `curp`, `imss_number`, `job_title`.
   - Lines 243-261: `document_types` table seeded with `DC3`, `REPSE`, `SUA`, `SAT_32D`, `INE`, `MEDICAL`.
   - Lines 284-294: `plant_document_requirements` table (per-plant required documents and validity days).
   - Lines 314-345: `documents` table with `status`: `'pending_review' | 'approved' | 'rejected' | 'expired'`, `issued_at`, `expires_at`.
   - Lines 363-367: `access_result` enum: `'granted'`, `'granted_with_warning'`, `'denied'`.
   - Lines 369-395: `access_logs` table recording `result`, `denial_reason`, `document_snapshot` (JSONB).

3. **Phase 1 MVP Scope & Acceptance Criteria (`ORIGINAL_REQUEST.md`):**
   - Lines 23-28: R2 Frontend Web Interface requires guard login screen, caseta check-in interface with worker UUID submission/scanning, and visual compliance status indicators (Green/Granted, Yellow/Warning, Red/Denied).
   - Lines 45-46: Acceptance criteria: Guard signs in through Next.js UI using test credentials; submitting test worker UUID executes check-in and renders granted/denied compliance card in browser.
   - Line 8: Integrity mode: demo.

## 2. Logic Chain
1. Based on Observation 1 (`DEVELOPMENT_RULES.md`), the frontend must be an isolated Next.js 14+ App Router project (`asesamo-web`) with TypeScript strict mode, Tailwind CSS, shadcn/ui, TanStack Query v5, and Sonner for toasts.
2. Based on Observation 1 (lines 141, 177-184), all HTTP communications must flow through a centralized client in `lib/api/` that automatically unwraps the `{ "data": ... }` and `{ "error": ... }` envelopes, injects JWT Bearer tokens from cookies, and exposes Spanish error messages to users.
3. Based on Observation 2 (lines 101-127), the QR code scanned by the guard encodes the worker's user UUID. Barcode scanners operate as HID keyboard wedges typing the UUID and sending an Enter keypress. Therefore, the Caseta check-in input must feature persistent auto-focus, rapid keystroke detection, and automatic mutation submission upon Enter.
4. Based on Observation 2 (lines 363-367) and Observation 3 (lines 27, 46), the visual status feedback must support three distinct tiers:
   - Green / Granted: Emerald-600 background, CheckCircle icon, full approval.
   - Yellow / Warning: Amber-500 background, AlertTriangle icon, warning message.
   - Red / Denied: Rose-600 background, Ban icon, prominent denial reason alert box.
5. Based on Observation 2 (lines 158-229, 243-345), the compliance result must display both:
   - A Worker Profile Card showing photo, name, CURP, NSS, job title, and contractor company with RFC/REPSE status.
   - A Document Validation Breakdown Table displaying each required document, ownership (worker vs contractor), validation status, and expiration date, explicitly flagging expired/missing documents that caused denial.
6. Based on Observation 3 (line 8: Integrity mode demo), the Caseta screen must include quick-fill preset buttons allowing instant demonstration and automated E2E testing of all compliance branches (Granted, Warning, Denied - DC3, Denied - Contractor Blocked, 404 Unknown Worker).

## 3. Caveats
- No active camera QR scanning library (e.g. `html5-qrcode`) is strictly required for Phase 1 MVP because industrial gates use hardware HID wedge scanners or typed UUIDs; however, the component architecture cleanly supports adding camera streaming in Phase 2.
- Web Audio sound effects (beeps/buzzers) depend on user browser interaction policies; Web Audio context must be resumed on first user click.
- Multi-tenancy at the frontend layer is primarily derived from the authenticated guard's JWT claims (`tenant_id`); guard users cannot cross tenants.

## 4. Conclusion
The frontend specification for `asesamo-web` is completely surveyed, detailed, and documented in:
`/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_spec_miner_survey_2/survey_report.md`

It delivers:
1. Complete Next.js 14+ App Router directory structure and component hierarchy.
2. Comprehensive Security Guard Login screen specification with Zod validation, cookie session handling, and role verification.
3. Caseta Check-in interface specification with auto-focused scanner input, HID keyboard wedge support, access point context, and demo quick-fill presets.
4. Detailed specifications for Green/Yellow/Red visual status indicators, Worker Profile Card, and Document Validation Breakdown Table.
5. Industrial UI/UX standards, dark/light theme tokens, Sonner toast recipes, loading skeletons, and WCAG 2.1 AA accessibility guidelines.
6. Robust typed API client layer (`lib/api`) with full TypeScript interfaces matching backend DTOs.

## 5. Verification Method
1. Inspect the survey report:
   ```bash
   test -f /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_spec_miner_survey_2/survey_report.md
   ```
2. Verify all 6 mandatory areas are documented:
   - Architecture: Search `## 1. Frontend Architecture`
   - Guard Login: Search `## 2. Security Guard Login Interface Specification`
   - Caseta Check-in: Search `## 3. Caseta Check-In Interface Specification`
   - Visual Feedback: Search `## 4. Visual Compliance Feedback & Status Indicators`
   - UI/UX Standards: Search `## 5. UI/UX Design Standards`
   - API Client Layer: Search `## 6. API Client Layer Structure`
3. Verify Specification Miner compliance:
   - Discovered Features table present (`## Features Discovered`)
   - Edge Cases table present (`## Edge Cases`)
