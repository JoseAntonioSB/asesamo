-- ============================================================
-- ASESAMO — Rollback Migration: 000003_create_work_permits.down.sql
-- Description: Drops PTAR tables, triggers, and types in reverse order
-- ============================================================

-- 1. DROP TRIGGERS
DROP TRIGGER IF EXISTS trg_permit_checklists_updated_at ON permit_checklists;
DROP TRIGGER IF EXISTS trg_work_permits_updated_at ON work_permits;

-- 2. DROP TABLES IN REVERSE DEPENDENCY ORDER
DROP TABLE IF EXISTS permit_signatures;
DROP TABLE IF EXISTS permit_checklists;
DROP TABLE IF EXISTS permit_workers;
DROP TABLE IF EXISTS work_permits;

-- 3. DROP DOMAIN & ENUM TYPES
DROP DOMAIN IF EXISTS permit_signer_role;
DROP TYPE IF EXISTS permit_signature_role;
DROP TYPE IF EXISTS permit_status;
DROP TYPE IF EXISTS permit_type;

-- 4. CLEANUP SEEDED SPECIALIZED DOCUMENT TYPES
DELETE FROM document_types 
WHERE tenant_id IS NULL 
  AND code IN (
    'DC3_ALTURAS', 
    'DC3_ESPACIOS_CONFINADOS', 
    'DC3_CORTE_SOLDADURA', 
    'DC3_LOTO',
    'DC3_CONFINED',
    'DC3_HOT_WORK',
    'DC3_ELECTRICAL'
  )
  AND is_system = TRUE
  AND NOT EXISTS (
    SELECT 1 FROM documents d WHERE d.document_type_id = document_types.id
  );
