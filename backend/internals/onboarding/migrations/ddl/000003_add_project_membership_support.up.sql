CREATE EXTENSION IF NOT EXISTS pgcrypto;

DO $$ BEGIN
    CREATE TYPE project_type AS ENUM ('personal', 'conjugal', 'shared');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE project_member_role AS ENUM ('read_only', 'update', 'write');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

ALTER TABLE projects
    ADD COLUMN IF NOT EXISTS owner_id UUID,
    ADD COLUMN IF NOT EXISTS type project_type NOT NULL DEFAULT 'personal';

ALTER TABLE projects ALTER COLUMN id SET DEFAULT gen_random_uuid();
ALTER TABLE projects ALTER COLUMN owner_user_id DROP NOT NULL;

UPDATE projects
SET owner_id = COALESCE(owner_id, owner_user_id)
WHERE owner_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_projects_owner ON projects (owner_id);

CREATE TABLE IF NOT EXISTS project_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    role project_member_role NOT NULL DEFAULT 'read_only',
    invited_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_project_members_project_user UNIQUE (project_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_project_members_project_id ON project_members (project_id);
CREATE INDEX IF NOT EXISTS idx_project_members_user_id ON project_members (user_id);
