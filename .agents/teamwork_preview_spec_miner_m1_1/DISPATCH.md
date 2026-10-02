# Dispatch for teamwork_preview_spec_miner_m1_1

## Scope: Milestone 1 — Multitenant Schema & Data Models Specification
Read `/Users/jasbotello/devel/APPCITAS/ORIGINAL_REQUEST.md` and `/Users/jasbotello/devel/APPCITAS/PROJECT.md`.
Also review `/Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_spec_miner_survey_1/analysis.md`.

Your mission for Milestone 1:
1. Formulate the exact, production-ready `prisma/schema.prisma` definitions for all 10 models:
   - `Clinic` (tenant root: id, slug, name, phone, address, createdAt, updatedAt)
   - `User` (id, clinicId, name, email, passwordHash, specialty, phone, role: OWNER | SPECIALIST, isActive, googleAccessToken, googleRefreshToken, createdAt, updatedAt)
   - `Service` (id, clinicId, name, description, price: Decimal, durationMin: Int, isActive, createdAt, updatedAt)
   - `SpecialistService` (id, clinicId, specialistId, serviceId, createdAt; unique @@unique([specialistId, serviceId]))
   - `Product` (id, clinicId, name, description, price: Decimal, stock: Int, sku, isActive, createdAt, updatedAt)
   - `Customer` (id, clinicId, name, email, phone, notes, createdAt, updatedAt)
   - `Appointment` (id, clinicId, specialistId, serviceId, customerId, dateTime, status: PENDING|CONFIRMED|COMPLETED|CANCELLED|NO_SHOW, notes, googleEventId, createdAt, updatedAt)
   - `Order` (id, clinicId, customerId, total: Decimal, paymentMethod: CASH|CARD|TRANSFER, status: PENDING|PAID|CANCELLED, paidAt, createdAt, updatedAt)
   - `OrderItem` (id, orderId, serviceId, productId, quantity: Int, unitPrice: Decimal, createdAt)
   - `Waitlist` (id, clinicId, customerId, serviceId, specialistId, preferredDate, status: WAITING|OFFERED|ACCEPTED|EXPIRED, priority: Int, createdAt, updatedAt)
2. Detail the exact `prisma/seed.ts` script provisioning:
   - "Clínica Fisia" (slug: `clinica-fisia`, phone: `+52 81 8000 1122`, address: `Av. Humberto Lobo 555, San Pedro Garza García, NL`)
   - Owner Dr. Alejandro Morales (`alejandro.morales@fisia.mx`)
   - 5 specialists: Lic. Sofia Garza (Deportiva), Lic. Mateo Rivas (Ortopédica), Lic. Camila Torres (Neurológica), Lic. David Salgado (Pediátrica), Lic. Valeria Peña (Miofascial)
   - Core services: Evaluación Inicial ($850, 60m), Sesión Terapia Física ($750, 45m), Terapia Manual ($800, 45m), Punción Seca ($900, 45m), Masaje Descarga ($850, 60m), Readaptación Funcional ($800, 50m)
   - Assignments in `SpecialistService` linking each specialist to their appropriate services
   - Products: Kinesiotape ($250), Theraband ($350), Gel Crioterápico ($180)
   - Sample patient, completed appointment, and paid order with cash closing verification data.

Write your report to `/Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_spec_miner_m1_1/analysis.md` and handoff to `/Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_spec_miner_m1_1/handoff.md`.
Communicate back via send_message when done.

## 2026-09-14T04:10:00Z
Received user request:
Produce the exact production Prisma schema (all 10 models, fields, enums, relations, cascade rules, clinicId multitenancy) and idempotent seed script blueprint for Milestone 1.
Write your analysis to /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_spec_miner_m1_1/analysis.md and handoff to /Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_spec_miner_m1_1/handoff.md.
Communicate back via send_message when done.
