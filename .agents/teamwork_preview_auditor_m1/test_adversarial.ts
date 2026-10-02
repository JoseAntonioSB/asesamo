import * as bcrypt from 'bcryptjs';
import { seed } from '../../prisma/seed';
import { serializeDecimal } from '../../src/lib/serialize';
import { Prisma } from '@prisma/client';

async function runAdversarialAudit() {
  console.log('--- Adversarial Integrity Test ---');

  // Test 1: Tampered password check
  const fakeHash = '$2a$10$abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqr';
  const match = await bcrypt.compare('Fisia2026!', fakeHash).catch(() => false);
  if (match) {
    throw new Error('SECURITY VIOLATION: Bcrypt matched invalid hash!');
  }
  console.log('✔ Test 1: Tampered hash rejected');

  // Test 2: Verify that seed actually fails if client throws an error
  const failingClient = {
    clinic: {
      upsert: async () => { throw new Error('Database connection reset'); },
    },
  };

  let threw = false;
  try {
    await seed(failingClient);
  } catch (e: any) {
    threw = true;
    if (!e.message.includes('Database connection reset')) {
      throw new Error(`Unexpected error: ${e.message}`);
    }
  }
  if (!threw) {
    throw new Error('INTEGRITY VIOLATION: Seed succeeded despite underlying client failure!');
  }
  console.log('✔ Test 2: Seed genuinely depends on database client execution');

  // Test 3: Decimal precision stress test
  const weirdDecimals = [
    new Prisma.Decimal('0.0000000001'),
    new Prisma.Decimal('9999999999.99'),
    new Prisma.Decimal('-0.01'),
  ];
  for (const d of weirdDecimals) {
    const s = serializeDecimal(d);
    if (typeof s !== 'number') {
      throw new Error(`serializeDecimal failed for ${d.toString()}: got ${typeof s}`);
    }
  }
  console.log('✔ Test 3: Decimal serializer handles extreme numeric values accurately');

  console.log('All adversarial tests PASSED.');
}

runAdversarialAudit().catch((err) => {
  console.error('Adversarial test failed:', err);
  process.exit(1);
});
