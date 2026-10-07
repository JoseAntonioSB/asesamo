# Analysis: Next.js 14+ App Router & UI System Architecture for `asesamo-web/`

**Milestone**: M4 (Frontend Web Interface Architecture)  
**Agent**: `teamwork_preview_explorer_m4_1`  
**Timestamp**: 2026-10-07T16:31:00Z  
**Status**: Ready for Implementation  

---

## 1. Executive Summary & Architectural Overview

The `asesamo-web/` frontend is the critical human-machine interface (HMI) for security guards and plant compliance personnel in industrial facilities. Operating in harsh environmental conditions (sun glare, handheld rugged scanners, multi-meter viewing distances), the user interface demands:
1. **Ultra-High Visual Contrast**: An industrial dark slate substrate (`#0a0f1d`) that eliminates eye fatigue while delivering crisp luminance contrast against text and status badges.
2. **Instant Status Feedback**: Unambiguous 3-tier visual signaling using industrial safety color semantics:
   - **Emerald Green (`#10b981`)**: Access Granted (`granted`).
   - **Amber Yellow (`#f59e0b`)**: Access Granted with Warning (`granted_with_warning`).
   - **Crimson Red (`#ef4444`)**: Access Denied (`denied`).
3. **Deterministic Typographic Architecture**: Dual font stack comprising **Inter** for clean human interfaces and **JetBrains Mono** for monospaced telemetry (worker UUIDs, CURP, NSS, RFC, REPSE registration numbers).
4. **App Router & Server Component Discipline**: Built on **Next.js 14.2+ App Router**, strict TypeScript (`strict: true`), Tailwind CSS, and shadcn/ui-compatible component primitives.
5. **Container-Ready Compilation**: Zero-dependency standalone output mode (`output: "standalone"`) perfectly aligned with the multi-stage Docker deployment required by Milestone 5 and `docker-compose.yml`.

---

## 2. Target File Tree Blueprint

The following structure represents the complete foundational architecture designed for `asesamo-web/`:

```
asesamo-web/
├── package.json               # Next 14, React 18, Tailwind, Lucide, CVA, TanStack Query
├── tsconfig.json              # Strict TypeScript with "@/": ["./*"] path aliases
├── postcss.config.js          # Standard Tailwind + Autoprefixer pipeline
├── tailwind.config.ts         # Industrial color tokens, font variables, and animations
├── next.config.mjs            # Standalone output & internal API proxy routing
├── components.json            # Official shadcn/ui configuration descriptor
├── app/
│   ├── globals.css            # Industrial CSS custom properties, substrate, and scrollbars
│   └── layout.tsx             # RootLayout Server Component with lang="es" and Inter/Mono fonts
├── lib/
│   └── utils.ts               # Core utility: cn() helper + Mexican date formatting
└── components/
    └── ui/
        ├── button.tsx         # CVA Button with industrial status variants & Radix Slot
        ├── card.tsx           # Elevated slate card container primitives
        ├── input.tsx          # High-contrast input supporting barcode scanner autofocus
        ├── badge.tsx          # Compliance status badges (granted, warning, denied)
        ├── table.tsx          # High-density tabular layout for document inspection
        └── alert.tsx          # Industrial alert banner primitive
```

---

## 3. Section 1: Next.js 14+ Project Configuration Files

### 3.1 `package.json`
Configures production and development dependencies. Pinned to stable Next.js 14.2+, React 18, `@radix-ui/react-slot`, `lucide-react`, and `@tanstack/react-query`.

```json
{
  "name": "asesamo-web",
  "version": "0.1.0",
  "private": true,
  "scripts": {
    "dev": "next dev",
    "build": "next build",
    "start": "next start",
    "lint": "next lint",
    "typecheck": "tsc --noEmit"
  },
  "dependencies": {
    "@radix-ui/react-slot": "^1.1.0",
    "@tanstack/react-query": "^5.59.16",
    "class-variance-authority": "^0.7.0",
    "clsx": "^2.1.1",
    "lucide-react": "^0.453.0",
    "next": "^14.2.15",
    "react": "^18.3.1",
    "react-dom": "^18.3.1",
    "tailwind-merge": "^2.5.4",
    "tailwindcss-animate": "^1.0.7"
  },
  "devDependencies": {
    "@types/node": "^20.16.11",
    "@types/react": "^18.3.11",
    "@types/react-dom": "^18.3.0",
    "autoprefixer": "^10.4.20",
    "postcss": "^8.4.47",
    "tailwindcss": "^3.4.14",
    "typescript": "^5.6.3"
  }
}
```

---

### 3.2 `tsconfig.json`
Strict TypeScript configuration enforcing no explicit `any`, ES2022 output, bundler module resolution, and root path mapping `"@/*": ["./*"]`.

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
    "baseUrl": ".",
    "paths": {
      "@/*": ["./*"]
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

---

### 3.3 `postcss.config.js`
Standard PostCSS configuration pipeline required by Tailwind CSS.

```javascript
module.exports = {
  plugins: {
    tailwindcss: {},
    autoprefixer: {},
  },
};
```

---

### 3.4 `tailwind.config.ts`
Full Tailwind CSS configuration mapping all CSS variable tokens, radius rules, custom font families, and animation keyframes.

```typescript
import type { Config } from "tailwindcss";

const config: Config = {
  darkMode: ["class"],
  content: [
    "./pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
    "./lib/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    container: {
      center: true,
      padding: "2rem",
      screens: {
        "2xl": "1400px",
      },
    },
    extend: {
      colors: {
        border: "hsl(var(--border))",
        input: "hsl(var(--input))",
        ring: "hsl(var(--ring))",
        background: "hsl(var(--background))",
        foreground: "hsl(var(--foreground))",
        primary: {
          DEFAULT: "hsl(var(--primary))",
          foreground: "hsl(var(--primary-foreground))",
        },
        secondary: {
          DEFAULT: "hsl(var(--secondary))",
          foreground: "hsl(var(--secondary-foreground))",
        },
        destructive: {
          DEFAULT: "hsl(var(--destructive))",
          foreground: "hsl(var(--destructive-foreground))",
        },
        muted: {
          DEFAULT: "hsl(var(--muted))",
          foreground: "hsl(var(--muted-foreground))",
        },
        accent: {
          DEFAULT: "hsl(var(--accent))",
          foreground: "hsl(var(--accent-foreground))",
        },
        popover: {
          DEFAULT: "hsl(var(--popover))",
          foreground: "hsl(var(--popover-foreground))",
        },
        card: {
          DEFAULT: "hsl(var(--card))",
          foreground: "hsl(var(--card-foreground))",
        },
        // Dedicated Industrial Access Status tokens:
        granted: {
          DEFAULT: "hsl(var(--granted))",
          foreground: "hsl(var(--granted-foreground))",
          light: "hsl(var(--granted-light))",
        },
        warning: {
          DEFAULT: "hsl(var(--warning))",
          foreground: "hsl(var(--warning-foreground))",
          light: "hsl(var(--warning-light))",
        },
        denied: {
          DEFAULT: "hsl(var(--denied))",
          foreground: "hsl(var(--denied-foreground))",
          light: "hsl(var(--denied-light))",
        },
      },
      borderRadius: {
        lg: "var(--radius)",
        md: "calc(var(--radius) - 2px)",
        sm: "calc(var(--radius) - 4px)",
      },
      fontFamily: {
        sans: ["var(--font-inter)", "system-ui", "sans-serif"],
        mono: ["var(--font-mono)", "ui-monospace", "SFMono-Regular", "monospace"],
      },
      keyframes: {
        "accordion-down": {
          from: { height: "0" },
          to: { height: "var(--radix-accordion-content-height)" },
        },
        "accordion-up": {
          from: { height: "var(--radix-accordion-content-height)" },
          to: { height: "0" },
        },
        "pulse-subtle": {
          "0%, 100%": { opacity: "1" },
          "50%": { opacity: "0.8" },
        },
      },
      animation: {
        "accordion-down": "accordion-down 0.2s ease-out",
        "accordion-up": "accordion-up 0.2s ease-out",
        "pulse-subtle": "pulse-subtle 2s cubic-bezier(0.4, 0, 0.6, 1) infinite",
      },
    },
  },
  plugins: [require("tailwindcss-animate")],
};

export default config;
```

---

### 3.5 `next.config.mjs`
Enables `output: "standalone"` for production Docker containers, strict React mode, and internal proxy routing to the Go API backend.

```javascript
/** @type {import('next').NextConfig} */
const nextConfig = {
  output: "standalone",
  reactStrictMode: true,
  poweredByHeader: false,
  async rewrites() {
    const internalApi = process.env.INTERNAL_API_URL || "http://127.0.0.1:8080/api/v1";
    return [
      {
        source: "/api/proxy/:path*",
        destination: `${internalApi}/:path*`,
      },
    ];
  },
};

export default nextConfig;
```

---

### 3.6 `components.json`
Standard shadcn/ui descriptor enabling compatibility with standard component tooling.

```json
{
  "$schema": "https://ui.shadcn.com/schema.json",
  "style": "default",
  "rsc": true,
  "tsx": true,
  "tailwind": {
    "config": "tailwind.config.ts",
    "css": "app/globals.css",
    "baseColor": "slate",
    "cssVariables": true,
    "prefix": ""
  },
  "aliases": {
    "components": "@/components",
    "utils": "@/lib/utils",
    "ui": "@/components/ui",
    "lib": "@/lib",
    "hooks": "@/hooks"
  }
}
```

---

## 4. Section 2: Global Styles & Design System Architecture

### 4.1 `app/globals.css`
Declares the dark slate substrate, high-contrast foreground, and dedicated status tokens for industrial telemetry.

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
  :root {
    /* Industrial High-Contrast Dark Slate Substrate (Default) */
    --background: 222 47% 7%;        /* #0a0f1d - Deep industrial slate */
    --foreground: 210 40% 98%;       /* #f8fafc - Crisp white */

    --card: 222 47% 10%;             /* #0e1626 - Elevated technical surface */
    --card-foreground: 210 40% 98%;

    --popover: 222 47% 10%;
    --popover-foreground: 210 40% 98%;

    --primary: 210 40% 98%;          /* High-contrast element fill */
    --primary-foreground: 222 47% 7%;

    --secondary: 217 33% 17%;        /* Dark tactical secondary button / panel */
    --secondary-foreground: 210 40% 98%;

    --muted: 217 33% 15%;            /* Muted telemetry indicators */
    --muted-foreground: 215 20% 65%; /* Readable secondary labels */

    --accent: 217 33% 20%;
    --accent-foreground: 210 40% 98%;

    --destructive: 0 84% 60%;        /* Denied crimson red */
    --destructive-foreground: 210 40% 98%;

    --border: 217 33% 20%;           /* Crisp precision dividing lines */
    --input: 217 33% 20%;
    --ring: 212 96% 57%;             /* High-visibility active focus ring */

    --radius: 0.375rem;              /* 6px - Precision industrial corner */

    /* Dedicated Industrial Telemetry Tokens */
    --granted: 142 71% 45%;          /* Emerald Green #10b981 */
    --granted-foreground: 144 80% 98%;
    --granted-light: 142 71% 45% / 0.15;

    --warning: 38 92% 50%;           /* Amber Yellow #f59e0b */
    --warning-foreground: 48 96% 12%;
    --warning-light: 38 92% 50% / 0.15;

    --denied: 0 84% 60%;             /* Crimson Red #ef4444 */
    --denied-foreground: 0 100% 98%;
    --denied-light: 0 84% 60% / 0.15;
  }

  /* Explicit .dark class alias */
  .dark {
    --background: 222 47% 7%;
    --foreground: 210 40% 98%;
    --card: 222 47% 10%;
    --card-foreground: 210 40% 98%;
    --popover: 222 47% 10%;
    --popover-foreground: 210 40% 98%;
    --primary: 210 40% 98%;
    --primary-foreground: 222 47% 7%;
    --secondary: 217 33% 17%;
    --secondary-foreground: 210 40% 98%;
    --muted: 217 33% 15%;
    --muted-foreground: 215 20% 65%;
    --accent: 217 33% 20%;
    --accent-foreground: 210 40% 98%;
    --destructive: 0 84% 60%;
    --destructive-foreground: 210 40% 98%;
    --border: 217 33% 20%;
    --input: 217 33% 20%;
    --ring: 212 96% 57%;
    --radius: 0.375rem;

    --granted: 142 71% 45%;
    --granted-foreground: 144 80% 98%;
    --granted-light: 142 71% 45% / 0.15;

    --warning: 38 92% 50%;
    --warning-foreground: 48 96% 12%;
    --warning-light: 38 92% 50% / 0.15;

    --denied: 0 84% 60%;
    --denied-foreground: 0 100% 98%;
    --denied-light: 0 84% 60% / 0.15;
  }
}

@layer base {
  * {
    @apply border-border;
  }

  html {
    color-scheme: dark;
  }

  body {
    @apply bg-background text-foreground antialiased;
    font-feature-settings: "rlig" 1, "calt" 1, "tnum" 1;
  }
}

/* Precision Tactical Scrollbars */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: hsl(222 47% 7%);
}

::-webkit-scrollbar-thumb {
  background: hsl(217 33% 25%);
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: hsl(217 33% 35%);
}
```

---

### 4.2 `app/layout.tsx`
Configures root layout with Spanish language tag `<html lang="es">`, Inter font for UI readability, and JetBrains Mono for industrial telemetry codes.

```tsx
import type { Metadata, Viewport } from "next";
import { Inter, JetBrains_Mono } from "next/font/google";
import "./globals.css";

const inter = Inter({
  subsets: ["latin"],
  variable: "--font-inter",
  display: "swap",
});

const jetbrainsMono = JetBrains_Mono({
  subsets: ["latin"],
  variable: "--font-mono",
  display: "swap",
});

export const metadata: Metadata = {
  title: {
    default: "ASESAMO | Control de Acceso Industrial",
    template: "%s | ASESAMO",
  },
  description:
    "Sistema de control de acceso vehicular y peatonal para plantas industriales con verificación documental REPSE, DC-3, IMSS y NOM en tiempo real.",
  authors: [{ name: "ASESAMO Engineering" }],
  icons: {
    icon: "/favicon.ico",
  },
};

export const viewport: Viewport = {
  themeColor: "#0a0f1d",
  width: "device-width",
  initialScale: 1,
  maximumScale: 1,
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="es" className="dark" suppressHydrationWarning>
      <body
        className={`${inter.variable} ${jetbrainsMono.variable} font-sans min-h-screen bg-background text-foreground antialiased selection:bg-slate-700 selection:text-white`}
      >
        <div className="relative flex min-h-screen flex-col">
          {children}
        </div>
      </body>
    </html>
  );
}
```

---

## 5. Section 3: Utility & Helper Layer

### 5.1 `lib/utils.ts`
Combines `clsx` and `tailwind-merge` to resolve conflicting Tailwind utility classes cleanly, plus Mexican Spanish date formatting.

```typescript
import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

/**
 * Utility helper combining clsx condition handling with tailwind-merge conflict resolution.
 */
export function cn(...inputs: ClassValue[]): string {
  return twMerge(clsx(inputs));
}

/**
 * Format date in Mexican Spanish locale (e.g. "01/Oct/2026").
 */
export function formatDate(dateString: string | null | undefined): string {
  if (!dateString) return "N/A";
  try {
    const date = new Date(dateString);
    if (isNaN(date.getTime())) return dateString;
    return new Intl.DateTimeFormat("es-MX", {
      day: "2-digit",
      month: "short",
      year: "numeric",
    }).format(date);
  } catch {
    return dateString;
  }
}
```

---

## 6. Section 4: Core UI Component Primitives

### 6.1 `components/ui/button.tsx`
Built with `class-variance-authority`, `@radix-ui/react-slot`, and `forwardRef`. Includes variants for standard actions as well as high-visibility industrial buttons (`granted`, `warning`, `denied`).

```tsx
import * as React from "react";
import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const buttonVariants = cva(
  "inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 select-none",
  {
    variants: {
      variant: {
        default:
          "bg-primary text-primary-foreground hover:bg-primary/90 active:scale-[0.99] transition-transform",
        destructive:
          "bg-destructive text-destructive-foreground hover:bg-destructive/90 active:scale-[0.99] transition-transform",
        outline:
          "border border-input bg-background hover:bg-accent hover:text-accent-foreground",
        secondary:
          "bg-secondary text-secondary-foreground hover:bg-secondary/80",
        ghost:
          "hover:bg-accent hover:text-accent-foreground",
        link:
          "text-primary underline-offset-4 hover:underline",
        // Industrial Telemetry Action Buttons
        granted:
          "bg-emerald-600 text-white hover:bg-emerald-500 font-semibold shadow-sm active:scale-[0.99] transition-transform",
        warning:
          "bg-amber-500 text-slate-950 hover:bg-amber-400 font-semibold shadow-sm active:scale-[0.99] transition-transform",
        denied:
          "bg-red-600 text-white hover:bg-red-500 font-semibold shadow-sm active:scale-[0.99] transition-transform",
      },
      size: {
        default: "h-10 px-4 py-2",
        sm: "h-8 rounded-md px-3 text-xs",
        lg: "h-12 rounded-md px-8 text-base",
        icon: "h-10 w-10",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
);

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean;
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, asChild = false, ...props }, ref) => {
    const Comp = asChild ? Slot : "button";
    return (
      <Comp
        className={cn(buttonVariants({ variant, size, className }))}
        ref={ref}
        {...props}
      />
    );
  }
);
Button.displayName = "Button";

export { Button, buttonVariants };
```

---

### 6.2 `components/ui/card.tsx`
Modular card surface with subtle industrial border, backdrop blur, and dark slate substrate elevation.

```tsx
import * as React from "react";
import { cn } from "@/lib/utils";

const Card = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
  <div
    ref={ref}
    className={cn(
      "rounded-lg border border-border bg-card text-card-foreground shadow-sm backdrop-blur-sm",
      className
    )}
    {...props}
  />
));
Card.displayName = "Card";

const CardHeader = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
  <div
    ref={ref}
    className={cn("flex flex-col space-y-1.5 p-6", className)}
    {...props}
  />
));
CardHeader.displayName = "CardHeader";

const CardTitle = React.forwardRef<
  HTMLHeadingElement,
  React.HTMLAttributes<HTMLHeadingElement>
>(({ className, ...props }, ref) => (
  <h3
    ref={ref}
    className={cn(
      "text-xl font-semibold leading-none tracking-tight text-foreground",
      className
    )}
    {...props}
  />
));
CardTitle.displayName = "CardTitle";

const CardDescription = React.forwardRef<
  HTMLParagraphElement,
  React.HTMLAttributes<HTMLParagraphElement>
>(({ className, ...props }, ref) => (
  <p
    ref={ref}
    className={cn("text-sm text-muted-foreground", className)}
    {...props}
  />
));
CardDescription.displayName = "CardDescription";

const CardContent = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
  <div ref={ref} className={cn("p-6 pt-0", className)} {...props} />
));
CardContent.displayName = "CardContent";

const CardFooter = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
  <div
    ref={ref}
    className={cn("flex items-center p-6 pt-0", className)}
    {...props}
  />
));
CardFooter.displayName = "CardFooter";

export {
  Card,
  CardHeader,
  CardFooter,
  CardTitle,
  CardDescription,
  CardContent,
};
```

---

### 6.3 `components/ui/input.tsx`
Input primitive supporting barcode scanner input, autofocus, and optional `isTelemetry` monospace styling for UUID and CURP inputs.

```tsx
import * as React from "react";
import { cn } from "@/lib/utils";

export interface InputProps
  extends React.InputHTMLAttributes<HTMLInputElement> {
  isTelemetry?: boolean;
}

const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type, isTelemetry = false, ...props }, ref) => {
    return (
      <input
        type={type}
        className={cn(
          "flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50",
          isTelemetry && "font-mono uppercase tracking-wider",
          className
        )}
        ref={ref}
        {...props}
      />
    );
  }
);
Input.displayName = "Input";

export { Input };
```

---

### 6.4 `components/ui/badge.tsx`
High-contrast status badges for document compliance results (`granted`, `warning`, `denied`).

```tsx
import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const badgeVariants = cva(
  "inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 select-none",
  {
    variants: {
      variant: {
        default:
          "border-transparent bg-primary text-primary-foreground hover:bg-primary/80",
        secondary:
          "border-transparent bg-secondary text-secondary-foreground hover:bg-secondary/80",
        destructive:
          "border-transparent bg-destructive text-destructive-foreground hover:bg-destructive/80",
        outline:
          "text-foreground border-border",
        // Industrial Compliance Badges
        granted:
          "border-emerald-500/30 bg-emerald-500/15 text-emerald-400 font-medium",
        warning:
          "border-amber-500/30 bg-amber-500/15 text-amber-400 font-medium",
        denied:
          "border-red-500/30 bg-red-500/15 text-red-400 font-medium",
        info:
          "border-blue-500/30 bg-blue-500/15 text-blue-400 font-medium",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
);

export interface BadgeProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, ...props }: BadgeProps) {
  return (
    <div className={cn(badgeVariants({ variant }), className)} {...props} />
  );
}

export { Badge, badgeVariants };
```

---

### 6.5 `components/ui/table.tsx`
Tabular component system for dense industrial data: document catalogs, expiration dates, contractor registrations, and compliance failure reasons.

```tsx
import * as React from "react";
import { cn } from "@/lib/utils";

const Table = React.forwardRef<
  HTMLTableElement,
  React.HTMLAttributes<HTMLTableElement>
>(({ className, ...props }, ref) => (
  <div className="relative w-full overflow-auto rounded-md border border-border">
    <table
      ref={ref}
      className={cn("w-full caption-bottom text-sm", className)}
      {...props}
    />
  </div>
));
Table.displayName = "Table";

const TableHeader = React.forwardRef<
  HTMLTableSectionElement,
  React.HTMLAttributes<HTMLTableSectionElement>
>(({ className, ...props }, ref) => (
  <thead ref={ref} className={cn("[&_tr]:border-b bg-muted/40", className)} {...props} />
));
TableHeader.displayName = "TableHeader";

const TableBody = React.forwardRef<
  HTMLTableSectionElement,
  React.HTMLAttributes<HTMLTableSectionElement>
>(({ className, ...props }, ref) => (
  <tbody
    ref={ref}
    className={cn("[&_tr:last-child]:border-0 divide-y divide-border", className)}
    {...props}
  />
));
TableBody.displayName = "TableBody";

const TableFooter = React.forwardRef<
  HTMLTableSectionElement,
  React.HTMLAttributes<HTMLTableSectionElement>
>(({ className, ...props }, ref) => (
  <tfoot
    ref={ref}
    className={cn(
      "border-t bg-muted/50 font-medium [&>tr]:last:border-b-0",
      className
    )}
    {...props}
  />
));
TableFooter.displayName = "TableFooter";

const TableRow = React.forwardRef<
  HTMLTableRowElement,
  React.HTMLAttributes<HTMLTableRowElement>
>(({ className, ...props }, ref) => (
  <tr
    ref={ref}
    className={cn(
      "border-b border-border transition-colors hover:bg-muted/30 data-[state=selected]:bg-muted",
      className
    )}
    {...props}
  />
));
TableRow.displayName = "TableRow";

const TableHead = React.forwardRef<
  HTMLTableCellElement,
  React.ThHTMLAttributes<HTMLTableCellElement>
>(({ className, ...props }, ref) => (
  <th
    ref={ref}
    className={cn(
      "h-10 px-4 text-left align-middle text-xs font-semibold uppercase tracking-wider text-muted-foreground [&:has([role=checkbox])]:pr-0",
      className
    )}
    {...props}
  />
));
TableHead.displayName = "TableHead";

const TableCell = React.forwardRef<
  HTMLTableCellElement,
  React.TdHTMLAttributes<HTMLTableCellElement>
>(({ className, ...props }, ref) => (
  <td
    ref={ref}
    className={cn(
      "p-4 align-middle text-sm text-foreground [&:has([role=checkbox])]:pr-0",
      className
    )}
    {...props}
  />
));
TableCell.displayName = "TableCell";

const TableCaption = React.forwardRef<
  HTMLTableCaptionElement,
  React.HTMLAttributes<HTMLTableCaptionElement>
>(({ className, ...props }, ref) => (
  <caption
    ref={ref}
    className={cn("mt-4 text-xs text-muted-foreground", className)}
    {...props}
  />
));
TableCaption.displayName = "TableCaption";

export {
  Table,
  TableHeader,
  TableBody,
  TableFooter,
  TableHead,
  TableRow,
  TableCell,
  TableCaption,
};
```

---

### 6.6 `components/ui/alert.tsx`
Industrial notification and warning banner component. Supports `granted`, `warning`, and `destructive` (denied) variants with automatic icon layout adjustments.

```tsx
import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const alertVariants = cva(
  "relative w-full rounded-lg border p-4 [&>svg~*]:pl-7 [&>svg+div]:translate-y-[-3px] [&>svg]:absolute [&>svg]:left-4 [&>svg]:top-4 [&>svg]:text-foreground",
  {
    variants: {
      variant: {
        default: "bg-background text-foreground border-border",
        destructive:
          "border-red-500/40 bg-red-950/20 text-red-300 [&>svg]:text-red-400",
        granted:
          "border-emerald-500/40 bg-emerald-950/20 text-emerald-300 [&>svg]:text-emerald-400",
        warning:
          "border-amber-500/40 bg-amber-950/20 text-amber-300 [&>svg]:text-amber-400",
        info:
          "border-blue-500/40 bg-blue-950/20 text-blue-300 [&>svg]:text-blue-400",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
);

const Alert = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement> & VariantProps<typeof alertVariants>
>(({ className, variant, ...props }, ref) => (
  <div
    ref={ref}
    role="alert"
    className={cn(alertVariants({ variant }), className)}
    {...props}
  />
));
Alert.displayName = "Alert";

const AlertTitle = React.forwardRef<
  HTMLParagraphElement,
  React.HTMLAttributes<HTMLHeadingElement>
>(({ className, ...props }, ref) => (
  <h5
    ref={ref}
    className={cn("mb-1 font-semibold leading-none tracking-tight", className)}
    {...props}
  />
));
AlertTitle.displayName = "AlertTitle";

const AlertDescription = React.forwardRef<
  HTMLParagraphElement,
  React.HTMLAttributes<HTMLParagraphElement>
>(({ className, ...props }, ref) => (
  <div
    ref={ref}
    className={cn("text-sm [&_p]:leading-relaxed opacity-90", className)}
    {...props}
  />
));
AlertDescription.displayName = "AlertDescription";

export { Alert, AlertTitle, AlertDescription };
```

---

## 7. Section 5: Integration Guide for Peer Explorers & Workers

### 7.1 Integration with Explorer 2 (`lib/api.ts`, `app/login/page.tsx`, `middleware.ts`)
- **Login Card**: Uses `@/components/ui/card` (`Card`, `CardHeader`, `CardTitle`, `CardDescription`, `CardContent`, `CardFooter`).
- **Input Fields**: Uses `@/components/ui/input` for `email` and `password` with auto-complete and dark mode styling.
- **Submit Button**: Uses `@/components/ui/button` (`variant="default"` or `variant="primary"`).
- **Error Banner**: Uses `@/components/ui/alert` (`variant="destructive"`) to display Spanish backend error messages.

### 7.2 Integration with Explorer 3 (`app/caseta/page.tsx`, Caseta components)
- **Scanner Input**: Uses `@/components/ui/input` (`isTelemetry={true}`) for high-speed HID wedge autofocus.
- **Status Banner**: Uses the industrial palette (`bg-emerald-600` / `bg-amber-500` / `bg-red-600`) or `@/components/ui/alert` (`variant="granted"` / `variant="warning"` / `variant="destructive"`).
- **Document Breakdown**: Uses `@/components/ui/table` with `@/components/ui/badge` for statuses:
  - `Aprobado` -> `<Badge variant="granted">APROBADO</Badge>`
  - `Por vencer` -> `<Badge variant="warning">POR VENCER</Badge>`
  - `Vencido` / `Rechazado` -> `<Badge variant="denied">VENCIDO</Badge>`
  - `Faltante` -> `<Badge variant="destructive">FALTANTE</Badge>`
- **Quick-Fill Presets**: Uses `@/components/ui/button` with `variant="outline"` or `variant="secondary"`.

---

## 8. Implementation Checklist for Worker Agents

When implementing these files into `asesamo-web/`:
1. [ ] Create directory `asesamo-web/` and subdirectories `app/`, `lib/`, `components/ui/`.
2. [ ] Write configuration files: `package.json`, `tsconfig.json`, `postcss.config.js`, `tailwind.config.ts`, `next.config.mjs`, `components.json`.
3. [ ] Write styling and root layout: `app/globals.css`, `app/layout.tsx`.
4. [ ] Write utilities: `lib/utils.ts`.
5. [ ] Write primitives: `components/ui/button.tsx`, `components/ui/card.tsx`, `components/ui/input.tsx`, `components/ui/badge.tsx`, `components/ui/table.tsx`, `components/ui/alert.tsx`.
6. [ ] Verify compilation with `npm install` (or verify that syntax passes TypeScript check `tsc --noEmit`).
