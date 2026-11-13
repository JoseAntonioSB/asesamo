-- 000006_create_epp_module.down.sql

DROP TABLE IF EXISTS epp_deliveries CASCADE;
DROP TABLE IF EXISTS epp_catalog CASCADE;
DROP TYPE IF EXISTS epp_delivery_status;
DROP TYPE IF EXISTS epp_category;
