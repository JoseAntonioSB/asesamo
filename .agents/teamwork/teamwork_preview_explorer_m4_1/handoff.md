# Handoff Report: Next.js 14+ App Router & UI System Architecture for `asesamo-web/`

**Milestone**: M4 (Frontend Web Interface Architecture)  
**Agent**: `teamwork_preview_explorer_m4_1`  
**Working Directory**: `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m4_1`  
**Handoff Type**: Hard (Investigation & Architectural Blueprint Complete)  
**Timestamp**: 2026-10-07T16:32:00Z  

---

## 1. Observation

1. **Repository Layout and Directory State**:
   - Inspected repository root `/Users/jasbotello/devel/asesamo`: `asesamo-web/` does not yet exist.
   - `docker-compose.yml` (lines 108–135) specifies service `web` with context `./asesamo-web`, dockerfile `Dockerfile`, environment variables `NODE_ENV: production`, `PORT: 3000`, `NEXT_PUBLIC_API_URL: ${NEXT_PUBLIC_API_URL:-http://localhost:8080/api/v1}`, `INTERNAL_API_URL: ${INTERNAL_API_URL:-http://api:8080/api/v1}`, and port forwarding `3000:3000`.
   - `Makefile` (lines 142–144) targets `asesamo-web`:
     ```makefile
     @if [ -d "asesamo-web" ]; then \
         echo "==> Linting asesamo-web..."; \
         (cd asesamo-web && if [ -f "package.json" ]; then npm run lint || true; fi); \
     ```
2. **Authoritative Requirements in `PROJECT.md` & `ORIGINAL_REQUEST.md`**:
   - `PROJECT.md` Feature #12 (line 40): "Frontend Next.js 14 App Router Setup: TypeScript strict, Tailwind CSS, shadcn/ui component scaffolding, TanStack Query".
   - `PROJECT.md` Feature #16 (line 44): "3-Tier Visual Status Feedback: High-visibility Green (Granted), Yellow (Warning), Red (Denied) banner cards".
   - `ORIGINAL_REQUEST.md` R2 (lines 23–28): "Implement a Next.js (App Router) web application styled with Tailwind CSS and shadcn/ui. Display an immediate, clear visual status indicator (Green/Granted, Yellow/Warning, Red/Denied)".
   - `DEVELOPMENT_RULES.md` (lines 24–33, 135–149):
     - Framework: Next.js 14+ App Router (Server Components by default).
     - Styles: Tailwind CSS.
     - Components: shadcn/ui (over Radix UI).
     - Strict TypeScript: `strict: true` in `tsconfig.json`, no `any`.
     - Code & comments in English; user-facing error messages in Spanish.
3. **Dispatch Directives for Explorer 1**:
   - Next.js 14+ project configuration: `package.json`, `tsconfig.json` with `@/*: ./*`, `tailwind.config.ts`, `postcss.config.js`, `next.config.mjs`.
   - Global styles & design system: `app/globals.css` with industrial modern palette (dark slate, high-contrast, emerald green, amber yellow, crimson red), `app/layout.tsx` with Inter font and `<html lang="es">`.
   - Core UI component primitives: `button.tsx`, `card.tsx`, `input.tsx`, `badge.tsx`, `table.tsx`, `alert.tsx`, and `lib/utils.ts` (`cn`).

---

## 2. Logic Chain

1. **Scaffolding Configuration (Step 1 -> Observation 1 & 2)**:
   - Next.js 14.2+ App Router requires modern dependency alignment: React 18.3, `@radix-ui/react-slot` (enabling `asChild` composition), `class-variance-authority` (for typed variant handling), `clsx` and `tailwind-merge` (for utility resolution), `lucide-react` (for icons), and `@tanstack/react-query` (as mandated by `PROJECT.md`).
   - `tsconfig.json` must enforce `strict: true` and map `@/*` to `./*` so that imports like `@/components/ui/button` and `@/lib/utils` resolve consistently across the entire app tree.
   - `next.config.mjs` must specify `output: "standalone"` to satisfy Docker containerization in Milestone 5.
2. **Visual Design System & Industrial Ergonomics (Step 2 -> Observation 2 & 3)**:
   - Security guards working in casetas experience extreme lighting variations (direct sunlight to nocturnal low light). A dark slate substrate (`hsl(222 47% 7%)` / `#0a0f1d`) provides an anti-glare canvas where status colors achieve maximum luminance contrast.
   - Dedicated semantic CSS variables (`--granted`, `--warning`, `--denied`) decouple business access outcomes from generic UI colors, ensuring that granted (emerald green), warning (amber yellow), and denied (crimson red) are uniformly recognized across badges, alerts, and buttons.
   - Dual-font architecture: Inter provides high-legibility sans-serif text for labels and body; JetBrains Mono provides high-density tabular and scan telemetry alignment for UUIDs, CURP, NSS, and RFC.
   - `<html lang="es">` guarantees compliance with Mexican industrial accessibility and browser localization standards.
3. **Core UI Primitives (Step 3 -> Observation 2 & 3)**:
   - `button.tsx`: ForwardRef component supporting standard variants plus specialized `granted`, `warning`, `denied` buttons for caseta gate control actions.
   - `card.tsx`: ForwardRef component suite (`Card`, `CardHeader`, `CardTitle`, `CardDescription`, `CardContent`, `CardFooter`) styled with technical borders and elevated backgrounds for worker inspection dossiers.
   - `input.tsx`: ForwardRef input supporting barcode scanner autofocus and `isTelemetry` monospace styling for barcode wedge entry.
   - `badge.tsx`: Pill badge with `granted`, `warning`, `denied` status variants for document verification states.
   - `table.tsx`: ForwardRef table components (`Table`, `TableHeader`, `TableBody`, `TableRow`, `TableHead`, `TableCell`, `TableCaption`) with crisp borders for DC-3, IMSS, and REPSE compliance lists.
   - `alert.tsx`: Banner component with status variants for guard feedback.
   - `lib/utils.ts`: Contains the canonical `cn()` helper combining `clsx` and `twMerge`.

---

## 3. Caveats

1. **Direct Implementation Scope**: Explorer agents operate in read-only investigation mode regarding the main application codebase. All files have been authored and documented in `analysis.md` and `handoff.md` within the working directory for worker agents to copy into `asesamo-web/`.
2. **Node Engine Version in Environment**: The configuration assumes Node.js 18.17+ or 20+ runtime for Next.js 14.
3. **No External Fonts at Build Time**: Google Fonts (`Inter`, `JetBrains_Mono`) use `next/font/google` which downloads fonts at build time. In air-gapped Docker environments without internet, a system font fallback is provided in `tailwind.config.ts`.

---

## 4. Conclusion

The Next.js 14+ App Router project scaffolding, industrial design system, and core UI component primitives have been designed, typed, and documented with zero placeholders. All specifications in `analysis.md` are directly actionable by worker agents to scaffold `asesamo-web/`. Peer explorers (Explorer 2 for Auth/API and Explorer 3 for Caseta UI) can build directly on top of these primitives.

---

## 5. Verification Method

To verify the design once implemented by worker agents:

1. **File Existence Check**:
   ```bash
   test -f asesamo-web/package.json
   test -f asesamo-web/tsconfig.json
   test -f asesamo-web/tailwind.config.ts
   test -f asesamo-web/postcss.config.js
   test -f asesamo-web/next.config.mjs
   test -f asesamo-web/components.json
   test -f asesamo-web/app/globals.css
   test -f asesamo-web/app/layout.tsx
   test -f asesamo-web/lib/utils.ts
   test -f asesamo-web/components/ui/button.tsx
   test -f asesamo-web/components/ui/card.tsx
   test -f asesamo-web/components/ui/input.tsx
   test -f asesamo-web/components/ui/badge.tsx
   test -f asesamo-web/components/ui/table.tsx
   test -f asesamo-web/components/ui/alert.tsx
   ```
2. **Dependency & Build Validation**:
   ```bash
   cd /Users/jasbotello/devel/asesamo/asesamo-web
   npm install
   npm run typecheck
   npm run build
   ```
3. **Invalidation Conditions**:
   - Any missing import or syntax error in `@/components/ui/*`.
   - TypeScript compilation error (`tsc --noEmit` exits non-zero).
   - Inability to build with `next build` due to missing `next.config.mjs` or Tailwind syntax issues.
