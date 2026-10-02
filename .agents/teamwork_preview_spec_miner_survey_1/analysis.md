# Exhaustive Specification Mining Report: APPCITAS Phase 1 (El Cimiento)

**Date**: 2026-09-14  
**Author**: teamwork_preview_spec_miner_survey_1  
**Target System**: APPCITAS — Multitenant Physiotherapy Clinic Management Platform  
**Authoritative Sources**:
1. `/Users/jasbotello/devel/APPCITAS/ORIGINAL_REQUEST.md`
2. `/Users/jasbotello/devel/APPCITAS/Roadmap.md`
3. `.agents/teamwork_preview_orchestrator/BRIEFING.md`

---

## 1. Executive Summary

Phase 1 ("El Cimiento") establishes the core multi-tenant architectural foundation for APPCITAS. It provides:
1. A 10-model PostgreSQL database schema managed with Prisma ORM, enforcing strict tenant isolation via `clinicId`.
2. NextAuth.js credentials-based authentication with session-level tenant and role context.
3. Direct clinic owner registration and frictionless onboarding of up to 5 specialist physiotherapists without requiring external email SMTP infrastructure.
4. An automated database seed script (`prisma/seed.ts`) preloading "Clínica Fisia", its owner, 5 specialists with realistic clinical specialties, core physiotherapy services with staff assignments, retail products, and sample appointment/order data.
5. A modern, responsive clinic dashboard (Overview, Team Roster, Services Catalog) adhering to Apple-inspired and minimalist healthcare design standards, with real-time feedback via Sonner toasts.

---

## 2. Complete Prisma Schema Specification (10 Models)

### 2.1 Multitenancy Architecture
- **Root Tenant Model**: `Clinic`.
- **Tenant Key**: `clinicId` (String / CUID).
- **Tenant Scope Rule**: Every model directly or indirectly belonging to a clinic MUST store `clinicId` and enforce foreign key relations to `Clinic(id)` with `onDelete: Cascade`.
- **Security Boundary**: All API routes, server actions, and Prisma queries MUST include `where: { clinicId: session.user.clinicId }` to prevent cross-tenant data leakage.

### 2.2 Enums Specification

```prisma
enum Role {
  OWNER
  SPECIALIST
}

enum AppointmentStatus {
  PENDING
  CONFIRMED
  COMPLETED
  CANCELLED
  NO_SHOW
}

enum PaymentMethod {
  CASH
  CARD
  TRANSFER
}

enum OrderStatus {
  PENDING
  PAID
  CANCELLED
}

enum WaitlistStatus {
  WAITING
  NOTIFIED
  BOOKED
  EXPIRED
  CANCELLED
}
```

### 2.3 Field-by-Field Specification of the 10 Models

#### 1. Model: `Clinic`
*The primary tenant entity representing a physiotherapy clinic.*
- `id`: `String` (`@id @default(cuid())`) — Unique tenant identifier.
- `slug`: `String` (`@unique`) — URL-friendly slug used for public booking (e.g. `tuapp.com/clinica-fisia` in Phase 2).
- `name`: `String` — Official business/clinic name (e.g. "Clínica Fisia").
- `phone`: `String?` — Primary contact telephone/WhatsApp for the clinic.
- `address`: `String?` — Physical address of the clinic.
- `createdAt`: `DateTime` (`@default(now())`) — Timestamp of creation.
- `updatedAt`: `DateTime` (`@updatedAt`) — Timestamp of last update.
- **Relations**:
  - `users`: `User[]` — Staff members (owners and specialists).
  - `services`: `Service[]` — Services offered by the clinic.
  - `specialistServices`: `SpecialistService[]` — Specialist-to-service assignments.
  - `products`: `Product[]` — Inventory/supplies catalog.
  - `customers`: `Customer[]` — Registered patients.
  - `appointments`: `Appointment[]` — Scheduled therapy sessions.
  - `orders`: `Order[]` — Billing and cash transactions.
  - `waitlists`: `Waitlist[]` — Queue for cancelled slot backfilling.

#### 2. Model: `User`
*Represents clinic staff: the clinic Owner and Specialist Physiotherapists.*
- `id`: `String` (`@id @default(cuid())`) — Unique user identifier.
- `clinicId`: `String` — Multitenancy foreign key referencing `Clinic.id`.
- `name`: `String` — Full name of the professional (e.g. "Dr. Alejandro Morales", "Lic. Sofia Mendoza").
- `email`: `String` (`@unique`) — Unique email address used for NextAuth credentials login.
- `password`: `String` — Securely hashed credentials (bcrypt).
- `role`: `Role` (`@default(SPECIALIST)`) — `OWNER` or `SPECIALIST`.
- `specialty`: `String?` — Clinical specialty (e.g. "Fisioterapia Deportiva", "Rehabilitación Neurológica").
- `phone`: `String?` — Direct contact phone for scheduling coordination.
- `isActive`: `Boolean` (`@default(true)`) — Active status flag; soft deletion if disabled.
- `googleAccessToken`: `String?` — Optional token storage for Phase 2 Google Calendar sync.
- `googleRefreshToken`: `String?` — Optional refresh token storage for Phase 2 Google Calendar sync.
- `createdAt`: `DateTime` (`@default(now())`) — Timestamp of account creation.
- `updatedAt`: `DateTime` (`@updatedAt`) — Timestamp of account update.
- **Relations**:
  - `clinic`: `Clinic` (`@relation(fields: [clinicId], references: [id], onDelete: Cascade)`)
  - `specialistServices`: `SpecialistService[]` — Services this specialist is assigned to perform.
  - `appointments`: `Appointment[]` — Appointments assigned to this specialist.
  - `waitlists`: `Waitlist[]` — Waitlist requests requesting this specialist.
- **Indexes**:
  - `@@index([clinicId])`
  - `@@index([email])`

#### 3. Model: `Service`
*Clinic-wide catalog of healthcare/physiotherapy services.*
- `id`: `String` (`@id @default(cuid())`) — Unique service identifier.
- `clinicId`: `String` — Multitenancy foreign key referencing `Clinic.id`.
- `name`: `String` — Name of the service (e.g. "Valoración Inicial", "Sesión de Fisioterapia Deportiva").
- `description`: `String?` — Detailed clinical description of treatment.
- `duration`: `Int` — Session duration in minutes (e.g. 45, 50, 60).
- `price`: `Decimal` (`@db.Decimal(10, 2)`) — Base price in local currency (MXN).
- `isActive`: `Boolean` (`@default(true)`) — Service active status flag.
- `createdAt`: `DateTime` (`@default(now())`) — Timestamp of creation.
- `updatedAt`: `DateTime` (`@updatedAt`) — Timestamp of update.
- **Relations**:
  - `clinic`: `Clinic` (`@relation(fields: [clinicId], references: [id], onDelete: Cascade)`)
  - `specialistServices`: `SpecialistService[]` — Join table records linking specialists.
  - `appointments`: `Appointment[]` — Appointments booked for this service.
  - `orderItems`: `OrderItem[]` — Order billing line items.
  - `waitlists`: `Waitlist[]` — Waitlist entries for this service.
- **Indexes**:
  - `@@index([clinicId])`
  - `@@unique([clinicId, name])` — Prevents duplicate service names within a clinic.

#### 4. Model: `SpecialistService`
*Explicit join table supporting flexible many-to-many assignments between Specialists and Services.*
- `id`: `String` (`@id @default(cuid())`) — Unique assignment identifier.
- `clinicId`: `String` — Multitenancy foreign key referencing `Clinic.id`.
- `specialistId`: `String` — Foreign key referencing `User.id`.
- `serviceId`: `String` — Foreign key referencing `Service.id`.
- `createdAt`: `DateTime` (`@default(now())`) — Timestamp of assignment creation.
- **Relations**:
  - `clinic`: `Clinic` (`@relation(fields: [clinicId], references: [id], onDelete: Cascade)`)
  - `specialist`: `User` (`@relation(fields: [specialistId], references: [id], onDelete: Cascade)`)
  - `service`: `Service` (`@relation(fields: [serviceId], references: [id], onDelete: Cascade)`)
- **Indexes & Constraints**:
  - `@@unique([specialistId, serviceId])` — Guarantees no duplicate assignment.
  - `@@index([clinicId])`
  - `@@index([specialistId])`
  - `@@index([serviceId])`

#### 5. Model: `Product`
*Physical supplies, retail goods, and clinic inventory.*
- `id`: `String` (`@id @default(cuid())`) — Unique product identifier.
- `clinicId`: `String` — Multitenancy foreign key referencing `Clinic.id`.
- `name`: `String` — Product name (e.g. "Kinesiotape CureTape Pro", "Gel Criogénico").
- `description`: `String?` — Product description and usage.
- `price`: `Decimal` (`@db.Decimal(10, 2)`) — Retail selling price.
- `stock`: `Int` (`@default(0)`) — Available inventory quantity.
- `sku`: `String?` — Stock Keeping Unit / barcode identifier.
- `isActive`: `Boolean` (`@default(true)`) — Product active status flag.
- `createdAt`: `DateTime` (`@default(now())`)
- `updatedAt`: `DateTime` (`@updatedAt`)
- **Relations**:
  - `clinic`: `Clinic` (`@relation(fields: [clinicId], references: [id], onDelete: Cascade)`)
  - `orderItems`: `OrderItem[]` — Order billing line items.
- **Indexes**:
  - `@@index([clinicId])`
  - `@@unique([clinicId, sku])`

#### 6. Model: `Customer`
*Patient records including contact details for appointments and WhatsApp integrations.*
- `id`: `String` (`@id @default(cuid())`) — Unique patient identifier.
- `clinicId`: `String` — Multitenancy foreign key referencing `Clinic.id`.
- `name`: `String` — Patient full name.
- `email`: `String?` — Patient email.
- `phone`: `String` — Patient phone number (E.164 format, required for Phase 3 WhatsApp Meta Cloud API).
- `notes`: `String?` — Clinical observations, history notes, or alerts.
- `createdAt`: `DateTime` (`@default(now())`)
- `updatedAt`: `DateTime` (`@updatedAt`)
- **Relations**:
  - `clinic`: `Clinic` (`@relation(fields: [clinicId], references: [id], onDelete: Cascade)`)
  - `appointments`: `Appointment[]`
  - `orders`: `Order[]`
  - `waitlists`: `Waitlist[]`
- **Indexes**:
  - `@@index([clinicId])`
  - `@@index([clinicId, phone])`

#### 7. Model: `Appointment`
*Scheduled physiotherapy session linking clinic, patient, specialist, and service.*
- `id`: `String` (`@id @default(cuid())`) — Unique appointment identifier.
- `clinicId`: `String` — Multitenancy foreign key referencing `Clinic.id`.
- `specialistId`: `String` — Foreign key referencing `User.id` (assigned physiotherapist).
- `serviceId`: `String` — Foreign key referencing `Service.id`.
- `customerId`: `String` — Foreign key referencing `Customer.id` (patient).
- `startTime`: `DateTime` — Start timestamp of session.
- `endTime`: `DateTime` — End timestamp of session.
- `status`: `AppointmentStatus` (`@default(CONFIRMED)`) — Status lifecycle (`PENDING`, `CONFIRMED`, `COMPLETED`, `CANCELLED`, `NO_SHOW`).
- `notes`: `String?` — Appointment-specific notes.
- `googleEventId`: `String?` — Google Calendar event identifier (for Phase 2 bidirectional sync).
- `createdAt`: `DateTime` (`@default(now())`)
- `updatedAt`: `DateTime` (`@updatedAt`)
- **Relations**:
  - `clinic`: `Clinic` (`@relation(fields: [clinicId], references: [id], onDelete: Cascade)`)
  - `specialist`: `User` (`@relation(fields: [specialistId], references: [id], onDelete: Restrict)`)
  - `service`: `Service` (`@relation(fields: [serviceId], references: [id], onDelete: Restrict)`)
  - `customer`: `Customer` (`@relation(fields: [customerId], references: [id], onDelete: Restrict)`)
  - `order`: `Order?` — Optional associated billing order.
- **Indexes**:
  - `@@index([clinicId])`
  - `@@index([clinicId, specialistId, startTime])`
  - `@@index([clinicId, status])`

#### 8. Model: `Order`
*Order and payment record foundation for daily cash closing (Corte de Caja).*
- `id`: `String` (`@id @default(cuid())`) — Unique order identifier.
- `clinicId`: `String` — Multitenancy foreign key referencing `Clinic.id`.
- `customerId`: `String?` — Foreign key referencing `Customer.id` (nullable for anonymous retail sales).
- `appointmentId`: `String?` (`@unique`) — Foreign key referencing `Appointment.id` (one-to-one).
- `total`: `Decimal` (`@db.Decimal(10, 2)`) — Total order amount.
- `paymentMethod`: `PaymentMethod` — Method of payment (`CASH`, `CARD`, `TRANSFER`).
- `status`: `OrderStatus` (`@default(PAID)`) — Order payment status (`PENDING`, `PAID`, `CANCELLED`).
- `notes`: `String?` — Payment reference or transaction notes.
- `paidAt`: `DateTime` (`@default(now())`) — Timestamp of payment completion (critical for daily cash closing).
- `createdAt`: `DateTime` (`@default(now())`)
- `updatedAt`: `DateTime` (`@updatedAt`)
- **Relations**:
  - `clinic`: `Clinic` (`@relation(fields: [clinicId], references: [id], onDelete: Cascade)`)
  - `customer`: `Customer?` (`@relation(fields: [customerId], references: [id], onDelete: SetNull)`)
  - `appointment`: `Appointment?` (`@relation(fields: [appointmentId], references: [id], onDelete: SetNull)`)
  - `items`: `OrderItem[]` — Itemized billing details.
- **Indexes**:
  - `@@index([clinicId])`
  - `@@index([clinicId, paidAt, status])`

#### 9. Model: `OrderItem`
*Line item in an order representing either a clinical service or a retail product.*
- `id`: `String` (`@id @default(cuid())`) — Unique item identifier.
- `orderId`: `String` — Foreign key referencing `Order.id`.
- `serviceId`: `String?` — Foreign key referencing `Service.id` (nullable).
- `productId`: `String?` — Foreign key referencing `Product.id` (nullable).
- `quantity`: `Int` (`@default(1)`) — Number of units or sessions.
- `unitPrice`: `Decimal` (`@db.Decimal(10, 2)`) — Price snapshot at time of checkout.
- `subtotal`: `Decimal` (`@db.Decimal(10, 2)`) — Calculated line item total (`quantity * unitPrice`).
- `createdAt`: `DateTime` (`@default(now())`)
- **Relations**:
  - `order`: `Order` (`@relation(fields: [orderId], references: [id], onDelete: Cascade)`)
  - `service`: `Service?` (`@relation(fields: [serviceId], references: [id], onDelete: SetNull)`)
  - `product`: `Product?` (`@relation(fields: [productId], references: [id], onDelete: SetNull)`)
- **Indexes**:
  - `@@index([orderId])`

#### 10. Model: `Waitlist`
*Patient cancellation and waitlist queue for automatic slot backfilling via WhatsApp (Phase 3).*
- `id`: `String` (`@id @default(cuid())`) — Unique waitlist entry identifier.
- `clinicId`: `String` — Multitenancy foreign key referencing `Clinic.id`.
- `customerId`: `String` — Foreign key referencing `Customer.id`.
- `serviceId`: `String?` — Preferred service requested.
- `specialistId`: `String?` — Preferred specialist requested (null = any available).
- `preferredDate`: `DateTime?` — Target date for appointment.
- `preferredTimeRange`: `String?` — Desired time window (e.g. "Mañanas", "16:00 - 19:00").
- `status`: `WaitlistStatus` (`@default(WAITING)`) — Lifecycle (`WAITING`, `NOTIFIED`, `BOOKED`, `EXPIRED`, `CANCELLED`).
- `priority`: `Int` (`@default(0)`) — Priority score for queue ordering.
- `notes`: `String?` — Patient symptoms or urgency remarks.
- `notifiedAt`: `DateTime?` — Timestamp when WhatsApp alert was dispatched.
- `createdAt`: `DateTime` (`@default(now())`)
- `updatedAt`: `DateTime` (`@updatedAt`)
- **Relations**:
  - `clinic`: `Clinic` (`@relation(fields: [clinicId], references: [id], onDelete: Cascade)`)
  - `customer`: `Customer` (`@relation(fields: [customerId], references: [id], onDelete: Cascade)`)
  - `service`: `Service?` (`@relation(fields: [serviceId], references: [id], onDelete: SetNull)`)
  - `specialist`: `User?` (`@relation(fields: [specialistId], references: [id], onDelete: SetNull)`)
- **Indexes**:
  - `@@index([clinicId, status])`
  - `@@index([clinicId, preferredDate])`

---

## 3. Pilot Seed Data Specifications (`prisma/seed.ts`)

The seed script MUST be idempotent (using `upsert` or clear-and-insert) and populate the following canonical dataset:

### 3.1 Pilot Clinic
- **Name**: `Clínica Fisia`
- **Slug**: `clinica-fisia`
- **Phone**: `+52 55 8432 1098`
- **Address**: `Av. Insurgentes Sur 1234, Col. Del Valle, Benito Juárez, CDMX`

### 3.2 Clinic Owner Account
- **Name**: `Dr. Alejandro Morales`
- **Email**: `admin@fisia.com`
- **Password**: Hashed version of `Fisia2026!` (bcrypt salt 10)
- **Role**: `OWNER`
- **Specialty**: `Director Clínico & Fisioterapia Integral`
- **Phone**: `+52 55 8432 1000`
- **IsActive**: `true`

### 3.3 The 5 Specialists
1. **Sofia Mendoza**:
   - Email: `sofia.mendoza@fisia.com`
   - Role: `SPECIALIST`
   - Specialty: `Fisioterapia Deportiva y Readaptación`
   - Phone: `+52 55 8432 1001`
   - Password: Hashed `Fisia2026!`
   - IsActive: `true`
2. **Mateo Valencia**:
   - Email: `mateo.valencia@fisia.com`
   - Role: `SPECIALIST`
   - Specialty: `Terapia Manual Ortopédica y Columna`
   - Phone: `+52 55 8432 1002`
   - Password: Hashed `Fisia2026!`
   - IsActive: `true`
3. **Camila Herrera**:
   - Email: `camila.herrera@fisia.com`
   - Role: `SPECIALIST`
   - Specialty: `Rehabilitación Neurológica`
   - Phone: `+52 55 8432 1003`
   - Password: Hashed `Fisia2026!`
   - IsActive: `true`
4. **Daniel Rios**:
   - Email: `daniel.rios@fisia.com`
   - Role: `SPECIALIST`
   - Specialty: `Punción Seca y Dolor Miofascial`
   - Phone: `+52 55 8432 1004`
   - Password: Hashed `Fisia2026!`
   - IsActive: `true`
5. **Valentina Gómez**:
   - Email: `valentina.gomez@fisia.com`
   - Role: `SPECIALIST`
   - Specialty: `Fisioterapia Pediátrica y Postural`
   - Phone: `+52 55 8432 1005`
   - Password: Hashed `Fisia2026!`
   - IsActive: `true`

### 3.4 Core Physiotherapy Services & Specialist Assignments
1. **Valoración Inicial de Fisioterapia**:
   - Duration: `60` min | Price: `$850.00 MXN`
   - Assigned: All 5 Specialists + Owner
2. **Sesión de Fisioterapia Deportiva**:
   - Duration: `50` min | Price: `$750.00 MXN`
   - Assigned: Sofia Mendoza, Mateo Valencia, Daniel Rios
3. **Terapia Manual Ortopédica**:
   - Duration: `45` min | Price: `$700.00 MXN`
   - Assigned: Mateo Valencia, Daniel Rios, Sofia Mendoza
4. **Punción Seca y Electroterapia**:
   - Duration: `45` min | Price: `$800.00 MXN`
   - Assigned: Daniel Rios, Sofia Mendoza
5. **Rehabilitación Neurológica**:
   - Duration: `60` min | Price: `$900.00 MXN`
   - Assigned: Camila Herrera
6. **Fisioterapia Pediátrica**:
   - Duration: `45` min | Price: `$750.00 MXN`
   - Assigned: Valentina Gómez

### 3.5 Core Retail Products (Inventory)
1. **Kinesiotape CureTape Pro (5cm x 5m)**: Price: `$280.00 MXN` | Stock: `25` | SKU: `KT-PRO-01`
2. **Banda Elástica Theraband Nivel Medio (Azul)**: Price: `$220.00 MXN` | Stock: `15` | SKU: `TB-MED-BL`
3. **Gel Criogénico Desinflamatorio Fisia 250ml**: Price: `$190.00 MXN` | Stock: `30` | SKU: `GEL-CRIO-250`
4. **Foam Roller de Alta Densidad 45cm**: Price: `$450.00 MXN` | Stock: `10` | SKU: `FR-HD-45`

### 3.6 Sample Operational Data (Appointments, Orders & Waitlist)
- **Patient**: Carlos Fernández (`+52 55 9876 5432`, `carlos.fernandez@gmail.com`)
- **Appointment**: Valoración Inicial with Sofia Mendoza (Tomorrow 10:00 - 11:00 AM, `CONFIRMED`)
- **Order**:
  - Total: `$1,030.00 MXN`
  - PaymentMethod: `CASH`
  - Status: `PAID`
  - Items: 1x Sesión Deportiva ($750.00) + 1x Kinesiotape ($280.00)
- **Waitlist**: Ana Lucía Castillo (`+52 55 1122 3344`, waiting for Fisioterapia Deportiva with Sofia Mendoza, Priority `1`, `WAITING`)

---

## 4. Authentication, Authorization & Onboarding Specifications

### 4.1 NextAuth Credentials Architecture
- **Provider**: NextAuth `CredentialsProvider`.
- **Session Strategy**: `jwt`.
- **Callbacks**:
  - `jwt({ token, user })`: When `user` is returned from `authorize`, populate `token.id`, `token.role`, `token.clinicId`, `token.clinicSlug`, `token.clinicName`.
  - `session({ session, token })`: Expose `session.user.id`, `session.user.role`, `session.user.clinicId`, `session.user.clinicSlug`, `session.user.clinicName`.
- **Type Augmentation**: Augment `Session` and `JWT` interfaces in `types/next-auth.d.ts` to ensure strict TypeScript compilation.

### 4.2 Route Protection
- **Middleware**: `middleware.ts` guarding `/dashboard/:path*` and `/api/dashboard/:path*`.
- **Unauthenticated Handling**: Redirect unauthenticated requests to `/login?callbackUrl=...`.
- **Role-Based Access Control**:
  - `OWNER`: Full access to clinic settings, financial closing, service creation, and specialist management.
  - `SPECIALIST`: Access restricted to their appointments, assigned services, and daily schedule. Attempting to access owner endpoints returns `403 Forbidden`.

### 4.3 Direct Specialist Onboarding Without External SMTP
- **Rationale**: Eliminates reliance on external transactional mail services (SendGrid/Resend) during Phase 1 deployment.
- **Workflow**:
  1. Owner accesses `/dashboard/team`.
  2. Clicks "Onboard Specialist" button which opens a modal dialog.
  3. Owner inputs:
     - Full Name
     - Email
     - Specialty
     - Phone Number
     - Temporary Password (with an automated "Generar Contraseña Segura" generator button).
  4. Form submits to `POST /api/specialists`.
  5. Server validates:
     - Active session with `role === 'OWNER'`.
     - Max specialists quota (up to 5 specialists).
     - Unique email.
  6. Server hashes password and persists `User` with `clinicId = session.user.clinicId` and `role = 'SPECIALIST'`.
  7. Server responds with created user details.
  8. Modal updates to show a "Credenciales Creadas" card with a prominent "Copiar al portapapeles" (Copy to clipboard) button, enabling immediate delivery via WhatsApp or direct communication.
  9. A success toast is displayed via Sonner.

---

## 5. Acceptance Criteria Checklist

| Category | Requirement | Validation Method |
|---|---|---|
| **Data & Architecture** | 10 Prisma models defined with correct types and relations | `npx prisma validate` |
| **Data & Architecture** | Multitenancy enforced via `clinicId` across tenant models | Code inspection of schema & indexes |
| **Data & Architecture** | Seed script provisions Clínica Fisia, owner, 5 specialists, services, products, order | `npx prisma db seed` |
| **Auth & Security** | Owner registration & NextAuth credentials login | End-to-end authentication flow |
| **Auth & Security** | Dashboard route protection redirecting to `/login` | Middleware route test on `/dashboard` |
| **Auth & Security** | Direct specialist onboarding without external SMTP | Create specialist in modal & verify DB |
| **Build & UI** | 0 TypeScript & Next.js packaging errors | `npm run build` |
| **Build & UI** | Responsive UI on desktop and mobile viewports | Viewport inspection (Overview, Team, Services) |
| **Build & UI** | Real-time notifications and feedback | Sonner toast triggers on mutation |

---

## 6. Features Discovered

## Features Discovered
| # | Category | Feature | Description | Inputs | Outputs | Error Behavior | Discovered Via |
|---|----------|---------|-------------|--------|---------|----------------|----------------|
| 1 | Database | Clinic Tenant Root | Multi-tenant organization container with slug, contact, address | Clinic name, slug, phone, address | Clinic record | Unique slug violation returns 409 | ORIGINAL_REQUEST R1 |
| 2 | Database | User Credentials & Profile | Multi-role user accounts (OWNER, SPECIALIST) with hashed passwords | Name, email, hashed password, role, specialty, phone | User record | Duplicate email returns 409 | ORIGINAL_REQUEST R1 & R2 |
| 3 | Database | Service Catalog | Clinic service offering with duration in minutes and decimal price | Name, duration, price, description | Service record | Duplicate service name per clinic returns 400 | ORIGINAL_REQUEST R1 |
| 4 | Database | SpecialistService Join Model | Explicit N:M relationship mapping specialists to eligible services | specialistId, serviceId, clinicId | SpecialistService record | Duplicate assignment returns constraint error | ORIGINAL_REQUEST R1 |
| 5 | Database | Product Inventory | Therapy supplies and retail products for sales and cash closing | Name, price, stock, sku | Product record | Negative stock prevented by validation | ORIGINAL_REQUEST R1 |
| 6 | Database | Customer Patient Registry | Patient contact records with phone required for WhatsApp in Phase 3 | Name, phone, email, notes | Customer record | Missing phone returns 400 | ORIGINAL_REQUEST R1 & Roadmap Fase 3 |
| 7 | Database | Appointment Booking | Clinical session linking patient, specialist, and service | specialistId, serviceId, customerId, startTime, endTime, status | Appointment record | Conflicting appointment slot detected | ORIGINAL_REQUEST R1 & Roadmap Fase 2 |
| 8 | Database | Order & OrderItem Foundation | Order tracking with CASH, CARD, TRANSFER for daily cash closing | customerId, paymentMethod, items list, total | Order with OrderItems | Invalid paymentMethod enum returns 400 | ORIGINAL_REQUEST R1 & Roadmap Fase 3 |
| 9 | Database | Waitlist Queue | Queue for automated backfilling of cancelled appointments | customerId, serviceId, specialistId, preferredDate, status | Waitlist record | Missing customerId returns 400 | ORIGINAL_REQUEST R1 & Roadmap Fase 3 |
| 10 | Seed | Pilot Seed Script | Automated data seeding for Clínica Fisia with 5 specialists & catalog | `npx prisma db seed` | Seeded database with pilot data | Script failure with non-zero exit code | ORIGINAL_REQUEST R1 |
| 11 | Auth | NextAuth Credentials Login | Secure authentication using email and bcrypt-hashed password | email, password | JWT token with session { id, role, clinicId, clinicSlug } | Invalid credentials returns 401 | ORIGINAL_REQUEST R2 |
| 12 | Auth | Route Protection Middleware | Prevents unauthorized access to `/dashboard/*` | Request headers / cookies | Allowed access or redirect to `/login` | Unauthenticated redirected to `/login` | ORIGINAL_REQUEST R2 |
| 13 | Auth | Clinic Registration | Public owner and clinic registration flow | Clinic name, owner name, email, password | Clinic and User records created | Email already registered returns 400 | ORIGINAL_REQUEST R2 |
| 14 | Onboarding | Direct Specialist Onboarding | Owner adds therapist with temporary password, no SMTP needed | Name, email, specialty, phone, tempPassword | Created specialist User record | Quota exceeded (>5) returns 403/400 | ORIGINAL_REQUEST R2 |
| 15 | UI | Overview Dashboard | KPI statistics cards, schedule preview, quick action shortcuts | None (authenticated session) | Responsive dashboard page | Auth failure redirects to login | ORIGINAL_REQUEST R3 |
| 16 | UI | Team Roster Page | Interactive grid of 5 specialists, status badges, onboard modal | Specialist data | Roster table and modal form | Form validation errors highlighted | ORIGINAL_REQUEST R3 |
| 17 | UI | Services Catalog Page | Service cards/table with duration, price, and specialist assignment | Service assignments | Updated SpecialistService links | Unauthorized assignment returns 403 | ORIGINAL_REQUEST R3 |
| 18 | UI | Sonner Feedback System | Toast notifications for async actions, errors, and clipboard copies | Toast call message & type | Floating notification toast | Fallback gracefully on dismiss | ORIGINAL_REQUEST Acceptance Criteria |
| 19 | Cash Close | Daily Cash Closing (Foundation) | Aggregation of orders by payment method (CASH, CARD, TRANSFER) | Target date | Total revenue and breakdown by method | Invalid date returns 400 | ORIGINAL_REQUEST R1 & Roadmap Fase 3 |
| 20 | Calendar | Google Calendar OAuth Prep | Schema storage for Google access and refresh tokens | OAuth callback tokens | Stored tokens on User model | Invalid token returns error | Roadmap Fase 2 |

---

## 7. Edge Cases & Boundary Conditions

## Edge Cases
| # | Feature | Input | Observed Behavior |
|---|---------|-------|-------------------|
| 1 | Multitenancy Isolation | Specialist from Clinic A queries `/api/services` | Query MUST filter by `where: { clinicId: session.user.clinicId }`. Returns ONLY Clinic A services. Never leaks Clinic B data. |
| 2 | Multitenancy Cross-Linking | Owner tries to assign a Specialist from Clinic B to a Service in Clinic A | Validation detects `specialist.clinicId !== session.user.clinicId` and rejects with `403 Forbidden`. |
| 3 | Specialist Onboarding Quota | Owner tries to onboard a 6th specialist when 5 already exist | Application checks current count of active specialists; rejects or warns that Phase 1 pilot tier is limited to 5 specialists. |
| 4 | Direct Specialist Onboarding | Owner generates temporary credentials without SMTP | Password is encrypted into DB; plaintext is returned ONCE to owner modal with a "Copiar Credenciales" button. Never logged to console in production. |
| 5 | Unauthenticated Route Access | Guest accesses `/dashboard/services` | Middleware intercepts request, preserves path in `callbackUrl`, and redirects to `/login?callbackUrl=/dashboard/services`. |
| 6 | Specialist Accessing Admin Area | Specialist accesses `/dashboard/team` onboarding action | Server action/API verifies `session.user.role === 'OWNER'`; returns `403 Forbidden` for non-owner. |
| 7 | Duplicate Email Registration | New user attempts registration with existing email | DB unique constraint catches conflict; API returns clean `409 Conflict` with "El correo electrónico ya está registrado". |
| 8 | Inactive Staff Login | Specialist has `isActive = false` and attempts login | NextAuth `authorize` validates `user.isActive`; rejects login with "Cuenta desactivada. Contacte al administrador". |
| 9 | Monetary Calculations | Aggregation of `Order.total` for cash closing | Decimal precision prevents floating point rounding errors (e.g. `0.1 + 0.2 !== 0.3`). Prisma `Decimal` maps to Postgres `DECIMAL(10, 2)`. |
| 10 | Service Deletion Cascade | Admin deactivates or deletes a service that has past orders/appointments | Hard delete is prevented by foreign key restriction (`onDelete: Restrict`); system utilizes soft delete (`isActive: false`) to protect historical accounting records. |
| 11 | Order with Mixed Items | Order containing both a physiotherapy session and a retail product (e.g. Kinesiotape) | `OrderItem` cleanly models either `serviceId` or `productId` with respective quantities, unit prices, and subtotals. |
| 12 | Seed Idempotency | Running `npx prisma db seed` multiple times | Script uses `upsert` or deletes existing test records cleanly so multiple runs succeed without unique constraint collisions. |
| 13 | Mobile Viewport Responsiveness | User opens Team Roster on a 375px mobile screen | Table shifts to responsive card layout or horizontal scroll; modal adapts to full screen width; buttons maintain min 44px touch target. |
| 14 | Public Booking Slug Collisions | Two clinics attempt to register the slug "clinica-fisia" | Slug uniqueness constraint triggers; slug generator appends random hash or ordinal (e.g. `clinica-fisia-2`). |
