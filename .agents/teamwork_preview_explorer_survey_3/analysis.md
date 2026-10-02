# UI/UX Design System & Component Architecture Survey — APPCITAS Phase 1 (El Cimiento)

**Document ID**: `APPCITAS-SURVEY-03-UIUX`  
**Date**: 2026-09-14  
**Author**: `teamwork_preview_explorer_survey_3`  
**Target Project**: APPCITAS — Multitenant Physiotherapy Clinic Management Platform  
**Target Milestone**: Phase 1 ("El Cimiento") MVP  
**Reference Directives**: `ORIGINAL_REQUEST.md`, `Roadmap.md`, installed skills (`ui-ux-pro-max`, `apple-design`, `minimalist-ui`, `ask-sonner`).

---

## Executive Summary

APPCITAS Phase 1 requires an exceptional, medical-grade web interface for physiotherapy clinic owners and staff. This survey delivers the complete UI/UX design tokens, layout shell, page component architectures, interactive dialogs, feedback systems, and responsive viewport strategies.

Synthesizing principles from:
1. **`ui-ux-pro-max`**: "Accessible & Ethical" pattern, WCAG 2.1 AA/AAA compliance, high-contrast typography (16px+ base, 4.5:1+ text contrast), 44x44px touch targets, zero emoji-as-icon anti-patterns, and semantic DOM structure.
2. **`apple-design`**: WWDC *Designing Fluid Interfaces* principles — instant touch-down micro-response (`active:scale-[0.98]`), translucent glassmorphic surfaces (`backdrop-blur-md bg-white/80 border-slate-200/80`), spring-eased transitions, spatial consistency, and calm depth.
3. **`minimalist-ui`**: Utilitarian editorial restraint, warm clean canvas (`#F8FAFC`), crisp flat bento cards (`1px solid #E2E8F0`, rounded corners, zero murky drop shadows), and absolute prohibition of generic AI tropes (no purple/pink glowing gradients, no generic placeholders).
4. **`ask-sonner`**: Single root `<Toaster />` integration, rich semantic status toasts, clipboard copy integrations for instant temporary credentials, and zero cascade conflicts.

---

## 1. Design System & Aesthetic Foundation

### 1.1 Aesthetic Philosophy: "Calm Clinical Precision"
Physiotherapy software is used by clinic directors, physiotherapists with busy patient schedules, and administrative staff. It must inspire clinical trust, serenity, and lightning-fast efficiency. 
- **Palette Mood**: Calm neutrals (slate/zinc) paired with soft medical teal accents (`#0D9488` / `#0891B2`), clean white card surfaces, and subtle emerald badges for active health statuses.
- **Depth & Surfaces**: Layered translucent chrome over an ultra-light matte canvas (`bg-slate-50`). No aggressive skeuomorphism; depth is communicated via subtle 1px border dividers (`border-slate-200/80`) and ultra-diffuse ambient elevations (`shadow-sm` max).
- **Typography Hierarchy**: Crisp, legible, modern geometric sans-serif (Inter / Geist Sans) with deliberate tabular figures for currency and duration metrics.

### 1.2 Color Token Architecture (Tailwind CSS Mapping)

| Token Name | Hex Code | Tailwind Class | Semantic Usage & Context |
| :--- | :--- | :--- | :--- |
| **Canvas Background** | `#F8FAFC` | `bg-slate-50` | Primary application canvas backdrop behind cards |
| **Surface Card** | `#FFFFFF` | `bg-white` | Content containers, bento cards, modal dialog shells |
| **Surface Hover** | `#F1F5F9` | `hover:bg-slate-100` | Table row hover, list item hover, subtle buttons |
| **Border Subtle** | `#E2E8F0` | `border-slate-200` | Card borders, dividers, table cell grid lines |
| **Border Hairline** | `#F1F5F9` | `border-slate-100` | Secondary internal section dividers |
| **Text Heading** | `#0F172A` | `text-slate-900` | High-contrast titles, metrics, clinic names (16:1 contrast) |
| **Text Body** | `#334155` | `text-slate-700` | Primary readable paragraphs, descriptions, input labels |
| **Text Muted** | `#64748B` | `text-slate-500` | Secondary metadata, dates, durations, helper captions |
| **Text Subtle** | `#94A3B8` | `text-slate-400` | Placeholders, inactive icons, breadcrumb separators |
| **Medical Teal Primary** | `#0D9488` | `bg-teal-600` / `text-teal-600` | Primary CTAs, active tab accents, key clinic highlights |
| **Medical Teal Hover** | `#0F766E` | `hover:bg-teal-700` | Button hover state (150ms ease-out) |
| **Teal Light Tint** | `#F0FDFA` | `bg-teal-50` | Active navigation pill, selected table rows, icon badge bg |
| **Teal Border Accent** | `#CCFBF1` | `border-teal-100` | Active selection border highlight |
| **Status Active (Green)** | `#16A34A` | `text-emerald-700 bg-emerald-50 border-emerald-200` | Active specialist status, paid order badge |
| **Status Pending (Amber)**| `#D97706` | `text-amber-700 bg-amber-50 border-amber-200` | Waitlist item, pending appointment, temporary credential |
| **Status Inactive/Danger**| `#DC2626` | `text-rose-700 bg-rose-50 border-rose-200` | Deactivated service, canceled appointment, error toast |
| **Interactive Focus** | `#14B8A6` | `focus-visible:ring-teal-500` | Accessible keyboard focus ring (2px offset) |

### 1.3 Typography Scale & Weights

- **Display Title (Dashboard Top)**: `text-2xl font-bold tracking-tight text-slate-900` (24px, 1.2 line height)
- **Section Heading**: `text-lg font-semibold tracking-tight text-slate-900` (18px)
- **Card Subheading / Metric Label**: `text-xs font-semibold uppercase tracking-wider text-slate-500`
- **Large Metric Figure**: `text-3xl font-extrabold tracking-tight text-slate-900 tabular-nums` (e.g. `$4,200`, `5 / 5`)
- **Body Standard**: `text-sm font-normal leading-relaxed text-slate-700` (14px)
- **Captions & Tags**: `text-xs font-medium text-slate-500` (12px)
- **Monospace Code/Credentials**: `font-mono text-xs text-slate-800 bg-slate-100 px-2 py-0.5 rounded border border-slate-200`

### 1.4 Fluid Interaction & Feedback Rules (`apple-design` & `minimalist-ui`)
- **Active State Scale**: All clickable buttons and cards feature `active:scale-[0.98] transition-transform duration-150 ease-out`.
- **Hover Transitions**: Smooth 150ms-200ms transitions on color, background, and border (`transition-colors duration-150`).
- **Focus Rings**: Strict WCAG compliance with visible keyboard focus rings (`focus:outline-none focus-visible:ring-2 focus-visible:ring-teal-500 focus-visible:ring-offset-2`).
- **Banned Elements Checklist**:
  - NO emojis anywhere in the UI. All icons are crisp SVG primitives from `lucide-react` with `strokeWidth={1.75}`.
  - NO purple/pink "AI SaaS" radial gradients or neon cards.
  - NO blurry heavy drop shadows (`shadow-2xl` banned on cards; use `shadow-sm` or `border border-slate-200`).
  - NO placeholder copy like "Lorem Ipsum" or "John Doe". All seed mockups use authentic physiotherapy terms ("Clínica Fisia", "Dra. Sofía Navarro", "Rehabilitación Deportiva", "Punción Seca").

---

## 2. Dashboard Shell & Layout Architecture

### 2.1 File Path & Component Structure
- **Root Dashboard Layout**: `src/app/dashboard/layout.tsx`
- **Sidebar Component**: `src/components/dashboard/Sidebar.tsx`
- **Topbar Component**: `src/components/dashboard/Topbar.tsx`
- **Mobile Nav Sheet**: `src/components/dashboard/MobileNav.tsx`
- **Tenant Context Indicator**: `src/components/dashboard/ClinicBadge.tsx`
- **User Profile Menu**: `src/components/dashboard/UserNav.tsx`

### 2.2 Wireframe / Layout Composition

```
+-----------------------------------------------------------------------------------------------+
| APPCITAS Clinic Shell                                                                         |
+----------------------+------------------------------------------------------------------------+
| [Logo] CLÍNICA FISIA | [☰ Mobile] Dashboard > Equipo     [Ver Enlace Público ↗] [Dr. Carlos ▼] |
| Fisioterapia & Rehab | (Topbar: bg-white/80 backdrop-blur-md border-b border-slate-200)        |
+----------------------+------------------------------------------------------------------------+
| ❖ Resumen (Overview) | MAIN CONTENT CONTAINER (max-w-7xl mx-auto p-6 lg:p-8 space-y-8)         |
|                      |                                                                        |
| 👥 Equipo (5 Spec.)  | [Page Title & Breadcrumb]                 [+ Primary Action Button]    |
|                      |                                                                        |
| 🩺 Catálogo Servicios| +------------------+------------------+------------------+-------------+ |
|                      | | Metric Card 1    | Metric Card 2    | Metric Card 3    | Metric 4    | |
| 📋 Corte de Caja     | +------------------+------------------+------------------+-------------+ |
|                      |                                                                        |
| ⚙ Configuración      | [Interactive Bento Section / Data Tables / Filtering Grid]             |
|                      |                                                                        |
+----------------------+                                                                        |
| [User Card]          |                                                                        |
| Dr. Carlos Fisia     |                                                                        |
| [Cerrar Sesión ⎋]    |                                                                        |
+----------------------+------------------------------------------------------------------------+
```

### 2.3 Responsive Layout Engine
- **Desktop Viewport (`>= 1024px` / `lg`)**:
  - Persistent left sidebar: `w-64 fixed inset-y-0 left-0 z-30 bg-white border-r border-slate-200 flex flex-col justify-between`.
  - Main container offset: `pl-64 flex flex-col min-h-screen bg-slate-50`.
  - Sticky header: `sticky top-0 z-20 h-16 bg-white/80 backdrop-blur-md border-b border-slate-200/80 px-8 flex items-center justify-between`.
- **Tablet / Mobile Viewport (`< 1024px`)**:
  - Sidebar collapses off-canvas (`transform -translate-x-full` transition with spring easing).
  - Topbar renders a 44x44px accessible hamburger trigger (`Menu` icon) with `aria-label="Abrir menú de navegación"`.
  - Clicking hamburger triggers `MobileNav` sheet with backdrop scrim (`bg-slate-900/40 backdrop-blur-sm fixed inset-0 z-40`).
  - Swiping left or tapping outside dismisses the sheet instantly without latency.

### 2.4 Navigation Items Specification

| Navigation Item | Route | Lucide Icon | Active Styling | Inactive Styling |
| :--- | :--- | :--- | :--- | :--- |
| **Resumen General** | `/dashboard` | `LayoutDashboard` | `bg-teal-50 text-teal-700 font-semibold border-r-2 border-teal-600` | `text-slate-600 hover:bg-slate-50 hover:text-slate-900 font-medium` |
| **Equipo Especialistas** | `/dashboard/specialists` | `Users` | `bg-teal-50 text-teal-700 font-semibold border-r-2 border-teal-600` | `text-slate-600 hover:bg-slate-50 hover:text-slate-900 font-medium` |
| **Catálogo de Servicios**| `/dashboard/services` | `Stethoscope` | `bg-teal-50 text-teal-700 font-semibold border-r-2 border-teal-600` | `text-slate-600 hover:bg-slate-50 hover:text-slate-900 font-medium` |
| **Caja & Cierre Diario** | `/dashboard/orders` | `Receipt` | `bg-teal-50 text-teal-700 font-semibold border-r-2 border-teal-600` | `text-slate-600 hover:bg-slate-50 hover:text-slate-900 font-medium` |
| **Configuración Clínica**| `/dashboard/settings` | `Settings` | `bg-teal-50 text-teal-700 font-semibold border-r-2 border-teal-600` | `text-slate-600 hover:bg-slate-50 hover:text-slate-900 font-medium` |

---

## 3. Detailed Page Component Specifications

### 3.1 Overview Page (`/dashboard/page.tsx`)

#### Purpose
Provides the clinic owner with an instant situational snapshot: team capacity (5 specialists), service availability, mock appointment flow (Phase 2 foundation), and daily payment totals (Phase 1 cash closing foundation).

#### Key Components & Sections
1. **Header Banner with Clinic Identity**:
   - Welcome greeting: "Bienvenido, Dr. Carlos" with live date formatted in Spanish ("Lunes, 14 de Septiembre de 2026").
   - Clinic badge: `Clínica Fisia` with a green pulsing pulse dot ("En línea / Operando").
   - Quick Action: Direct button to "Ver enlace público" with `ExternalLink` icon.
2. **KPI Metric Bento Cards (4-Column Responsive Grid)**:
   - Card 1: **Equipo de Especialistas**
     - Metric: `5 / 5`
     - Status: `Activos al 100%`
     - Icon: `Users` in `bg-teal-50 text-teal-600 p-2.5 rounded-xl`
     - Helper: "Capacidad de plantilla completa"
     - Action: Link to `/dashboard/specialists`
   - Card 2: **Servicios en Catálogo**
     - Metric: `6 servicios`
     - Status: `Todos asignados`
     - Icon: `Activity` in `bg-sky-50 text-sky-600 p-2.5 rounded-xl`
     - Helper: "Tarifa promedio: $720 MXN"
     - Action: Link to `/dashboard/services`
   - Card 3: **Citas Programadas (Hoy)**
     - Metric: `8 citas`
     - Status: `Próxima en 45 min`
     - Icon: `CalendarClock` in `bg-emerald-50 text-emerald-600 p-2.5 rounded-xl`
     - Helper: "Fase 2: Conexión Google Calendar"
   - Card 4: **Corte de Caja Diario (Estimado)**
     - Metric: `$4,200 MXN`
     - Status: `8 órdenes registradas`
     - Icon: `WalletCards` in `bg-purple-50 text-purple-600 p-2.5 rounded-xl`
     - Helper: "Efectivo: $2,100 | Tarjeta: $1,400 | Transfer: $700"
3. **Quick Action Launcher Bar**:
   - Three high-visibility action tiles:
     - `+ Incorporar Especialista` (opens instant onboarding modal)
     - `+ Registrar Nuevo Servicio` (opens service creation modal)
     - `🔗 Compartir Enlace de Citas WhatsApp` (copies `appcitas.com/clinica-fisia` to clipboard with Sonner feedback)
4. **Dual Bento Grid: Team Roster Snapshot + Active Service Distribution**:
   - Left Column (Team Preview): Compact list of the 5 specialists with avatar, specialty tag, and live availability badge.
   - Right Column (Services Snapshot): Top booked services with duration tag, price tag, and specialist assignment avatars.

---

### 3.2 Team Roster Page (`/dashboard/specialists/page.tsx`)

#### Purpose
Dedicated workstation for managing the 5 specialist physiotherapists, tracking their specialty certifications, active/inactive statuses, and onboarding new therapists with instant temporary credentials.

#### Page Layout & Elements
1. **Action Bar**:
   - Search & filter: Real-time search bar ("Filtrar por nombre o especialidad...") + specialty drop-down.
   - Counter Badge: `5 de 5 Fisioterapeutas Registrados` (with visual indicator of maximum Phase 1 quota).
   - Primary CTA: `+ Incorporar Especialista` (`bg-teal-600 hover:bg-teal-700 text-white font-medium rounded-xl px-4 py-2.5 shadow-sm active:scale-[0.98] transition flex items-center gap-2`).
2. **Specialist Bento Cards Grid (3 Columns on Desktop, 2 on Tablet, 1 on Mobile)**:
   - Each specialist card features:
     - **Card Header**:
       - Avatar: Circular 48x48px container with initials in bold medical pastel (`bg-teal-100 text-teal-800 font-bold`).
       - Status Indicator: Subtle emerald status pill `Activo` with small SVG dot.
     - **Physiotherapist Details**:
       - Full Name: `text-base font-semibold text-slate-900` (e.g. "Dra. Sofía Navarro").
       - Email: `text-xs text-slate-500 font-normal flex items-center gap-1.5` (`Mail` icon).
       - Phone: `text-xs text-slate-500 font-normal flex items-center gap-1.5` (`Phone` icon).
     - **Specialty Badge**:
       - Custom pastel pill: `bg-slate-100 text-slate-800 text-xs font-semibold px-2.5 py-1 rounded-full w-fit border border-slate-200/60` (e.g., "Fisioterapia Deportiva y Readaptación").
     - **Services Assigned Counter**:
       - `Layers` icon + "3 servicios vinculados en catálogo".
     - **Action Footer**:
       - Button: "Copiar Credenciales" (`KeyRound` icon, copies login email + temp password with Sonner toast).
       - Button: "Editar" (`Pencil` icon).
       - Status Toggle: Accessible switch to toggle `Activo / Inactivo`.

#### Onboard Specialist Modal Dialog (`src/components/specialists/OnboardModal.tsx`)
1. **UX Problem Solved**: Eliminates external SMTP dependency. The clinic owner creates the account directly and is immediately presented with temporary credentials to share via WhatsApp or in-person.
2. **Visual & Interaction Specs**:
   - Scrim: `fixed inset-0 bg-slate-900/35 backdrop-blur-sm z-50 flex items-center justify-center p-4 animate-in fade-in duration-200`.
   - Dialog Container: `bg-white rounded-2xl border border-slate-200 shadow-2xl max-w-lg w-full p-6 space-y-6 animate-in zoom-in-95 duration-200`.
   - Header:
     - Title: `Incorporar Nuevo Especialista`
     - Subtitle: `Crea la cuenta del fisioterapeuta y obtén credenciales inmediatas para entrega directa.`
     - Close Button: `X` icon in `hover:bg-slate-100 rounded-lg p-1.5`.
   - Form Inputs:
     - `Nombre y Apellidos`: Input (placeholder: "ej. Dr. Roberto Méndez").
     - `Correo Electrónico`: Input type="email" (placeholder: "ej. roberto@fisia.com").
     - `Especialidad Principal`: Select or Combobox ("Fisioterapia Deportiva", "Terapia Manual Ortopédica", "Rehabilitación Neurológica", "Fisioterapia Pediátrica", "Readaptación Funcional").
     - `Teléfono Móvil`: Input type="tel" (placeholder: "+52 55 1234 5678").
   - **Auto-Generated Credential Panel**:
     - Visual callout box: `bg-slate-50 border border-slate-200 rounded-xl p-4 space-y-2`.
     - Label: `Contraseña Temporal Generada`.
     - Code Display: `font-mono text-sm font-semibold text-slate-800 tracking-wider bg-white px-3 py-2 rounded-lg border border-slate-200 flex items-center justify-between`.
     - Value: Auto-populated string (e.g. `Fisia-2026-Tr8#`).
     - Refresh Button: `RotateCw` icon to randomize a new temporary password.
     - "Copiar Datos de Acceso" Button: Copies formatted message:
       ```
       Hola Dr(a). Roberto, tus accesos a APPCITAS Clínica Fisia:
       Portal: https://appcitas.com/login
       Usuario: roberto@fisia.com
       Contraseña temporal: Fisia-2026-Tr8#
       ```
   - Modal Actions:
     - Secondary: "Cancelar" (`hover:bg-slate-100 text-slate-600 px-4 py-2 rounded-xl text-sm font-medium`).
     - Primary: "Registrar Especialista" (`bg-teal-600 hover:bg-teal-700 text-white px-5 py-2 rounded-xl text-sm font-semibold active:scale-[0.98] transition shadow-sm`).

---

### 3.3 Services Catalog Page (`/dashboard/services/page.tsx`)

#### Purpose
Allows the clinic owner to manage treatment offerings, configure pricing and appointment durations, and selectively assign which specialists can deliver each service.

#### Key Components & Sections
1. **Catalog Controls Bar**:
   - Search input: "Buscar servicio o tratamiento..."
   - Filter pills: "Todos (6)", "Terapia Manual (2)", "Deportiva (2)", "Especializada (2)".
   - Action Button: `+ Registrar Servicio` (opens modal to create a new clinical offering).
2. **Service Cards / Catalog List**:
   - Each service card contains:
     - **Service Title & Tag**:
       - Title: `text-lg font-semibold text-slate-900` (e.g., "Rehabilitación Deportiva y Biomecánica").
       - Category Badge: `text-xs font-medium text-slate-600 bg-slate-100 px-2.5 py-0.5 rounded-full`.
     - **Clinical Description**:
       - `text-sm text-slate-600 leading-relaxed` (e.g., "Sesión intensiva enfocada en lesiones musculares, tendinosas y reincorporación al entrenamiento deportivo").
     - **Key Attributes Strip**:
       - Duration: `Clock` icon + `50 min` in `bg-sky-50 text-sky-700 px-2.5 py-1 rounded-lg text-xs font-semibold flex items-center gap-1.5`.
       - Price: `Tag` icon + `$850 MXN` in `bg-emerald-50 text-emerald-700 px-2.5 py-1 rounded-lg text-xs font-bold flex items-center gap-1.5`.
     - **Specialist Assignment Matrix (Interactive Multi-Select / Checkbox Group)**:
       - Header: `Fisioterapeutas Habilitados:` (`text-xs font-semibold text-slate-500 uppercase tracking-wider`).
       - Specialist Selection Tiles: Checkbox or pill toggle for each of the 5 specialists.
       - Each option shows:
         - Avatar initials.
         - Specialist name ("Dra. Sofía Navarro").
         - Checked indicator (Teal checkbox with checkmark).
       - **Immediate Feedback**: Toggling a specialist triggers an optimistic UI update, posts to `/api/services/[id]/specialists`, and displays an unobtrusive Sonner toast (`toast.success("Asignación actualizada: Dra. Sofía habilitada para Rehabilitación Deportiva")`).
     - **Card Footer**:
       - "Editar Parámetros" (opens edit dialog).
       - "Estado": Switch toggle (`Activo / Pausado`).

#### Service Creation/Edit Modal (`src/components/services/ServiceModal.tsx`)
- Fields:
  1. `Nombre del Servicio` (text).
  2. `Descripción Clínica` (textarea).
  3. `Duración en Minutos` (number or radio buttons: 30 min, 45 min, 50 min, 60 min, 90 min).
  4. `Precio al Público ($ MXN)` (currency input with decimal format).
  5. `Asignación Inicial de Especialistas` (multi-select checklist of the 5 staff members, with "Seleccionar todos" shortcut).

---

## 4. Feedback & Notification System (`ask-sonner`)

### 4.1 Root Configuration
In accordance with `ask-sonner` best practices, **one `<Toaster />` is mounted once at the root layout** (`src/app/layout.tsx`). It is never mounted conditionally or per-page.

```tsx
// src/app/layout.tsx
import { Toaster } from 'sonner';

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="es" className="h-full bg-slate-50">
      <body className="h-full font-sans antialiased text-slate-900 selection:bg-teal-100 selection:text-teal-900">
        {children}
        <Toaster
          position="top-right"
          richColors
          closeButton
          duration={3500}
          offset={24}
          mobileOffset={{ top: 16, right: 16, left: 16 }}
          toastOptions={{
            className: "!rounded-2xl !border !border-slate-200 !shadow-lg !font-sans",
          }}
        />
      </body>
    </html>
  );
}
```

### 4.2 Standardized Toast Recipe Catalog

| Action Event | Sonner Call | Title & Description | Visual Feedback |
| :--- | :--- | :--- | :--- |
| **Authentication Success** | `toast.success()` | "Sesión iniciada" / "Bienvenido a Clínica Fisia" | Green success icon, auto-closes in 3s |
| **Authentication Failure** | `toast.error()` | "Error de acceso" / "Credenciales inválidas. Verifica tu correo y contraseña." | Red error icon, persistent for 5s |
| **Specialist Onboarded** | `toast.success()` | "Especialista incorporado" / "Dr. Roberto Méndez registrado. Credenciales temporales copiadas." | Green success + Action button: "Ver equipo" |
| **Credentials Copied** | `toast.info()` | "Credenciales copiadas" / "Información de acceso lista para enviar por WhatsApp." | Teal info icon, auto-closes in 2.5s |
| **Service Created** | `toast.success()` | "Servicio registrado" / "El tratamiento ya está disponible para agendamiento." | Green success icon |
| **Specialist Assigned** | `toast.success()` | "Asignación actualizada" / "Dra. Sofía Navarro vinculada al servicio." | Instant feedback on checkbox click |
| **Async Save Action** | `toast.promise()` | Loading: "Guardando cambios...", Success: "Cambios guardados con éxito", Error: "Error al guardar" | Loading spinner transitions into checkmark |
| **Public Link Copied** | `toast.success()` | "Enlace de citas copiado" / "appcitas.com/clinica-fisia listo para compartir." | Green success icon |

---

## 5. Viewport & Responsive Design Breakdown

### 5.1 Breakpoint Matrix

| Viewport Tier | Breakpoint Width | Layout Adaptation Strategy |
| :--- | :--- | :--- |
| **Mobile Portrait** | `< 640px` (`sm`) | Single column layout; full-width stacked bento cards; 44x44px minimum touch targets; fixed bottom action sheet or off-canvas drawer; topbar with hamburger and clinic title; toasts centered at top. |
| **Mobile Landscape / Small Tablet** | `640px - 767px` | 2-column KPI metric cards; full-width tables with horizontal swipe or compact card presentation; modal dialogs fill 90% width. |
| **Tablet Portrait** | `768px - 1023px` (`md`) | 2-column specialist grid; top navigation remains sticky; sidebar is off-canvas with accessible slide-over drawer; 2-column overview widgets. |
| **Desktop Standard** | `1024px - 1279px` (`lg`) | Persistent left sidebar (260px); sticky glassmorphic topbar; 4-column metric grid; 3-column specialist card grid; dual-column overview bento layout. |
| **Large Desktop / Widescreen** | `>= 1280px` (`xl`) | Max content container `max-w-7xl mx-auto`; generous whitespace (32px padding); balanced breathing room adhering to minimalist-ui macro-spacing principles. |

### 5.2 Mobile Navigation Drawer Interaction
- Triggered by `aria-expanded` hamburger button in topbar.
- Springs smoothly from `translateX(-100%)` to `translateX(0)` using cubic-bezier easing (`cubic-bezier(0.16, 1, 0.3, 1)` over 250ms).
- Traps keyboard focus within drawer when open.
- Closes automatically on route change or ESC key press.

---

## 6. Implementation Checklist & Verification Gates

### 6.1 Downstream Implementer Guidance
Downstream developer agents should implement the component tree with these explicit files:

```
src/
├── app/
│   ├── (auth)/
│   │   ├── login/page.tsx
│   │   └── register/page.tsx
│   ├── dashboard/
│   │   ├── layout.tsx                # Shell: Sidebar + Topbar + Toast Root
│   │   ├── page.tsx                  # Overview Page: Metrics + Quick Actions
│   │   ├── specialists/page.tsx      # Team Roster Page: 5 Specialists
│   │   ├── services/page.tsx         # Services Catalog & Specialist Assignments
│   │   └── settings/page.tsx         # Clinic profile & public link settings
│   ├── api/
│   │   ├── specialists/route.ts      # GET / POST specialist onboarding
│   │   └── services/route.ts         # GET / POST / PUT service & assignments
│   └── layout.tsx                    # Root layout mounting Sonner <Toaster />
└── components/
    ├── dashboard/
    │   ├── Sidebar.tsx               # Desktop navigation bar
    │   ├── Topbar.tsx                # Glassmorphic header & breadcrumbs
    │   ├── MobileNav.tsx             # Responsive off-canvas slide-out
    │   ├── MetricCard.tsx            # KPI bento container
    │   └── QuickActions.tsx          # Action button launchers
    ├── specialists/
    │   ├── SpecialistCard.tsx        # Individual therapist card
    │   ├── OnboardModal.tsx          # Instant credentials onboarding dialog
    │   └── CredentialsBox.tsx        # Monospace temp password + copy button
    ├── services/
    │   ├── ServiceCard.tsx           # Service card with price & duration
    │   ├── SpecialistSelector.tsx    # Interactive assignment checkbox grid
    │   └── ServiceModal.tsx          # Add/edit service dialog
    └── ui/
        ├── Button.tsx                # Fluid press scale button
        ├── Input.tsx                 # Accessible bordered input
        ├── Badge.tsx                 # Medical status pills
        └── Modal.tsx                 # Accessible backdrop-blur dialog shell
```

### 6.2 Acceptance Verification Commands
1. **Component Type Safety & Build**:
   `npm run build` must complete with 0 TypeScript, JSX, or Tailwind compilation errors.
2. **Lighthouse Accessibility Score**:
   Dashboard routes must score >= 95 on Lighthouse Accessibility (color contrast, ARIA landmarks, focus rings).
3. **Viewport Responsiveness**:
   Verify layout at `375px` (iPhone), `768px` (iPad), and `1440px` (Desktop) without horizontal overflow or clipped cards.

---
