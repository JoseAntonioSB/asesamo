# Dispatch for teamwork_preview_worker_m2

## Task: Implement Milestone 2 — Authentication & Clinic Onboarding
Read `/Users/jasbotello/devel/APPCITAS/ORIGINAL_REQUEST.md` and `/Users/jasbotello/devel/APPCITAS/PROJECT.md`.
Read Explorer findings:
- `/Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_explorer_survey_2_gen2/analysis.md`
- `/Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_spec_miner_m1_1/analysis.md`

## Mandatory Integrity Warning
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. A teamwork_preview_auditor will independently verify your work. Integrity violations WILL be detected and your work WILL be rejected.

## File Write Ownership
You exclusively own and will create/modify the following files:
- `src/lib/serialize.ts` (apply advisory fix: `if (data instanceof Date) return data;`)
- `src/types/next-auth.d.ts`
- `src/lib/auth.ts`
- `src/middleware.ts`
- `src/app/api/auth/[...nextauth]/route.ts`
- `src/app/api/clinic/register/route.ts`
- `src/app/api/specialists/route.ts`
- `src/app/api/services/route.ts`
- `src/app/api/services/[id]/assign/route.ts`
- `src/app/(auth)/login/page.tsx`
- `src/app/(auth)/register/page.tsx`
- `tests/verify-m2.ts`

## Concrete Implementation Requirements
1. In `src/lib/serialize.ts`: add `if (data instanceof Date) return data;` before the plain object loop.
2. In `src/types/next-auth.d.ts`: declare module "next-auth" and "next-auth/jwt" augmenting `Session` and `JWT` with:
   - `id: string`
   - `role: 'OWNER' | 'SPECIALIST'`
   - `clinicId: string`
   - `clinicName: string`
   - `clinicSlug: string`
3. In `src/lib/auth.ts`: NextAuth configuration with:
   - `session: { strategy: "jwt" }`
   - `providers: [CredentialsProvider({ ... })]`
   - Query user from database via `prisma.user.findFirst` matching email with include `clinic: true`.
   - Verify password using `bcrypt.compare`.
   - Enforce `user.isActive === true`.
   - Populate JWT token in `jwt` callback with user properties.
   - Populate `session.user` in `session` callback with token properties.
   - Set `pages: { signIn: '/login' }`.
   - `secret: process.env.NEXTAUTH_SECRET`.
4. In `src/app/api/auth/[...nextauth]/route.ts`: export `const handler = NextAuth(authOptions); export { handler as GET, handler as POST };`.
5. In `src/middleware.ts`: protect `/dashboard/:path*`.
   - Inspect token using `getToken({ req, secret: process.env.NEXTAUTH_SECRET })`.
   - If not authenticated, redirect to `/login?callbackUrl=...`.
   - If authenticated user accesses `/login` or `/register`, redirect to `/dashboard`.
   - Export `config = { matcher: ['/dashboard/:path*', '/login', '/register'] }`.
6. In `src/app/api/clinic/register/route.ts`:
   - `POST`: Create tenant Clinic and Owner User in a transaction (`prisma.$transaction`).
   - Hash password with `bcrypt.hash(password, 10)`.
   - Set user role to `OWNER`.
   - Return `{ success: true, clinic, user }`.
7. In `src/app/api/specialists/route.ts`:
   - Enforce session authentication (must be authenticated OWNER of the clinic).
   - `GET`: list specialists for `session.user.clinicId` including specialty, active status, assignments count.
   - `POST`: create specialist under `session.user.clinicId`.
     - Count active specialists: `const count = await prisma.user.count({ where: { clinicId, role: 'SPECIALIST', isActive: true } });`
     - If `count >= 5`, return 409 Conflict with `{ error: "Specialist quota reached (maximum 5 active specialists allowed)." }`.
     - Generate temporary password if not provided (`Temp${Math.random().toString(36).slice(-8)}!`).
     - Hash password and create user with role `SPECIALIST`.
     - Return `{ success: true, specialist, temporaryPassword }`.
8. In `src/app/api/services/route.ts` & `src/app/api/services/[id]/assign/route.ts`:
   - Enforce session authentication.
   - `GET /api/services`: list services for clinic with assigned specialists.
   - `POST /api/services`: create service for clinic.
   - `PUT /api/services/[id]/assign`: replace assigned specialists in `SpecialistService` for that service and clinic.
9. In `src/app/(auth)/login/page.tsx` & `src/app/(auth)/register/page.tsx`:
   - Beautiful client components using Tailwind and Lucide icons.
   - Medical teal palette (`text-teal-600`, `bg-teal-600`, `hover:bg-teal-700`).
   - Form submission handling with Sonner toasts (`toast.success`, `toast.error`).
10. Create `tests/verify-m2.ts` that tests the auth options, middleware logic, specialist quota enforcement, and registration logic.
11. Run `npx tsc --noEmit` and `npx tsx tests/verify-m2.ts`.
12. Write your handoff to `/Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_worker_m2/handoff.md`.

## 2026-09-14T04:24:34Z
User request:
Implement Milestone 2 per your write ownership:
1. Fix `src/lib/serialize.ts` (Date check).
2. NextAuth types and options: `src/types/next-auth.d.ts`, `src/lib/auth.ts`, `src/app/api/auth/[...nextauth]/route.ts`.
3. Middleware: `src/middleware.ts` protecting `/dashboard/*`.
4. APIs: `src/app/api/clinic/register/route.ts`, `src/app/api/specialists/route.ts` (with 5-specialist quota & temporary password generation), `src/app/api/services/route.ts`, `src/app/api/services/[id]/assign/route.ts`.
5. Auth UI: `src/app/(auth)/login/page.tsx`, `src/app/(auth)/register/page.tsx`.
6. Verification script: `tests/verify-m2.ts`.
