-- 000004_create_safety_inspections.down.sql

DROP TABLE IF EXISTS inspection_action_plans CASCADE;
DROP TABLE IF EXISTS safety_inspection_items CASCADE;
DROP TABLE IF EXISTS safety_inspections CASCADE;
DROP TABLE IF EXISTS inspection_template_items CASCADE;
DROP TABLE IF EXISTS inspection_templates CASCADE;

DROP TYPE IF EXISTS capa_status;
DROP TYPE IF EXISTS finding_severity;
DROP TYPE IF EXISTS item_compliance_status;
DROP TYPE IF EXISTS inspection_status;
DROP TYPE IF EXISTS inspection_type;
