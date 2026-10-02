import sqlite3
import uuid
import re
import datetime

def run_verification():
    print("=== EMPIRICAL SEED HARNESS START ===")
    conn = sqlite3.connect(":memory:")
    conn.execute("PRAGMA foreign_keys = ON;")

    # Custom functions for SQLite to match PostgreSQL
    conn.create_function("gen_random_uuid", 0, lambda: str(uuid.uuid4()))
    conn.create_function("crypt", 2, lambda pw, salt: f"$2a$10$simulatedhashfor_{pw}")
    conn.create_function("gen_salt", 2, lambda algo, cost: "simulated_salt")
    conn.create_function("now", 0, lambda: "2026-10-01 12:00:00")

    cursor = conn.cursor()

    # 1. Create tables according to 000001_init.up.sql
    cursor.executescript("""
    CREATE TABLE tenants (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL,
      rfc TEXT,
      logo_r2_key TEXT,
      plan TEXT NOT NULL DEFAULT 'starter' CHECK (plan IN ('starter', 'pro', 'enterprise')),
      is_active BOOLEAN NOT NULL DEFAULT 1,
      created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      deleted_at TEXT
    );

    CREATE TABLE plants (
      id TEXT PRIMARY KEY,
      tenant_id TEXT NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
      name TEXT NOT NULL,
      address TEXT,
      city TEXT,
      state TEXT,
      block_on_expired_docs BOOLEAN NOT NULL DEFAULT 1,
      created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      deleted_at TEXT
    );

    CREATE TABLE access_points (
      id TEXT PRIMARY KEY,
      plant_id TEXT NOT NULL REFERENCES plants(id) ON DELETE RESTRICT,
      tenant_id TEXT NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
      name TEXT NOT NULL,
      description TEXT,
      is_active BOOLEAN NOT NULL DEFAULT 1,
      created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      deleted_at TEXT
    );

    CREATE TABLE users (
      id TEXT PRIMARY KEY,
      tenant_id TEXT REFERENCES tenants(id) ON DELETE RESTRICT,
      email TEXT NOT NULL UNIQUE,
      password_hash TEXT NOT NULL,
      full_name TEXT NOT NULL,
      phone TEXT,
      role TEXT NOT NULL,
      avatar_r2_key TEXT,
      is_active BOOLEAN NOT NULL DEFAULT 1,
      last_login_at TEXT,
      created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      deleted_at TEXT
    );

    CREATE TABLE contractors (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL,
      rfc TEXT,
      repse_number TEXT,
      contractor_type TEXT NOT NULL DEFAULT 'company',
      admin_user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
      is_active BOOLEAN NOT NULL DEFAULT 1,
      created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      deleted_at TEXT
    );

    CREATE TABLE plant_contractors (
      id TEXT PRIMARY KEY,
      plant_id TEXT NOT NULL REFERENCES plants(id) ON DELETE RESTRICT,
      tenant_id TEXT NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
      contractor_id TEXT NOT NULL REFERENCES contractors(id) ON DELETE RESTRICT,
      status TEXT NOT NULL DEFAULT 'pending',
      invited_by TEXT REFERENCES users(id) ON DELETE SET NULL,
      invited_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      activated_at TEXT,
      created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      UNIQUE (plant_id, contractor_id)
    );

    CREATE TABLE contractor_workers (
      id TEXT PRIMARY KEY,
      contractor_id TEXT NOT NULL REFERENCES contractors(id) ON DELETE RESTRICT,
      user_id TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
      job_title TEXT,
      curp TEXT,
      imss_number TEXT,
      is_active BOOLEAN NOT NULL DEFAULT 1,
      created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      deleted_at TEXT,
      UNIQUE (contractor_id, user_id)
    );

    CREATE TABLE document_types (
      id TEXT PRIMARY KEY,
      tenant_id TEXT REFERENCES tenants(id) ON DELETE RESTRICT,
      code TEXT NOT NULL,
      name TEXT NOT NULL,
      description TEXT,
      applies_to TEXT NOT NULL DEFAULT 'worker' CHECK (applies_to IN ('worker', 'contractor')),
      default_validity_days INT,
      is_required_default BOOLEAN NOT NULL DEFAULT 1,
      is_system BOOLEAN NOT NULL DEFAULT 0,
      is_active BOOLEAN NOT NULL DEFAULT 1,
      created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      UNIQUE (tenant_id, code)
    );

    CREATE TABLE plant_document_requirements (
      id TEXT PRIMARY KEY,
      plant_id TEXT NOT NULL REFERENCES plants(id) ON DELETE RESTRICT,
      tenant_id TEXT NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
      document_type_id TEXT NOT NULL REFERENCES document_types(id) ON DELETE RESTRICT,
      is_required BOOLEAN NOT NULL DEFAULT 1,
      validity_days INT,
      created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      UNIQUE (plant_id, document_type_id)
    );

    CREATE TABLE documents (
      id TEXT PRIMARY KEY,
      user_id TEXT REFERENCES users(id) ON DELETE RESTRICT,
      contractor_id TEXT REFERENCES contractors(id) ON DELETE RESTRICT,
      document_type_id TEXT NOT NULL REFERENCES document_types(id) ON DELETE RESTRICT,
      r2_key TEXT NOT NULL,
      file_name TEXT NOT NULL,
      file_size_bytes INT,
      mime_type TEXT,
      issued_at TEXT,
      expires_at TEXT,
      status TEXT NOT NULL DEFAULT 'pending_review',
      reviewed_by TEXT REFERENCES users(id) ON DELETE SET NULL,
      reviewed_at TEXT,
      rejection_reason TEXT,
      ai_extracted_data TEXT,
      ai_confidence NUMERIC,
      created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      deleted_at TEXT,
      CONSTRAINT chk_document_owner CHECK (
        (user_id IS NOT NULL AND contractor_id IS NULL) OR
        (user_id IS NULL AND contractor_id IS NOT NULL)
      )
    );
    """)

    # Pre-seed document_types as in 000001_init.up.sql
    cursor.executescript("""
    INSERT INTO document_types (id, tenant_id, code, name, applies_to, default_validity_days, is_required_default, is_system) VALUES
      (gen_random_uuid(), NULL, 'DC3',      'Constancia de Capacitación DC-3',          'worker',      365,  1,  1),
      (gen_random_uuid(), NULL, 'REPSE',    'Registro REPSE',                            'contractor',  365,  1,  1),
      (gen_random_uuid(), NULL, 'SUA',      'Comprobante de Pago SUA (IMSS/INFONAVIT)', 'contractor',  30,   1,  1),
      (gen_random_uuid(), NULL, 'SAT_32D',  'Opinión de Cumplimiento SAT (32-D)',        'contractor',  30,   1,  1),
      (gen_random_uuid(), NULL, 'INE',      'Identificación Oficial (INE/IFE)',          'worker',      NULL, 1,  1),
      (gen_random_uuid(), NULL, 'MEDICAL',  'Certificado Médico de Aptitud',             'worker',      365,  1,  1);
    """)
    print("[INIT] Schema and initial document_types created.")

    # Read and parse seed file
    with open("db/seeds/000001_seed_dev_data.sql") as f:
        raw_sql = f.read()

    # Clean SQL comments
    clean_sql = re.sub(r"--.*$", "", raw_sql, flags=re.MULTILINE)
    raw_stmts = [s.strip() for s in clean_sql.split(";") if s.strip()]

    # Filter out BEGIN and COMMIT
    stmts = [s for s in raw_stmts if s.upper() not in ("BEGIN", "COMMIT")]
    print(f"[PARSE] Total seed statements extracted: {len(stmts)}")

    def adapt_pg_to_sqlite(stmt):
        # Adapt INTERVAL syntax to SQLite date syntax
        # e.g.: CURRENT_DATE - INTERVAL '60 days' -> date('now', '-60 days')
        # e.g.: CURRENT_DATE + INTERVAL '305 days' -> date('now', '+305 days')
        # e.g.: now() - INTERVAL '59 days' -> datetime('now', '-59 days')
        stmt = re.sub(r"CURRENT_DATE\s*-\s*INTERVAL\s*'(\d+)\s*days'", r"date('now', '-\1 days')", stmt)
        stmt = re.sub(r"CURRENT_DATE\s*\+\s*INTERVAL\s*'(\d+)\s*days'", r"date('now', '+\1 days')", stmt)
        stmt = re.sub(r"now\(\)\s*-\s*INTERVAL\s*'(\d+)\s*days'", r"datetime('now', '-\1 days')", stmt)
        stmt = re.sub(r"now\(\)\s*\+\s*INTERVAL\s*'(\d+)\s*days'", r"datetime('now', '+\1 days')", stmt)
        return stmt

    # Execute Pass 1
    print("\n--- EXECUTING PASS 1 (Initial Seed) ---")
    for idx, stmt in enumerate(stmts):
        adapted = adapt_pg_to_sqlite(stmt)
        try:
            cursor.execute(adapted)
        except Exception as e:
            print(f"FAILED on statement {idx+1}:\n{stmt}")
            raise e
    conn.commit()
    print("[PASS 1 SUCCESS] All 25 statements executed without error.")

    # Check row counts after Pass 1
    tables = [
        "tenants", "plants", "access_points", "users",
        "contractors", "plant_contractors", "contractor_workers",
        "document_types", "plant_document_requirements", "documents"
    ]
    counts_pass1 = {}
    for t in tables:
        cursor.execute(f"SELECT COUNT(*) FROM {t}")
        cnt = cursor.fetchone()[0]
        counts_pass1[t] = cnt
        print(f"  Table {t:<30}: {cnt} rows")

    # Specific count assertions
    assert counts_pass1["tenants"] == 1, "Expected 1 tenant"
    assert counts_pass1["plants"] == 1, "Expected 1 plant"
    assert counts_pass1["access_points"] == 1, "Expected 1 access point"
    assert counts_pass1["users"] == 5, f"Expected 5 users, got {counts_pass1['users']}"
    assert counts_pass1["contractors"] == 1, "Expected 1 contractor"
    assert counts_pass1["plant_contractors"] == 1, "Expected 1 plant_contractor"
    assert counts_pass1["contractor_workers"] == 3, "Expected 3 contractor_workers"
    assert counts_pass1["document_types"] == 6, "Expected 6 document_types"
    assert counts_pass1["plant_document_requirements"] == 6, "Expected 6 plant_document_requirements"
    assert counts_pass1["documents"] == 12, "Expected 12 documents"
    print("[PASS 1 COUNTS] All row counts match expectations.")

    # Execute Pass 2 (Idempotency check)
    print("\n--- EXECUTING PASS 2 (Idempotency Re-run) ---")
    for idx, stmt in enumerate(stmts):
        adapted = adapt_pg_to_sqlite(stmt)
        try:
            cursor.execute(adapted)
        except Exception as e:
            print(f"IDEMPOTENCY FAILURE on statement {idx+1}:\n{stmt}")
            raise e
    conn.commit()
    print("[PASS 2 SUCCESS] Second execution succeeded with zero errors.")

    # Check row counts after Pass 2 (must be identical)
    counts_pass2 = {}
    for t in tables:
        cursor.execute(f"SELECT COUNT(*) FROM {t}")
        cnt = cursor.fetchone()[0]
        counts_pass2[t] = cnt
        assert cnt == counts_pass1[t], f"Table {t} row count changed from {counts_pass1[t]} to {cnt}!"

    print("[IDEMPOTENCY VERIFIED] 100% identical row counts across both runs.")

    # Foreign key integrity check via PRAGMA foreign_key_check
    cursor.execute("PRAGMA foreign_key_check;")
    fk_violations = cursor.fetchall()
    if fk_violations:
        print("FK VIOLATIONS FOUND:", fk_violations)
        raise AssertionError("Foreign key violations detected")
    print("[FK CHECK] PRAGMA foreign_key_check returned 0 violations.")

    print("\n=== ALL EMPIRICAL VERIFICATIONS PASSED ===")

if __name__ == "__main__":
    run_verification()
