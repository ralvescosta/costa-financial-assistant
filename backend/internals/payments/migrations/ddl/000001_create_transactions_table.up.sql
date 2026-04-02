CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL,
    amount_cents BIGINT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
