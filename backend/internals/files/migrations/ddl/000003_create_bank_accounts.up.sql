CREATE TABLE IF NOT EXISTS bank_accounts (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID        NOT NULL,
    label      TEXT        NOT NULL,
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_bank_accounts_project_label UNIQUE (project_id, label)
);

CREATE INDEX IF NOT EXISTS idx_bank_accounts_project_id ON bank_accounts (project_id);