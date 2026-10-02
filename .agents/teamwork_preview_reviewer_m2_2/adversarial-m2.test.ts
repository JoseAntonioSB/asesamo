/// <reference path="../../src/types/next-auth.d.ts" />

import { Prisma, Role } from '@prisma/client';
import * as bcrypt from 'bcryptjs';
import { NextRequest } from 'next/server';
import { encode } from 'next-auth/jwt';

async function runAdversarialM2Suite() {
  console.log('╔══════════════════════════════════════════════════════════════╗');
  console.log('║       APPCITAS Milestone 2: Adversarial Security Suite       ║');
  console.log('╚══════════════════════════════════════════════════════════════╝\n');

  let passed = 0;
  let failed = 0;
  const failures: string[] = [];

  function assertAdv(condition: boolean, testId: string, description: string) {
    if (condition) {
      passed++;
      console.log(`  ✔ [PASS] ${testId}: ${description}`);
    } else {
      failed++;
      failures.push(`${testId}: ${description}`);
      console.error(`  ✖ [FAIL] ${testId}: ${description}`);
    }
  }

  // Initialize mock store for clean isolation
  const store = {
    clinics: new Map<string, any>(),
    users: new Map<string, any>(),
    services: new Map<string, any>(),
    specialistServices: new Map<string, any>(),
  };

  let idCounter = 1;
  const generateId = (prefix: string) => `${prefix}_${idCounter++}_${Date.now()}`;

  function matches(record: any, where: any): boolean {
    if (!where) return true;
    for (const [key, value] of Object.entries(where)) {
      if (value === undefined) continue;
      if (key === 'id' && typeof value === 'object' && value !== null && 'in' in value) {
        if (!Array.isArray((value as any).in) || !(value as any).in.includes(record.id)) return false;
      } else if (typeof value === 'object' && value !== null && !Array.isArray(value)) {
        for (const [subKey, subVal] of Object.entries(value)) {
          if (record[subKey] !== subVal) return false;
        }
      } else if (record[key] !== value) {
        return false;
      }
    }
    return true;
  }

  function applySelect(record: any, select?: any, include?: any) {
    if (!record) return record;
    let result: any;
    if (select) {
      result = {};
      for (const [field, isIncluded] of Object.entries(select)) {
        if (isIncluded) {
          if (field === 'specialistServices') {
            result.specialistServices = Array.from(store.specialistServices.values())
              .filter((ss) => ss.specialistId === record.id || ss.serviceId === record.id);
          } else {
            result[field] = record[field];
          }
        }
      }
    } else {
      result = { ...record };
    }
    if (include?.clinic && record.clinicId) {
      result.clinic = store.clinics.get(record.clinicId) || null;
    }
    return result;
  }

  function createOps(map: Map<string, any>, prefix: string) {
    return {
      async upsert({ where, update, create }: any) {
        let existing: any = null;
        for (const r of map.values()) {
          if (matches(r, where)) { existing = r; break; }
        }
        if (existing) {
          const u = { ...existing, ...update, updatedAt: new Date() };
          map.set(existing.id, u);
          return u;
        } else {
          const id = create.id ?? generateId(prefix);
          const n = { ...create, id, createdAt: new Date(), updatedAt: new Date() };
          map.set(id, n);
          return n;
        }
      },
      async create({ data, select, include }: any) {
        const id = data.id ?? generateId(prefix);
        const r = { ...data, id, createdAt: new Date(), updatedAt: new Date() };
        map.set(id, r);
        return applySelect(r, select, include);
      },
      async findFirst({ where, select, include }: any) {
        for (const r of map.values()) {
          if (matches(r, where)) return applySelect(r, select, include);
        }
        return null;
      },
      async findUnique({ where, select, include }: any) {
        for (const r of map.values()) {
          if (matches(r, where)) return applySelect(r, select, include);
        }
        return null;
      },
      async findMany({ where, select, include }: any = {}) {
        const out = [];
        for (const r of map.values()) {
          if (matches(r, where)) out.push(applySelect(r, select, include));
        }
        return out;
      },
      async count({ where }: any = {}) {
        let c = 0;
        for (const r of map.values()) {
          if (matches(r, where)) c++;
        }
        return c;
      },
      async deleteMany({ where }: any = {}) {
        let d = 0;
        for (const [id, r] of Array.from(map.entries())) {
          if (matches(r, where)) {
            map.delete(id);
            d++;
          }
        }
        return { count: d };
      },
    };
  }

  const mockDb: any = {
    _store: store,
    clinic: createOps(store.clinics, 'cln'),
    user: createOps(store.users, 'usr'),
    service: createOps(store.services, 'srv'),
    specialistService: createOps(store.specialistServices, 'ss'),
    async $transaction(fn: any) {
      return await fn(mockDb);
    },
    async $disconnect() {},
  };

  (globalThis as any).prisma = mockDb;
  process.env.NEXTAUTH_SECRET = 'adversarial-secret-key-32-chars-long-2026';
  process.env.NEXTAUTH_URL = 'http://localhost:3000';

  // Dynamically import project modules after setting globalThis.prisma
  const { serializeDecimal } = await import('../../src/lib/serialize');
  const { authOptions } = await import('../../src/lib/auth');
  const { middleware } = await import('../../src/middleware');
  const { POST: registerClinic } = await import('../../src/app/api/clinic/register/route');
  const { GET: getSpecialists, POST: createSpecialist } = await import('../../src/app/api/specialists/route');
  const { GET: getServices, POST: createService } = await import('../../src/app/api/services/route');
  const { PUT: assignSpecialists } = await import('../../src/app/api/services/[id]/assign/route');

  // --- ATTACK VECTOR 1: Registration Sanitization & Privilege Escalation ---
  console.log('\n▶ Vector 1: Registration Password Sanitization & Privilege Escalation');
  const regPayload = {
    clinicName: 'Clinica Adversarial Defense',
    ownerName: 'Dr. Hacker Defense',
    email: 'Hacker@Defense.COM', // Uppercase test
    password: 'SuperSecretPassword2026!',
    role: 'SUPERADMIN', // Injection attempt
    clinicId: 'cln_stolen_tenant', // Injection attempt
  };

  const regRes = await registerClinic(new Request('http://localhost:3000/api/clinic/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(regPayload),
  }));

  assertAdv(regRes.status === 201, 'ADV-01', 'Registration succeeds with 201');
  const regJson = await regRes.json();
  assertAdv(regJson.owner.role === Role.OWNER, 'ADV-02', 'Injected role is ignored; role is strictly OWNER');
  assertAdv(regJson.owner.password === undefined, 'ADV-03', 'Plaintext password is not in response');
  assertAdv(regJson.owner.passwordHash === undefined, 'ADV-04', 'Password hash is not in response');
  assertAdv(regJson.clinic.id !== 'cln_stolen_tenant', 'ADV-05', 'Injected clinicId ignored; newly generated tenant ID used');
  assertAdv(regJson.owner.email === 'hacker@defense.com', 'ADV-06', 'Email is normalized to lowercase in registration');

  const createdOwnerInDb = await mockDb.user.findFirst({ where: { email: 'hacker@defense.com' } });
  assertAdv(createdOwnerInDb.passwordHash.startsWith('$2'), 'ADV-07', 'Password stored as bcrypt hash');
  const bcryptVerify = await bcrypt.compare('SuperSecretPassword2026!', createdOwnerInDb.passwordHash);
  assertAdv(bcryptVerify === true, 'ADV-08', 'Stored hash accurately matches original password');

  // --- ATTACK VECTOR 2: NextAuth Credentials Inactive User & Normalization ---
  console.log('\n▶ Vector 2: NextAuth Authentication Bounds & Inactive Users');
  const credProvider = authOptions.providers.find((p: any) => p.id === 'credentials' || p.name === 'Credenciales');
  const authorize = (credProvider as any).options?.authorize || (credProvider as any).authorize;

  // Case insensitive login
  const upperLogin = await authorize({ email: 'HACKER@defense.com', password: 'SuperSecretPassword2026!' });
  assertAdv(upperLogin !== null && upperLogin.id === createdOwnerInDb.id, 'ADV-09', 'Email case insensitivity allows login');

  // Deactivated user cannot log in
  store.users.get(createdOwnerInDb.id).isActive = false;
  let deactivatedThrew = false;
  try {
    await authorize({ email: 'hacker@defense.com', password: 'SuperSecretPassword2026!' });
  } catch (err: any) {
    deactivatedThrew = true;
    assertAdv(err.message.includes('inválidas') || err.message.includes('inactivo'), 'ADV-10', 'Deactivated user rejected with generic error');
  }
  assertAdv(deactivatedThrew, 'ADV-11', 'Deactivated user is strictly prevented from authenticating');
  store.users.get(createdOwnerInDb.id).isActive = true; // restore

  // --- ATTACK VECTOR 3: Specialist Cross-Tenant Contamination & Privilege Escalation ---
  console.log('\n▶ Vector 3: Specialist Cross-Tenant Contamination & Quota Bounds');
  // Create Clinic B and Owner B
  const clinicB = await mockDb.clinic.create({ data: { name: 'Clinica B', slug: 'clinica-b' } });
  const ownerB = await mockDb.user.create({
    data: {
      clinicId: clinicB.id,
      name: 'Owner B',
      email: 'owner.b@test.com',
      passwordHash: 'hash',
      role: Role.OWNER,
      isActive: true,
    },
  });

  // Owner B attempts to create specialist in Clinic A
  (globalThis as any).__TEST_SESSION__ = {
    user: {
      id: ownerB.id,
      email: ownerB.email,
      role: Role.OWNER,
      clinicId: clinicB.id,
      clinicName: clinicB.name,
      clinicSlug: clinicB.slug,
    },
  };

  const maliciousSpecPayload = {
    name: 'Infiltrator Specialist',
    email: 'infiltrator@test.com',
    specialty: 'Traumatologia',
    clinicId: regJson.clinic.id, // Attempt to inject into Clinic A
    role: 'OWNER', // Attempt to escalate role to OWNER
  };

  const infiltratorRes = await createSpecialist(new Request('http://localhost:3000/api/specialists', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(maliciousSpecPayload),
  }));

  assertAdv(infiltratorRes.status === 201, 'ADV-12', 'Specialist creation responds 201');
  const infiltratorJson = await infiltratorRes.json();
  assertAdv(infiltratorJson.specialist.role === Role.SPECIALIST, 'ADV-13', 'Role forced to SPECIALIST; elevation blocked');
  
  const infiltratorInDb = await mockDb.user.findFirst({ where: { email: 'infiltrator@test.com' } });
  assertAdv(infiltratorInDb.clinicId === clinicB.id, 'ADV-14', 'Specialist is saved under session tenant (Clinic B), NOT Clinic A');
  assertAdv(infiltratorJson.specialist.passwordHash === undefined, 'ADV-15', 'Specialist password hash is sanitized from response');

  // --- ATTACK VECTOR 4: Cross-Tenant Service Assignment Injection ---
  console.log('\n▶ Vector 4: Service Assignment Cross-Tenant Defense');
  // Clinic A Service
  const serviceA = await mockDb.service.create({
    data: {
      clinicId: regJson.clinic.id,
      name: 'Punción Seca',
      durationMin: 45,
      price: 600,
      isActive: true,
    },
  });

  // Owner A session
  (globalThis as any).__TEST_SESSION__ = {
    user: {
      id: createdOwnerInDb.id,
      email: createdOwnerInDb.email,
      role: Role.OWNER,
      clinicId: regJson.clinic.id,
      clinicName: regJson.clinic.name,
      clinicSlug: regJson.clinic.slug,
    },
  };

  // Attempt to assign Clinic B's specialist to Clinic A's service
  const crossAssignReq = new Request(`http://localhost:3000/api/services/${serviceA.id}/assign`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ specialistIds: [infiltratorInDb.id] }), // specialist from Clinic B
  });
  const crossAssignRes = await assignSpecialists(crossAssignReq, { params: { id: serviceA.id } });
  assertAdv(crossAssignRes.status === 400, 'ADV-16', 'Assigning foreign specialist returns 400 Bad Request');
  const crossAssignJson = await crossAssignRes.json();
  assertAdv(crossAssignJson.error?.includes('no pertenecen'), 'ADV-17', 'Error confirms specialists do not belong to tenant');

  // Attempt to assign specialists to service belonging to another clinic
  (globalThis as any).__TEST_SESSION__ = {
    user: {
      id: ownerB.id,
      email: ownerB.email,
      role: Role.OWNER,
      clinicId: clinicB.id,
      clinicName: clinicB.name,
      clinicSlug: clinicB.slug,
    },
  };
  const foreignServiceReq = new Request(`http://localhost:3000/api/services/${serviceA.id}/assign`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ specialistIds: [infiltratorInDb.id] }),
  });
  const foreignServiceRes = await assignSpecialists(foreignServiceReq, { params: { id: serviceA.id } });
  assertAdv(foreignServiceRes.status === 404, 'ADV-18', 'Modifying foreign service returns 404 Not Found');

  // --- ATTACK VECTOR 5: Middleware Open Redirect and Query Tampering ---
  console.log('\n▶ Vector 5: Middleware Open Redirect and Callback URL Defense');
  // Attempt to pass malicious query into /dashboard
  const evilQueryReq = new NextRequest('http://localhost:3000/dashboard?redirect=https://evil.com&token=secret');
  const evilQueryRes = await middleware(evilQueryReq);
  assertAdv(evilQueryRes.status === 307, 'ADV-19', 'Unauthenticated request redirected with 307');
  const evilLoc = evilQueryRes.headers.get('location') || '';
  const parsedEvilLoc = new URL(evilLoc);
  assertAdv(parsedEvilLoc.pathname === '/login', 'ADV-20', 'Redirect destination is local /login');
  assertAdv(parsedEvilLoc.searchParams.get('callbackUrl') === '/dashboard?redirect=https://evil.com&token=secret', 'ADV-21', 'CallbackUrl correctly preserves path and query as safe relative parameter string');
  assertAdv(!evilLoc.startsWith('https://evil.com'), 'ADV-22', 'Middleware does NOT perform external open redirect');

  // Authenticated user hitting /login with malicious query
  const fakeToken = await encode({
    token: { id: 'u-1', email: 'u@test.com', role: Role.OWNER, clinicId: 'c-1', clinicName: 'C1', clinicSlug: 'c1' },
    secret: process.env.NEXTAUTH_SECRET!,
  });
  const authLoginReq = new NextRequest('http://localhost:3000/login?callbackUrl=https://evil.com', {
    headers: { cookie: `next-auth.session-token=${fakeToken}` },
  });
  const authLoginRes = await middleware(authLoginReq);
  assertAdv(authLoginRes.status === 307, 'ADV-23', 'Authenticated user on /login redirected with 307');
  const authLoc = authLoginRes.headers.get('location') || '';
  const parsedAuthLoc = new URL(authLoc);
  assertAdv(parsedAuthLoc.pathname === '/dashboard', 'ADV-24', 'Authenticated user strictly redirected to /dashboard, ignoring callbackUrl attack parameter');

  // --- ATTACK VECTOR 6: Date & Decimal Integrity Verification ---
  console.log('\n▶ Vector 6: Date Object & Decimal Structural Integrity');
  const targetDate = new Date('2026-09-14T04:30:00.000Z');
  const complexPrismaModel = {
    id: 'rec_99',
    createdAt: targetDate,
    updatedAt: new Date(targetDate.getTime() + 60000),
    price: new Prisma.Decimal('1999.99'),
    discount: new Prisma.Decimal('0.00'),
    items: [
      { name: 'Item 1', fee: new Prisma.Decimal('50.00'), expiry: targetDate },
      { name: 'Item 2', fee: new Prisma.Decimal('25.50'), expiry: null },
    ],
  };

  const serialized = serializeDecimal(complexPrismaModel);
  assertAdv(serialized.createdAt instanceof Date, 'ADV-25', 'createdAt remains an instance of Date');
  assertAdv(serialized.createdAt.getTime() === targetDate.getTime(), 'ADV-26', 'createdAt timestamp value is intact');
  assertAdv(serialized.items[0].expiry instanceof Date, 'ADV-27', 'Nested expiry date remains Date');
  assertAdv(typeof serialized.price === 'number' && serialized.price === 1999.99, 'ADV-28', 'Decimal price converted to number 1999.99');
  assertAdv(typeof serialized.discount === 'number' && serialized.discount === 0, 'ADV-29', 'Decimal 0.00 converted to number 0');
  assertAdv(typeof serialized.items[1].fee === 'number' && serialized.items[1].fee === 25.5, 'ADV-30', 'Nested array decimal converted to number');

  console.log('\n══════════════════════════════════════════════════════════════');
  console.log(`Total Adversarial Vectors Tested: ${passed + failed} | Passed: ${passed} | Failed: ${failed}`);
  console.log('══════════════════════════════════════════════════════════════\n');

  if (failed > 0) {
    console.error('ADVERSARIAL SUITE ENCOUNTERED FAILURES:');
    for (const f of failures) console.error(`  - ${f}`);
    process.exit(1);
  }
}

runAdversarialM2Suite().catch((err) => {
  console.error('Fatal unhandled error in adversarial suite:', err);
  process.exit(1);
});
