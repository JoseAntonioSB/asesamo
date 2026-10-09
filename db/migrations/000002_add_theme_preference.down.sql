-- Migration: 000002_add_theme_preference.down.sql
-- Description: Drops theme_preference column from users table.

ALTER TABLE users DROP COLUMN IF EXISTS theme_preference;
