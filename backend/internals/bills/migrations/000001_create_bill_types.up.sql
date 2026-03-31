-- 000001_create_bill_types.up.sql
-- Creates the bill_types table for project-scoped bill categorization labels.

CREATE TABLE IF NOT EXISTS bill_types (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID        NOT NULL,
    name       TEXT        NOT NULL,
    created_by UUID        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_bill_types_project_name UNIQUE (project_id, name)
);

CREATE INDEX IF NOT EXISTS idx_bill_types_project_id ON bill_types (project_id);
