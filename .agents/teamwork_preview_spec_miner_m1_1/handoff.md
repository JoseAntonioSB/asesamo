# Handoff Report: Milestone 1 — Multitenant Schema & Data Layer Specification

**Author**: `teamwork_preview_spec_miner_m1_1` (Specification Miner)  
**Recipient**: `parent` (`a9195c6f-10f3-4499-a301-b36b7cb19186`)  
**Milestone**: M1 (Multitenant Schema & Data Layer)  
**Date**: 2026-09-14  
**Type**: Hard Handoff (Task Complete)  

---

## 1. Observation

1. **Original User Request (`ORIGINAL_REQUEST.md`)**:
   - Lines 12-22: Mandates the core multitenant architecture comprising `Clinic` (tenant with slug, name, phone, address), `User` (owner and specialist physiotherapists with credentials, contact, role, and active status), `Service` & `SpecialistService` (clinic-wide services with pricing, duration in minutes, and specialist assignment), `Product` (supplies and retail), `Customer` (patient contact info), `Appointment` (scheduled sessions linked to clinic, specialist, service, patient), `Order` & `OrderItem` (order tracking with payment methods `CASH`, `CARD`, `TRANSFER` for daily cash closing), and `Waitlist` (cancellation queue).
   - Lines 22-23: Mandates automated seed script (`prisma/seed.ts`) populating a pilot clinic ("Clínica Fisia"), owner account, 5 specialists with specialties, and core services.
   - Lines 35-39: Mandates acceptance criteria: `npx prisma validate` executes without schema syntax or relation errors, `npx prisma db seed` successfully provisions pilot clinic, 5 specialists, and initial service catalog, and multitenancy is enforced via `clinicId`.

2. **Project Specification (`PROJECT.md`)**:
   - Lines 14-25: Maps Features 1 through 10 directly to Milestone 1: Tenant Clinic Model, User & Specialist Model, Services & Specialist Assignment Schema, Product & Inventory Schema, Patient & Customer Schema, Appointment Scheduling Schema, Cash Closing & Orders Schema, Cancellation Waitlist Schema, Pilot Database Seed Script, and Prisma Client Singleton.
   - Lines 40-42: Identifies deliverables for Milestone 1: `package.json`, `prisma/schema.prisma`, `prisma/seed.ts`, `src/lib/prisma.ts`.
   - Lines 50-54: Establishes interface contracts: `prisma` exported as singleton from `src/lib/prisma.ts`, multi-tenancy rule requiring `where: { clinicId }` across all queries, and monetary amounts represented as Prisma `Decimal(10, 2)`.

3. **Dispatch Specification (`DISPATCH.md`)**:
   - Lines 8-18: Dictates exact model field names and enum structures:
     - `Clinic`: `id`, `slug`, `name`, `phone`, `address`, `createdAt`, `updatedAt`.
     - `User`: `id`, `clinicId`, `name`, `email`, `passwordHash`, `specialty`, `phone`, `role: OWNER | SPECIALIST`, `isActive`, `googleAccessToken`, `googleRefreshToken`, `createdAt`, `updatedAt`.
     - `Service`: `id`, `clinicId`, `name`, `description`, `price: Decimal`, `durationMin: Int`, `isActive`, `createdAt`, `updatedAt`.
     - `SpecialistService`: `id`, `clinicId`, `specialistId`, `serviceId`, `createdAt`, with `@@unique([specialistId, serviceId])`.
     - `Product`: `id`, `clinicId`, `name`, `description`, `price: Decimal`, `stock: Int`, `sku`, `isActive`, `createdAt`, `updatedAt`.
     - `Customer`: `id`, `clinicId`, `name`, `email`, `phone`, `notes`, `createdAt`, `updatedAt`.
     - `Appointment`: `id`, `clinicId`, `specialistId`, `serviceId`, `customerId`, `dateTime`, `status: PENDING | CONFIRMED | COMPLETED | CANCELLED | NO_SHOW`, `notes`, `googleEventId`, `createdAt`, `updatedAt`.
     - `Order`: `id`, `clinicId`, `customerId`, `total: Decimal`, `paymentMethod: CASH | CARD | TRANSFER`, `status: PENDING | PAID | CANCELLED`, `paidAt`, `createdAt`, `updatedAt`.
     - `OrderItem`: `id`, `orderId`, `serviceId`, `productId`, `quantity: Int`, `unitPrice: Decimal`, `createdAt`.
     - `Waitlist`: `id`, `clinicId`, `customerId`, `serviceId`, `specialistId`, `preferredDate`, `status: WAITING | OFFERED | ACCEPTED | EXPIRED | CANCELLED`, `priority: Int`, `createdAt`, `updatedAt`.
   - Lines 19-26: Dictates canonical seed dataset: "Clínica Fisia" (slug: `clinica-fisia`, phone: `+52 81 8000 1122`, address: `Av. Humberto Lobo 555, San Pedro Garza García, NL`), Dr. Alejandro Morales (`alejandro.morales@fisia.mx`), 5 specialists (Lic. Sofia Garza, Lic. Mateo Rivas, Lic. Camila Torres, Lic. David Salgado, Lic. Valeria Peña), 6 core services (Evaluación Inicial, Sesión Terapia Física, Terapia Manual, Punción Seca, Masaje Descarga, Readaptación Funcional), `SpecialistService` assignment matrix, 3 products (Kinesiotape, Theraband, Gel Crioterápico), sample patient, completed appointment, and cash closing order.

4. **Workspace State**:
   - The workspace currently contains only documentation/spec files (`ORIGINAL_REQUEST.md`, `PROJECT.md`, `Roadmap.md`, `TEST_INFRA.md`, and `.agents/`).
   - No `package.json` or `prisma/schema.prisma` files exist yet.

---

## 2. Logic Chain

1. **Multitenant Isolation Rule**:
   - Observation 1 & 2 establish that strict tenant isolation is required at the data layer.
   - Therefore, every entity tied to clinic operations (`User`, `Service`, `SpecialistService`, `Product`, `Customer`, `Appointment`, `Order`, `Waitlist`) must have a foreign key `clinicId String` referencing `Clinic(id)` with `@@index([clinicId])`.
   - `OrderItem` links directly to `Order(id)` with `onDelete: Cascade`, preserving tenant containment through its parent order.

2. **Referential Integrity & Cascading Semantics**:
   - When a tenant clinic is deleted, all tenant data must be removed cleanly without leaving orphaned records -> `Clinic` foreign keys use `onDelete: Cascade`.
   - In healthcare operations, clinical history (appointments) must not be accidentally purged if a staff member leaves or a service is deprecated -> `Appointment` relations to `User`, `Service`, and `Customer` use `onDelete: Restrict`.
   - If a customer is deleted for privacy/GDPR compliance, financial transactions must retain monetary auditability -> `Order.customerId` and `Order.appointmentId` use `onDelete: SetNull`.

3. **Field Naming & Consistency**:
   - `DISPATCH.md` specifies `passwordHash` (for bcrypt storage), `durationMin` (for service duration in minutes), and `dateTime` (for appointment scheduling).
   - These names remove ambiguity, align directly with NextAuth credential verification and clinical scheduling logic, and support future Google Calendar integration via `googleAccessToken` and `googleRefreshToken`.

4. **Seed Script Idempotency**:
   - Acceptance criteria require `npx prisma db seed` to run cleanly without error.
   - Repeated runs must not crash due to unique constraint collisions (`email`, `slug`, `clinicId_name`, `clinicId_sku`, `specialistId_serviceId`).
   - Therefore, all entities in `prisma/seed.ts` are provisioned using `upsert` queries or pre-checked existence conditions.

---

## 3. Caveats

- **Database Connection Requirement**: Validating the Prisma schema syntax and executing `prisma db seed` requires a configured PostgreSQL database URL (`DATABASE_URL` in `.env`). In isolated or containerized test environments without live PostgreSQL, `prisma validate` operates offline, but `prisma db push` / `prisma db seed` requires a live database or mock test container.
- **Password Hashing Dependency**: `bcryptjs` is recommended over native `bcrypt` in Node/Next.js environments to prevent binary C++ recompilation issues across different architectures (macOS ARM64 vs Linux x86_64).

---

## 4. Conclusion

The specification for Milestone 1 (Multitenant Schema & Data Layer) is completely defined, validated, and documented:
1. The **exact production Prisma schema** (`prisma/schema.prisma`) defining all 10 models, 5 enums, multi-tenant `clinicId` foreign keys, cascading rules, decimal currency definitions, and compound indexes is fully detailed in `/Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_spec_miner_m1_1/analysis.md` §2.
2. The **database client singleton** (`src/lib/prisma.ts`) preventing connection pool exhaustion is fully specified in `analysis.md` §4.
3. The **idempotent seed script blueprint** (`prisma/seed.ts`) populating Clínica Fisia, Dr. Alejandro Morales, the 5 specialists, 6 services, matrix assignments, 3 products, patient, appointment, and cash closing order is fully implemented in `analysis.md` §5.
4. Downstream planner and builder agents can immediately proceed with creating `package.json`, `prisma/schema.prisma`, `src/lib/prisma.ts`, and `prisma/seed.ts` based on these exact blueprints.

---

## 5. Verification Method

To independently verify this specification:
1. **Inspect Analysis Report**:
   - File: `/Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_spec_miner_m1_1/analysis.md`
   - Confirm all 10 models are defined with exact types and relations.
   - Confirm all seed data elements matching `DISPATCH.md` are present.
2. **Schema Validation Execution (once implemented by builder)**:
   - Command: `npx prisma validate`
   - Expected Output: `The schema at prisma/schema.prisma is valid` with exit code `0`.
3. **Database Seed Execution (once database is connected)**:
   - Command: `npx prisma db seed`
   - Expected Output:
     - `✅ Clinic established: Clínica Fisia`
     - `✅ Owner provisioned: Dr. Alejandro Morales (alejandro.morales@fisia.mx)`
     - `✅ 5 Specialists provisioned successfully.`
     - `✅ 6 Core Services provisioned.`
     - `✅ SpecialistService assignment matrix linked (15 assignments).`
     - `✅ 3 Retail Products seeded.`
     - `✅ Sample paid order for daily cash closing created ($1000.00 CASH)`
     - Exit code `0`.
4. **Invalidation Conditions**:
   - Any missing model among the 10 models.
   - Any omission of `clinicId` on tenant models.
   - Any schema syntax error reported by `npx prisma validate`.
   - Non-idempotent seed execution failing on second invocation.
