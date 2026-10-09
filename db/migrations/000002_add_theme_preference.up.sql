-- Migration: 000002_add_theme_preference.up.sql
-- Description: Adds theme_preference column to users table with check constraint.

ALTER TABLE users 
ADD COLUMN IF NOT EXISTS theme_preference TEXT DEFAULT 'system' 
CHECK (theme_preference IN ('light', 'dark', 'system'));
