# Specification Mining Report: Milestone 1 — Multitenant Schema & Data Layer

**Milestone**: M1 (Multitenant Schema & Data Layer)  
**Author**: `teamwork_preview_spec_miner_m1_1` (Specification Miner)  
**Target System**: APPCITAS — Multitenant Physiotherapy Clinic Management Platform  
**Date**: 2026-09-14  
**Working Directory**: `/Users/jasbotello/devel/APPCITAS/.agents/teamwork_preview_spec_miner_m1_1`  

---

## 1. Executive Overview

This specification establishes the authoritative database architecture and pilot data provisioning for Milestone 1 of APPCITAS. It defines:
1. **The Production Prisma Schema (`prisma/schema.prisma`)**: A 10-model PostgreSQL database schema enforcing strict multi-tenant isolation via `clinicId`, referential integrity, cascading rules, monetary decimals, and optimized indexing.
2. **The Idempotent Database Seed Script (`prisma/seed.ts`)**: An automated, repeatable seeding blueprint provisioning "Clínica Fisia", its clinic owner (Dr. Alejandro Morales), 5 certified specialists with clinical specialties, 6 core physiotherapy services, 3 retail products, many-to-many specialist-service assignment matrix, patient records, and operational transaction data for daily cash closing (Corte de Caja).
3. **Database Client Singleton (`src/lib/prisma.ts`)**: Global singleton contract preventing connection pool exhaustion in Next.js App Router serverless and hot-reloading development environments.

---

## 2. Complete Production Prisma Schema (`prisma/schema.prisma`)

```prisma
// =============================================================================
// APPCITAS: Phase 1 (El Cimiento) Database Schema
// Multitenant PostgreSQL Schema for Physiotherapy Clinic Operations
// =============================================================================

generator client {
  provider = "prisma-client-js"
}

datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}

// -----------------------------------------------------------------------------
// Enums
// -----------------------------------------------------------------------------

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
  OFFERED
  ACCEPTED
  EXPIRED
  CANCELLED
}

// -----------------------------------------------------------------------------
// Model 1: Clinic (Root Tenant Entity)
// -----------------------------------------------------------------------------
model Clinic {
  id                 String              @id @default(cuid())
  slug               String              @unique
  name               String
  phone              String?
  address            String?
  createdAt          DateTime            @default(now())
  updatedAt          DateTime            @updatedAt

  // Tenant Child Relationships (Cascading on Tenant Teardown)
  users              User[]
  services           Service[]
  specialistServices SpecialistService[]
  products           Product[]
  customers          Customer[]
  appointments       Appointment[]
  orders             Order[]
  waitlists          Waitlist[]

  @@map("clinics")
}

// -----------------------------------------------------------------------------
// Model 2: User (Clinic Owner & Specialist Physiotherapists)
// -----------------------------------------------------------------------------
model User {
  id                 String              @id @default(cuid())
  clinicId           String
  name               String
  email              String              @unique
  passwordHash       String
  specialty          String?
  phone              String?
  role               Role                @default(SPECIALIST)
  isActive           Boolean             @default(true)
  googleAccessToken  String?
  googleRefreshToken String?
  createdAt          DateTime            @default(now())
  updatedAt          DateTime            @updatedAt

  // Relations
  clinic             Clinic              @relation(fields: [clinicId], references: [id], onDelete: Cascade)
  specialistServices SpecialistService[]
  appointments       Appointment[]
  waitlists          Waitlist[]

  @@index([clinicId])
  @@index([email])
  @@map("users")
}

// -----------------------------------------------------------------------------
// Model 3: Service (Clinic-wide Physiotherapy Services Catalog)
// -----------------------------------------------------------------------------
model Service {
  id                 String              @id @default(cuid())
  clinicId           String
  name               String
  description        String?
  price              Decimal             @db.Decimal(10, 2)
  durationMin        Int
  isActive           Boolean             @default(true)
  createdAt          DateTime            @default(now())
  updatedAt          DateTime            @updatedAt

  // Relations
  clinic             Clinic              @relation(fields: [clinicId], references: [id], onDelete: Cascade)
  specialistServices SpecialistService[]
  appointments       Appointment[]
  orderItems         OrderItem[]
  waitlists          Waitlist[]

  @@unique([clinicId, name])
  @@index([clinicId])
  @@map("services")
}

// -----------------------------------------------------------------------------
// Model 4: SpecialistService (Explicit M:N Join for Specialist Assignments)
// -----------------------------------------------------------------------------
model SpecialistService {
  id           String   @id @default(cuid())
  clinicId     String
  specialistId String
  serviceId    String
  createdAt    DateTime @default(now())

  // Relations
  clinic       Clinic   @relation(fields: [clinicId], references: [id], onDelete: Cascade)
  specialist   User     @relation(fields: [specialistId], references: [id], onDelete: Cascade)
  service      Service  @relation(fields: [serviceId], references: [id], onDelete: Cascade)

  @@unique([specialistId, serviceId])
  @@index([clinicId])
  @@index([specialistId])
  @@index([serviceId])
  @@map("specialist_services")
}

// -----------------------------------------------------------------------------
// Model 5: Product (Physical Therapy Supplies & Retail Goods)
// -----------------------------------------------------------------------------
model Product {
  id          String      @id @default(cuid())
  clinicId    String
  name        String
  description String?
  price       Decimal     @db.Decimal(10, 2)
  stock       Int         @default(0)
  sku         String?
  isActive    Boolean     @default(true)
  createdAt   DateTime    @default(now())
  updatedAt   DateTime    @updatedAt

  // Relations
  clinic      Clinic      @relation(fields: [clinicId], references: [id], onDelete: Cascade)
  orderItems  OrderItem[]

  @@unique([clinicId, sku])
  @@index([clinicId])
  @@map("products")
}

// -----------------------------------------------------------------------------
// Model 6: Customer (Clinic Patients & Contact Directory)
// -----------------------------------------------------------------------------
model Customer {
  id           String        @id @default(cuid())
  clinicId     String
  name         String
  email        String?
  phone        String
  notes        String?
  createdAt    DateTime      @default(now())
  updatedAt    DateTime      @updatedAt

  // Relations
  clinic       Clinic        @relation(fields: [clinicId], references: [id], onDelete: Cascade)
  appointments Appointment[]
  orders       Order[]
  waitlists    Waitlist[]

  @@index([clinicId])
  @@index([clinicId, phone])
  @@map("customers")
}

// -----------------------------------------------------------------------------
// Model 7: Appointment (Clinical Session Booking)
// -----------------------------------------------------------------------------
model Appointment {
  id            String            @id @default(cuid())
  clinicId      String
  specialistId  String
  serviceId     String
  customerId    String
  dateTime      DateTime
  status        AppointmentStatus @default(CONFIRMED)
  notes         String?
  googleEventId String?
  createdAt     DateTime          @default(now())
  updatedAt     DateTime          @updatedAt

  // Relations
  clinic        Clinic            @relation(fields: [clinicId], references: [id], onDelete: Cascade)
  specialist    User              @relation(fields: [specialistId], references: [id], onDelete: Restrict)
  service       Service           @relation(fields: [serviceId], references: [id], onDelete: Restrict)
  customer      Customer          @relation(fields: [customerId], references: [id], onDelete: Restrict)
  order         Order?

  @@index([clinicId])
  @@index([clinicId, specialistId, dateTime])
  @@index([clinicId, status])
  @@map("appointments")
}

// -----------------------------------------------------------------------------
// Model 8: Order (Billing & Daily Cash Closing Foundation)
// -----------------------------------------------------------------------------
model Order {
  id            String        @id @default(cuid())
  clinicId      String
  customerId    String?
  appointmentId String?       @unique
  total         Decimal       @db.Decimal(10, 2)
  paymentMethod PaymentMethod
  status        OrderStatus   @default(PAID)
  notes         String?
  paidAt        DateTime      @default(now())
  createdAt     DateTime      @default(now())
  updatedAt     DateTime      @updatedAt

  // Relations
  clinic        Clinic        @relation(fields: [clinicId], references: [id], onDelete: Cascade)
  customer      Customer?     @relation(fields: [customerId], references: [id], onDelete: SetNull)
  appointment   Appointment?  @relation(fields: [appointmentId], references: [id], onDelete: SetNull)
  items         OrderItem[]

  @@index([clinicId])
  @@index([clinicId, paidAt, status])
  @@map("orders")
}

// -----------------------------------------------------------------------------
// Model 9: OrderItem (Itemized Order Line Items)
// -----------------------------------------------------------------------------
model OrderItem {
  id        String   @id @default(cuid())
  orderId   String
  serviceId String?
  productId String?
  quantity  Int      @default(1)
  unitPrice Decimal  @db.Decimal(10, 2)
  createdAt DateTime @default(now())

  // Relations
  order     Order    @relation(fields: [orderId], references: [id], onDelete: Cascade)
  service   Service? @relation(fields: [serviceId], references: [id], onDelete: SetNull)
  product   Product? @relation(fields: [productId], references: [id], onDelete: SetNull)

  @@index([orderId])
  @@map("order_items")
}

// -----------------------------------------------------------------------------
// Model 10: Waitlist (Cancellation Backfill & Queue Foundation)
// -----------------------------------------------------------------------------
model Waitlist {
  id            String         @id @default(cuid())
  clinicId      String
  customerId    String
  serviceId     String?
  specialistId  String?
  preferredDate DateTime?
  status        WaitlistStatus @default(WAITING)
  priority      Int            @default(0)
  notes         String?
  createdAt     DateTime       @default(now())
  updatedAt     DateTime       @updatedAt

  // Relations
  clinic        Clinic         @relation(fields: [clinicId], references: [id], onDelete: Cascade)
  customer      Customer       @relation(fields: [customerId], references: [id], onDelete: Cascade)
  service       Service?       @relation(fields: [serviceId], references: [id], onDelete: SetNull)
  specialist    User?          @relation(fields: [specialistId], references: [id], onDelete: SetNull)

  @@index([clinicId, status])
  @@index([clinicId, preferredDate])
  @@map("waitlists")
}
```

---

## 3. Multitenancy Architecture & Cascading Rules Analysis

### 3.1 Tenant Partitioning
Every entity in the system belongs directly or transitively to a `Clinic`. 
- **Direct Tenant Links**: `User`, `Service`, `SpecialistService`, `Product`, `Customer`, `Appointment`, `Order`, and `Waitlist` all maintain a non-nullable `clinicId String` referencing `Clinic.id`.
- **Transitive Tenant Links**: `OrderItem` links directly to `Order.id`. Because `Order` contains `clinicId`, child items inherit tenant isolation while reducing denormalization overhead.

### 3.2 Referential Integrity & Cascade Semantics
1. **Clinic Teardown (`Clinic` -> Children)**:
   - Rule: `onDelete: Cascade`.
   - Rationale: If a tenant clinic is offboarded or purged, all associated users, catalog items, patient records, appointments, transactions, and queues must be cleaned up atomically without leaving orphaned records.
2. **Clinical History Protection (`Appointment` -> `User`, `Service`, `Customer`)**:
   - Rule: `onDelete: Restrict`.
   - Rationale: A specialist, service, or customer cannot be permanently hard-deleted from the database if they have associated appointment records. Deactivation via `isActive = false` (soft deletion) preserves medical liability histories and historical auditability.
3. **Financial & Accounting Integrity (`Order` -> `Customer`, `Appointment`)**:
   - Rule: `onDelete: SetNull`.
   - Rationale: If a customer record is removed under GDPR/ARCO rights, past fiscal records and daily cash closing transactions must retain their monetary totals and audit trails, with customer reference safely set to `null`.
4. **Order Line Items (`Order` -> `OrderItem`)**:
   - Rule: `onDelete: Cascade`.
   - Rationale: An order line item has no semantic meaning outside its parent order. Deleting an order removes its line items.
5. **Specialist Assignments (`SpecialistService` -> `User`, `Service`)**:
   - Rule: `onDelete: Cascade`.
   - Rationale: If a specialist or service is deleted, the join row connecting them is automatically cleaned up.

### 3.3 Indexing Strategy
- **Tenant Scope Indexing**: Every single tenant table includes `@@index([clinicId])` to guarantee that multi-tenant filter clauses (`where: { clinicId }`) execute with index scans rather than full table sequential scans.
- **Compound Lookup Indexes**:
  - `Appointment`: `@@index([clinicId, specialistId, dateTime])` accelerates calendar and schedule collision queries.
  - `Order`: `@@index([clinicId, paidAt, status])` accelerates daily cash closing (Corte de Caja) aggregations.
  - `Customer`: `@@index([clinicId, phone])` accelerates phone lookups for appointments and WhatsApp messaging.
  - `Waitlist`: `@@index([clinicId, status])` and `@@index([clinicId, preferredDate])` optimize queue processing upon appointment cancellations.
- **Tenant Scoped Uniqueness**:
  - `Service`: `@@unique([clinicId, name])` prevents duplicate service titles within the same clinic while permitting distinct clinics to name services identically.
  - `Product`: `@@unique([clinicId, sku])` guarantees unique barcode/SKU within a clinic inventory.
  - `SpecialistService`: `@@unique([specialistId, serviceId])` prevents duplicate assignment records.

---

## 4. Database Client Singleton Specification (`src/lib/prisma.ts`)

To prevent PostgreSQL connection pool exhaustion during Next.js App Router hot module reloading, the client must be exported as a global singleton:

```typescript
// src/lib/prisma.ts
import { PrismaClient } from '@prisma/client';

const prismaClientSingleton = () => {
  return new PrismaClient({
    log: process.env.NODE_ENV === 'development' ? ['query', 'error', 'warn'] : ['error'],
  });
};

declare const globalThis: {
  prismaGlobal: ReturnType<typeof prismaClientSingleton>;
} & typeof global;

export const prisma = globalThis.prismaGlobal ?? prismaClientSingleton();

if (process.env.NODE_ENV !== 'production') {
  globalThis.prismaGlobal = prisma;
}

export default prisma;
```

---

## 5. Complete Idempotent Seed Script Specification (`prisma/seed.ts`)

The database seed script is fully idempotent using Prisma `upsert` queries. It populates "Clínica Fisia" in San Pedro Garza García, NL, with realistic Mexican healthcare credentials, pricing, and operational data.

```typescript
// prisma/seed.ts
import { PrismaClient, Role, AppointmentStatus, PaymentMethod, OrderStatus, WaitlistStatus } from '@prisma/client';
import * as bcrypt from 'bcryptjs';

const prisma = new PrismaClient();

async function main() {
  console.log('🌱 Starting APPCITAS Phase 1 database seed...');

  // ---------------------------------------------------------------------------
  // 1. Pilot Clinic: Clínica Fisia
  // ---------------------------------------------------------------------------
  const clinic = await prisma.clinic.upsert({
    where: { slug: 'clinica-fisia' },
    update: {
      name: 'Clínica Fisia',
      phone: '+52 81 8000 1122',
      address: 'Av. Humberto Lobo 555, San Pedro Garza García, NL',
    },
    create: {
      slug: 'clinica-fisia',
      name: 'Clínica Fisia',
      phone: '+52 81 8000 1122',
      address: 'Av. Humberto Lobo 555, San Pedro Garza García, NL',
    },
  });
  console.log(`✅ Clinic established: ${clinic.name} (${clinic.id})`);

  // Default initial password hash: "Fisia2026!"
  const defaultPasswordHash = await bcrypt.hash('Fisia2026!', 10);

  // ---------------------------------------------------------------------------
  // 2. Clinic Owner: Dr. Alejandro Morales
  // ---------------------------------------------------------------------------
  const owner = await prisma.user.upsert({
    where: { email: 'alejandro.morales@fisia.mx' },
    update: {
      name: 'Dr. Alejandro Morales',
      role: Role.OWNER,
      specialty: 'Director Clínico & Fisioterapia Avanzada',
      phone: '+52 81 8000 1100',
      isActive: true,
      clinicId: clinic.id,
      passwordHash: defaultPasswordHash,
    },
    create: {
      clinicId: clinic.id,
      name: 'Dr. Alejandro Morales',
      email: 'alejandro.morales@fisia.mx',
      passwordHash: defaultPasswordHash,
      role: Role.OWNER,
      specialty: 'Director Clínico & Fisioterapia Avanzada',
      phone: '+52 81 8000 1100',
      isActive: true,
    },
  });
  console.log(`✅ Owner provisioned: ${owner.name} (${owner.email})`);

  // ---------------------------------------------------------------------------
  // 3. The 5 Specialist Physiotherapists
  // ---------------------------------------------------------------------------
  const specialistsData = [
    {
      name: 'Lic. Sofia Garza',
      email: 'sofia.garza@fisia.mx',
      specialty: 'Fisioterapia Deportiva',
      phone: '+52 81 8000 1101',
    },
    {
      name: 'Lic. Mateo Rivas',
      email: 'mateo.rivas@fisia.mx',
      specialty: 'Terapia Manual y Ortopédica',
      phone: '+52 81 8000 1102',
    },
    {
      name: 'Lic. Camila Torres',
      email: 'camila.torres@fisia.mx',
      specialty: 'Rehabilitación Neurológica',
      phone: '+52 81 8000 1103',
    },
    {
      name: 'Lic. David Salgado',
      email: 'david.salgado@fisia.mx',
      specialty: 'Fisioterapia Pediátrica',
      phone: '+52 81 8000 1104',
    },
    {
      name: 'Lic. Valeria Peña',
      email: 'valeria.pena@fisia.mx',
      specialty: 'Punción Seca y Dolor Miofascial',
      phone: '+52 81 8000 1105',
    },
  ];

  const specialists = [];
  for (const s of specialistsData) {
    const specialist = await prisma.user.upsert({
      where: { email: s.email },
      update: {
        name: s.name,
        role: Role.SPECIALIST,
        specialty: s.specialty,
        phone: s.phone,
        isActive: true,
        clinicId: clinic.id,
        passwordHash: defaultPasswordHash,
      },
      create: {
        clinicId: clinic.id,
        name: s.name,
        email: s.email,
        passwordHash: defaultPasswordHash,
        role: Role.SPECIALIST,
        specialty: s.specialty,
        phone: s.phone,
        isActive: true,
      },
    });
    specialists.push(specialist);
  }
  console.log(`✅ 5 Specialists provisioned successfully.`);

  // ---------------------------------------------------------------------------
  // 4. Core Services Catalog (6 Services)
  // ---------------------------------------------------------------------------
  const servicesData = [
    {
      name: 'Evaluación Inicial',
      description: 'Valoración biomecánica completa, historia clínica y plan personalizado de rehabilitación.',
      price: 850.00,
      durationMin: 60,
    },
    {
      name: 'Sesión Terapia Física',
      description: 'Sesión terapéutica integral con agentes físicos, electroterapia y ejercicios guiados.',
      price: 750.00,
      durationMin: 45,
    },
    {
      name: 'Terapia Manual',
      description: 'Técnicas especializadas de movilización articular y liberación de tejidos blandos.',
      price: 800.00,
      durationMin: 45,
    },
    {
      name: 'Punción Seca',
      description: 'Tratamiento invasivo con agujas estériles para la desactivación de puntos gatillo miofasciales.',
      price: 900.00,
      durationMin: 45,
    },
    {
      name: 'Masaje Descarga',
      description: 'Terapia manual descontracturante profunda enfocada en recuperación muscular y atletas.',
      price: 850.00,
      durationMin: 60,
    },
    {
      name: 'Readaptación Funcional',
      description: 'Entrenamiento neuromuscular y reeducación del gesto motor para reincorporación deportiva.',
      price: 800.00,
      durationMin: 50,
    },
  ];

  const services = [];
  for (const s of servicesData) {
    const service = await prisma.service.upsert({
      where: {
        clinicId_name: {
          clinicId: clinic.id,
          name: s.name,
        },
      },
      update: {
        description: s.description,
        price: s.price,
        durationMin: s.durationMin,
        isActive: true,
      },
      create: {
        clinicId: clinic.id,
        name: s.name,
        description: s.description,
        price: s.price,
        durationMin: s.durationMin,
        isActive: true,
      },
    });
    services.push(service);
  }
  console.log(`✅ 6 Core Services provisioned.`);

  // Helper map for linking
  const serviceByName = Object.fromEntries(services.map((s) => [s.name, s]));
  const specialistByName = Object.fromEntries(specialists.map((sp) => [sp.name, sp]));

  // ---------------------------------------------------------------------------
  // 5. Specialist-Service Assignments (SpecialistService Matrix)
  // ---------------------------------------------------------------------------
  const assignments: { specialistId: string; serviceId: string }[] = [
    // Evaluación Inicial: all specialists + owner
    { specialistId: owner.id, serviceId: serviceByName['Evaluación Inicial'].id },
    ...specialists.map((sp) => ({ specialistId: sp.id, serviceId: serviceByName['Evaluación Inicial'].id })),

    // Sesión Terapia Física: Sofia Garza, Mateo Rivas, Valeria Peña
    { specialistId: specialistByName['Lic. Sofia Garza'].id, serviceId: serviceByName['Sesión Terapia Física'].id },
    { specialistId: specialistByName['Lic. Mateo Rivas'].id, serviceId: serviceByName['Sesión Terapia Física'].id },
    { specialistId: specialistByName['Lic. Valeria Peña'].id, serviceId: serviceByName['Sesión Terapia Física'].id },

    // Terapia Manual: Mateo Rivas, Sofia Garza, Valeria Peña
    { specialistId: specialistByName['Lic. Mateo Rivas'].id, serviceId: serviceByName['Terapia Manual'].id },
    { specialistId: specialistByName['Lic. Sofia Garza'].id, serviceId: serviceByName['Terapia Manual'].id },
    { specialistId: specialistByName['Lic. Valeria Peña'].id, serviceId: serviceByName['Terapia Manual'].id },

    // Punción Seca: Valeria Peña, Sofia Garza
    { specialistId: specialistByName['Lic. Valeria Peña'].id, serviceId: serviceByName['Punción Seca'].id },
    { specialistId: specialistByName['Lic. Sofia Garza'].id, serviceId: serviceByName['Punción Seca'].id },

    // Masaje Descarga: Sofia Garza, Mateo Rivas
    { specialistId: specialistByName['Lic. Sofia Garza'].id, serviceId: serviceByName['Masaje Descarga'].id },
    { specialistId: specialistByName['Lic. Mateo Rivas'].id, serviceId: serviceByName['Masaje Descarga'].id },

    // Readaptación Funcional: Sofia Garza, Camila Torres
    { specialistId: specialistByName['Lic. Sofia Garza'].id, serviceId: serviceByName['Readaptación Funcional'].id },
    { specialistId: specialistByName['Lic. Camila Torres'].id, serviceId: serviceByName['Readaptación Funcional'].id },
  ];

  for (const a of assignments) {
    await prisma.specialistService.upsert({
      where: {
        specialistId_serviceId: {
          specialistId: a.specialistId,
          serviceId: a.serviceId,
        },
      },
      update: {
        clinicId: clinic.id,
      },
      create: {
        clinicId: clinic.id,
        specialistId: a.specialistId,
        serviceId: a.serviceId,
      },
    });
  }
  console.log(`✅ SpecialistService assignment matrix linked (${assignments.length} assignments).`);

  // ---------------------------------------------------------------------------
  // 6. Core Retail Products (Inventory)
  // ---------------------------------------------------------------------------
  const productsData = [
    {
      name: 'Kinesiotape CureTape Pro (5cm x 5m)',
      description: 'Cinta neuromuscular elástica de algodón hipoalergénico resistente al agua.',
      price: 250.00,
      stock: 40,
      sku: 'KT-PRO-5M',
    },
    {
      name: 'Banda Elástica Theraband (Nivel Medio)',
      description: 'Banda de resistencia látex color azul para rehabilitación neuromuscular progresiva.',
      price: 350.00,
      stock: 25,
      sku: 'TB-BLUE-MED',
    },
    {
      name: 'Gel Crioterápico Fisia 250ml',
      description: 'Gel desinflamatorio con extracto de árnica y mentol para alivio analgésico muscular.',
      price: 180.00,
      stock: 30,
      sku: 'GEL-CRIO-250',
    },
  ];

  const products = [];
  for (const p of productsData) {
    const product = await prisma.product.upsert({
      where: {
        clinicId_sku: {
          clinicId: clinic.id,
          sku: p.sku,
        },
      },
      update: {
        name: p.name,
        description: p.description,
        price: p.price,
        stock: p.stock,
        isActive: true,
      },
      create: {
        clinicId: clinic.id,
        name: p.name,
        description: p.description,
        price: p.price,
        stock: p.stock,
        sku: p.sku,
        isActive: true,
      },
    });
    products.push(product);
  }
  console.log(`✅ 3 Retail Products seeded.`);

  // ---------------------------------------------------------------------------
  // 7. Sample Customer (Patient)
  // ---------------------------------------------------------------------------
  let customer = await prisma.customer.findFirst({
    where: {
      clinicId: clinic.id,
      phone: '+52 81 1999 8877',
    },
  });

  if (!customer) {
    customer = await prisma.customer.create({
      data: {
        clinicId: clinic.id,
        name: 'Carlos Benítez',
        email: 'carlos.benitez@gmail.com',
        phone: '+52 81 1999 8877',
        notes: 'Paciente activo, corredor amateur. Esguince de tobillo derecho grado 2.',
      },
    });
  }
  console.log(`✅ Sample patient created: ${customer.name}`);

  // ---------------------------------------------------------------------------
  // 8. Sample Completed Appointment
  // ---------------------------------------------------------------------------
  const appointmentDate = new Date();
  appointmentDate.setHours(10, 0, 0, 0);

  let appointment = await prisma.appointment.findFirst({
    where: {
      clinicId: clinic.id,
      customerId: customer.id,
      specialistId: specialistByName['Lic. Sofia Garza'].id,
    },
  });

  if (!appointment) {
    appointment = await prisma.appointment.create({
      data: {
        clinicId: clinic.id,
        customerId: customer.id,
        specialistId: specialistByName['Lic. Sofia Garza'].id,
        serviceId: serviceByName['Sesión Terapia Física'].id,
        dateTime: appointmentDate,
        status: AppointmentStatus.COMPLETED,
        notes: 'Sesión terapéutica inicial completada satisfactoriamente con Sofia Garza.',
      },
    });
  }
  console.log(`✅ Sample appointment created: ${appointment.id}`);

  // ---------------------------------------------------------------------------
  // 9. Sample Order & Items (Cash Closing Foundation)
  // ---------------------------------------------------------------------------
  let order = await prisma.order.findUnique({
    where: { appointmentId: appointment.id },
  });

  if (!order) {
    order = await prisma.order.create({
      data: {
        clinicId: clinic.id,
        customerId: customer.id,
        appointmentId: appointment.id,
        total: 1000.00, // $750 service + $250 product
        paymentMethod: PaymentMethod.CASH,
        status: OrderStatus.PAID,
        paidAt: new Date(),
        notes: 'Pago en efectivo de sesión ($750) + Kinesiotape ($250). Corte de caja inicial.',
        items: {
          create: [
            {
              serviceId: serviceByName['Sesión Terapia Física'].id,
              quantity: 1,
              unitPrice: 750.00,
            },
            {
              productId: products[0].id,
              quantity: 1,
              unitPrice: 250.00,
            },
          ],
        },
      },
    });
  }
  console.log(`✅ Sample paid order for daily cash closing created: ${order.id} ($${order.total} CASH)`);

  // ---------------------------------------------------------------------------
  // 10. Sample Waitlist Queue Entry
  // ---------------------------------------------------------------------------
  let waitlistPatient = await prisma.customer.findFirst({
    where: {
      clinicId: clinic.id,
      phone: '+52 81 2233 4455',
    },
  });

  if (!waitlistPatient) {
    waitlistPatient = await prisma.customer.create({
      data: {
        clinicId: clinic.id,
        name: 'Mariana Lozano',
        email: 'mariana.lozano@gmail.com',
        phone: '+52 81 2233 4455',
        notes: 'Requiere punción seca por contractura severa.',
      },
    });
  }

  const tomorrow = new Date();
  tomorrow.setDate(tomorrow.getDate() + 1);
  tomorrow.setHours(16, 0, 0, 0);

  const existingWaitlist = await prisma.waitlist.findFirst({
    where: {
      clinicId: clinic.id,
      customerId: waitlistPatient.id,
    },
  });

  if (!existingWaitlist) {
    await prisma.waitlist.create({
      data: {
        clinicId: clinic.id,
        customerId: waitlistPatient.id,
        serviceId: serviceByName['Punción Seca'].id,
        specialistId: specialistByName['Lic. Valeria Peña'].id,
        preferredDate: tomorrow,
        status: WaitlistStatus.WAITING,
        priority: 1,
        notes: 'Disponible por la tarde para cualquier cancelación de punción seca con Lic. Valeria Peña.',
      },
    });
  }
  console.log(`✅ Sample waitlist queue record created.`);

  console.log('🎉 APPCITAS Milestone 1 database seed completed successfully!');
}

main()
  .catch((e) => {
    console.error('❌ Error during seeding:', e);
    process.exit(1);
  })
  .finally(async () => {
    await prisma.$disconnect();
  });
```

---

## 6. Features Discovered

| # | Category | Feature | Description | Inputs | Outputs | Error Behavior | Discovered Via |
|---|----------|---------|-------------|--------|---------|----------------|----------------|
| 1 | Multitenancy | Clinic Tenant Root | Container entity for organization; isolates all clinic data via unique slug and id | `name`, `slug`, `phone`, `address` | `Clinic` record | Rejects duplicate slug with P2002 unique constraint error | ORIGINAL_REQUEST R1, PROJECT.md F1 |
| 2 | Auth / Staff | User & Credentials | Multi-role user accounts (`OWNER`, `SPECIALIST`) with bcrypt `passwordHash` and active flag | `clinicId`, `name`, `email`, `passwordHash`, `role`, `specialty`, `phone` | `User` record | Rejects duplicate email with P2002; rejects null required fields with P2011 | ORIGINAL_REQUEST R1 & R2, PROJECT.md F2 |
| 3 | Calendar Sync | Google OAuth Tokens | Preparation fields on `User` (`googleAccessToken`, `googleRefreshToken`) for Phase 2 bidirectional Google Calendar sync | OAuth callback access/refresh tokens | Persisted token strings on User | None (nullable fields) | Roadmap.md Fase 2, DISPATCH.md |
| 4 | Catalog | Service Catalog | Clinic service catalog with `price: Decimal(10,2)` and `durationMin: Int` | `clinicId`, `name`, `price`, `durationMin`, `description` | `Service` record | Rejects duplicate `(clinicId, name)` with P2002 | ORIGINAL_REQUEST R1, PROJECT.md F3, DISPATCH.md |
| 5 | Assignment | SpecialistService Join Model | Explicit join model linking specialists to eligible services with composite uniqueness `@@unique([specialistId, serviceId])` | `clinicId`, `specialistId`, `serviceId` | `SpecialistService` record | Rejects duplicate assignment with P2002; rejects non-existent specialist or service with P2003 | ORIGINAL_REQUEST R1, PROJECT.md F3, DISPATCH.md |
| 6 | Inventory | Product Inventory | Physical therapy supplies and retail items for clinic sales | `clinicId`, `name`, `price`, `stock`, `sku` | `Product` record | Rejects duplicate `(clinicId, sku)` with P2002 | ORIGINAL_REQUEST R1, PROJECT.md F4, DISPATCH.md |
| 7 | Directory | Customer (Patient) Model | Patient contact information and clinical notes | `clinicId`, `name`, `phone`, `email`, `notes` | `Customer` record | Missing required name or phone triggers P2011 | ORIGINAL_REQUEST R1, PROJECT.md F5, DISPATCH.md |
| 8 | Scheduling | Appointment Scheduling | Clinical appointment linking patient, specialist, service, and scheduled `dateTime` | `clinicId`, `customerId`, `specialistId`, `serviceId`, `dateTime`, `status`, `notes` | `Appointment` record | Missing foreign keys triggers P2003 | ORIGINAL_REQUEST R1, PROJECT.md F6, DISPATCH.md |
| 9 | Billing | Order & Daily Cash Closing | Payment tracking recording `paymentMethod` (`CASH`, `CARD`, `TRANSFER`), `total`, `paidAt`, and `status` | `clinicId`, `customerId`, `appointmentId`, `total`, `paymentMethod`, `paidAt`, `status` | `Order` record | Rejects invalid `paymentMethod` or `status` enum values | ORIGINAL_REQUEST R1, PROJECT.md F7, DISPATCH.md |
| 10 | Billing | OrderItem Line Item | Granular order lines representing either a clinical service or a retail product with unit price | `orderId`, `serviceId?`, `productId?`, `quantity`, `unitPrice` | `OrderItem` record | Missing `orderId` triggers P2003 | ORIGINAL_REQUEST R1, PROJECT.md F7, DISPATCH.md |
| 11 | Queue | Waitlist Queue | Cancellation waitlist queue linking patient, desired service, optional specialist, and priority | `clinicId`, `customerId`, `serviceId?`, `specialistId?`, `preferredDate?`, `status`, `priority` | `Waitlist` record | Missing `clinicId` or `customerId` triggers P2003 | ORIGINAL_REQUEST R1, PROJECT.md F8, DISPATCH.md |
| 12 | Database | Seed Script Automation | Automated idempotent provisioning script populating Clínica Fisia, 5 specialists, catalog, and sample data | `npx prisma db seed` | Fully seeded database | Script exits with code 1 upon uncaught exception | ORIGINAL_REQUEST R1, Acceptance Criteria, DISPATCH.md |
| 13 | Infrastructure | Prisma Client Singleton | Global singleton instance avoiding connection exhaustion in Next.js development server | Node `globalThis` | Single shared `PrismaClient` | Connection pool errors if not used | PROJECT.md F10 |

---

## 7. Edge Cases & Boundary Conditions

| # | Feature | Input | Observed Behavior |
|---|---------|-------|-------------------|
| 1 | Multi-Tenant Data Leakage | Cross-tenant query without `clinicId` filter | Schema enforces `clinicId` across all child entities; query layer MUST filter by `where: { clinicId }` to prevent tenant cross-contamination. |
| 2 | Duplicate Specialist Assignment | Creating a duplicate `SpecialistService` row for the same specialist and service | Blocked by PostgreSQL `@@unique([specialistId, serviceId])` constraint; Prisma raises P2002 error. |
| 3 | Clinic Teardown | `prisma.clinic.delete({ where: { id } })` | Cascades delete to all child tables (`users`, `services`, `specialistServices`, `products`, `customers`, `appointments`, `orders`, `waitlists`), cleanly purging tenant data without orphaned rows. |
| 4 | Specialist Deletion with Appointments | Attempting hard delete on a `User` who has historical `Appointment` records | Blocked by `onDelete: Restrict` on `Appointment.specialistId`; database integrity prevents destruction of patient appointment records. |
| 5 | Service Deletion with Appointments | Attempting hard delete on a `Service` linked to existing `Appointment` records | Blocked by `onDelete: Restrict` on `Appointment.serviceId`; preserves medical procedure history. |
| 6 | Customer Deletion with Orders | Customer requests deletion under privacy laws | Past `Order` rows retain payment totals and line items, setting `customerId: null` via `onDelete: SetNull` to maintain financial audit trails. |
| 7 | Floating Point Rounding in Orders | Calculating total across multiple services and products | Avoided by using Prisma `Decimal` with PostgreSQL `DECIMAL(10, 2)` precision for all monetary columns (`price`, `total`, `unitPrice`). |
| 8 | Multiple Seed Executions | Running `npx prisma db seed` multiple times sequentially | All seed operations use `upsert` or check existing records before creating, ensuring 100% idempotency without duplicate key collisions. |
| 9 | Duplicate Service Name within Clinic | Attempting to create two services named "Sesión Terapia Física" in the same clinic | Blocked by `@@unique([clinicId, name])`; distinct clinics can still use identical service names independently. |
| 10 | Duplicate SKU within Clinic | Attempting to assign the same SKU to two products in the same clinic | Blocked by `@@unique([clinicId, sku])`; SKU uniqueness is scoped to the tenant. |
| 11 | Appointment to Order One-to-One | Creating two separate `Order` records referencing the same `appointmentId` | Blocked by `@unique` attribute on `Order.appointmentId`, ensuring each appointment is billed at most once. |
| 12 | OrderItem Heterogeneity | Order item referencing both a service and product or neither | Schema allows `serviceId: null` when billing a retail product, or `productId: null` when billing a clinical session. |
