-- 000001_create_identity_and_projects.up.sql
-- Multi-tenant bootstrap: users, projects, and project_members tables.
-- Applies to: onboarding service database.

BEGIN;

-- ─── Extensions ──────────────────────────────────────────────────────────────
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ─── Enums ───────────────────────────────────────────────────────────────────
CREATE TYPE user_status AS ENUM ('active', 'inactive');
CREATE TYPE project_type AS ENUM ('personal', 'conjugal', 'shared');
CREATE TYPE project_member_role AS ENUM ('read_only', 'update', 'write');

-- ─── users ───────────────────────────────────────────────────────────────────
CREATE TABLE users (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email        TEXT        NOT NULL,
    display_name TEXT        NOT NULL,
    status       user_status NOT NULL DEFAULT 'active',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_users_email UNIQUE (email)
);

-- ─── projects ────────────────────────────────────────────────────────────────
CREATE TABLE projects (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id   UUID         NOT NULL REFERENCES users (id),
    name       TEXT         NOT NULL,
    type       project_type NOT NULL DEFAULT 'personal',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_projects_owner_id ON projects (owner_id);

-- ─── project_members ─────────────────────────────────────────────────────────
CREATE TABLE project_members (
    id           UUID                 PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id   UUID                 NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    user_id      UUID                 NOT NULL REFERENCES users (id),
    role         project_member_role  NOT NULL DEFAULT 'read_only',
    invited_by   UUID                 REFERENCES users (id),
    created_at   TIMESTAMPTZ          NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ          NOT NULL DEFAULT now(),
    CONSTRAINT uq_project_members_project_user UNIQUE (project_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_project_members_project_id ON project_members (project_id);
CREATE INDEX IF NOT EXISTS idx_project_members_user_id    ON project_members (user_id);
CREATE INDEX IF NOT EXISTS idx_project_members_project_role ON project_members (project_id, role);

COMMIT;
