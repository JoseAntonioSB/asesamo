-- 000007_create_emergency_evacuation_module.down.sql

DROP TABLE IF EXISTS evacuation_roll_call CASCADE;
DROP TABLE IF EXISTS emergency_evacuations CASCADE;
DROP TABLE IF EXISTS muster_points CASCADE;
DROP TYPE IF EXISTS roll_call_status;
DROP TYPE IF EXISTS evacuation_status;
DROP TYPE IF EXISTS evacuation_type;
