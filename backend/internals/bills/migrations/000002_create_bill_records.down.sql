-- 000002_create_bill_records.down.sql
-- Removes bill_type_id from bill_records.

DROP INDEX IF EXISTS idx_bill_records_bill_type_id;

ALTER TABLE bill_records
    DROP COLUMN IF EXISTS bill_type_id;
