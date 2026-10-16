-- ============================================================
-- ASESAMO — Permisos de Trabajo de Alto Riesgo (PTAR Digital)
-- Migration: 000003_create_work_permits.up.sql
-- Standards: STPS NOM-009, NOM-033, NOM-027, NOM-029
-- ============================================================

-- ------------------------------------------------------------
-- 1. ENUM TYPES
-- ------------------------------------------------------------
DO $$ 
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'permit_type') THEN
    CREATE TYPE permit_type AS ENUM (
      'heights',           -- NOM-009-STPS-2011: Trabajos en altura
      'confined_space',    -- NOM-033-STPS-2015: Espacios confinados
      'hot_work',          -- NOM-027-STPS-2008: Soldadura y corte
      'electrical_loto'    -- NOM-029-STPS-2011: Mantenimiento eléctrico y LOTO
    );
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'permit_status') THEN
    CREATE TYPE permit_status AS ENUM (
      'draft',             -- Borrador inicial
      'pending_approval',  -- Enviado a revisión y autorización EHS
      'approved_active',   -- Autorizado y activo en campo
      'suspended',         -- Suspendido temporalmente (clima, emergencia)
      'closed',            -- Concluido y cerrado formalmente
      'cancelled'          -- Cancelado antes o durante ejecución
    );
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'permit_signature_role') THEN
    CREATE TYPE permit_signature_role AS ENUM (
      'requester',             -- Solicitante / Encargado de cuadrilla
      'contractor_supervisor', -- Supervisor de la empresa contratista
      'ehs_authorizer'         -- Responsable / Supervisor EHS de planta
    );
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'permit_signer_role') THEN
    CREATE DOMAIN permit_signer_role AS permit_signature_role;
  END IF;
END $$;

-- ------------------------------------------------------------
-- 2. WORK PERMITS TABLE
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS work_permits (
  id                  UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id           UUID          NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
  plant_id            UUID          REFERENCES plants(id) ON DELETE RESTRICT,
  permit_number       TEXT          NOT NULL,
  permit_type         permit_type   NOT NULL,
  status              permit_status NOT NULL DEFAULT 'draft',
  contractor_id       UUID          NOT NULL REFERENCES contractors(id) ON DELETE RESTRICT,
  work_area           TEXT          NOT NULL,
  description         TEXT          NOT NULL,
  hazard_description  TEXT,
  start_time          TIMESTAMPTZ   NOT NULL,
  end_time            TIMESTAMPTZ   NOT NULL,
  created_by_user_id  UUID          NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  approved_by_user_id UUID          REFERENCES users(id) ON DELETE SET NULL,
  approved_at         TIMESTAMPTZ,
  suspended_at        TIMESTAMPTZ,
  closed_at           TIMESTAMPTZ,
  cancelled_at        TIMESTAMPTZ,
  status_reason       TEXT,
  created_at          TIMESTAMPTZ   NOT NULL DEFAULT now(),
  updated_at          TIMESTAMPTZ   NOT NULL DEFAULT now(),
  deleted_at          TIMESTAMPTZ,
  CONSTRAINT uq_work_permits_tenant_number UNIQUE (tenant_id, permit_number),
  CONSTRAINT chk_work_permits_time_range CHECK (end_time > start_time)
);

CREATE INDEX IF NOT EXISTS idx_work_permits_tenant_id        ON work_permits(tenant_id);
CREATE INDEX IF NOT EXISTS idx_work_permits_plant_id         ON work_permits(plant_id);
CREATE INDEX IF NOT EXISTS idx_work_permits_contractor_id    ON work_permits(contractor_id);
CREATE INDEX IF NOT EXISTS idx_work_permits_status           ON work_permits(status);
CREATE INDEX IF NOT EXISTS idx_work_permits_permit_type      ON work_permits(permit_type);
CREATE INDEX IF NOT EXISTS idx_work_permits_created_by       ON work_permits(created_by_user_id);
CREATE INDEX IF NOT EXISTS idx_work_permits_tenant_status    ON work_permits(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_work_permits_date_range       ON work_permits(start_time, end_time);
CREATE INDEX IF NOT EXISTS idx_work_permits_active_window    ON work_permits(tenant_id, status, start_time, end_time)
  WHERE status = 'approved_active';

DROP TRIGGER IF EXISTS trg_work_permits_updated_at ON work_permits;
CREATE TRIGGER trg_work_permits_updated_at
  BEFORE UPDATE ON work_permits
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ------------------------------------------------------------
-- 3. PERMIT WORKERS (CREW ASSIGNMENT & COMPLIANCE SNAPSHOT)
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS permit_workers (
  id                  UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id           UUID          NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
  permit_id           UUID          NOT NULL REFERENCES work_permits(id) ON DELETE CASCADE,
  worker_id           UUID          NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  compliance_status   TEXT          NOT NULL DEFAULT 'compliant'
                                    CHECK (compliance_status IN ('compliant', 'non_compliant', 'warning')),
  compliance_snapshot JSONB         NOT NULL DEFAULT '{}'::jsonb,
  assigned_by         UUID          REFERENCES users(id) ON DELETE SET NULL,
  assigned_at         TIMESTAMPTZ   NOT NULL DEFAULT now(),
  created_at          TIMESTAMPTZ   NOT NULL DEFAULT now(),
  CONSTRAINT uq_permit_workers_permit_worker UNIQUE (permit_id, worker_id)
);

CREATE INDEX IF NOT EXISTS idx_permit_workers_permit_id  ON permit_workers(permit_id);
CREATE INDEX IF NOT EXISTS idx_permit_workers_worker_id  ON permit_workers(worker_id);
CREATE INDEX IF NOT EXISTS idx_permit_workers_tenant_id  ON permit_workers(tenant_id);

-- ------------------------------------------------------------
-- 4. PERMIT CHECKLISTS (STPS NOM SAFETY VERIFICATION ITEMS)
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS permit_checklists (
  id                  UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id           UUID          NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
  permit_id           UUID          NOT NULL REFERENCES work_permits(id) ON DELETE CASCADE,
  item_key            TEXT          NOT NULL,
  nom_standard        TEXT          NOT NULL,
  title               TEXT          NOT NULL,
  description         TEXT,
  category            TEXT          NOT NULL DEFAULT 'safety'
                                    CHECK (category IN ('safety', 'ppe', 'equipment', 'atmospheric', 'emergency')),
  is_mandatory        BOOLEAN       NOT NULL DEFAULT TRUE,
  is_verified         BOOLEAN       NOT NULL DEFAULT FALSE,
  verified_by         UUID          REFERENCES users(id) ON DELETE SET NULL,
  verified_at         TIMESTAMPTZ,
  measurements        JSONB         NOT NULL DEFAULT '{}'::jsonb,
  notes               TEXT,
  display_order       INT           NOT NULL DEFAULT 0,
  created_at          TIMESTAMPTZ   NOT NULL DEFAULT now(),
  updated_at          TIMESTAMPTZ   NOT NULL DEFAULT now(),
  CONSTRAINT uq_permit_checklists_permit_item UNIQUE (permit_id, item_key)
);

CREATE INDEX IF NOT EXISTS idx_permit_checklists_permit_id    ON permit_checklists(permit_id);
CREATE INDEX IF NOT EXISTS idx_permit_checklists_tenant_id    ON permit_checklists(tenant_id);
CREATE INDEX IF NOT EXISTS idx_permit_checklists_unverified   ON permit_checklists(permit_id, is_mandatory)
  WHERE is_verified = FALSE;

DROP TRIGGER IF EXISTS trg_permit_checklists_updated_at ON permit_checklists;
CREATE TRIGGER trg_permit_checklists_updated_at
  BEFORE UPDATE ON permit_checklists
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ------------------------------------------------------------
-- 5. PERMIT SIGNATURES (DIGITAL SIGN-OFF & NON-REPUDIATION)
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS permit_signatures (
  id                  UUID                  PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id           UUID                  NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
  permit_id           UUID                  NOT NULL REFERENCES work_permits(id) ON DELETE CASCADE,
  user_id             UUID                  NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  role                permit_signature_role NOT NULL,
  signature_data      TEXT,
  ip_address          TEXT,
  user_agent          TEXT,
  notes               TEXT,
  signed_at           TIMESTAMPTZ           NOT NULL DEFAULT now(),
  created_at          TIMESTAMPTZ           NOT NULL DEFAULT now(),
  CONSTRAINT uq_permit_signatures_permit_role UNIQUE (permit_id, role)
);

CREATE INDEX IF NOT EXISTS idx_permit_signatures_permit_id ON permit_signatures(permit_id);
CREATE INDEX IF NOT EXISTS idx_permit_signatures_user_id   ON permit_signatures(user_id);
CREATE INDEX IF NOT EXISTS idx_permit_signatures_tenant_id ON permit_signatures(tenant_id);

-- ------------------------------------------------------------
-- 6. SEED SPECIALIZED DOCUMENT TYPES FOR CANDADO INTELIGENTE
-- ------------------------------------------------------------
INSERT INTO document_types (id, tenant_id, code, name, applies_to, default_validity_days, is_required_default, is_system)
SELECT gen_random_uuid(), NULL, 'DC3_ALTURAS', 'Constancia DC-3 Trabajos en Altura (NOM-009)', 'worker', 365, FALSE, TRUE
WHERE NOT EXISTS (SELECT 1 FROM document_types WHERE tenant_id IS NULL AND code = 'DC3_ALTURAS');

INSERT INTO document_types (id, tenant_id, code, name, applies_to, default_validity_days, is_required_default, is_system)
SELECT gen_random_uuid(), NULL, 'DC3_ESPACIOS_CONFINADOS', 'Constancia DC-3 Espacios Confinados (NOM-033)', 'worker', 365, FALSE, TRUE
WHERE NOT EXISTS (SELECT 1 FROM document_types WHERE tenant_id IS NULL AND code = 'DC3_ESPACIOS_CONFINADOS');

INSERT INTO document_types (id, tenant_id, code, name, applies_to, default_validity_days, is_required_default, is_system)
SELECT gen_random_uuid(), NULL, 'DC3_CORTE_SOLDADURA', 'Constancia DC-3 Corte y Soldadura (NOM-027)', 'worker', 365, FALSE, TRUE
WHERE NOT EXISTS (SELECT 1 FROM document_types WHERE tenant_id IS NULL AND code = 'DC3_CORTE_SOLDADURA');

INSERT INTO document_types (id, tenant_id, code, name, applies_to, default_validity_days, is_required_default, is_system)
SELECT gen_random_uuid(), NULL, 'DC3_LOTO', 'Constancia DC-3 Seguridad Eléctrica y LOTO (NOM-029)', 'worker', 365, FALSE, TRUE
WHERE NOT EXISTS (SELECT 1 FROM document_types WHERE tenant_id IS NULL AND code = 'DC3_LOTO');

INSERT INTO document_types (id, tenant_id, code, name, applies_to, default_validity_days, is_required_default, is_system)
SELECT gen_random_uuid(), NULL, 'DC3_CONFINED', 'Constancia DC-3 Espacios Confinados (NOM-033)', 'worker', 365, FALSE, TRUE
WHERE NOT EXISTS (SELECT 1 FROM document_types WHERE tenant_id IS NULL AND code = 'DC3_CONFINED');

INSERT INTO document_types (id, tenant_id, code, name, applies_to, default_validity_days, is_required_default, is_system)
SELECT gen_random_uuid(), NULL, 'DC3_HOT_WORK', 'Constancia DC-3 Corte y Soldadura (NOM-027)', 'worker', 365, FALSE, TRUE
WHERE NOT EXISTS (SELECT 1 FROM document_types WHERE tenant_id IS NULL AND code = 'DC3_HOT_WORK');

INSERT INTO document_types (id, tenant_id, code, name, applies_to, default_validity_days, is_required_default, is_system)
SELECT gen_random_uuid(), NULL, 'DC3_ELECTRICAL', 'Constancia DC-3 Seguridad Eléctrica y LOTO (NOM-029)', 'worker', 365, FALSE, TRUE
WHERE NOT EXISTS (SELECT 1 FROM document_types WHERE tenant_id IS NULL AND code = 'DC3_ELECTRICAL');
