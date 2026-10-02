-- ============================================================
-- ASESAMO — Rollback Migration
-- Migration: 000001_init.down.sql
-- Description: Drop all MVP tables in reverse dependency order
-- ============================================================

DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS access_logs;
DROP TABLE IF EXISTS documents;
DROP TABLE IF EXISTS plant_document_requirements;
DROP TABLE IF EXISTS document_types;
DROP TABLE IF EXISTS contractor_workers;
DROP TABLE IF EXISTS plant_contractors;
DROP TABLE IF EXISTS contractors;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS access_points;
DROP TABLE IF EXISTS plants;
DROP TABLE IF EXISTS tenants;

DROP TYPE IF EXISTS notification_type;
DROP TYPE IF EXISTS access_result;
DROP TYPE IF EXISTS document_status;
DROP TYPE IF EXISTS plant_contractor_status;
DROP TYPE IF EXISTS contractor_type;
DROP TYPE IF EXISTS user_role;

DROP FUNCTION IF EXISTS set_updated_at();

DROP EXTENSION IF EXISTS "citext";
DROP EXTENSION IF EXISTS "pgcrypto";
