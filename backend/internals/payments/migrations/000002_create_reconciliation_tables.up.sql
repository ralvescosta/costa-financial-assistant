-- 000002_create_reconciliation_tables.up.sql
-- Creates reconciliation_links table.
-- Note: reconciliation_status enum and the reconciliation_status column on transaction_lines
-- are already created by backend/internals/files/migrations/000002_create_analysis_tables.up.sql.

-- Link type enum (idempotent)
DO $$ BEGIN
  CREATE TYPE reconciliation_link_type_enum AS ENUM ('auto', 'manual');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- Reconciliation links table
CREATE TABLE IF NOT EXISTS reconciliation_links (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  transaction_line_id UUID NOT NULL REFERENCES transaction_lines(id) ON DELETE CASCADE,
  bill_record_id      UUID NOT NULL REFERENCES bill_records(id) ON DELETE CASCADE,
  link_type           reconciliation_link_type_enum NOT NULL,
  linked_by           UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT uq_reconciliation_link UNIQUE (transaction_line_id, bill_record_id)
);

CREATE INDEX IF NOT EXISTS idx_reconciliation_links_project_id ON reconciliation_links (project_id);
CREATE INDEX IF NOT EXISTS idx_reconciliation_links_transaction_line_id ON reconciliation_links (transaction_line_id);
CREATE INDEX IF NOT EXISTS idx_reconciliation_links_bill_record_id ON reconciliation_links (bill_record_id);
