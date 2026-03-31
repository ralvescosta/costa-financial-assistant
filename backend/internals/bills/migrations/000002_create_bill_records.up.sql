-- 000002_create_bill_records.up.sql
-- Adds bill_type_id to bill_records linking categorisation labels (bill_types)
-- with the extracted bill data created by the files service.

ALTER TABLE bill_records
    ADD COLUMN IF NOT EXISTS bill_type_id UUID REFERENCES bill_types(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_bill_records_bill_type_id ON bill_records (bill_type_id);
