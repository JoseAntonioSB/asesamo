-- 000005_create_incidents_module.down.sql

DROP TABLE IF EXISTS safety_incidents CASCADE;
DROP TYPE IF EXISTS incident_status;
DROP TYPE IF EXISTS incident_severity;
DROP TYPE IF EXISTS incident_type;
