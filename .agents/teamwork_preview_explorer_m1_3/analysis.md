# Database Validation, Seed Verification & Error Prevention Strategy
## APPCITAS Phase 1 (El Cimiento) — Milestone 1

**Document ID**: `APPCITAS-M1-VAL-01`  
**Author**: `teamwork_preview_explorer_m1_3`  
**Role**: Teamwork Preview Explorer — Database Integrity, Seed Verification & Error Prevention  
**Date**: 2026-09-14  
**Workspace**: `/Users/jasbotello/devel/APPCITAS`  
**Target Milestone**: Milestone 1 (Multitenant Schema & Data Layer)  
**Authoritative References**: `ORIGINAL_REQUEST.md`, `PROJECT.md`, `TEST_INFRA.md`, `Roadmap.md`

---

## 1. Executive Summary

Milestone 1 ("El Cimiento") lays the database and data access foundation for the entire APPCITAS multi-tenant platform. If this layer has hidden flaws — such as unindexed foreign keys, ambiguous Prisma relations, brittle seed scripts, silent Decimal serialization bugs in Next.js, or data leakage across clinic tenants — every subsequent milestone (M2 Authentication & Onboarding, M3 Dashboard UI, M-E2E Testing) will inherit compounding defects.

This document establishes:
1. **A Rigorous Database Validation Strategy**: How to programmatically and syntactically verify `prisma/schema.prisma` via `npx prisma validate` and Prisma Client type compilation (`prisma generate`).
2. **An Exhaustive Seed Verification Protocol**: How to verify that `npx prisma db seed` provisions the pilot clinic ("Clínica Fisia"), its owner, 5 specialists, 6 clinical services with specialist assignments, 4 retail products, and operational records with 100% accuracy and complete idempotency.
3. **Strict Multitenancy Isolation Rules**: Database-level foreign keys, index coverage, and validation rules preventing cross-tenant leakage.
4. **Deep-Dive Error Prevention & Edge Case Mitigation**:
   - Exact foreign key deletion and reset ordering to prevent PostgreSQL `ForeignKeyViolation` (P2003 / 23503).
   - Defensive solutions for Prisma `Decimal` class stringification and `BigInt` serialization failures in Next.js App Router RSC and Route Handlers.
   - Blueprint for 100% idempotent seeding using deterministic unique composites and hybrid wipe/upsert workflows.
   - Production-safe environment variable validation (`DATABASE_URL`, `NEXTAUTH_SECRET`, connection pooling / PgBouncer flags).

---

## 2. Schema Validation Strategy (`npx prisma validate`)

### 2.1 What `prisma validate` Guarantees (and What It Does Not)

`npx prisma validate` performs static Abstract Syntax Tree (AST) validation of the `prisma/schema.prisma` file without requiring an active database connection.

| Checks Performed by `prisma validate` | Checks NOT Performed by `prisma validate` |
| :--- | :--- |
| Schema syntax validity (keywords, braces, types) | Database connectivity or network reachability |
| Datasource & Generator configuration syntax | Table/column existence in physical PostgreSQL DB |
| Relation consistency (matching `fields` and `references`) | Runtime constraints (e.g. negative prices, empty strings) |
| Attribute validity (`@id`, `@unique`, `@@unique`, `@@index`) | Foreign key cross-tenant logical enforcement |
| Native type annotations (e.g. `@db.Decimal(10, 2)`) | Next.js runtime type serialization |
| Unique relation naming when multiple relations link same models | Seed script execution integrity |

### 2.2 Concrete Verification Procedure

To verify schema syntax and integrity during development and CI:

```bash
# 1. Execute schema AST validation
npx prisma validate --schema=prisma/schema.prisma

# 2. Verify client generation (verifies TypeScript types compile cleanly without name collisions)
npx prisma generate --schema=prisma/schema.prisma
```

#### Automated Script Check (`scripts/validate-schema.sh`)
```bash
#!/usr/bin/env bash
set -eo pipefail

echo "🔍 Validating Prisma schema syntax and relations..."
if ! npx prisma validate --schema=prisma/schema.prisma; then
  echo "❌ Prisma schema validation failed!"
  exit 1
fi

echo "⚙️ Generating Prisma Client to verify type bindings..."
if ! npx prisma generate --schema=prisma/schema.prisma; then
  echo "❌ Prisma Client generation failed!"
  exit 1
fi

echo "✅ Prisma schema is valid and TypeScript client generated successfully."
```

### 2.3 Critical Schema Constraints to Enforce During AST Review

When inspecting `prisma/schema.prisma`, the following relation rules must be satisfied:

1. **Relation Pair Symmetry**:
   - `User.clinic` ↔ `Clinic.users`
   - `Service.clinic` ↔ `Clinic.services`
   - `SpecialistService.clinic` ↔ `Clinic.specialistServices`
   - `SpecialistService.specialist` ↔ `User.specialistServices`
   - `SpecialistService.service` ↔ `Service.specialistServices`
   - `Appointment.specialist` ↔ `User.appointments`
   - `Appointment.service` ↔ `Service.appointments`
   - `Appointment.customer` ↔ `Customer.appointments`
   - `Order.appointment` ↔ `Appointment.order` (One-to-One: requires `@unique` on `Order.appointmentId`)
2. **Explicit Delete Rules**:
   - Root tenant cascade: Every tenant model has `onDelete: Cascade` to `Clinic`.
   - Historical protection: `Appointment` relations to `User`, `Service`, and `Customer` must use `onDelete: Restrict` so deleting a user or service cannot silently corrupt appointment history.
   - Financial protection: `OrderItem` relations to `Service` and `Product` must use `onDelete: SetNull` so deactivating or deleting catalog items preserves past sales orders.

---

## 3. Seed Verification Strategy (`npx prisma db seed`)

### 3.1 Pre-Flight Database Requirements

Before executing `npx prisma db seed`, the target PostgreSQL database must have the schema applied:
```bash
# Push schema directly to development PostgreSQL database
npx prisma db push --skip-generate
```

The `package.json` manifest must specify the runner in its `prisma` stanza:
```json
"prisma": {
  "seed": "tsx prisma/seed.ts"
}
```

### 3.2 Target Canonical Dataset

The seed script MUST provision the exact dataset required by `ORIGINAL_REQUEST §R1` and `PROJECT.md`:

```
Tenant: Clínica Fisia (slug: clinica-fisia)
├── Owner: Dr. Alejandro Morales (admin@fisia.com, Role: OWNER)
├── 5 Specialists (Role: SPECIALIST):
│   ├── Lic. Sofia Mendoza (Fisioterapia Deportiva y Readaptación)
│   ├── Lic. Mateo Valencia (Terapia Manual Ortopédica y Columna)
│   ├── Lic. Camila Herrera (Rehabilitación Neurológica)
│   ├── Lic. Daniel Rios (Punción Seca y Dolor Miofascial)
│   └── Lic. Valentina Gómez (Fisioterapia Pediátrica y Postural)
├── 6 Core Services:
│   ├── Valoración Inicial de Fisioterapia ($850, 60m) -> All 5 Specialists + Owner
│   ├── Sesión de Fisioterapia Deportiva ($750, 50m) -> Sofia, Mateo, Daniel
│   ├── Terapia Manual Ortopédica ($700, 45m) -> Mateo, Daniel, Sofia
│   ├── Punción Seca y Electroterapia ($800, 45m) -> Daniel, Sofia
│   ├── Rehabilitación Neurológica ($900, 60m) -> Camila
│   └── Fisioterapia Pediátrica ($750, 45m) -> Valentina
├── 4 Products:
│   ├── Kinesiotape CureTape Pro (SKU: KT-PRO-01, $280, Stock: 25)
│   ├── Banda Elástica Theraband (SKU: TB-MED-BL, $220, Stock: 15)
│   ├── Gel Criogénico Desinflamatorio (SKU: GEL-CRIO-250, $190, Stock: 30)
│   └── Foam Roller de Alta Densidad (SKU: FR-HD-45, $450, Stock: 10)
└── Operational Test Records:
    ├── 1 Patient: Carlos Fernández (phone: +525598765432)
    ├── 1 Appointment: CONFIRMED with Sofia Mendoza
    ├── 1 Order: PAID, CASH ($1,030 MXN) with 2 OrderItems (Service + Kinesiotape)
    └── 1 Waitlist Entry: WAITING, Priority 1
```

### 3.3 Quantitative & Qualitative Verification Matrix

After running `npx prisma db seed`, an automated verification test must query the database and validate each assertion:

| Entity | Query Filter | Expected Count / Value | Verification Rule |
| :--- | :--- | :--- | :--- |
| **Clinic** | `slug: 'clinica-fisia'` | Exactly 1 record | `name === 'Clínica Fisia'`, `phone !== null`, `address !== null` |
| **Owner** | `email: 'admin@fisia.com'` | Exactly 1 record | `role === 'OWNER'`, `clinicId === clinic.id`, bcrypt password matches `'Fisia2026!'` |
| **Specialists** | `clinicId: clinic.id, role: 'SPECIALIST'` | Exactly 5 records | All 5 emails match canonical list; all have `isActive: true`; valid bcrypt passwords |
| **Total Users** | `clinicId: clinic.id` | Exactly 6 records | 1 Owner + 5 Specialists |
| **Services** | `clinicId: clinic.id` | Exactly 6 records | Every service has `duration > 0`, `price > 0`, `isActive: true` |
| **SpecialistService** | `clinicId: clinic.id` | ≥ 16 assignments | All `specialistId` belong to clinic users; all `serviceId` belong to clinic services |
| **Products** | `clinicId: clinic.id` | Exactly 4 records | All SKUs match canonical list; stock ≥ 10 for each; price > 0 |
| **Customer** | `clinicId: clinic.id` | ≥ 1 record | Phone format starts with `+52` (E.164 compliant) |
| **Appointment** | `clinicId: clinic.id` | ≥ 1 record | Status `CONFIRMED`; links customer, specialist, service |
| **Order** | `clinicId: clinic.id` | ≥ 1 record | Status `PAID`, `paymentMethod: 'CASH'`, `total == 1030.00` |
| **OrderItems** | `orderId: order.id` | Exactly 2 items | 1 Service item ($750) + 1 Product item ($280); sum matches order total |
| **Waitlist** | `clinicId: clinic.id` | ≥ 1 record | Status `WAITING`, `priority >= 1` |

### 3.4 Standalone Seed Verification Script (`tests/verify-seed.ts`)

This standalone executable script verifies the seeded database programmatically:

```typescript
// tests/verify-seed.ts
import { PrismaClient, Role, PaymentMethod, OrderStatus, AppointmentStatus } from '@prisma/client';
import bcrypt from 'bcryptjs';

const prisma = new PrismaClient();

async function verifySeed() {
  console.log('🔎 Executing APPCITAS Milestone 1 Database Verification...');
  let errors: string[] = [];

  try {
    // 1. Verify Clinic
    const clinic = await prisma.clinic.findUnique({ where: { slug: 'clinica-fisia' } });
    if (!clinic) {
      errors.push('Pilot clinic "clinica-fisia" not found.');
      throw new Error(errors.join('\n'));
    }
    console.log(`✅ Clinic verified: ${clinic.name} (${clinic.id})`);

    // 2. Verify Owner
    const owner = await prisma.user.findUnique({ where: { email: 'admin@fisia.com' } });
    if (!owner || owner.role !== Role.OWNER || owner.clinicId !== clinic.id) {
      errors.push('Owner "admin@fisia.com" invalid or not assigned to pilot clinic.');
    } else {
      const passwordOk = await bcrypt.compare('Fisia2026!', owner.password);
      if (!passwordOk) errors.push('Owner password does not match default "Fisia2026!".');
      else console.log(`✅ Owner verified: ${owner.name} (Password hash verified)`);
    }

    // 3. Verify Exactly 5 Specialists
    const specialists = await prisma.user.findMany({
      where: { clinicId: clinic.id, role: Role.SPECIALIST },
    });
    if (specialists.length !== 5) {
      errors.push(`Expected exactly 5 specialists, found ${specialists.length}.`);
    } else {
      console.log(`✅ Exactly 5 specialists verified: ${specialists.map(s => s.name).join(', ')}`);
    }

    // 4. Verify 6 Services
    const services = await prisma.service.findMany({ where: { clinicId: clinic.id } });
    if (services.length !== 6) {
      errors.push(`Expected 6 services, found ${services.length}.`);
    } else {
      console.log(`✅ Exactly 6 services verified.`);
    }

    // 5. Verify SpecialistService Join Table
    const assignments = await prisma.specialistService.findMany({ where: { clinicId: clinic.id } });
    if (assignments.length < 15) {
      errors.push(`Expected at least 15 service assignments, found ${assignments.length}.`);
    } else {
      console.log(`✅ Specialist assignments verified: ${assignments.length} assignments active.`);
    }

    // 6. Verify 4 Products
    const products = await prisma.product.findMany({ where: { clinicId: clinic.id } });
    if (products.length !== 4) {
      errors.push(`Expected 4 products, found ${products.length}.`);
    } else {
      console.log(`✅ Exactly 4 products verified.`);
    }

    // 7. Verify Cash Closing & Order Pipeline
    const orders = await prisma.order.findMany({
      where: { clinicId: clinic.id },
      include: { items: true },
    });
    if (orders.length === 0) {
      errors.push('No orders found for cash closing verification.');
    } else {
      const paidCashOrder = orders.find(o => o.paymentMethod === PaymentMethod.CASH && o.status === OrderStatus.PAID);
      if (!paidCashOrder) {
        errors.push('No paid cash order found for daily closing foundation.');
      } else {
        const itemSum = paidCashOrder.items.reduce((sum, it) => sum + Number(it.subtotal), 0);
        if (Math.abs(Number(paidCashOrder.total) - itemSum) > 0.01) {
          errors.push(`Order total (${paidCashOrder.total}) does not equal sum of items (${itemSum}).`);
        } else {
          console.log(`✅ Cash closing order verified: Total $${paidCashOrder.total} MXN matches items.`);
        }
      }
    }

    // Report Summary
    if (errors.length > 0) {
      console.error('\n❌ SEED VERIFICATION FAILED:');
      errors.forEach(err => console.error(`  - ${err}`));
      process.exit(1);
    }

    console.log('\n🎉 ALL MILESTONE 1 SEED VERIFICATIONS PASSED (100% SUCCESS)');
  } finally {
    await prisma.$disconnect();
  }
}

verifySeed().catch((err) => {
  console.error('Fatal error during seed verification:', err);
  process.exit(1);
});
```

---

## 4. Multitenancy Enforcement & Isolation Verification Strategy

### 4.1 Strict Database-Level Isolation via `clinicId`

Every tenant-bound entity in APPCITAS MUST reference `Clinic(id)` via a foreign key `clinicId`.

```
                    ┌─────────────────────────┐
                    │      Clinic (Root)      │
                    │   id: String @id CUID   │
                    └────────────┬────────────┘
         ┌───────────────────────┼───────────────────────┐
         │ (onDelete: Cascade)   │ (onDelete: Cascade)   │ (onDelete: Cascade)
         ▼                       ▼                       ▼
    ┌─────────┐             ┌─────────┐             ┌──────────┐
    │  User   │             │ Service │             │ Customer │
    └────┬────┘             └────┬────┘             └────┬─────┘
         │                       │                       │
         └───────────┬───────────┘                       │
                     ▼                                   │
          ┌─────────────────────┐                        │
          │  SpecialistService  │                        │
          └─────────────────────┘                        │
                     ▲                                   │
                     └─────────────────┬─────────────────┘
                                       ▼
                              ┌─────────────────┐
                              │   Appointment   │
                              └────────┬────────┘
                                       ▼
                              ┌─────────────────┐
                              │      Order      │
                              └────────┬────────┘
                                       ▼
                              ┌─────────────────┐
                              │    OrderItem    │
                              └─────────────────┘
```

### 4.2 Indexing Coverage for Multi-Tenant Query Performance

Multi-tenant queries always filter by `WHERE clinicId = ?`. Without leading indexes on `clinicId`, PostgreSQL will execute sequential table scans on every tenant query, causing massive performance degradation as clinic data scales.

#### Mandatory Index Matrix
| Model | Index Definition | Purpose |
| :--- | :--- | :--- |
| `User` | `@@index([clinicId])`, `@@index([email])` | Fast specialist listing per clinic; fast auth lookup |
| `Service` | `@@index([clinicId])`, `@@unique([clinicId, name])` | Fast catalog listing; prevents duplicate service names per clinic |
| `SpecialistService` | `@@index([clinicId])`, `@@unique([specialistId, serviceId])` | Fast assignment query; guarantees uniqueness |
| `Product` | `@@index([clinicId])`, `@@unique([clinicId, sku])` | Fast inventory fetch; prevents SKU collisions per clinic |
| `Customer` | `@@index([clinicId])`, `@@index([clinicId, phone])` | Fast patient search and WhatsApp phone lookup |
| `Appointment` | `@@index([clinicId])`, `@@index([clinicId, specialistId, startTime])` | Instant calendar agenda retrieval for specialist |
| `Order` | `@@index([clinicId])`, `@@index([clinicId, paidAt, status])` | High-speed daily cash closing aggregation |
| `Waitlist` | `@@index([clinicId, status])`, `@@index([clinicId, preferredDate])` | Fast queue processing on cancellation |

### 4.3 The "Cross-Tenant Foreign Key Trap" & Architectural Defense

#### The Vulnerability
In a standard Prisma schema:
```prisma
model Appointment {
  id           String   @id @default(cuid())
  clinicId     String
  specialistId String
  clinic       Clinic   @relation(fields: [clinicId], references: [id])
  specialist   User     @relation(fields: [specialistId], references: [id])
}
```
At the PostgreSQL foreign key level:
1. `clinicId` is validated to exist in `Clinic`.
2. `specialistId` is validated to exist in `User`.
**PostgreSQL DOES NOT automatically check that `specialist.clinicId === appointment.clinicId`!**
If an application bug or malicious request passes `specialistId` belonging to Clinic B into an appointment creation payload for Clinic A, standard single-column foreign keys WILL NOT prevent the record from being inserted!

#### Two-Layer Defense Strategy
1. **Layer 1: Application-Layer Tenant Guard (Mandatory in all API Route Handlers)**:
   Whenever an entity references another entity within a tenant context, the API route must verify that both share the same `clinicId`:
   ```typescript
   // Example: POST /api/appointments
   const specialist = await prisma.user.findFirst({
     where: { id: specialistId, clinicId: session.user.clinicId, isActive: true },
   });
   if (!specialist) {
     return NextResponse.json({ error: 'Especialista no encontrado en su clínica' }, { status: 404 });
   }
   ```
2. **Layer 2: Multi-Tenant Query Wrapper Rule**:
   Every database query outside the public login/registration endpoint MUST explicitly specify `clinicId: session.user.clinicId`:
   ```typescript
   // Anti-pattern (SECURITY VULNERABILITY):
   const services = await prisma.service.findMany(); // LEAKS ALL TENANTS!

   // Compliant Pattern:
   const services = await prisma.service.findMany({
     where: { clinicId: session.user.clinicId, isActive: true },
   });
   ```

### 4.4 Automated Cross-Tenant Isolation Test Scenario

To verify multi-tenancy in automated test suites:
1. Provision Clinic A (`clinica-fisia`) and Clinic B (`clinica-norte`).
2. Create 3 services in Clinic A and 2 services in Clinic B.
3. Authenticate as Clinic A owner and request `GET /api/services`.
4. **Assert**: Response array length is exactly 3; none of the returned IDs or names belong to Clinic B.
5. Attempt `PUT /api/services/[clinicB_serviceId]/assign` using Clinic A owner session.
6. **Assert**: API returns `404 Not Found` or `403 Forbidden`; Clinic B service remains unchanged.

---

## 5. Edge Cases & Common Traps in Prisma with PostgreSQL

### 5.1 Trap 1: Foreign Key Constraint Ordering & Safe Resets

#### Root Cause
PostgreSQL strictly validates foreign key constraints on every `DELETE` statement. If a script attempts to delete a parent table before child tables, PostgreSQL aborts the entire transaction with error:
`ERROR: update or delete on table "X" violates foreign key constraint "fk_Y_X" on table "Y"` (Prisma error `P2003`).

Furthermore, `onDelete: Restrict` is placed on `Appointment` relations to `User`, `Service`, and `Customer` to protect clinical records. Therefore, deleting a `User` or `Service` while an `Appointment` references them will fail.

#### Model Dependency Hierarchy
```
Level 0 (Leaf nodes):
  - OrderItem (references Order, Service, Product)
  - Waitlist (references Customer, Service, User)
Level 1:
  - Order (references Customer, Appointment)
Level 2:
  - Appointment (references Customer, User, Service)
Level 3:
  - SpecialistService (references User, Service)
Level 4:
  - Service (references Clinic)
  - Product (references Clinic)
  - Customer (references Clinic)
  - User (references Clinic)
Level 5 (Root):
  - Clinic
```

#### Safe Tenant Reset Sequence (Prisma Transaction)
When resetting a tenant or resetting the database prior to re-seeding, execute deletions in this exact inverse-topological order:

```typescript
// Safe reset transaction
await prisma.$transaction(async (tx) => {
  // 1. Delete OrderItems
  await tx.orderItem.deleteMany({ where: { order: { clinicId } } });
  // 2. Delete Orders
  await tx.order.deleteMany({ where: { clinicId } });
  // 3. Delete Waitlist entries
  await tx.waitlist.deleteMany({ where: { clinicId } });
  // 4. Delete Appointments (must precede User, Service, Customer)
  await tx.appointment.deleteMany({ where: { clinicId } });
  // 5. Delete SpecialistService joins
  await tx.specialistService.deleteMany({ where: { clinicId } });
  // 6. Delete Services
  await tx.service.deleteMany({ where: { clinicId } });
  // 7. Delete Products
  await tx.product.deleteMany({ where: { clinicId } });
  // 8. Delete Users (Owner & Specialists)
  await tx.user.deleteMany({ where: { clinicId } });
  // 9. Delete Customers
  await tx.customer.deleteMany({ where: { clinicId } });
  // 10. Delete Clinic (optional if re-creating)
  await tx.clinic.deleteMany({ where: { id: clinicId } });
});
```

*Note on TRUNCATE*: In local integration tests where the entire database can be wiped, `TRUNCATE TABLE "Clinic" CASCADE;` is valid because PostgreSQL `CASCADE` automatically truncates all dependent tables. However, in shared databases or multi-tenant test environments, **never** use global TRUNCATE; use the tenant-scoped transaction above.

---

### 5.2 Trap 2: Decimal & BigInt Serialization in Next.js App Router

#### The Core Problem
1. **Prisma `Decimal` Class**:
   Prisma returns `@db.Decimal(10, 2)` fields as instances of `Prisma.Decimal` (from `decimal.js`).
   When serialized to JSON across the Next.js Server-to-Client Component boundary (RSC payload) or in Route Handlers (`NextResponse.json(data)`):
   - `Decimal` serializes as a **string** (e.g. `"850.00"`).
   - If client components perform math operations (e.g. `price + 50`), JavaScript will perform string concatenation: `"850.0050"`!
2. **`BigInt` Fatal TypeError**:
   If any model includes a `BigInt` field, standard `JSON.stringify` throws an unhandled exception:
   `TypeError: Do not know how to serialize a BigInt`
   This crashes the Next.js App Router rendering engine instantly.

#### Architectural Prevention Strategies

##### Strategy A: Serialization Utility Helper (`src/lib/serialize.ts`)
Provide a lightweight utility that recursively converts `Prisma.Decimal` to `number` and `Date` to ISO strings:

```typescript
// src/lib/serialize.ts
import { Prisma } from '@prisma/client';

export function serializeDecimal<T>(data: T): T {
  if (data === null || data === undefined) return data;

  if (data instanceof Prisma.Decimal) {
    return data.toNumber() as unknown as T;
  }

  if (Array.isArray(data)) {
    return data.map(serializeDecimal) as unknown as T;
  }

  if (typeof data === 'object') {
    const result: any = {};
    for (const [key, value] of Object.entries(data)) {
      result[key] = serializeDecimal(value);
    }
    return result as T;
  }

  return data;
}

export function formatCurrency(amount: number | Prisma.Decimal | string): string {
  const num = typeof amount === 'number' ? amount : Number(amount);
  return new Intl.NumberFormat('es-MX', {
    style: 'currency',
    currency: 'MXN',
    minimumFractionDigits: 2,
  }).format(isNaN(num) ? 0 : num);
}
```

##### Strategy B: Global BigInt JSON Polyfill (`src/lib/prisma.ts`)
Add a safety polyfill in the Prisma Client singleton to prevent BigInt serialization panics across the entire app:

```typescript
// Prevent BigInt serialization crashes in Next.js
if (typeof BigInt !== 'undefined') {
  (BigInt.prototype as any).toJSON = function () {
    return this.toString();
  };
}
```

---

### 5.3 Trap 3: Idempotency in `prisma/seed.ts`

#### The Problem
If `prisma/seed.ts` uses plain `prisma.<model>.create()`, running `npx prisma db seed` a second time triggers unique constraint violations:
- `P2002: Unique constraint failed on the fields: (slug)` for `Clinic`
- `P2002: Unique constraint failed on the fields: (email)` for `User`
- `P2002: Unique constraint failed on the fields: (clinicId, name)` for `Service`
- `P2002: Unique constraint failed on the fields: (specialistId, serviceId)` for `SpecialistService`

#### Solution: 100% Upsert Pattern with Deterministic Unique Keys
Every entity created in `prisma/seed.ts` must use `prisma.<model>.upsert()` referencing a guaranteed unique identifier:

| Model | Upsert `where` Criteria | Schema Requirement |
| :--- | :--- | :--- |
| `Clinic` | `where: { slug: 'clinica-fisia' }` | `slug String @unique` |
| `User` | `where: { email: user.email }` | `email String @unique` |
| `Service` | `where: { clinicId_name: { clinicId, name } }` | `@@unique([clinicId, name])` |
| `SpecialistService` | `where: { specialistId_serviceId: { specialistId, serviceId } }` | `@@unique([specialistId, serviceId])` |
| `Product` | `where: { clinicId_sku: { clinicId, sku } }` | `@@unique([clinicId, sku])` |
| `Customer` | `where: { id: 'sample-customer-fisia' }` | Fixed deterministic ID |
| `Appointment` | `where: { id: 'sample-appointment-fisia' }` | Fixed deterministic ID |
| `Order` | `where: { id: 'sample-order-fisia' }` | Fixed deterministic ID |
| `Waitlist` | `where: { id: 'sample-waitlist-fisia' }` | Fixed deterministic ID |

#### Complete Idempotent Seed Implementation Blueprint
```typescript
// prisma/seed.ts blueprint
import { PrismaClient, Role, PaymentMethod, AppointmentStatus, OrderStatus, WaitlistStatus } from '@prisma/client';
import bcrypt from 'bcryptjs';

const prisma = new PrismaClient();

async function main() {
  console.log('🌱 Starting idempotent seed...');
  const passwordHash = await bcrypt.hash('Fisia2026!', 10);

  // 1. Upsert Clinic
  const clinic = await prisma.clinic.upsert({
    where: { slug: 'clinica-fisia' },
    update: {
      name: 'Clínica Fisia',
      phone: '+52 55 8432 1098',
      address: 'Av. Insurgentes Sur 1234, Col. Del Valle, Benito Juárez, CDMX',
    },
    create: {
      slug: 'clinica-fisia',
      name: 'Clínica Fisia',
      phone: '+52 55 8432 1098',
      address: 'Av. Insurgentes Sur 1234, Col. Del Valle, Benito Juárez, CDMX',
    },
  });

  // 2. Upsert Owner
  const owner = await prisma.user.upsert({
    where: { email: 'admin@fisia.com' },
    update: {
      name: 'Dr. Alejandro Morales',
      clinicId: clinic.id,
      role: Role.OWNER,
      specialty: 'Director Clínico & Fisioterapia Integral',
      phone: '+52 55 8432 1000',
      password: passwordHash,
      isActive: true,
    },
    create: {
      email: 'admin@fisia.com',
      name: 'Dr. Alejandro Morales',
      clinicId: clinic.id,
      role: Role.OWNER,
      specialty: 'Director Clínico & Fisioterapia Integral',
      phone: '+52 55 8432 1000',
      password: passwordHash,
      isActive: true,
    },
  });

  // 3. Upsert 5 Specialists
  const specialistsData = [
    { name: 'Lic. Sofia Mendoza', email: 'sofia.mendoza@fisia.com', specialty: 'Fisioterapia Deportiva y Readaptación', phone: '+52 55 8432 1001' },
    { name: 'Lic. Mateo Valencia', email: 'mateo.valencia@fisia.com', specialty: 'Terapia Manual Ortopédica y Columna', phone: '+52 55 8432 1002' },
    { name: 'Lic. Camila Herrera', email: 'camila.herrera@fisia.com', specialty: 'Rehabilitación Neurológica', phone: '+52 55 8432 1003' },
    { name: 'Lic. Daniel Rios', email: 'daniel.rios@fisia.com', specialty: 'Punción Seca y Dolor Miofascial', phone: '+52 55 8432 1004' },
    { name: 'Lic. Valentina Gómez', email: 'valentina.gomez@fisia.com', specialty: 'Fisioterapia Pediátrica y Postural', phone: '+52 55 8432 1005' },
  ];

  const specialists = [];
  for (const s of specialistsData) {
    const sp = await prisma.user.upsert({
      where: { email: s.email },
      update: {
        name: s.name,
        clinicId: clinic.id,
        role: Role.SPECIALIST,
        specialty: s.specialty,
        phone: s.phone,
        password: passwordHash,
        isActive: true,
      },
      create: {
        email: s.email,
        name: s.name,
        clinicId: clinic.id,
        role: Role.SPECIALIST,
        specialty: s.specialty,
        phone: s.phone,
        password: passwordHash,
        isActive: true,
      },
    });
    specialists.push(sp);
  }

  // 4. Upsert Services & Specialist Assignments (using @@unique([clinicId, name]))
  const servicesData = [
    { name: 'Valoración Inicial de Fisioterapia', duration: 60, price: 850.0, description: 'Evaluación funcional y diseño de plan terapéutico.', emails: ['admin@fisia.com', ...specialistsData.map(s => s.email)] },
    { name: 'Sesión de Fisioterapia Deportiva', duration: 50, price: 750.0, description: 'Tratamiento de lesiones y reinserción al ejercicio.', emails: ['sofia.mendoza@fisia.com', 'mateo.valencia@fisia.com', 'daniel.rios@fisia.com'] },
    { name: 'Terapia Manual Ortopédica', duration: 45, price: 700.0, description: 'Movilizaciones y manipulaciones articulares.', emails: ['mateo.valencia@fisia.com', 'daniel.rios@fisia.com', 'sofia.mendoza@fisia.com'] },
    { name: 'Punción Seca y Electroterapia', duration: 45, price: 800.0, description: 'Desactivación de puntos gatillo y neuromodulación.', emails: ['daniel.rios@fisia.com', 'sofia.mendoza@fisia.com'] },
    { name: 'Rehabilitación Neurológica', duration: 60, price: 900.0, description: 'Terapia neuromotora en secuelas neurológicas.', emails: ['camila.herrera@fisia.com'] },
    { name: 'Fisioterapia Pediátrica', duration: 45, price: 750.0, description: 'Atención especializada en neurodesarrollo infantil.', emails: ['valentina.gomez@fisia.com'] },
  ];

  for (const s of servicesData) {
    const service = await prisma.service.upsert({
      where: { clinicId_name: { clinicId: clinic.id, name: s.name } },
      update: { description: s.description, duration: s.duration, price: s.price, isActive: true },
      create: { clinicId: clinic.id, name: s.name, description: s.description, duration: s.duration, price: s.price, isActive: true },
    });

    for (const email of s.emails) {
      const u = email === owner.email ? owner : specialists.find(spec => spec.email === email);
      if (u) {
        await prisma.specialistService.upsert({
          where: { specialistId_serviceId: { specialistId: u.id, serviceId: service.id } },
          update: {},
          create: { clinicId: clinic.id, specialistId: u.id, serviceId: service.id },
        });
      }
    }
  }

  // 5. Upsert Products
  const productsData = [
    { name: 'Kinesiotape CureTape Pro (5cm x 5m)', sku: 'KT-PRO-01', price: 280.0, stock: 25 },
    { name: 'Banda Elástica Theraband Nivel Medio', sku: 'TB-MED-BL', price: 220.0, stock: 15 },
    { name: 'Gel Criogénico Desinflamatorio 250ml', sku: 'GEL-CRIO-250', price: 190.0, stock: 30 },
    { name: 'Foam Roller de Alta Densidad 45cm', sku: 'FR-HD-45', price: 450.0, stock: 10 },
  ];

  for (const p of productsData) {
    await prisma.product.upsert({
      where: { clinicId_sku: { clinicId: clinic.id, sku: p.sku } },
      update: { name: p.name, price: p.price, stock: p.stock },
      create: { clinicId: clinic.id, name: p.name, sku: p.sku, price: p.price, stock: p.stock },
    });
  }

  // 6. Upsert Sample Patient, Appointment, Order & Waitlist (Deterministic IDs)
  const customer = await prisma.customer.upsert({
    where: { id: 'sample-customer-fisia' },
    update: {},
    create: {
      id: 'sample-customer-fisia',
      clinicId: clinic.id,
      name: 'Carlos Fernández',
      email: 'carlos.fernandez@gmail.com',
      phone: '+525598765432',
      notes: 'Paciente referido por esguince de tobillo.',
    },
  });

  const firstService = await prisma.service.findFirst({ where: { clinicId: clinic.id, name: 'Valoración Inicial de Fisioterapia' } });
  if (firstService && specialists[0]) {
    const startTime = new Date();
    startTime.setDate(startTime.getDate() + 1);
    startTime.setHours(10, 0, 0, 0);

    const endTime = new Date(startTime);
    endTime.setMinutes(startTime.getMinutes() + firstService.duration);

    const appointment = await prisma.appointment.upsert({
      where: { id: 'sample-appointment-fisia' },
      update: {},
      create: {
        id: 'sample-appointment-fisia',
        clinicId: clinic.id,
        specialistId: specialists[0].id,
        serviceId: firstService.id,
        customerId: customer.id,
        startTime,
        endTime,
        status: AppointmentStatus.CONFIRMED,
      },
    });

    // Check if order already exists before creating items
    const existingOrder = await prisma.order.findUnique({ where: { id: 'sample-order-fisia' } });
    if (!existingOrder) {
      const productSample = await prisma.product.findFirst({ where: { clinicId: clinic.id, sku: 'KT-PRO-01' } });
      await prisma.order.create({
        data: {
          id: 'sample-order-fisia',
          clinicId: clinic.id,
          customerId: customer.id,
          appointmentId: appointment.id,
          paymentMethod: PaymentMethod.CASH,
          status: OrderStatus.PAID,
          total: 1030.0,
          paidAt: new Date(),
          notes: 'Pago en efectivo para cierre diario.',
          items: {
            create: [
              { serviceId: firstService.id, quantity: 1, unitPrice: 750.0, subtotal: 750.0 },
              ...(productSample ? [{ productId: productSample.id, quantity: 1, unitPrice: 280.0, subtotal: 280.0 }] : []),
            ],
          },
        },
      });
    }
  }

  await prisma.waitlist.upsert({
    where: { id: 'sample-waitlist-fisia' },
    update: {},
    create: {
      id: 'sample-waitlist-fisia',
      clinicId: clinic.id,
      customerId: customer.id,
      status: WaitlistStatus.WAITING,
      priority: 1,
      preferredTimeRange: '16:00 - 19:00',
    },
  });

  console.log('🎉 Idempotent seed completed successfully.');
}

main()
  .catch((e) => {
    console.error('Error during seed:', e);
    process.exit(1);
  })
  .finally(async () => {
    await prisma.$disconnect();
  });
```

---

### 5.4 Trap 4: Environment Variable Handling & Connection Pooling

#### Safe Environment Validation (`src/lib/env.ts`)
Missing or improperly formatted environment variables during deployment or test runs cause cryptic failures. We mandate a central runtime validator:

```typescript
// src/lib/env.ts
export function validateEnv() {
  const requiredEnvVars = ['DATABASE_URL', 'NEXTAUTH_SECRET'];
  const missing: string[] = [];

  for (const envVar of requiredEnvVars) {
    if (!process.env[envVar] || process.env[envVar]!.trim() === '') {
      missing.push(envVar);
    }
  }

  if (missing.length > 0) {
    throw new Error(
      `❌ Critical configuration error: Missing required environment variables:\n` +
      missing.map(m => `   - ${m}`).join('\n') +
      `\nPlease check your .env file or deployment environment.`
    );
  }
}
```

#### PostgreSQL Connection String & Special Characters Trap
If database passwords contain special characters (e.g. `@`, `:`, `/`, `#`, `%`, `&`), the URL parser in Prisma and PostgreSQL client fails with `P1000: Authentication failed against database server` or `Invalid connection string`.
- **Mitigation**: Database passwords must be URL-encoded (e.g., `#` becomes `%23`, `@` becomes `%40`).

#### Connection Pooling on DigitalOcean / Managed PostgreSQL
Managed PostgreSQL instances (like DigitalOcean Managed Databases) use PgBouncer for connection pooling:
1. **Pooled Port (6543)**: Requires adding `?pgbouncer=true&connection_limit=1` to `DATABASE_URL`.
2. **Direct Port (5432)**: Prisma migrations and `prisma db push` require direct connections because PgBouncer does not support prepared statements across migration transactions.
3. In development / Phase 1: Standard direct PostgreSQL URL on port 5432 is recommended:
   `DATABASE_URL="postgresql://postgres:postgres@localhost:5432/appcitas?schema=public"`

#### Global Prisma Client Singleton (`src/lib/prisma.ts`)
In Next.js development mode, Hot Module Replacement (HMR) re-evaluates server files on every code save. Without a global singleton, a new `PrismaClient` instance is instantiated on every save, quickly exceeding PostgreSQL's `max_connections` (Prisma error `P2024: Timed out fetching a new connection from the connection pool`).

```typescript
// src/lib/prisma.ts
import { PrismaClient } from '@prisma/client';

const globalForPrisma = globalThis as unknown as {
  prisma: PrismaClient | undefined;
};

export const prisma =
  globalForPrisma.prisma ??
  new PrismaClient({
    log: process.env.NODE_ENV === 'development' ? ['query', 'error', 'warn'] : ['error'],
  });

if (process.env.NODE_ENV !== 'production') globalForPrisma.prisma = prisma;
```

---

## 6. Actionable Implementation Checklist for Milestone 1 Implementers

To guarantee 100% first-time pass on all acceptance gates, implementers must execute against this verification checklist:

- [ ] **1. Schema Definition (`prisma/schema.prisma`)**:
  - [ ] All 10 models defined (`Clinic`, `User`, `Service`, `SpecialistService`, `Product`, `Customer`, `Appointment`, `Order`, `OrderItem`, `Waitlist`).
  - [ ] Every tenant model has `clinicId String` and `@relation(fields: [clinicId], references: [id], onDelete: Cascade)`.
  - [ ] Unique constraints declared: `Clinic(slug)`, `User(email)`, `Service(clinicId, name)`, `SpecialistService(specialistId, serviceId)`, `Product(clinicId, sku)`.
  - [ ] Execute `npx prisma validate` -> Exit Code `0`.
- [ ] **2. Prisma Client Generation**:
  - [ ] Execute `npx prisma generate` -> Output generates `@prisma/client` cleanly.
- [ ] **3. Singleton & Serialization**:
  - [ ] Create `src/lib/prisma.ts` using the global singleton pattern.
  - [ ] Add `serializeDecimal` helper to avoid stringified decimals on UI.
- [ ] **4. Seed Script (`prisma/seed.ts`)**:
  - [ ] Use `bcryptjs` with salt round 10 for password hashing.
  - [ ] Use `upsert` for all models to guarantee 100% idempotency.
  - [ ] Seed "Clínica Fisia", owner, 5 specialists, 6 services with assignments, 4 products, and operational records.
  - [ ] Execute `npx prisma db seed` -> Exit Code `0`.
  - [ ] Execute `npx prisma db seed` a second time -> Exit Code `0` (Zero constraint collision errors).
- [ ] **5. Verification Script Execution**:
  - [ ] Run `npx tsx tests/verify-seed.ts` -> All 7 entity checks pass with green assertions.

---

*Report formulated and verified by `teamwork_preview_explorer_m1_3`.*
