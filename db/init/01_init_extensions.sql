-- ============================================================
-- ASESAMO — PostgreSQL 16 Initial Extensions Bootstrap
-- Mounted to /docker-entrypoint-initdb.d/ to ensure extensions
-- exist immediately upon database cluster creation.
-- ============================================================

CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "citext";
