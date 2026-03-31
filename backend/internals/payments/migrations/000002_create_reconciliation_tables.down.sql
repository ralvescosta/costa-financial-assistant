-- 000002_create_reconciliation_tables.down.sql
-- Reverses reconciliation tables migration.

DROP TABLE IF EXISTS reconciliation_links;

DROP TYPE IF EXISTS reconciliation_link_type_enum;
