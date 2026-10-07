# Dispatch Assignment: Milestone M4 Worker (asesamo-web)
Timestamp: 2026-10-07T16:35:00Z
Agent: teamwork_preview_worker_m4_1
Working Directory: /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_worker_m4_1

## Objective
Implement the complete, production-grade Next.js 14+ frontend application in `/Users/jasbotello/devel/asesamo/asesamo-web/` based on the architectural blueprints from Explorers 1, 2, and 3.

## Authoritative Inputs
- /Users/jasbotello/devel/asesamo/.agents/teamwork/ORIGINAL_REQUEST.md
- /Users/jasbotello/devel/asesamo/PROJECT.md (Milestone 4, Features #12-17)
- /Users/jasbotello/devel/asesamo/DEVELOPMENT_RULES.md
- /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m4_1/analysis.md & handoff.md
- /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m4_2/analysis.md & handoff.md
- /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m4_3/analysis.md & handoff.md

## Scope & File Ownership
You exclusively own and must implement the entire `asesamo-web/` directory tree:
1. Configuration & Project Foundation:
   - `package.json` (Next 14+, React 18, lucide-react, clsx, tailwind-merge, class-variance-authority, tailwindcss, typescript, etc.)
   - `tsconfig.json` (strict: true, "@/*": ["./*"])
   - `next.config.mjs` (output: "standalone")
   - `tailwind.config.ts`, `postcss.config.js`, `components.json`
2. Core Styles, Utilities & Layout:
   - `app/globals.css` (Industrial theme, status colors: emerald granted, amber warning, crimson denied)
   - `app/layout.tsx` (Root layout, metadata, `<html lang="es">`, AuthProvider)
   - `lib/utils.ts` (`cn` helper)
   - `lib/utils/audio.ts` (Web Audio API feedback chimes)
3. UI Primitives (`components/ui/`):
   - `button.tsx`, `card.tsx`, `input.tsx`, `badge.tsx`, `table.tsx`, `alert.tsx`
4. API & Authentication Layer:
   - `lib/api/types.ts` & `lib/types/caseta.ts`
   - `lib/api/client.ts` & `lib/api.ts` (Base URL normalization, envelope unwrapping, Bearer token injection, Spanish error handling)
   - `lib/auth.ts` (Dual cookie/storage persistence)
   - `contexts/auth-context.tsx` (`AuthProvider` and `useAuth()`)
   - `middleware.ts` (Edge route guard protecting `/caseta` and redirecting `/login`)
5. Pages & Workstations:
   - `app/page.tsx` (Redirect to `/caseta` or `/login`)
   - `app/login/page.tsx` (Guard authentication form, validation, quick-fill demo credentials, Spanish error banners)
   - `app/caseta/page.tsx` (Gate workstation layout, header, digital clock, active access point, guard logout)
6. Caseta Inspection Components (`components/caseta/`):
   - `scanner-input.tsx` (Auto-focus, HID wedge keyboard barcode support on Enter, auto-clear & refocus, global shortcut refocus)
   - `quick-fill-presets.tsx` (1-click buttons for Carlos Hernández, María Rodríguez, Jorge Treviño, dual seed database alignment)
   - `status-banner.tsx` (3-tier high-visibility feedback cards: Green/Granted, Yellow/Warning, Red/Denied with Spanish reason)
   - `worker-card.tsx` (Worker profile, avatar, CURP, NSS, contractor info)
   - `document-table.tsx` (Document breakdown table with badges and expiration dates)

## Build & Verification Commands
Inside `/Users/jasbotello/devel/asesamo/asesamo-web`:
- `npm install`
- `npm run lint` (or `npx next lint`)
- `npm run build` (or `npx next build`)
Both lint and build must exit with code 0.

Write your report to `handoff.md` in your working directory and call send_message when done.


## 2026-10-07T16:35:17Z
You are teamwork_preview_worker_m4_1.
Your working directory is: /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_worker_m4_1

MANDATORY: You MUST read /Users/jasbotello/devel/asesamo/.agents/teamwork/ORIGINAL_REQUEST.md before starting work.
Also read:
- /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_worker_m4_1/DISPATCH.md
- /Users/jasbotello/devel/asesamo/PROJECT.md
- /Users/jasbotello/devel/asesamo/DEVELOPMENT_RULES.md
- /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m4_1/analysis.md and handoff.md
- /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m4_2/analysis.md and handoff.md
- /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m4_3/analysis.md and handoff.md

MANDATORY INTEGRITY WARNING:
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. A teamwork_preview_auditor will independently verify your work. Integrity violations WILL be detected and your work WILL be rejected.

Scope & File Ownership:
You exclusively own and must implement the entire `asesamo-web/` application tree at `/Users/jasbotello/devel/asesamo/asesamo-web`:
1. Foundation & Configuration:
   - `package.json`: Next.js 14, React 18, React DOM 18, lucide-react, clsx, tailwind-merge, class-variance-authority, tailwindcss, postcss, autoprefixer, typescript, @types/react, @types/node, @types/react-dom.
   - `tsconfig.json`: strict: true, path alias "@/*": ["./*"].
   - `next.config.mjs`: output: "standalone".
   - `tailwind.config.ts` and `postcss.config.js`.
   - `components.json`.
2. Global Styling & Root Layout:
   - `app/globals.css`: Dark slate background, industrial contrast, semantic status tokens (granted: emerald, warning: amber, denied: crimson).
   - `app/layout.tsx`: Root layout with font definitions, metadata, `<html lang="es">`, wrapped with `AuthProvider`.
   - `lib/utils.ts`: Standard `cn` helper function (`clsx` + `twMerge`).
   - `lib/utils/audio.ts`: Web Audio API synthetic feedback chimes.
3. UI Primitives (`components/ui/`):
   - `button.tsx`, `card.tsx`, `input.tsx`, `badge.tsx`, `table.tsx`, `alert.tsx`.
4. API & Authentication Layer:
   - `lib/api/types.ts` & `lib/types/caseta.ts`: Complete TypeScript contracts.
   - `lib/api/client.ts` & `lib/api.ts`: Centralized `ApiClient` with base URL normalization (`NEXT_PUBLIC_API_URL` handling), envelope unwrapping (`{ data, meta }`), error translation (`{ error: { code, message } }`), and Bearer token injection.
   - `lib/auth.ts`: Dual cookie (`asesamo_token`) and localStorage persistence.
   - `contexts/auth-context.tsx`: `AuthProvider` and `useAuth()` hook.
   - `middleware.ts`: Next.js Edge route guard protecting `/caseta` and redirecting `/login`.
5. Pages & Components:
   - `app/page.tsx`: Landing redirect to `/caseta` or `/login`.
   - `app/login/page.tsx`: Guard authentication screen with demo auto-fill button (`guard@asesamo.com` / `Password123!`), validation, and Spanish error alerts.
   - `app/caseta/page.tsx`: Guard workstation layout, header with active access point (`Caseta Principal Norte`), plant name (`Planta Querétaro Norte`), live digital clock, sound chime toggle, and guard logout.
   - `components/caseta/scanner-input.tsx`: Auto-focused HID wedge input (keyboard scanner sends Enter, auto-submits, clears and refocuses; global F2 or / refocus listener).
   - `components/caseta/quick-fill-presets.tsx`: 1-click test buttons for Carlos Hernández (`66666666-6666-6666-6666-666666666661`), María Rodríguez (`66666666-6666-6666-6666-666666666662`), and Jorge Treviño (`66666666-6666-6666-6666-666666666663`) with dual seed database compatibility.
   - `components/caseta/status-banner.tsx`: 3-tier high-visibility feedback banners (Green/Granted, Yellow/Warning, Red/Denied with Spanish denial reason).
   - `components/caseta/worker-card.tsx`: Worker profile, avatar/initials, CURP, NSS, contractor name, RFC, REPSE number.
   - `components/caseta/document-table.tsx`: Breakdown table of worker documents with color-coded status badges and expiration dates.

Build and Test Verification:
Inside `/Users/jasbotello/devel/asesamo/asesamo-web`:
- Run `npm install`
- Run `npm run lint` (or `npx next lint` - must exit with code 0)
- Run `npm run build` (or `npx next build` - must exit with code 0)

Document all created files, build output, and verification results in `handoff.md` in your working directory. Call send_message when done.
