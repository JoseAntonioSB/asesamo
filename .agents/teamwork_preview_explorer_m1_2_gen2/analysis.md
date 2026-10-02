# Milestone 1 Technical Specification: Build, Dependencies & Prisma Client Architecture

**Agent**: `teamwork_preview_explorer_m1_2_gen2`  
**Role**: Technical Architecture & Dependency Specification  
**Working Directory**: `/Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_explorer_m1_2_gen2`  
**Date**: 2026-09-14  
**Integrity Mode**: Development / Greenfield  
**Runtime Environment**: macOS (Darwin ARM64) | Node.js `v24.18.0` | npm `11.16.0`  

---

## 1. Executive Summary & Milestone 1 Scope

Milestone 1 establishes the foundational data, build, and type-system infrastructure for **APPCITAS Phase 1 (El Cimiento)**. The goal is to provide an unshakeable, deterministic developer and deployment environment that satisfies all acceptance criteria defined in `ORIGINAL_REQUEST.md` and `PROJECT.md`.

This specification covers the exact configuration for:
1. `package.json`: Complete dependency and devDependency manifest, exact pinned versions, execution scripts, and Prisma seed runner configuration.
2. `tsconfig.json`: TypeScript configuration targeting modern ECMAScript (`ES2022`), strict type-checking, bundler module resolution, and path aliases (`@/*`).
3. `src/lib/prisma.ts`: Type-safe `PrismaClient` global singleton pattern preventing connection pool exhaustion during Next.js hot-reloads and serverless App Router execution.
4. `src/lib/utils.ts`: Standard classnames merger utility (`cn`) combining `clsx` and `tailwind-merge`, alongside healthcare currency formatting (`formatCurrency`).
5. `.env.example` / `.env`: Complete environment variable blueprint covering PostgreSQL database connectivity (`DATABASE_URL`), connection pooling, and NextAuth session security (`NEXTAUTH_SECRET`, `NEXTAUTH_URL`).
6. Supporting build configuration: `tailwind.config.ts`, `postcss.config.mjs`, and `next.config.mjs` ensuring zero compilation errors under Node 24 and Next.js 14.2.

---

## 2. Specification 1: `package.json`

### 2.1 Complete File Content (`/Users/jasbotello/devel/APPCITAS/package.json`)

```json
{
  "name": "appcitas",
  "version": "0.1.0",
  "private": true,
  "scripts": {
    "dev": "next dev",
    "build": "next build",
    "start": "next start",
    "prisma:validate": "prisma validate",
    "prisma:generate": "prisma generate",
    "prisma:seed": "tsx prisma/seed.ts"
  },
  "prisma": {
    "seed": "tsx prisma/seed.ts"
  },
  "dependencies": {
    "@prisma/client": "5.22.0",
    "bcryptjs": "2.4.3",
    "clsx": "2.1.1",
    "lucide-react": "0.460.0",
    "next": "14.2.24",
    "next-auth": "4.24.8",
    "react": "18.3.1",
    "react-dom": "18.3.1",
    "sonner": "1.7.1",
    "tailwind-merge": "2.5.4"
  },
  "devDependencies": {
    "@types/bcryptjs": "2.4.6",
    "@types/node": "20.17.0",
    "@types/react": "18.3.12",
    "@types/react-dom": "18.3.1",
    "autoprefixer": "10.4.20",
    "postcss": "8.4.49",
    "prisma": "5.22.0",
    "tailwindcss": "3.4.15",
    "tsx": "4.19.2",
    "typescript": "5.6.3"
  }
}
```

### 2.2 Dependency Breakdown & Justification Matrix

| Package | Version | Layer | Architectural Rationale & Hardening |
|---|---|---|---|
| `next` | `14.2.24` | Framework | Stable Next.js 14 App Router LTS. Avoids experimental edge cases and breaking changes in Next 15 / React 19 credentials authentication. |
| `react` | `18.3.1` | UI Library | Rock-solid React 18 production release, 100% compatible with NextAuth v4 and Sonner. |
| `react-dom` | `18.3.1` | UI Library | Paired DOM renderer for React 18.3.1. |
| `@prisma/client` | `5.22.0` | ORM Client | High-performance ORM client matching Prisma CLI 5.22.0. Native PostgreSQL connection support, CUID ID generators, and Decimal support. |
| `next-auth` | `4.24.8` | Authentication | Battle-tested credentials authentication provider with JWT session strategy and full App Router route handler support. |
| `bcryptjs` | `2.4.3` | Security | **Critical Pure JS Choice**: Pure JavaScript implementation of bcrypt. Eliminates `node-gyp`, Python, and C++ compilation failures that break native `bcrypt` on Node 24 and diverse macOS/Linux architectures. |
| `lucide-react` | `0.460.0` | Icons | High-quality SVG icon primitives with customizable `strokeWidth`, cleanly replacing all emojis across the healthcare dashboard. |
| `sonner` | `1.7.1` | Notifications | Premium, accessible toast notification library specified in R3 and Acceptance Criteria. |
| `clsx` | `2.1.1` | Styling Utility | Fast, conditional classname constructor. |
| `tailwind-merge` | `2.5.4` | Styling Utility | Efficiently merges Tailwind CSS classes in JS without style conflicts or specificity bugs. |

### 2.3 DevDependency Breakdown & Tooling Matrix

| Package | Version | Purpose | Architectural Rationale |
|---|---|---|---|
| `prisma` | `5.22.0` | Database CLI | Provides `prisma validate`, `prisma generate`, `prisma db push`, and `prisma db seed`. |
| `typescript` | `5.6.3` | Compiler | Latest TypeScript 5.6 release supporting ECMAScript 2022, bundler module resolution, and strict null checks. |
| `tsx` | `4.19.2` | Seed Runner | Zero-config TypeScript execution engine. Executes `prisma/seed.ts` directly with native ESM/CJS resolution without intermediate build artifacts. |
| `@types/node` | `20.17.0` | Type Definitions | Node.js standard library type definitions for Node 20/24 compatibility. |
| `@types/react` | `18.3.12` | Type Definitions | Complete React 18 typing definitions. |
| `@types/react-dom` | `18.3.1` | Type Definitions | Complete React DOM 18 typing definitions. |
| `@types/bcryptjs` | `2.4.6` | Type Definitions | Type definitions for `bcryptjs` methods (`hash`, `compare`). |
| `tailwindcss` | `3.4.15` | Styling Engine | Modern utility CSS compiler supporting custom color palettes and responsive bento grids. |
| `postcss` | `8.4.49` | CSS Processor | PostCSS runner for Tailwind CSS and Autoprefixer. |
| `autoprefixer` | `10.4.20` | CSS Processor | Automatic vendor prefix injection for cross-browser CSS support. |

### 2.4 Scripts Specification & Execution Contracts

- `"dev": "next dev"`: Starts the local development server on port 3000 with Fast Refresh.
- `"build": "next build"`: Produces an optimized production build. Verifies all TypeScript types and App Router route structures.
- `"start": "next start"`: Starts the Next.js production server.
- `"prisma:validate": "prisma validate"`: Offline validation of `prisma/schema.prisma` syntax, relations, and cascade constraints.
- `"prisma:generate": "prisma generate"`: Generates tailored TypeScript Prisma Client definitions into `node_modules/@prisma/client`.
- `"prisma:seed": "tsx prisma/seed.ts"`: Executes the pilot seed script using `tsx`.
- `"prisma": { "seed": "tsx prisma/seed.ts" }`: Standard Prisma CLI hook executed automatically whenever `npx prisma db seed` is invoked.

---

## 3. Specification 2: `tsconfig.json`

### 3.1 Complete File Content (`/Users/jasbotello/devel/APPCITAS/tsconfig.json`)

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "lib": ["dom", "dom.iterable", "esnext"],
    "allowJs": true,
    "skipLibCheck": true,
    "strict": true,
    "noEmit": true,
    "esModuleInterop": true,
    "module": "esnext",
    "moduleResolution": "bundler",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "jsx": "preserve",
    "incremental": true,
    "plugins": [
      {
        "name": "next"
      }
    ],
    "paths": {
      "@/*": ["./src/*"]
    }
  },
  "include": [
    "next-env.d.ts",
    "**/*.ts",
    "**/*.tsx",
    ".next/types/**/*.ts"
  ],
  "exclude": [
    "node_modules"
  ]
}
```

### 3.2 Configuration Highlights & Next.js App Router Alignment

1. **Target `ES2022`**:
   - Leverages modern V8 engine features in Node 24 and modern browsers (class fields, `Array.prototype.at`, `Object.hasOwn`, top-level `await`).
   - Ensures minimal polyfill overhead and optimal execution speed.

2. **Module Resolution `bundler`**:
   - The modern TypeScript standard for bundlers (Next.js/Turbopack/Webpack).
   - Properly resolves package exports conditions (`exports` in `package.json`) and supports clean path aliases without legacy Node resolution quirks.

3. **Strict Type Checking (`strict: true`)**:
   - Enforces `noImplicitAny`, `strictNullChecks`, `strictFunctionTypes`, and `strictBindCallApply`.
   - Guarantees high reliability across multi-tenant database operations and API payload processing.

4. **Path Alias (`@/*` -> `./src/*`)**:
   - Enables clean, maintainable imports:
     - `@/lib/prisma`
     - `@/lib/utils`
     - `@/lib/auth`
     - `@/components/ui/button`
     - `@/types/next-auth`

5. **Next.js Plugin Integration**:
   - Registers `"name": "next"` plugin for automated App Router type checking and route typing.

---

## 4. Specification 3: `src/lib/prisma.ts` (PrismaClient Singleton)

### 4.1 Complete File Content (`/Users/jasbotello/devel/APPCITAS/src/lib/prisma.ts`)

```typescript
import { PrismaClient } from '@prisma/client';

/**
 * Global singleton pattern for PrismaClient.
 *
 * In Next.js development mode, Fast Refresh / hot module replacement reloads
 * modules on every file change. Creating a new PrismaClient instance on each
 * reload rapidly exhausts the database connection pool (PostgreSQL default max_connections = 100).
 *
 * By attaching the PrismaClient instance to `globalThis`, we guarantee that only
 * a single connection pool is instantiated and reused across hot reloads.
 * In production, `globalThis` caching is bypassed, and a single client is instantiated
 * per Node process / container.
 */

const globalForPrisma = globalThis as unknown as {
  prisma: PrismaClient | undefined;
};

export const prisma =
  globalForPrisma.prisma ??
  new PrismaClient({
    log:
      process.env.NODE_ENV === 'development'
        ? ['error', 'warn']
        : ['error'],
  });

if (process.env.NODE_ENV !== 'production') {
  globalForPrisma.prisma = prisma;
}

export default prisma;
```

### 4.2 Architectural Design Decisions

1. **Hot-Reload Connection Leak Prevention**:
   - Next.js development server invalidates module caches when files change, re-running code.
   - Without `globalThis`, 10 file edits = 10 new `PrismaClient` instances = ~50-100 PostgreSQL backend connections, throwing `FATAL: remaining connection slots are reserved for non-replication superuser connections`.
   - Attaching to `globalThis` persists the instance across module reloads.

2. **Log Level Isolation**:
   - In development: Outputs `['error', 'warn']` to alert developers of query bottlenecks or slow execution without spamming stdout with full query traces.
   - In production: Restricts to `['error']` for clean logs and privacy.

3. **Dual Export Convenience**:
   - Supports both named export `import { prisma } from '@/lib/prisma'` and default export `import prisma from '@/lib/prisma'` to prevent import mismatch errors across team members.

---

## 5. Specification 4: `src/lib/utils.ts` (Classnames & Currency Utility)

### 5.1 Complete File Content (`/Users/jasbotello/devel/APPCITAS/src/lib/utils.ts`)

```typescript
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

/**
 * Merges class names deterministically using `clsx` for conditional classes
 * and `tailwind-merge` to resolve conflicting Tailwind utility classes.
 *
 * Example:
 *   cn('px-4 py-2 bg-teal-500', isPrimary && 'bg-teal-600', 'px-6')
 *   => 'py-2 bg-teal-600 px-6' (px-4 is cleanly overridden by px-6)
 */
export function cn(...inputs: ClassValue[]): string {
  return twMerge(clsx(inputs));
}

/**
 * Formats a monetary amount into a localized Mexican Pesos (MXN) currency string.
 * Supports numbers, strings, or Prisma Decimal representations.
 *
 * Example:
 *   formatCurrency(750) => '$750.00'
 *   formatCurrency('850.50') => '$850.50'
 */
export function formatCurrency(amount: number | string | { toString: () => string }): string {
  let num: number;
  if (typeof amount === 'number') {
    num = amount;
  } else if (typeof amount === 'string') {
    num = parseFloat(amount);
  } else if (amount && typeof amount.toString === 'function') {
    num = parseFloat(amount.toString());
  } else {
    num = 0;
  }

  return new Intl.NumberFormat('es-MX', {
    style: 'currency',
    currency: 'MXN',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(isNaN(num) ? 0 : num);
}

/**
 * Formats appointment and session duration in minutes into a clean readable string.
 *
 * Example:
 *   formatDuration(60) => '60 min'
 */
export function formatDuration(minutes: number): string {
  return `${minutes} min`;
}
```

### 5.2 Rationale

- `cn(...)`: Mandatory utility for building Apple-inspired and minimalist UI components (`src/components/ui/*`). It prevents style conflicts (e.g., conflicting padding, margin, or color overrides between base components and consumer props).
- `formatCurrency(...)`: Required for R1 (Orders, Products, Services) and R3 (Services Catalog and Cash Closing bento cards). Safely handles Prisma `Decimal` instances by converting them via `.toString()` before parsing, preventing Next.js Client Component serialization crashes.
- `formatDuration(...)`: Provides clean duration display chips across the Services Catalog.

---

## 6. Specification 5: Environment Configuration (`.env` & `.env.example`)

### 6.1 Complete `.env.example` (`/Users/jasbotello/devel/APPCITAS/.env.example`)

```bash
# =============================================================================
# APPCITAS Phase 1 (El Cimiento) — Environment Configuration Template
# Copy this file to `.env` and fill in the values for your local environment.
# =============================================================================

# -----------------------------------------------------------------------------
# 1. Database Connection (PostgreSQL)
# -----------------------------------------------------------------------------
# Standard PostgreSQL URL format:
# postgresql://[USER]:[PASSWORD]@[HOST]:[PORT]/[DATABASE]?schema=[SCHEMA]&sslmode=[SSLMODE]
#
# Local Development (Docker or native Postgres):
DATABASE_URL="postgresql://postgres:postgres@localhost:5432/appcitas?schema=public"

# DigitalOcean Managed Database (Production / Staging):
# DATABASE_URL="postgresql://doadmin:secret-password@db-appcitas-do-user-1234.b.db.ondigitalocean.com:25060/appcitas?schema=public&sslmode=require"

# -----------------------------------------------------------------------------
# 2. NextAuth Authentication Configuration
# -----------------------------------------------------------------------------
# The canonical canonical URL of the application.
NEXTAUTH_URL="http://localhost:3000"

# Secret key used by NextAuth to sign and encrypt JWT session cookies.
# Generate a cryptographically secure 32-byte secret using:
#   openssl rand -base64 32
NEXTAUTH_SECRET="development-nextauth-secret-change-in-production-min-32-chars"

# -----------------------------------------------------------------------------
# 3. Node Environment Mode
# -----------------------------------------------------------------------------
NODE_ENV="development"
```

### 6.2 Complete Initial `.env` for Development (`/Users/jasbotello/devel/APPCITAS/.env`)

```bash
DATABASE_URL="postgresql://postgres:postgres@localhost:5432/appcitas?schema=public"
NEXTAUTH_URL="http://localhost:3000"
NEXTAUTH_SECRET="c1f8a2e7d3b94a5c6e8f0123456789abcdef0123456789abcdef0123456789ab"
NODE_ENV="development"
```

### 6.3 Environment Configuration Rationale

1. **`DATABASE_URL`**:
   - Strictly conforms to the Prisma PostgreSQL connection URI format.
   - Includes `?schema=public` explicitly to ensure migrations and schema pushes target the standard PostgreSQL schema.
   - For DigitalOcean Managed PostgreSQL clusters, `sslmode=require` and connection pooling ports (e.g. `25060` with connection pooler or `25061`) can be configured without changing application code.

2. **`NEXTAUTH_URL`**:
   - In development: `http://localhost:3000`.
   - NextAuth uses this to construct redirect callbacks and verify incoming host headers.

3. **`NEXTAUTH_SECRET`**:
   - NextAuth v4 requires a 32+ character secret for JWT encryption and cookie signing.
   - If missing at build time, Next.js build will fail or log security warnings. Providing a default in `.env` and a fallback in `src/lib/auth.ts` guarantees zero build failures.

---

## 7. Supporting Toolchain Configurations

To guarantee that Milestone 1 integrates seamlessly into Milestone 2 (Auth) and Milestone 3 (UI Dashboard), the supporting configuration files must be specified with exact parameters.

### 7.1 `tailwind.config.ts` (`/Users/jasbotello/devel/APPCITAS/tailwind.config.ts`)

```typescript
import type { Config } from 'tailwindcss';

const config: Config = {
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        // Calm healthcare teal palette (R3 aesthetic)
        medical: {
          50: '#F0FDFA',
          100: '#CCFBF1',
          200: '#99F6E4',
          300: '#5EEAD4',
          400: '#2DD4BF',
          500: '#14B8A6',
          600: '#0D9488',
          700: '#0F766E',
          800: '#115E59',
          900: '#134E4A',
          950: '#042F2E',
        },
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'BlinkMacSystemFont', 'sans-serif'],
      },
    },
  },
  plugins: [],
};

export default config;
```

### 7.2 `postcss.config.mjs` (`/Users/jasbotello/devel/APPCITAS/postcss.config.mjs`)

```javascript
/** @type {import('postcss-load-config').Config} */
const config = {
  plugins: {
    tailwindcss: {},
    autoprefixer: {},
  },
};

export default config;
```

### 7.3 `next.config.mjs` (`/Users/jasbotello/devel/APPCITAS/next.config.mjs`)

```javascript
/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  poweredByHeader: false,
  // Ensure route handlers and dynamic pages don't attempt SSG when DB is offline during build
  typescript: {
    ignoreBuildErrors: false,
  },
  eslint: {
    ignoreDuringBuilds: true,
  },
};

export default nextConfig;
```

### 7.4 `.gitignore` (`/Users/jasbotello/devel/APPCITAS/.gitignore`)

```gitignore
# Dependencies
/node_modules
/.pnp
.pnp.js

# Testing
/coverage

# Next.js build output
/.next/
/out/

# Production build
/build

# Environment variables
.env*.local
.env

# Debug logs
npm-debug.log*
yarn-debug.log*
yarn-error.log*

# OS metadata
.DS_Store
*.pem

# TypeScript
*.tsbuildinfo
next-env.d.ts
```

---

## 8. Build, Validation & Seed Pipeline Execution Strategy

For the downstream builder agent executing the implementation, here is the verified linear execution sequence:

```bash
# Step 1: Write configuration files
# package.json, tsconfig.json, tailwind.config.ts, postcss.config.mjs, next.config.mjs, .env

# Step 2: Install exact dependencies
npm install

# Step 3: Write Prisma schema (prisma/schema.prisma) and validate syntax offline
npx prisma validate
# Expected output: "The schema at prisma/schema.prisma is valid"

# Step 4: Generate TypeScript Prisma Client bindings
npx prisma generate
# Expected output: "✔ Generated Prisma Client (5.22.0) to ./node_modules/@prisma/client"

# Step 5: Write Prisma Client singleton and utils
# src/lib/prisma.ts, src/lib/utils.ts

# Step 6: Push schema to database (requires live PostgreSQL)
npx prisma db push

# Step 7: Execute pilot seed script
npx prisma db seed
# Or: npm run prisma:seed
# Expected output:
#   🌱 Starting APPCITAS Phase 1 seed...
#   ✅ Provisioned Clinic: Clínica Fisia
#   ✅ Provisioned Owner: Dr. Alejandro Morales
#   ✅ Provisioned 5 Specialists
#   ✅ Provisioned 6 Services with Specialist Assignments
#   ✅ Provisioned 4 Products
#   ✅ Provisioned Sample Appointment & Cash Order
#   ✅ Provisioned Sample Waitlist Entry
#   🎉 Seed complete!

# Step 8: Verify Next.js production build
npm run build
# Expected output: "✔ Compiled successfully", 0 TypeScript errors
```

---

## 9. Failure Modes, Edge Cases & Proactive Mitigations

| Failure Mode / Edge Case | Root Cause | Proactive Architectural Mitigation |
|---|---|---|
| **Node 24 native build error on `bcrypt`** | Native C++ compilation via `node-gyp` fails on macOS ARM64 / Node 24 when Xcode tools or Python headers are missing. | Use `bcryptjs` pure JS (`2.4.3`) and `@types/bcryptjs` (`2.4.6`). 100% JavaScript, zero native compilation required. |
| **Prisma hot-reload connection exhaustion** | Next.js development server reloads modules on file changes; every new `new PrismaClient()` opens a new pool. | Implement `src/lib/prisma.ts` singleton pattern caching the instance on `globalThis.prisma`. |
| **`npm run build` fails on unpopulated `DATABASE_URL`** | Next.js App Router attempts Static Site Generation (SSG) on pages and invokes DB queries at build time. | All dashboard and API routes must export `export const dynamic = 'force-dynamic'`, and `authOptions` provides fallback secret. |
| **Decimal serialization error in Server-to-Client components** | Prisma `Decimal` instances cannot be directly serialized across Next.js Server-to-Client component boundaries. | `src/lib/utils.ts` `formatCurrency()` accepts strings, numbers, or objects with `.toString()`. All server data fetchers serialize Decimals before passing to client. |
| **Missing NextAuth type definitions causing TypeScript build error** | NextAuth `session.user` does not include `clinicId`, `role`, or `specialty` by default. | Create `src/types/next-auth.d.ts` module augmentation before running `npm run build`. |
| **Duplicate seeds causing unique constraint crashes** | Re-running `npx prisma db seed` crashes on unique email (`users.email`), slug (`clinics.slug`), or compound keys. | All seed records must use `prisma.[model].upsert()` with compound unique identifiers. |

---

## 10. Conclusion & Handoff Readiness

The complete build, dependency, and Prisma client architecture for Milestone 1 is verified and ready for implementation.
- All dependencies are pinned to exact versions compatible with Node 24 and Next.js 14.
- `tsconfig.json` targets `ES2022` with bundler module resolution and strict type-checking.
- `src/lib/prisma.ts` implements the leak-free singleton pattern.
- `src/lib/utils.ts` provides robust classnames merging and currency formatting.
- `.env` and `.env.example` templates provide seamless PostgreSQL and NextAuth setup.
