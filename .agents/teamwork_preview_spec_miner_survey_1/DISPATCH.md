# Dispatch for teamwork_preview_spec_miner_survey_1

## Task: Specification Mining for APPCITAS Phase 1
Read `/Users/jasbotello/devel/APPCITAS/ORIGINAL_REQUEST.md` and `/Users/jasbotello/devel/APPCITAS/Roadmap.md`.
Extract exhaustive, unambiguous specifications for:
1. Database Schema & Multitenancy:
   - Every required model: Clinic, User, Service, SpecialistService, Product, Customer, Appointment, Order, OrderItem, Waitlist.
   - Exact fields, types, relations, indexes, enums (e.g. Role: OWNER, SPECIALIST; PaymentMethod: CASH, CARD, TRANSFER; AppointmentStatus; etc.).
   - Explicit multitenancy boundary: which models have `clinicId`, foreign keys, cascade rules.
2. Pilot Seed Data Specifications:
   - Clinic "Clínica Fisia" details (slug, name, phone, address).
   - Owner account details and credentials.
   - 5 specialists details (names, specialties, contact, temporary credentials).
   - Core physiotherapy services (names, prices, durations, default assignments).
3. Authentication & Authorization Specifications:
   - NextAuth credentials provider flow.
   - Route protection for `/dashboard/*` (redirecting unauthenticated users to `/login`).
   - Specialist onboarding without external SMTP (direct password/credentials creation by owner).
4. Acceptance Criteria & Validation checklist.

Write your findings to `/Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_spec_miner_survey_1/analysis.md` and a final handoff to `/Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_spec_miner_survey_1/handoff.md`.

## 2026-09-14T04:04:55Z
You are teamwork_preview_spec_miner_survey_1.
Your working directory is: /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_spec_miner_survey_1
You MUST read /Users/jasbotello/devel/APPCITAS/ORIGINAL_REQUEST.md and /Users/jasbotello/devel/APPCITAS/Roadmap.md before starting work.
Also read your assignment in /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_spec_miner_survey_1/DISPATCH.md.

Perform exhaustive specification extraction for APPCITAS Phase 1 (El Cimiento):
1. Extract exact requirements for all 10 Prisma models: Clinic, User, Service, SpecialistService, Product, Customer, Appointment, Order, OrderItem, Waitlist. Define fields, types, enums, relations, and multitenancy boundary (clinicId).
2. Extract exact pilot seed data requirements (Clínica Fisia, owner, 5 specialists with specialties, core services).
3. Extract authentication, authorization, and specialist onboarding requirements (NextAuth credentials, /dashboard/* protection, direct specialist onboarding without external SMTP).
4. Extract all acceptance criteria and edge cases.

Write your detailed findings to /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_spec_miner_survey_1/analysis.md and your final handoff report to /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_spec_miner_survey_1/handoff.md.
Communicate back to parent when done via send_message.
