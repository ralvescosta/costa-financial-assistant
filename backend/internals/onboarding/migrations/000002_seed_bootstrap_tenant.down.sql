-- 000002_seed_bootstrap_tenant.down.sql
-- Removes the Phase-1 bootstrap seed data.

BEGIN;

DELETE FROM project_members
WHERE id = '00000000-0000-0000-0000-000000000100';

DELETE FROM projects
WHERE id = '00000000-0000-0000-0000-000000000010';

DELETE FROM users
WHERE id = '00000000-0000-0000-0000-000000000001';

COMMIT;
