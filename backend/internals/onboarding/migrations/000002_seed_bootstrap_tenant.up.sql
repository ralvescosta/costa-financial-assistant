-- 000002_seed_bootstrap_tenant.up.sql
-- Inserts the Phase-1 bootstrap user, project, and project_member.
-- All IDs are deterministic UUIDs so dependent services can reference them.

BEGIN;

-- Bootstrap user (identity-grpc issues JWT for this subject)
INSERT INTO users (id, email, display_name, status, created_at, updated_at)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'bootstrap@costa.local',
    'Bootstrap User',
    'active',
    now(),
    now()
)
ON CONFLICT (id) DO NOTHING;

-- Bootstrap project
INSERT INTO projects (id, owner_id, name, type, created_at, updated_at)
VALUES (
    '00000000-0000-0000-0000-000000000010',
    '00000000-0000-0000-0000-000000000001',
    'My Finances',
    'personal',
    now(),
    now()
)
ON CONFLICT (id) DO NOTHING;

-- Bootstrap member (owner gets write role)
INSERT INTO project_members (id, project_id, user_id, role, invited_by, created_at, updated_at)
VALUES (
    '00000000-0000-0000-0000-000000000100',
    '00000000-0000-0000-0000-000000000010',
    '00000000-0000-0000-0000-000000000001',
    'write',
    NULL,
    now(),
    now()
)
ON CONFLICT (project_id, user_id) DO NOTHING;

COMMIT;
