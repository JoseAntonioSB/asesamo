# ASESAMO Phase 1 MVP — Frontend Specification Survey (`asesamo-web`)
**Surveyor:** `teamwork_preview_spec_miner_survey_2` (Specification Miner)  
**Date:** 2026-10-01  
**Target Application:** `asesamo-web` (Next.js 14+ App Router)  
**Authoritative References:**  
- `ORIGINAL_REQUEST.md` (User Requirements & Phase 1 MVP Scope)
- `DEVELOPMENT_RULES.md` (Architectural Standards, Tech Stack & UI/UX Guidelines)
- `db/migrations/000001_init.up.sql` (Authoritative PostgreSQL Data Schema)
- `SAS Seguridad e Higiene asesamo.md` (Industrial Access Domain & Guard Operations)

---

## Executive Summary

The ASESAMO frontend application (`asesamo-web`) serves as the operational frontline interface for industrial security personnel (guards) positioned at factory access points (casetas / gates). Its primary Phase 1 MVP objective is to provide high-speed, zero-friction verification of contractor workers entering industrial facilities by evaluating worker and company compliance (REPSE, DC-3, IMSS/SUA, INE, Medical Certs) against plant access policies, rendering instantaneous, unambiguous visual compliance feedback (Green/Granted, Yellow/Warning, Red/Denied).

This survey defines the comprehensive architecture, interface contracts, visual state machines, component hierarchies, and data structures required to implement `asesamo-web` strictly compliant with `DEVELOPMENT_RULES.md`.

---

## Features Discovered

| # | Category | Feature | Description | Inputs | Outputs | Error Behavior | Discovered Via |
|---|----------|---------|-------------|--------|---------|----------------|----------------|
| 1 | Architecture | Next.js 14+ App Router Setup | Root layout with font optimization, ThemeProvider, Sonner Toaster, and TanStack Query provider | URL route requests | Rendered HTML / React Server & Client Components | Standard Next.js error boundary / 404 page | `DEVELOPMENT_RULES.md` §1, §4 |
| 2 | Architecture | Strict TypeScript Configuration | Strict type checking (`strict: true`), zero explicit `any`, fully typed API DTOs | TS source files | Compilation check, IntelliSense | Build fails on type errors | `DEVELOPMENT_RULES.md` §4 |
| 3 | Architecture | Tailwind + shadcn/ui Design Tokens | Semantic design system with CSS custom variables supporting dark and light industrial palettes | CSS variables, Tailwind classes | Styled UI components with accessible contrasts | Fallback CSS colors | `DEVELOPMENT_RULES.md` §1, §15 |
| 4 | Authentication | Security Guard Login Form | Dedicated login view with email & password fields, loading state, and validation | Email string, password string | JWT auth token, user profile, redirect to `/caseta` | Displays user-facing error message in Spanish | `ORIGINAL_REQUEST.md` R2, `DEVELOPMENT_RULES.md` §6, §7 |
| 5 | Authentication | Guard Role Authorization Guard | Verification that authenticated user has `role: 'guard'`, `plant_admin`, or `super_admin` | Authenticated user profile | Access granted to `/caseta` route | 403 Forbidden with prompt: "Acceso exclusivo para personal de seguridad" | `000001_init.up.sql` user_role enum |
| 6 | Authentication | JWT Storage & Session Persistence | Storage of JWT access token in secure cookie (for SSR/middleware) and client auth context | JWT string from backend | Injected Authorization header for API calls | Auto-logout and redirect to `/login` on token expiry (401) | `DEVELOPMENT_RULES.md` §7 |
| 7 | Caseta Check-in | Worker UUID Barcode/QR Scanner Input | High-speed text input with persistent auto-focus, paste listener, and HID scanner emulation | Worker UUID string (scanned or typed) | Check-in mutation trigger on Enter | Validation error if input is not a valid UUID string | `ORIGINAL_REQUEST.md` R2, `000001_init.up.sql` users table |
| 8 | Caseta Check-in | Demo / Test Worker Preset Selectors | One-click simulation buttons to populate compliant, warning, denied, and missing workers | Click on demo preset button | Auto-fills worker UUID and triggers check-in | Disabled in production environment | `ORIGINAL_REQUEST.md` Integrity mode: demo |
| 9 | Caseta Check-in | Access Point Context & Gate Switcher | Selector for active caseta (e.g., "Caseta Principal", "Caseta Norte") persisted in localStorage | Access Point ID, Plant ID | Context banner display and access_point_id in check-in payload | Warning if no access point is selected | `000001_init.up.sql` access_points table |
| 10 | Caseta Check-in | Double-Scan Debounce Guard | Prevention of duplicate scans within 3 seconds for the same worker UUID | Repeated scan events | Ignores redundant requests, keeps current card active | Toast alert: "Escaneo duplicado ignorado" | Guard ergonomic requirement |
| 11 | Compliance UI | Green / Granted Status Banner | Full-width visual indicator confirming all worker & contractor documents are compliant | Check-in result: `'granted'` | Emerald green banner (`bg-emerald-600`), shield check icon, success sound | N/A | `ORIGINAL_REQUEST.md` R2, `000001_init.up.sql` access_result |
| 12 | Compliance UI | Yellow / Warning Status Banner | Visual indicator confirming access permitted per plant policy despite non-critical document warning | Check-in result: `'granted_with_warning'` | Amber yellow banner (`bg-amber-500`), alert triangle icon, warning reason | N/A | `ORIGINAL_REQUEST.md` R2, `000001_init.up.sql` access_result |
| 13 | Compliance UI | Red / Denied Status Banner | Visual indicator blocking entry due to expired/missing documents or blocked contractor | Check-in result: `'denied'` | Rose red banner (`bg-rose-600`), ban icon, denial reason alert box, buzzer sound | N/A | `ORIGINAL_REQUEST.md` R2, `000001_init.up.sql` access_result |
| 14 | Compliance UI | Worker Profile Card | Summary card rendering worker avatar/photo, full name, CURP, NSS, contractor, and job title | Worker metadata object | Rendered high-contrast profile card with quick-copy buttons | Fallback placeholder avatar if photo is null | `ORIGINAL_REQUEST.md` R2, `000001_init.up.sql` contractor_workers |
| 15 | Compliance UI | Document Validation Breakdown Table | Tabular breakdown of every evaluated document (type, status, validity date, required flag) | Array of evaluated document objects | Structured table with colored status badges (Aprobado, Vencido, Pendiente, Faltante) | Highlights failing rows with red border/background | `ORIGINAL_REQUEST.md` R2, `000001_init.up.sql` documents table |
| 16 | Compliance UI | Contractor Compliance Status Tag | Tag displaying contractor name, RFC, REPSE registration number, and plant relationship status | Contractor metadata object | Badge showing status: Activo, Pendiente, Suspendido, Bloqueado | Red tag and warning if contractor is not active | `000001_init.up.sql` plant_contractors |
| 17 | Shift History | Recent Access Log Stream | Live feed of recent check-ins processed at current access point during active shift | Stream / query of recent access logs | Compact list with worker name, time, contractor, and result badge | Empty state illustration if no check-ins yet | `000001_init.up.sql` access_logs table |
| 18 | UI/UX Standards | Industrial High-Contrast Theme | Design system tuned for industrial lighting (bright sun glare or low-light night shifts) | CSS custom properties, Tailwind dark mode | Crisp legible typography (Inter/Geist + monospace for codes) | Fallback to standard contrast | `DEVELOPMENT_RULES.md` §15, `ui-ux-pro-max` |
| 19 | UI/UX Standards | Dark/Light Mode Switcher | Accessible theme toggle persisted in localStorage/cookie via `next-themes` | User toggle click | Flips theme class on `<html>` element | Defaults to system preference | `DEVELOPMENT_RULES.md` §15 |
| 20 | UI/UX Standards | Sonner Toast Notifications | Notification toasts for async events (login success, scan errors, copy to clipboard) | `toast.success()`, `toast.error()` | Floating non-intrusive notification cards with rich colors | Auto-dismisses after 4 seconds | `DEVELOPMENT_RULES.md` §15, `ask-sonner` |
| 21 | UI/UX Standards | Skeleton Loading & Optimistic UI | Visual skeleton placeholders displayed while check-in verification API call is in-flight | Check-in pending state | Skeleton cards for profile and document table | Replaced by result or error state | `DEVELOPMENT_RULES.md` §15 |
| 22 | UI/UX Standards | Accessibility & Screen Reader (a11y) | WCAG 2.1 AA compliance, keyboard navigation focus rings, `aria-live` status announcements | Keyboard Tab/Enter, screen reader | Spoken check-in result for visually impaired personnel | Standard HTML fallback | `DEVELOPMENT_RULES.md` §15, `ui-styling` |
| 23 | API Layer | Centralized API Client (`lib/api`) | Base fetch wrapper encapsulating base URL (`NEXT_PUBLIC_API_URL`), headers, and token injection | HTTP method, endpoint path, payload | Typed JSON response data envelope | Throws structured `ApiClientError` on HTTP != 2xx | `DEVELOPMENT_RULES.md` §4, §6 |
| 24 | API Layer | Unified Envelope Unwrapping | Client parser for standard backend envelope `{ "data": ... }` and `{ "error": ... }` | Raw HTTP response | Unwrapped payload data | Extracts error `code` and Spanish `message` | `DEVELOPMENT_RULES.md` §6 |

---

## Edge Cases

| # | Feature | Input / Condition | Observed / Required Behavior |
|---|---------|-------------------|-----------------------------|
| 1 | Check-in Input | Empty string submitted or accidental Enter key press | Input field is prevented from submitting; subtle border shake or toast: "Ingrese un UUID de trabajador válido." |
| 2 | Check-in Input | Malformed UUID string (e.g., "12345", invalid format) | Client validates UUID format via regex before sending API request; displays inline validation message without making wasteful network call. |
| 3 | Check-in Input | Worker UUID not found in database (404 Not Found) | Backend returns `{"error": {"code": "WORKER_NOT_FOUND", "message": "Trabajador no encontrado en el sistema."}}`; UI displays Red Denied banner with prompt: "El código escaneado no corresponde a ningún trabajador registrado." |
| 4 | Check-in Input | Worker inactive / disabled (`users.is_active = false`) | Backend returns `WORKER_INACTIVE`; UI displays Red Denied banner: "Trabajador dado de baja o inactivo." |
| 5 | Check-in Input | Rapid double scan (scanner bounces or worker scans twice) | Frontend debounce filter drops duplicate scan within 3000ms; Sonner toast shows: "Escaneo duplicado ignorado." Prevents duplicate records in `access_logs`. |
| 6 | Access Policy | All required docs valid and approved (`result: 'granted'`) | UI renders Green Granted banner; displays worker details, approved badges for all documents; auto-focuses back to input field for next worker. |
| 7 | Access Policy | Mandatory doc expired & plant policy `block_on_expired_docs = true` | UI renders Red Denied banner; prominently highlights the expired document row in red; displays clear denial reason (e.g., "DC-3 vencido el 2026-09-01"); access blocked. |
| 8 | Access Policy | Doc expired & plant policy `block_on_expired_docs = false` | UI renders Yellow Warning banner; allows access but displays warning badge and informs guard: "Acceso autorizado con advertencia por política de planta." |
| 9 | Access Policy | Worker documents valid, but Contractor REPSE/SUA expired | Backend denies access; UI renders Red Denied banner; explains failure at contractor level: "Empresa contratista no cumple con requerimiento de REPSE o SUA vigente." |
| 10 | Access Policy | Contractor status is `'suspended'` or `'blocked'` for this plant | UI renders Red Denied banner; contractor status badge rendered in bold red: "Contratista bloqueado para ingresar a esta planta." |
| 11 | Access Policy | Document uploaded but status is `'pending_review'` or `'rejected'` | Evaluated as non-compliant for required documents; table displays "En revisión" or "Rechazado" badge; denied if required. |
| 12 | Auth Session | Guard token expires while active at gate (401 Unauthorized) | API client intercepts 401; clears stored token; redirects guard to `/login?session_expired=true`; toast: "Su sesión ha expirado. Inicie sesión nuevamente." |
| 13 | Auth Session | Non-guard user (e.g., contractor worker) logs into Caseta | API returns user profile with `role: "worker"`; frontend auth guard blocks entry, terminates session, and renders: "Acceso no autorizado: Este portal es exclusivo para personal de seguridad." |
| 14 | Network State | Backend server offline or intermittent Wi-Fi / 4G drop | API client catches fetch failure; Sonner toast shows persistent error: "Error de conexión con el servidor. Verifique su red."; check-in button re-enabled. |
| 15 | Access Point | No access point selected or invalid access point UUID | Form requires selecting an active access point before enabling scanner; shows modal: "Seleccione una Caseta de Acceso para iniciar turno." |
| 16 | Profile Media | Worker has no avatar photo (`avatar_r2_key` is null) | WorkerProfileCard renders fallback avatar with worker initials and distinct background color based on name hash. |

---

## 1. Frontend Architecture (`asesamo-web`)

### 1.1 Technical Stack & Dependencies

Strictly compliant with `DEVELOPMENT_RULES.md`:
- **Framework:** Next.js 14.2+ (App Router, React 18 / 19 Server Components)
- **Language:** TypeScript 5.x (`strict: true`, `noImplicitAny: true`, `strictNullChecks: true`)
- **Styling:** Tailwind CSS 3.4+ with `@tailwindcss/typography`
- **Component Primitives:** `shadcn/ui` (Radix UI primitives + Lucide React icons)
- **Async State & Caching:** `@tanstack/react-query` (v5)
- **Toast Notifications:** `sonner`
- **Theme Management:** `next-themes` (Dark / Light / System)
- **Forms & Validation:** `react-hook-form` + `zod`
- **HTTP Layer:** Native `fetch` with typed abstraction layer in `lib/api/`

### 1.2 Directory Structure Specification

```
asesamo-web/
├── app/
│   ├── layout.tsx                   # Root layout (Fonts, ThemeProvider, QueryClientProvider, Sonner Toaster)
│   ├── globals.css                  # Tailwind directives, CSS custom variables (industrial tokens)
│   ├── page.tsx                     # Root router: checks auth cookie and redirects to /caseta or /login
│   ├── (auth)/
│   │   ├── layout.tsx               # Centered industrial security backdrop
│   │   └── login/
│   │       └── page.tsx             # Guard login page
│   └── (dashboard)/
│       ├── layout.tsx               # Authenticated layout (Header, Caseta badge, Plant badge, UserMenu)
│       ├── caseta/
│       │   └── page.tsx             # Main Caseta Check-in operational screen
│       └── history/
│           └── page.tsx             # Access logs history for current shift/guard
├── components/
│   ├── ui/                          # shadcn/ui base primitives
│   │   ├── button.tsx
│   │   ├── card.tsx
│   │   ├── input.tsx
│   │   ├── badge.tsx
│   │   ├── table.tsx
│   │   ├── avatar.tsx
│   │   ├── select.tsx
│   │   ├── dialog.tsx
│   │   ├── alert.tsx
│   │   ├── skeleton.tsx
│   │   └── separator.tsx
│   ├── auth/
│   │   └── LoginForm.tsx            # Client component for guard login
│   ├── caseta/
│   │   ├── CasetaCheckIn.tsx        # Orchestrator client component
│   │   ├── WorkerScannerInput.tsx   # Continuous barcode/QR scanner input
│   │   ├── QuickFillPresets.tsx     # Demo simulation buttons (complaint/warning/denied)
│   │   ├── AccessPointSelector.tsx  # Gate/Caseta selector dialog & badge
│   │   ├── ComplianceResultBanner.tsx # Large Green/Yellow/Red compliance hero
│   │   ├── WorkerProfileCard.tsx    # Worker identity, photo, CURP, NSS, contractor info
│   │   ├── DocumentBreakdownTable.tsx # Evaluated documents list & status badges
│   │   └── RecentAccessList.tsx     # Stream of recent scans during active shift
│   ├── layout/
│   │   ├── Header.tsx               # Top navigation bar
│   │   ├── ThemeToggle.tsx          # Light/dark mode toggle
│   │   └── GuardProfileMenu.tsx     # Active guard identity & logout
│   └── providers/
│       ├── QueryProvider.tsx        # TanStack Query client provider
│       └── ThemeProvider.tsx        # next-themes provider
├── hooks/
│   ├── use-auth.ts                  # Guard authentication state & helpers
│   ├── use-check-in.ts              # React Query mutation for check-in verification
│   └── use-caseta-config.ts         # Persistent access point selection in localStorage
├── lib/
│   ├── api/
│   │   ├── client.ts                # Base HTTP fetch wrapper with token injection & error parsing
│   │   ├── auth.ts                  # Login, logout, getMe API functions
│   │   ├── access.ts                # Check-in, recent logs API functions
│   │   ├── plants.ts                # Plant & access point queries
│   │   └── types.ts                 # Full TypeScript interfaces for request/response DTOs
│   ├── auth/
│   │   └── cookies.ts               # Cookie helper for JWT storage
│   └── utils.ts                     # cn() class merger, date formatting helpers
├── public/
│   ├── favicon.ico
│   ├── sounds/                      # Audio feedback (beep-granted.mp3, beep-denied.mp3)
│   └── logo.svg                     # ASESAMO industrial brand mark
├── package.json
├── tsconfig.json
├── tailwind.config.ts
├── components.json
└── .env.example
```

---

## 2. Security Guard Login Interface Specification

### 2.1 UI Layout & Interaction Design

The guard login screen must be minimalist, sturdy, and distraction-free:
- **Container:** Centered card on a dark slate industrial background (`bg-slate-950`).
- **Brand Header:** ASESAMO logo accompanied by "Módulo de Caseta — Control de Acceso Industrial".
- **Form Fields:**
  1. **Correo Electrónico:** Input field with email type, placeholder `guardia@empresa.com`, auto-focus.
  2. **Contraseña:** Password field with toggleable show/hide eye icon.
  3. **Recordar Turno:** Checkbox to keep session active across browser restarts on guard terminal.
  4. **Iniciar Sesión:** High-contrast primary button (`bg-primary text-primary-foreground`) with loading spinner during submission.
- **Support Footer:** Contact notice: *"¿Problemas con su acceso? Contacte a Seguridad Patrimonial / EHS."*

### 2.2 Form Validation Schema (Zod)

```typescript
import { z } from "zod";

export const loginSchema = z.object({
  email: z
    .string()
    .min(1, { message: "El correo electrónico es requerido." })
    .email({ message: "Ingrese un correo electrónico válido." }),
  password: z
    .string()
    .min(6, { message: "La contraseña debe tener al menos 6 caracteres." }),
});

export type LoginFormData = z.infer<typeof loginSchema>;
```

### 2.3 Authentication Workflow & JWT Handling

1. **Submission:** Guard submits credentials via `LoginForm.tsx` calling `authApi.login({ email, password })`.
2. **API Endpoint:** Calls `POST /api/v1/auth/login`.
3. **Response Envelope:**
   ```json
   {
     "data": {
       "token": "eyJhbGciOi...",
       "user": {
         "id": "c1f7a07c-3f5e-4bb5-a3ec-95d8869c9b58",
         "tenant_id": "84d59a88-fef5-46aa-83a3-b24eecbaae87",
         "email": "guardia.norte@grupoindustrial.mx",
         "full_name": "Sargento Juan Manuel Morales",
         "role": "guard"
       }
     }
   }
   ```
4. **Role Enforcement:** The client checks `user.role`.
   - Allowed roles: `'guard'`, `'plant_admin'`, `'super_admin'`.
   - If user has another role (e.g. `'worker'`, `'contractor_admin'`), the client immediately rejects:
     *Toast:* `"Acceso denegado: Su cuenta no cuenta con permisos de operador de caseta."`
5. **Storage:**
   - The JWT token is saved to an `httpOnly` or `SameSite=Lax` browser cookie named `asesamo_token` (with `maxAge = 3600` seconds / 1 hour per `DEVELOPMENT_RULES.md` §7).
   - In addition, the user identity is stored in React Query auth cache for immediate UI hydration.
6. **Redirect:**
   - Navigates immediately to `/caseta` via `router.push("/caseta")`.
   - Sonner toast announces: `"Bienvenido(a), Sargento Juan Manuel Morales"`.

### 2.4 Error Handling Matrix

| HTTP Status | Error Code | User-Facing Message (Spanish) | Visual Handling |
|-------------|------------|-------------------------------|-----------------|
| 401 | `INVALID_CREDENTIALS` | "Credenciales inválidas. Verifique su correo y contraseña." | Red alert box above form + Sonner error toast |
| 403 | `USER_INACTIVE` | "Su cuenta de guardia está inactiva. Contacte al administrador de planta." | Red alert box + disabled inputs |
| 429 | `RATE_LIMIT_EXCEEDED` | "Demasiados intentos fallidos. Espere 2 minutos antes de reintentar." | Yellow warning alert with countdown |
| 500 / Network | `SERVER_ERROR` | "Error al conectar con el servidor de ASESAMO. Verifique la conexión." | Sonner error toast with retry button |

---

## 3. Caseta Check-In Interface Specification

### 3.1 Worker UUID Entry & Scanner Integration

Industrial facilities utilize handheld 2D barcode / QR scanners (e.g., Honeywell, Zebra) operating as **HID Keyboard Wedge** devices. When a guard scans a worker's QR badge:
1. The scanner automatically types the encoded worker UUID character-by-character into whatever input element currently has focus.
2. The scanner sends a trailing `Enter` keypress (`keyCode 13`).

**Scanner Input Requirements (`WorkerScannerInput.tsx`):**
- **Continuous Auto-Focus:** The input field automatically maintains focus on mount and instantly recovers focus 50ms after every check-in submission or modal closure, unless the guard is explicitly typing in a search/modal input.
- **Fast Paste & Scan Detection:** Automatically detects rapid burst entry (< 80ms total input time) characteristic of hardware scanners.
- **Enter Key Handling:** On `Enter`, triggers validation and submission immediately without requiring mouse interaction.
- **Visual Status Rings:**
  - Neutral / Ready: Slate border with subtle pulsing dot: *"Listo para escanear credencial..."*
  - Scanning / Processing: Blue spinner indicator.
  - Success / Granted: Momentary emerald flash.
  - Denied: Momentary rose flash.
- **Audio Feedback:** Play distinct audio chime via Web Audio API:
  - High dual-tone bell (1200Hz + 1600Hz) on `granted`.
  - Staccato warning buzz (600Hz) on `granted_with_warning`.
  - Low descending buzzer (300Hz -> 180Hz) on `denied`.
  *(Includes a mute switch in header for quiet environments).*

### 3.2 Demo Quick-Fill Presets (`QuickFillPresets.tsx`)

Per `ORIGINAL_REQUEST.md` (Integrity mode: demo), the Caseta interface must provide rapid simulation buttons so operators and testers can verify all compliance branches instantly:
1. **🟢 Trabajador Conforme (Acceso Permitido):**
   - Populates UUID of a worker with 100% compliant documents (valid DC-3, valid REPSE, active SUA).
2. **🟡 Acceso con Advertencia (Política de Planta):**
   - Populates UUID of a worker whose non-blocking document is expired or expiring soon, evaluating to `granted_with_warning`.
3. **🔴 Acceso Denegado (DC-3 Vencido):**
   - Populates UUID of a worker whose mandatory DC-3 certificate has expired.
4. **🔴 Acceso Denegado (Contratista Bloqueado):**
   - Populates UUID of a worker whose contractor company is suspended or not registered in `plant_contractors`.
5. **❌ Trabajador Inexistente (404 Desconocido):**
   - Populates a random unseeded UUID (`00000000-0000-0000-0000-000000000000`) to test error handling.

### 3.3 Access Point Selection & Context (`AccessPointSelector.tsx`)

- **Context Bar:** Located at top of the Caseta screen:
  - **Planta:** "Planta Querétaro Norte (Bajío)"
  - **Punto de Acceso:** "Caseta Principal (Torniquete 1)"
  - **Operador:** "Guardia: Juan Manuel Morales"
  - **Botón Cambiar:** Opens a dialog to switch active access point.
- **Persistence:** Selected `access_point_id` and `plant_id` are saved in browser `localStorage` (`asesamo_active_access_point`).
- **Initial Load:** If no access point is stored, the interface automatically presents a selection modal populating access points from `GET /api/v1/plants/{plant_id}/access-points`.

### 3.4 Check-In API Request & Execution

- **Trigger:** On scanner Enter or click on "Verificar Acceso".
- **Payload:**
  ```json
  {
    "worker_id": "7b8f2d4e-1234-4a5b-9c8d-e0123456789a",
    "access_point_id": "11111111-2222-3333-4444-555555555555"
  }
  ```
- **Mutation:** Handled via TanStack Query `useMutation`:
  ```typescript
  const checkInMutation = useMutation({
    mutationFn: (data: CheckInRequest) => accessApi.checkIn(data),
    onSuccess: (data) => {
      // update active result card
      // invalidate recent access logs query
      // play appropriate audio chime
      // re-focus scanner input
    },
    onError: (error: ApiClientError) => {
      // show error toast
      // re-focus scanner input
    }
  });
  ```

---

## 4. Visual Compliance Feedback & Status Indicators

### 4.1 Three-Tier Visual Status Hero Banner (`ComplianceResultBanner.tsx`)

The compliance banner is the most critical visual element in the Caseta interface. It must be instantly legible from 2 to 3 meters away on a mounted tablet or guard monitor.

```
+-----------------------------------------------------------------------------------+
|  [ICON 64px]    ESTADO DE ACCESO: ACCESO PERMITIDO (GRANTED)                      |
|                 Verificado el 01/Oct/2026 16:45:12 | Caseta Principal             |
+-----------------------------------------------------------------------------------+
```

#### 4.1.1 Tier 1: Green / Granted (`granted`)
- **Semantic Meaning:** Worker and contractor documentation is 100% valid, approved, and up to date per plant policy. Entry is authorized.
- **Visual Styling:**
  - Container: High-contrast emerald green background (`bg-emerald-600` light / `bg-emerald-950` dark with `border-2 border-emerald-500`).
  - Text: White bold typography (`text-white font-extrabold tracking-wide`).
  - Icon: Prominent ShieldCheck / CheckCircle2 (64px, `text-emerald-200`).
  - Header: **"ACCESO AUTORIZADO"** (text-3xl / text-4xl uppercase).
  - Subtitle: *"Documentación en regla. Permitir el paso al torniquete."*
  - Badge: `[ 🟢 CUMPLIMIENTO TOTAL ]`

#### 4.1.2 Tier 2: Yellow / Warning (`granted_with_warning`)
- **Semantic Meaning:** Worker has one or more document anomalies (e.g. document expiring within 5 days, or plant policy allows temporary entry without blocking).
- **Visual Styling:**
  - Container: Industrial amber background (`bg-amber-500` light / `bg-amber-950` dark with `border-2 border-amber-400`).
  - Text: High-contrast dark amber/white typography (`text-amber-950 dark:text-amber-100 font-extrabold`).
  - Icon: AlertTriangle / ShieldAlert (64px, `text-amber-300`).
  - Header: **"ACCESO CON ADVERTENCIA"** (text-3xl uppercase).
  - Warning Callout: *"Paso autorizado según política de planta. Notificar al contratista para regularizar."*
  - Reason Box: Prominent amber callout stating exact warning (e.g. *"Aviso: Póliza SUA vence en 3 días (04/Oct/2026)"*).
  - Badge: `[ 🟡 ADVERTENCIA ADMINISTRATIVA ]`

#### 4.1.3 Tier 3: Red / Denied (`denied`)
- **Semantic Meaning:** Mandatory document expired, missing, rejected, or contractor company blocked. Worker **MUST NOT ENTER**.
- **Visual Styling:**
  - Container: High-impact crimson red background (`bg-rose-600` light / `bg-rose-950` dark with `border-2 border-rose-500 animate-pulse-subtle`).
  - Text: High-contrast white typography (`text-white font-extrabold`).
  - Icon: OctagonX / ShieldX / Ban (64px, `text-rose-200`).
  - Header: **"ACCESO DENEGADO"** (text-3xl / text-4xl uppercase).
  - Subtitle: *"ENTRADA PROHIBIDA. Detener al trabajador en caseta."*
  - Denial Reason Alert: Bold high-contrast box (`bg-rose-900/80 border border-rose-400 p-4 rounded-lg`):
    - **Motivo de Bloqueo:** *"Constancia DC-3 vencida el 15/Sep/2026. Requiere recertificación antes de ingresar."*
  - Action Directive for Guard: *"Indicar al trabajador comunicarse con su supervisor de contratista o el departamento de EHS."*
  - Badge: `[ 🔴 BLOQUEADO EN CASETA ]`

---

### 4.2 Worker Profile Card Specification (`WorkerProfileCard.tsx`)

The Worker Profile Card displays the comprehensive physical identity and corporate affiliation of the scanned individual:

```
+-----------------------------------------------------------------------------------+
|  [ PHOTO ]   CARLOS ENRIQUE RODRÍGUEZ MORALES                 [ ACTIVO ]          |
|   120x120    Puesto: Soldador Especialista en Alturas         ID: 7b8f2d4e...     |
|              -------------------------------------------------------------------- |
|              CURP: ROMC850412HDFR02   [Copy]  |  NSS: 12856478901   [Copy]        |
|              -------------------------------------------------------------------- |
|              Empresa Contratista: Construcciones y Montajes del Norte S.A. de C.V.|
|              RFC: CMN980512AB3        |  REPSE: AR4859/2022   [Vigente]           |
|              Estatus con Planta: ACTIVO AUTORIZADO                                |
+-----------------------------------------------------------------------------------+
```

#### Field Breakdown:
1. **Avatar / Photo:**
   - 120x120px rounded rectangle with high-contrast border.
   - Served via Cloudflare R2 presigned URL or placeholder with worker initials.
2. **Full Name:**
   - Large bold title font (`text-2xl font-bold tracking-tight text-foreground`).
3. **Puesto / Job Title:**
   - Worker job title (`users.job_title` from `contractor_workers`), e.g., *"Técnico Electricista Media Tensión"*.
4. **Official Identification Codes:**
   - **CURP:** Rendered in monospace font (`font-mono tracking-wider`), includes quick copy-to-clipboard button.
   - **NSS (IMSS):** Monospace 11-digit number with copy button.
5. **Contractor Company Affiliation:**
   - **Nombre de Empresa:** Full legal entity name (`contractors.name`).
   - **RFC:** Mexican Federal Tax ID in monospace.
   - **REPSE:** Registro de Prestadores de Servicios Especializados registration number and badge.
   - **Relación con Planta:** Relationship status badge (`plant_contractors.status`: Activo / Suspendido / Bloqueado).

---

### 4.3 Document Validation Breakdown Table (`DocumentBreakdownTable.tsx`)

A structured, accessible data table listing all evaluated documents that determined the access outcome:

```
+---------------------------------------------------------------------------------------------------------+
| Documento                 | Ámbito      | Requerido | Estatus       | Vigencia / Expiración | Alerta    |
+---------------------------+-------------+-----------+---------------+-----------------------+-----------+
| Constancia DC-3           | Trabajador  | OBLIGATORIO| 🔴 VENCIDO    | 15/Sep/2026 (Hace 16d)| CAUSA BLOQ|
| Registro REPSE            | Contratista | OBLIGATORIO| 🟢 APROBADO   | 20/Dic/2027           | Vigente   |
| Pago SUA (IMSS/INFONAVIT) | Contratista | OBLIGATORIO| 🟢 APROBADO   | 20/Oct/2026           | Vigente   |
| Opinión SAT 32-D          | Contratista | OBLIGATORIO| 🟢 APROBADO   | 15/Oct/2026           | Vigente   |
| Identificación Oficial INE| Trabajador  | OBLIGATORIO| 🟢 APROBADO   | Sin Expiración        | Vigente   |
| Certificado Médico Aptitud| Trabajador  | OBLIGATORIO| 🟢 APROBADO   | 30/Nov/2026           | Vigente   |
+---------------------------------------------------------------------------------------------------------+
```

#### Table Column Specification:
1. **Documento:** Display name of document type (`document_types.name`), accompanied by its legal code badge (`DC3`, `REPSE`, `SUA`, `SAT_32D`, `INE`, `MEDICAL`).
2. **Ámbito (Pertenencia):**
   - `Trabajador` (individual worker credential).
   - `Empresa Contratista` (contractor company fiscal/labor document).
3. **Requerido (Obligatoriedad):**
   - `OBLIGATORIO` (`is_required = true` from plant requirements or system default).
   - `OPCIONAL` (informative document).
4. **Estatus de Validación:**
   - `🟢 Aprobado / Vigente`: Approved by EHS manager and within validity date.
   - `🟡 Por Vencer`: Valid today, but expires within plant warning threshold (< 15 days).
   - `🔴 Vencido`: Expiration date is in the past (`expires_at < current_date`).
   - `🔴 Faltante`: Document required by plant but never uploaded by worker/contractor.
   - `🔵 En Revisión`: Uploaded, awaiting EHS manager review (`pending_review`).
   - `❌ Rechazado`: Rejected by EHS manager due to legibility/validity issues.
5. **Vigencia / Fecha de Expiración:**
   - Formatted localized date (e.g. `15/Sep/2026`).
   - Relative expiration counter (e.g., *"Vence en 12 días"*, *"Venció hace 5 días"*).
6. **Alerta / Causa de Bloqueo:**
   - Visual red tag flagging the specific document(s) causing access denial.

---

## 5. UI/UX Design Standards & Industrial Ergonomics

### 5.1 Industrial Aesthetics & Color Tokens

Per `DEVELOPMENT_RULES.md` §15 (`ui-styling`, `ui-ux-pro-max`), the Caseta interface avoids frivolous generic SaaS styling in favor of high-durability industrial clarity:
- **Typography:**
  - Headings & Primary Text: `Inter` or `Geist` with strict font-weight hierarchy (`font-semibold` / `font-bold`).
  - Machine Identifiers & Codes (UUIDs, CURP, NSS, RFC, REPSE): `JetBrains Mono` or `ui-monospace` for instant character differentiation (`0` vs `O`, `1` vs `l`).
- **Color Palette (CSS Variables):**
  - **Background Light:** `#F8FAFC` (Slate-50) | **Background Dark:** `#090D16` (Deep Charcoal Slate)
  - **Card Surface Light:** `#FFFFFF` | **Card Surface Dark:** `#0F172A` (Slate-900)
  - **Border Light:** `#E2E8F0` (Slate-200) | **Border Dark:** `#1E293B` (Slate-800)
  - **Compliance Granted (Green):**
    - Light: Background `#059669` (Emerald-600), Foreground `#FFFFFF`
    - Dark: Background `#064E3B` (Emerald-900), Border `#10B981` (Emerald-500), Foreground `#ECFDF5`
  - **Compliance Warning (Yellow):**
    - Light: Background `#D97706` (Amber-600), Foreground `#FFFFFF`
    - Dark: Background `#78350F` (Amber-900), Border `#F59E0B` (Amber-500), Foreground `#FFFBEB`
  - **Compliance Denied (Red):**
    - Light: Background `#E11D48` (Rose-600), Foreground `#FFFFFF`
    - Dark: Background `#4C0519` (Rose-950), Border `#F43F5E` (Rose-500), Foreground `#FFF1F2`

### 5.2 Dark / Light Mode Support

Industrial security booths frequently shift from intense daylight to complete darkness during night shifts.
- Supported via `next-themes` with `attribute="class"` on `<html>`.
- Default to `system` with a sticky `ThemeToggle` in the top header.
- Dark mode utilizes deep charcoal slate (`#0B0F19`) rather than pure `#000000` to prevent harsh OLED halation.

### 5.3 Sonner Toast Integration (`ask-sonner`)

Strictly adhering to `.agents/skills/ask-sonner/SKILL.md`:
1. **Single Root Mount:** In `app/layout.tsx`:
   ```tsx
   import { Toaster } from "sonner";

   export default function RootLayout({ children }: { children: React.ReactNode }) {
     return (
       <html lang="es" suppressHydrationWarning>
         <body className="min-h-screen bg-background font-sans antialiased">
           {children}
           <Toaster richColors position="top-right" closeButton />
         </body>
       </html>
     );
   }
   ```
2. **Notification Recipes in Caseta:**
   - Successful check-in: `toast.success("Acceso registrado en bitácora", { description: "Trabajador: Carlos Rodríguez" })`
   - Access denied: `toast.error("Acceso Denegado", { description: "DC-3 Vencido. Registro guardado en bitácora." })`
   - Copy CURP / NSS: `toast.info("CURP copiado al portapapeles")`
   - Gate change: `toast("Punto de acceso actualizado", { description: "Operando ahora en Caseta Sur" })`

### 5.4 Loading States & Skeletons

- During check-in mutation, the input disables briefly while displaying an inline spinner.
- Skeletons (`<Skeleton className="h-48 w-full" />`) replace profile and table areas during cold fetches.
- Status buttons transition smoothly without layout shifting.

### 5.5 Accessibility & Ergonomics (a11y)

- **Keyboard Navigation Sequence:**
  1. `Tab 1`: Barcode scanner input.
  2. `Tab 2`: "Verificar Acceso" button.
  3. `Tab 3`: Caseta switcher.
  4. `Tab 4`: Guard logout menu.
- **Screen Reader Announcements (`aria-live`):**
  - Result banner container includes `aria-live="assertive"` so that screen readers announce immediately:
    *"Acceso permitido para Carlos Rodríguez"* or *"Acceso denegado. Motivo: DC-3 vencido."*
- **Contrast Ratios:** All text elements exceed WCAG 2.1 AA requirement (minimum 4.5:1 for normal text, 3:1 for large text).

---

## 6. API Client Layer Structure (`lib/api`)

### 6.1 Core API Client (`lib/api/client.ts`)

Per `DEVELOPMENT_RULES.md` §4, components never execute raw `fetch` calls directly. All communications pass through a typed client layer encapsulating base URL, authorization headers, and error envelop unwrapping:

```typescript
// lib/api/client.ts
export interface ApiSuccessEnvelope<T> {
  data: T;
  meta?: {
    page?: number;
    total?: number;
  };
}

export interface ApiErrorEnvelope {
  error: {
    code: string;
    message: string;
  };
}

export class ApiClientError extends Error {
  code: string;
  status: number;

  constructor(code: string, message: string, status: number) {
    super(message);
    this.name = "ApiClientError";
    this.code = code;
    this.status = status;
  }
}

const BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

function getAuthToken(): string | null {
  if (typeof document === "undefined") return null;
  const match = document.cookie.match(new RegExp("(^| )asesamo_token=([^;]+)"));
  return match ? decodeURIComponent(match[2]) : null;
}

export async function apiClient<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const token = getAuthToken();
  const headers = new Headers(options.headers || {});
  headers.set("Content-Type", "application/json");
  headers.set("Accept", "application/json");

  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const url = `${BASE_URL}${endpoint}`;
  const response = await fetch(url, {
    ...options,
    headers,
  });

  const json = await response.json();

  if (!response.ok) {
    const errorBody = json as ApiErrorEnvelope;
    const code = errorBody?.error?.code || "UNKNOWN_ERROR";
    const message = errorBody?.error?.message || "Ocurrió un error inesperado.";
    throw new ApiClientError(code, message, response.status);
  }

  const successBody = json as ApiSuccessEnvelope<T>;
  return successBody.data;
}
```

### 6.2 Data Transfer Objects (DTO) & TypeScript Interfaces (`lib/api/types.ts`)

```typescript
// lib/api/types.ts

export type AccessResult = "granted" | "granted_with_warning" | "denied";

export type DocumentStatus = "pending_review" | "approved" | "rejected" | "expired";

export type ContractorStatus = "pending" | "active" | "suspended" | "blocked";

export interface User {
  id: string;
  tenant_id: string;
  email: string;
  full_name: string;
  role: "super_admin" | "plant_admin" | "ehs_manager" | "guard" | "contractor_admin" | "worker";
  avatar_url?: string | null;
}

export interface ContractorSummary {
  id: string;
  name: string;
  rfc?: string | null;
  repse_number?: string | null;
  status: ContractorStatus;
}

export interface WorkerSummary {
  id: string;
  full_name: string;
  curp?: string | null;
  imss_number?: string | null;
  job_title?: string | null;
  avatar_url?: string | null;
  contractor: ContractorSummary;
}

export interface EvaluatedDocument {
  document_type_code: string;
  document_type_name: string;
  applies_to: "worker" | "contractor";
  status: DocumentStatus | "missing";
  is_required: boolean;
  issued_at?: string | null;
  expires_at?: string | null;
  file_name?: string | null;
}

export interface CheckInRequest {
  worker_id: string;
  access_point_id: string;
}

export interface CheckInResponse {
  access_log_id: string;
  result: AccessResult;
  denial_reason?: string | null;
  checked_in_at: string;
  worker: WorkerSummary;
  plant: {
    id: string;
    name: string;
  };
  access_point: {
    id: string;
    name: string;
  };
  documents: EvaluatedDocument[];
  document_snapshot: EvaluatedDocument[];
}

export interface AccessLogItem {
  id: string;
  checked_in_at: string;
  result: AccessResult;
  denial_reason?: string | null;
  worker_name: string;
  contractor_name: string;
  access_point_name: string;
}

export interface AccessPoint {
  id: string;
  plant_id: string;
  tenant_id: string;
  name: string;
  description?: string | null;
  is_active: boolean;
}
```

### 6.3 Domain API Service Modules

#### `lib/api/auth.ts`:
```typescript
import { apiClient } from "./client";
import { User } from "./types";

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface LoginResult {
  token: string;
  user: User;
}

export const authApi = {
  login: (credentials: LoginCredentials) =>
    apiClient<LoginResult>("/api/v1/auth/login", {
      method: "POST",
      body: JSON.stringify(credentials),
    }),

  getCurrentUser: () => apiClient<User>("/api/v1/auth/me"),

  logout: async () => {
    if (typeof document !== "undefined") {
      document.cookie = "asesamo_token=; Max-Age=0; path=/;";
    }
  },
};
```

#### `lib/api/access.ts`:
```typescript
import { apiClient } from "./client";
import { CheckInRequest, CheckInResponse, AccessLogItem, AccessPoint } from "./types";

export const accessApi = {
  checkIn: (payload: CheckInRequest) =>
    apiClient<CheckInResponse>("/api/v1/access/check-in", {
      method: "POST",
      body: JSON.stringify(payload),
    }),

  getRecentAccessLogs: (accessPointId?: string) =>
    apiClient<AccessLogItem[]>(
      `/api/v1/access-logs?per_page=10${accessPointId ? `&access_point_id=${accessPointId}` : ""}`
    ),

  getAccessPoints: (plantId?: string) =>
    apiClient<AccessPoint[]>(
      `/api/v1/access-points${plantId ? `?plant_id=${plantId}` : ""}`
    ),
};
```

---

## 7. Concrete Next Steps for Implementation

1. **Scaffold Next.js 14 Web App (`asesamo-web`):**
   - Initialize with `create-next-app` (TypeScript strict, Tailwind CSS, App Router, ESLint).
   - Configure shadcn/ui components (`button`, `card`, `input`, `badge`, `table`, `dialog`, `skeleton`).
   - Install TanStack Query v5, `sonner`, `lucide-react`, `next-themes`.
2. **Build API Client & Types (`lib/api/*`):**
   - Implement typed client matching Go backend endpoints.
3. **Build Guard Login View (`(auth)/login`):**
   - Implement `LoginForm` with Zod validation, cookie storage, and guard role redirect.
4. **Build Caseta Check-In View (`(dashboard)/caseta`):**
   - Implement `CasetaCheckIn`, `WorkerScannerInput`, `ComplianceResultBanner`, `WorkerProfileCard`, and `DocumentBreakdownTable`.
   - Wire up Quick-fill demo simulation buttons for all test scenarios.
5. **Verify End-to-End with Backend & Docker Compose:**
   - Execute test check-ins matching backend seed data.

---
*Report compiled and verified by `teamwork_preview_spec_miner_survey_2` for Phase 1 MVP delivery.*
