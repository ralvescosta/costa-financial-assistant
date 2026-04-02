CREATE INDEX IF NOT EXISTS idx_bills_project_id ON bills (project_id);
CREATE INDEX IF NOT EXISTS idx_bills_status ON bills (status);
CREATE INDEX IF NOT EXISTS idx_bills_created_at ON bills (created_at DESC);
