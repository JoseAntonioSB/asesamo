## 2026-10-01T22:15:51Z

You are teamwork_preview_spec_miner_survey_2.
Your working directory is: /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_spec_miner_survey_2

MANDATORY: You MUST read /Users/jasbotello/devel/asesamo/.agents/teamwork/ORIGINAL_REQUEST.md before starting work.
Also read:
- /Users/jasbotello/devel/asesamo/DEVELOPMENT_RULES.md
- /Users/jasbotello/devel/asesamo/db/migrations/000001_init.up.sql

Your task:
Conduct a comprehensive frontend specification survey for Phase 1 MVP of ASESAMO.
Extract and document:
1. Frontend architecture for asesamo-web (Next.js 14+ App Router, TypeScript strict, Tailwind CSS, shadcn/ui components).
2. Security guard login screen specifications (form, validation, JWT storage in cookie/state, error handling, redirect).
3. Caseta Check-in interface specifications:
   - Worker UUID entry (input field for typing/barcode/QR scanner simulation).
   - Access point selection/context.
   - Check-in submission and API call.
4. Visual compliance feedback & status indicators:
   - Green / Granted: Access allowed, all docs compliant.
   - Yellow / Warning: Access granted with warning according to plant policy.
   - Red / Denied: Access denied, list of missing or expired mandatory documents and denial reason.
   - Worker profile card: Worker photo/avatar, full name, CURP/NSS, contractor company name, plant & access point.
   - Document validation breakdown table/list showing each evaluated document, type, status, and expiration.
5. UI/UX design standards: industrial look and feel, dark/light mode support, Sonner toasts for notifications, loading states, accessibility.
6. API client layer structure (lib/api encapsulating base URL and auth header).

Write your detailed findings to:
/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_spec_miner_survey_2/survey_report.md
Also write handoff.md in your working directory.
When finished, call send_message to report completion.
