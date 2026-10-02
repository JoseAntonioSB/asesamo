# Dispatch for teamwork_preview_explorer_survey_2_gen2

## Task: Technical Architecture Investigation for APPCITAS Phase 1
Read `/Users/jasbotello/devel/APPCITAS/ORIGINAL_REQUEST.md` and `/Users/jasbotello/devel/APPCITAS/Roadmap.md`.
Also review the predecessor's progress in `/Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_explorer_survey_2/progress.md`.

DO NOT execute any external network commands (e.g. do not curl npm registry).
Synthesize the complete technical architecture and code layout for APPCITAS Phase 1 based on verified standard requirements:
1. Environment & Stack:
   - Node 24.x, npm 11.x, Next.js 14.x App Router with TypeScript and Tailwind CSS.
   - Database: Prisma ORM with PostgreSQL provider. Prisma CLI, `@prisma/client`, and `tsx` for running `prisma/seed.ts`.
   - Authentication: NextAuth.js (v4) with credentials provider, JWT session strategy, custom session/jwt callbacks attaching `clinicId`, `role`, `clinicSlug`.
   - Password hashing: `bcryptjs` and `@types/bcryptjs`.
   - UI styling: Tailwind CSS, `lucide-react`, `sonner`, `clsx`, `tailwind-merge`.
2. Detailed Code Layout (`src/` structure):
   - Exact file tree for `prisma/schema.prisma`, `prisma/seed.ts`.
   - `src/lib/prisma.ts` (global PrismaClient singleton avoiding connection exhaustion).
   - `src/lib/auth.ts` (NextAuth options and configuration).
   - `src/app/api/auth/[...nextauth]/route.ts`.
   - `src/app/api/clinic/register/route.ts`.
   - `src/app/api/specialists/route.ts`.
   - `src/app/api/services/route.ts`.
   - `src/app/(auth)/login/page.tsx`, `src/app/(auth)/register/page.tsx`.
   - `src/app/dashboard/layout.tsx`, `src/app/dashboard/page.tsx`, `src/app/dashboard/specialists/page.tsx`, `src/app/dashboard/services/page.tsx`.
   - Middleware `src/middleware.ts` protecting `/dashboard/:path*`.
3. Validation & Build Pipeline:
   - Specific commands: `npx prisma validate`, `npx prisma db seed`, `npm run build`.
   - Edge cases for 0-error build (mocking/handling env vars during build, next auth secret, database url).

Write your detailed findings to `/Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_explorer_survey_2_gen2/analysis.md` and your final handoff report to `/Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_explorer_survey_2_gen2/handoff.md`.
Communicate back to parent via send_message when done.
