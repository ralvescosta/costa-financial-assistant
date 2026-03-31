-- 000002_create_analysis_tables.down.sql
-- Drops analysis_jobs, bill_records, statement_records, and transaction_lines tables.

DROP TABLE IF EXISTS transaction_lines;
DROP TABLE IF EXISTS statement_records;
DROP TABLE IF EXISTS bill_records;
DROP TABLE IF EXISTS analysis_jobs;

DROP TYPE IF EXISTS reconciliation_status;
DROP TYPE IF EXISTS transaction_direction;
DROP TYPE IF EXISTS payment_status;
DROP TYPE IF EXISTS job_status;
DROP TYPE IF EXISTS job_type;
