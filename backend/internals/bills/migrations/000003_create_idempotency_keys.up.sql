-- 000003_create_idempotency_keys.up.sql
-- Creates the shared idempotency_keys table for protecting mutating bill and
-- reconciliation workflows from duplicate processing under at-least-once delivery.

CREATE TABLE IF NOT EXISTS idempotency_keys (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id       UUID        NOT NULL,
    operation        TEXT        NOT NULL,
    idempotency_key  TEXT        NOT NULL,
    response_hash    TEXT,
    expires_at       TIMESTAMPTZ NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_idempotency_project_operation_key UNIQUE (project_id, operation, idempotency_key)
);

CREATE INDEX IF NOT EXISTS idx_idempotency_keys_project_id ON idempotency_keys (project_id);
CREATE INDEX IF NOT EXISTS idx_idempotency_keys_expires_at ON idempotency_keys (expires_at);
