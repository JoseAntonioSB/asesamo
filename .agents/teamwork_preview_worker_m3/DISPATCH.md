# Dispatch for teamwork_preview_worker_m3

## Task: Implement Milestone 3 — Modern Clinic Dashboard UI
Read `/Users/jasbotello/devel/APPCITAS/ORIGINAL_REQUEST.md` and `/Users/jasbotello/devel/APPCITAS/PROJECT.md`.
Read UI/UX blueprint at `/Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_explorer_survey_3/analysis.md`.

## Mandatory Integrity Warning
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. A teamwork_preview_auditor will independently verify your work. Integrity violations WILL be detected and your work WILL be rejected.

## File Write Ownership
You exclusively own and will create/modify the following files:
- `src/app/layout.tsx`
- `src/app/page.tsx`
- `src/app/dashboard/layout.tsx`
- `src/app/dashboard/page.tsx`
- `src/app/dashboard/specialists/page.tsx`
- `src/app/dashboard/services/page.tsx`
- `src/components/ui/button.tsx`
- `src/components/ui/modal.tsx`
- `src/components/ui/card.tsx`
- `src/components/ui/input.tsx`
- `src/components/ui/badge.tsx`
- `src/components/dashboard/sidebar.tsx`
- `src/components/dashboard/header.tsx`
- `src/components/dashboard/mobile-nav.tsx`
- `src/components/specialists/specialist-card.tsx`
- `src/components/specialists/onboard-modal.tsx`
- `src/components/services/service-card.tsx`
- `src/components/services/specialist-assignment-modal.tsx`
- `tests/verify-m3.ts`

## Concrete UI/UX Requirements
1. **Design System & Aesthetics**:
   - Apply `ui-ux-pro-max`, `apple-design`, `minimalist-ui`:
   - Primary: Medical teal (`bg-teal-600`, `text-teal-600`, `hover:bg-teal-700`, `active:scale-[0.98]`).
   - Surfaces: Crisp white cards with hairline borders (`border-slate-200`), subtle shadows, rounded-xl/rounded-2xl.
   - Typography: Crisp, modern sans-serif with strong hierarchy.
   - Icons: `lucide-react` icons (Users, Stethoscope, Calendar, DollarSign, Plus, Check, Copy, AlertCircle, Menu, X, LogOut, ChevronRight).
2. **Root Layout (`src/app/layout.tsx`)**:
   - Wrap application with Inter font, metadata ("APPCITAS — Gestión Integral para Clínicas de Fisioterapia"), and Sonner `<Toaster richColors closeButton position="top-right" />`.
3. **Landing Page (`src/app/page.tsx`)**:
   - Clean healthcare entrance with login/register links and redirection to `/dashboard` if authenticated.
4. **Dashboard Layout (`src/app/dashboard/layout.tsx`)**:
   - Sidebar for desktop (>= 1024px) with navigation links: Overview (`/dashboard`), Team Roster (`/dashboard/specialists`), Services Catalog (`/dashboard/services`), Logout button.
   - Header with clinic name badge, user info, and public booking link placeholder (`tuapp.com/clinica-fisia`).
   - Mobile responsive drawer with accessible hamburger button (< 1024px).
5. **Overview Page (`src/app/dashboard/page.tsx`)**:
   - KPI metric bento cards:
     - Total Specialists (e.g. 5/5 activos) with progress indicator.
     - Services Catalog (e.g. 6 servicios activos).
     - Citas de Hoy (placeholder/mock for Phase 2).
     - Corte de Caja / Cierre Diario (foundation for Phase 3).
   - Quick Actions: "Onboard Specialist", "Crear Servicio", "Ver Agenda Pública".
6. **Team Roster Page (`src/app/dashboard/specialists/page.tsx`)**:
   - Displays specialists with specialty tags, phone, email, active badge.
   - Shows active quota counter (X / 5 especialistas activos).
   - "Dar de alta especialista" button opening `OnboardSpecialistModal`.
   - Modal supports direct account creation (Name, Email, Specialty, Phone, Password).
   - Upon creation, shows temporary credentials card with "Copiar al portapapeles" (Zero-SMTP via Sonner `toast.success("Credenciales copiadas al portapapeles")`).
7. **Services Catalog Page (`src/app/dashboard/services/page.tsx`)**:
   - Lists services with formatted price (MXN), duration chip (e.g. "45 min"), description.
   - Flexible specialist assignment matrix with optimistic checkbox toggles or assignment modal.
   - Saving updates calls `PUT /api/services/[id]/assign` or `PATCH /api/services` with Sonner toast feedback.
8. **Verification & Build**:
   - Create `tests/verify-m3.ts` verifying rendering, component props, and API response integration.
   - Run `npx tsc --noEmit`.
   - Run `npm run build` — must complete with 0 TypeScript and Next.js packaging errors! (Ensure server components handle static evaluation cleanly, e.g. using `export const dynamic = 'force-dynamic'` where request headers/cookies are evaluated).

Write your handoff to `/Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_worker_m3/handoff.md`.

## 2026-09-14T04:36:22Z
You are teamwork_preview_worker_m3.
Your working directory is: /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_worker_m3
You MUST read /Users/jasbotello/devel/APPCITAS/ORIGINAL_REQUEST.md before starting work.
Also read /Users/jasbotello/devel/APPCITAS/PROJECT.md and your dispatch at /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_worker_m3/DISPATCH.md.

MANDATORY INTEGRITY WARNING:
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. A teamwork_preview_auditor will independently verify your work. Integrity violations WILL be detected and your work WILL be rejected.

Implement Milestone 3 files per your write ownership:
1. Reusable UI components: button, modal, card, input, badge.
2. Root layout (`src/app/layout.tsx`) with Sonner Toaster & root page (`src/app/page.tsx`).
3. Dashboard shell: `src/app/dashboard/layout.tsx` (sidebar, header, mobile-nav).
4. Overview page: `src/app/dashboard/page.tsx` (KPI bento cards, quick actions).
5. Team Roster page: `src/app/dashboard/specialists/page.tsx` (5 specialists, quota counter, OnboardSpecialistModal with credentials copy).
6. Services Catalog page: `src/app/dashboard/services/page.tsx` (pricing, duration, specialist assignment matrix).
7. Verification test: `tests/verify-m3.ts`.

Run `npx tsc --noEmit` and `npm run build` — ensure 0 TypeScript and Next.js packaging errors!
Write your handoff to /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_worker_m3/handoff.md and notify parent via send_message.
