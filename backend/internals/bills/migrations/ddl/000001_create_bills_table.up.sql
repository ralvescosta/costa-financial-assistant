CREATE TABLE IF NOT EXISTS bills (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL,
    bill_type TEXT NOT NULL,
    status TEXT NOT NULL,
    amount_cents BIGINT NOT NULL DEFAULT 0,
    due_date DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
