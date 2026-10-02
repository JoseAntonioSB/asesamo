-- ============================================================
-- ASESAMO — Database Schema
-- Migration: 000001_init.up.sql
-- Description: Full MVP schema — tenants, plants, users,
--              contractors, workers, documents, access control,
--              notifications.
-- ============================================================

-- ------------------------------------
-- EXTENSIONS
-- ------------------------------------
CREATE EXTENSION IF NOT EXISTS "pgcrypto";   -- gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS "citext";     -- case-insensitive text (emails)

-- ------------------------------------
-- HELPER: auto-update updated_at
-- ------------------------------------
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================================
-- 1. TENANTS
-- The paying customer. One company = one tenant.
-- A tenant can have multiple plants.
-- ============================================================
CREATE TABLE tenants (
  id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  name          TEXT        NOT NULL,                        -- "Grupo Industrial Norteño SA de CV"
  rfc           TEXT,                                        -- Mexican tax ID
  logo_r2_key   TEXT,                                        -- R2 key of their logo file
  plan          TEXT        NOT NULL DEFAULT 'starter'       -- 'starter' | 'pro' | 'enterprise'
                            CHECK (plan IN ('starter', 'pro', 'enterprise')),
  is_active     BOOLEAN     NOT NULL DEFAULT TRUE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at    TIMESTAMPTZ                                  -- soft delete
);

CREATE TRIGGER trg_tenants_updated_at
  BEFORE UPDATE ON tenants
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================
-- 2. PLANTS
-- A physical industrial facility. One tenant → many plants.
-- ============================================================
CREATE TABLE plants (
  id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id     UUID        NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
  name          TEXT        NOT NULL,                        -- "Planta Querétaro Norte"
  address       TEXT,
  city          TEXT,
  state         TEXT,
  -- Access control policy: each plant decides its own blocking behavior
  block_on_expired_docs BOOLEAN NOT NULL DEFAULT TRUE,       -- block worker if any required doc is expired
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at    TIMESTAMPTZ
);

CREATE INDEX idx_plants_tenant_id ON plants(tenant_id);

CREATE TRIGGER trg_plants_updated_at
  BEFORE UPDATE ON plants
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================
-- 3. ACCESS POINTS
-- A plant can have multiple gates / checkpoints.
-- e.g. "Caseta Norte", "Caseta Sur", "Entrada Vehicular"
-- ============================================================
CREATE TABLE access_points (
  id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  plant_id      UUID        NOT NULL REFERENCES plants(id) ON DELETE RESTRICT,
  tenant_id     UUID        NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
  name          TEXT        NOT NULL,                        -- "Caseta Principal"
  description   TEXT,
  is_active     BOOLEAN     NOT NULL DEFAULT TRUE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at    TIMESTAMPTZ
);

CREATE INDEX idx_access_points_plant_id   ON access_points(plant_id);
CREATE INDEX idx_access_points_tenant_id  ON access_points(tenant_id);

CREATE TRIGGER trg_access_points_updated_at
  BEFORE UPDATE ON access_points
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================
-- 4. USERS
-- Everyone with a login: plant admins, EHS managers,
-- guards, AND contractor workers (same table, different roles).
-- ============================================================
CREATE TYPE user_role AS ENUM (
  'super_admin',       -- ASESAMO platform operator
  'plant_admin',       -- Customer admin (manages their tenant)
  'ehs_manager',       -- EHS manager (approves docs, views dashboard)
  'guard',             -- Gate guard (only check-in/check-out)
  'contractor_admin',  -- Admin of a contractor company
  'worker'             -- Individual contractor worker
);

CREATE TABLE users (
  id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id       UUID        REFERENCES tenants(id) ON DELETE RESTRICT,
                              -- NULL only for super_admin
  email           CITEXT      NOT NULL UNIQUE,
  password_hash   TEXT        NOT NULL,
  full_name       TEXT        NOT NULL,
  phone           TEXT,
  role            user_role   NOT NULL,
  -- QR code: the UUID itself is used as the QR payload
  -- (no separate field needed — the QR encodes this id)
  avatar_r2_key   TEXT,                                      -- profile photo in R2
  is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
  last_login_at   TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_users_email     ON users(email);
CREATE INDEX idx_users_role      ON users(role);

CREATE TRIGGER trg_users_updated_at
  BEFORE UPDATE ON users
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================
-- 5. REFRESH TOKENS
-- Stored server-side to enable token revocation.
-- ============================================================
CREATE TABLE refresh_tokens (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash  TEXT        NOT NULL UNIQUE,                   -- SHA-256 of the actual token
  expires_at  TIMESTAMPTZ NOT NULL,
  revoked_at  TIMESTAMPTZ,                                   -- NULL = still valid
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);

-- ============================================================
-- 6. CONTRACTOR COMPANIES
-- A company (empresa contratista) that sends workers to plants.
-- An independent worker (persona física) also gets a record here
-- with company_type = 'individual'.
-- ============================================================
CREATE TYPE contractor_type AS ENUM ('company', 'individual');

CREATE TABLE contractors (
  id              UUID              PRIMARY KEY DEFAULT gen_random_uuid(),
  -- Contractors are GLOBAL (not tenant-scoped).
  -- They are linked to plants via plant_contractors (see below).
  name            TEXT              NOT NULL,
  rfc             TEXT,                                      -- RFC de la empresa contratista
  repse_number    TEXT,                                      -- Número de registro REPSE
  contractor_type contractor_type   NOT NULL DEFAULT 'company',
  -- The contractor_admin user who manages this contractor profile
  admin_user_id   UUID              REFERENCES users(id) ON DELETE SET NULL,
  is_active       BOOLEAN           NOT NULL DEFAULT TRUE,
  created_at      TIMESTAMPTZ       NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ       NOT NULL DEFAULT now(),
  deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_contractors_rfc          ON contractors(rfc);
CREATE INDEX idx_contractors_repse_number ON contractors(repse_number);

CREATE TRIGGER trg_contractors_updated_at
  BEFORE UPDATE ON contractors
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================
-- 7. PLANT ↔ CONTRACTOR RELATIONSHIP
-- A plant "activates" a contractor to allow their workers in.
-- Status managed per-plant independently.
-- ============================================================
CREATE TYPE plant_contractor_status AS ENUM ('pending', 'active', 'suspended', 'blocked');

CREATE TABLE plant_contractors (
  id              UUID                    PRIMARY KEY DEFAULT gen_random_uuid(),
  plant_id        UUID                    NOT NULL REFERENCES plants(id) ON DELETE RESTRICT,
  tenant_id       UUID                    NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
  contractor_id   UUID                    NOT NULL REFERENCES contractors(id) ON DELETE RESTRICT,
  status          plant_contractor_status NOT NULL DEFAULT 'pending',
  invited_by      UUID                    REFERENCES users(id) ON DELETE SET NULL,
  invited_at      TIMESTAMPTZ             NOT NULL DEFAULT now(),
  activated_at    TIMESTAMPTZ,
  created_at      TIMESTAMPTZ             NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ             NOT NULL DEFAULT now(),
  UNIQUE (plant_id, contractor_id)         -- one relationship per plant-contractor pair
);

CREATE INDEX idx_plant_contractors_plant_id      ON plant_contractors(plant_id);
CREATE INDEX idx_plant_contractors_tenant_id     ON plant_contractors(tenant_id);
CREATE INDEX idx_plant_contractors_contractor_id ON plant_contractors(contractor_id);

CREATE TRIGGER trg_plant_contractors_updated_at
  BEFORE UPDATE ON plant_contractors
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================
-- 8. CONTRACTOR ↔ WORKER RELATIONSHIP
-- Workers (users with role='worker') belong to a contractor.
-- ============================================================
CREATE TABLE contractor_workers (
  id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  contractor_id   UUID        NOT NULL REFERENCES contractors(id) ON DELETE RESTRICT,
  user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  -- Worker's job title within the contractor company
  job_title       TEXT,
  curp            TEXT,                                      -- CURP de la persona física
  imss_number     TEXT,                                      -- NSS del trabajador
  is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at      TIMESTAMPTZ,
  UNIQUE (contractor_id, user_id)
);

CREATE INDEX idx_contractor_workers_contractor_id ON contractor_workers(contractor_id);
CREATE INDEX idx_contractor_workers_user_id       ON contractor_workers(user_id);

CREATE TRIGGER trg_contractor_workers_updated_at
  BEFORE UPDATE ON contractor_workers
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================
-- 9. DOCUMENT TYPES
-- Configurable catalog. System defaults seeded on init.
-- Tenants can add custom document types for their plants.
-- ============================================================
CREATE TABLE document_types (
  id                    UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id             UUID        REFERENCES tenants(id) ON DELETE RESTRICT,
                                    -- NULL = global system-defined type
  code                  TEXT        NOT NULL,                -- 'DC3' | 'REPSE' | 'SUA' | 'SAT_32D' | 'INE' | 'MEDICAL' | custom
  name                  TEXT        NOT NULL,                -- "Constancia DC-3"
  description           TEXT,
  -- Who owns this document: the contractor company or the individual worker?
  applies_to            TEXT        NOT NULL DEFAULT 'worker'
                        CHECK (applies_to IN ('worker', 'contractor')),
  -- Default validity in days (plant can override per plant_document_requirements)
  default_validity_days INT,                                 -- NULL = no expiry (e.g. INE)
  is_required_default   BOOLEAN     NOT NULL DEFAULT TRUE,
  is_system             BOOLEAN     NOT NULL DEFAULT FALSE,  -- TRUE = created by ASESAMO, not editable
  is_active             BOOLEAN     NOT NULL DEFAULT TRUE,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, code)
);

CREATE INDEX idx_document_types_tenant_id ON document_types(tenant_id);
CREATE INDEX idx_document_types_code      ON document_types(code);

CREATE TRIGGER trg_document_types_updated_at
  BEFORE UPDATE ON document_types
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Seed: global system document types (tenant_id = NULL)
INSERT INTO document_types (id, tenant_id, code, name, applies_to, default_validity_days, is_required_default, is_system) VALUES
  (gen_random_uuid(), NULL, 'DC3',      'Constancia de Capacitación DC-3',          'worker',      365,  TRUE,  TRUE),
  (gen_random_uuid(), NULL, 'REPSE',    'Registro REPSE',                            'contractor',  365,  TRUE,  TRUE),
  (gen_random_uuid(), NULL, 'SUA',      'Comprobante de Pago SUA (IMSS/INFONAVIT)', 'contractor',  30,   TRUE,  TRUE),
  (gen_random_uuid(), NULL, 'SAT_32D',  'Opinión de Cumplimiento SAT (32-D)',        'contractor',  30,   TRUE,  TRUE),
  (gen_random_uuid(), NULL, 'INE',      'Identificación Oficial (INE/IFE)',          'worker',      NULL, TRUE,  TRUE),
  (gen_random_uuid(), NULL, 'MEDICAL',  'Certificado Médico de Aptitud',             'worker',      365,  TRUE,  TRUE);

-- ============================================================
-- 10. PLANT DOCUMENT REQUIREMENTS
-- Each plant can override which document types are required
-- and customize validity days.
-- ============================================================
CREATE TABLE plant_document_requirements (
  id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  plant_id          UUID        NOT NULL REFERENCES plants(id) ON DELETE RESTRICT,
  tenant_id         UUID        NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
  document_type_id  UUID        NOT NULL REFERENCES document_types(id) ON DELETE RESTRICT,
  is_required       BOOLEAN     NOT NULL DEFAULT TRUE,
  validity_days     INT,                                     -- overrides document_types.default_validity_days
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (plant_id, document_type_id)
);

CREATE INDEX idx_plant_doc_req_plant_id  ON plant_document_requirements(plant_id);
CREATE INDEX idx_plant_doc_req_tenant_id ON plant_document_requirements(tenant_id);

CREATE TRIGGER trg_plant_doc_req_updated_at
  BEFORE UPDATE ON plant_document_requirements
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================
-- 11. DOCUMENTS
-- Actual document files uploaded by contractors/workers.
-- Global to the user (not scoped to a single plant).
-- ============================================================
CREATE TYPE document_status AS ENUM (
  'pending_review',  -- uploaded, waiting for EHS manager approval
  'approved',        -- approved by EHS manager
  'rejected',        -- rejected (reason stored in rejection_reason)
  'expired'          -- was approved but validity period passed
);

CREATE TABLE documents (
  id                UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
  -- Owner: either a worker (user) or a contractor company
  user_id           UUID            REFERENCES users(id) ON DELETE RESTRICT,
  contractor_id     UUID            REFERENCES contractors(id) ON DELETE RESTRICT,
  document_type_id  UUID            NOT NULL REFERENCES document_types(id) ON DELETE RESTRICT,
  -- File stored in Cloudflare R2
  r2_key            TEXT            NOT NULL,                -- e.g. "contractors/uuid/dc3_cert.pdf"
  file_name         TEXT            NOT NULL,                -- original filename
  file_size_bytes   INT,
  mime_type         TEXT,
  -- Validity
  issued_at         DATE,                                    -- date on the document itself
  expires_at        DATE,                                    -- computed or manually set
  -- Review workflow
  status            document_status NOT NULL DEFAULT 'pending_review',
  reviewed_by       UUID            REFERENCES users(id) ON DELETE SET NULL,
  reviewed_at       TIMESTAMPTZ,
  rejection_reason  TEXT,
  -- AI OCR metadata (populated if AI processing was triggered)
  ai_extracted_data JSONB,                                   -- raw data extracted by AI
  ai_confidence     NUMERIC(4,3),                           -- 0.000 to 1.000
  created_at        TIMESTAMPTZ     NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ     NOT NULL DEFAULT now(),
  deleted_at        TIMESTAMPTZ,
  -- Exactly one owner: either user_id or contractor_id, never both, never neither
  CONSTRAINT chk_document_owner CHECK (
    (user_id IS NOT NULL AND contractor_id IS NULL) OR
    (user_id IS NULL AND contractor_id IS NOT NULL)
  )
);

CREATE INDEX idx_documents_user_id           ON documents(user_id);
CREATE INDEX idx_documents_contractor_id     ON documents(contractor_id);
CREATE INDEX idx_documents_document_type_id  ON documents(document_type_id);
CREATE INDEX idx_documents_status            ON documents(status);
CREATE INDEX idx_documents_expires_at        ON documents(expires_at);  -- for expiry queries

CREATE TRIGGER trg_documents_updated_at
  BEFORE UPDATE ON documents
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================
-- 12. ACCESS LOGS
-- Every check-in and check-out event at a gate.
-- Includes a snapshot of the worker's document status at
-- the moment of entry (for audit trail).
-- ============================================================
CREATE TYPE access_result AS ENUM (
  'granted',             -- all docs valid, access allowed
  'granted_with_warning',-- docs have issues but plant policy allows entry
  'denied'               -- blocked due to expired/missing docs
);

CREATE TABLE access_logs (
  id                UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id         UUID          NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
  plant_id          UUID          NOT NULL REFERENCES plants(id) ON DELETE RESTRICT,
  access_point_id   UUID          NOT NULL REFERENCES access_points(id) ON DELETE RESTRICT,
  -- Who entered
  user_id           UUID          NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  contractor_id     UUID          REFERENCES contractors(id) ON DELETE RESTRICT,
  -- Who processed the check-in (guard user)
  processed_by      UUID          REFERENCES users(id) ON DELETE SET NULL,
  -- Timestamps
  checked_in_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),
  checked_out_at    TIMESTAMPTZ,                             -- NULL until check-out
  -- Result
  result            access_result NOT NULL,
  denial_reason     TEXT,                                    -- e.g. "DC-3 expired on 2026-09-01"
  -- Snapshot: which documents were valid/invalid at check-in time
  -- Stored as JSONB for immutability (audit trail must not change if docs change later)
  document_snapshot JSONB         NOT NULL DEFAULT '[]',
  -- Example snapshot:
  -- [{"document_type_code": "DC3", "status": "approved", "expires_at": "2027-01-15"},
  --  {"document_type_code": "REPSE", "status": "expired", "expires_at": "2026-09-01"}]
  notes             TEXT,
  created_at        TIMESTAMPTZ   NOT NULL DEFAULT now()
  -- No updated_at: access logs are immutable after creation
);

CREATE INDEX idx_access_logs_tenant_id       ON access_logs(tenant_id);
CREATE INDEX idx_access_logs_plant_id        ON access_logs(plant_id);
CREATE INDEX idx_access_logs_user_id         ON access_logs(user_id);
CREATE INDEX idx_access_logs_contractor_id   ON access_logs(contractor_id);
CREATE INDEX idx_access_logs_checked_in_at   ON access_logs(checked_in_at);
-- Partial index to quickly find workers currently inside (no check-out yet)
CREATE INDEX idx_access_logs_active_entries  ON access_logs(plant_id, user_id)
  WHERE checked_out_at IS NULL;

-- ============================================================
-- 13. NOTIFICATIONS
-- In-app persistent notifications + email trigger flag.
-- ============================================================
CREATE TYPE notification_type AS ENUM (
  'document_expiring_soon',   -- doc expires in N days
  'document_expired',         -- doc is now expired
  'document_uploaded',        -- contractor uploaded a new doc (for EHS manager)
  'document_approved',        -- EHS manager approved a doc (for contractor)
  'document_rejected',        -- EHS manager rejected a doc (for contractor)
  'access_denied',            -- worker was blocked at gate
  'contractor_invited',       -- plant invited a contractor to their catalog
  'contractor_activated'      -- plant activated a contractor
);

CREATE TABLE notifications (
  id              UUID                PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id       UUID                REFERENCES tenants(id) ON DELETE RESTRICT,
  recipient_id    UUID                NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  type            notification_type   NOT NULL,
  title           TEXT                NOT NULL,
  body            TEXT                NOT NULL,
  -- Optional deep-link to the related entity
  resource_type   TEXT,               -- 'document' | 'access_log' | 'contractor' | ...
  resource_id     UUID,
  -- Read state
  read_at         TIMESTAMPTZ,        -- NULL = unread
  -- Email
  email_sent_at   TIMESTAMPTZ,        -- NULL = email not sent yet
  created_at      TIMESTAMPTZ         NOT NULL DEFAULT now()
);

CREATE INDEX idx_notifications_recipient_id ON notifications(recipient_id);
CREATE INDEX idx_notifications_tenant_id    ON notifications(tenant_id);
CREATE INDEX idx_notifications_read_at      ON notifications(recipient_id) WHERE read_at IS NULL;
CREATE INDEX idx_notifications_created_at   ON notifications(created_at);
