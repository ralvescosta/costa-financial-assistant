CREATE TABLE IF NOT EXISTS reconciliations (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL,
    transaction_id UUID NOT NULL,
    state TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_reconciliations_project_id ON reconciliations (project_id);
