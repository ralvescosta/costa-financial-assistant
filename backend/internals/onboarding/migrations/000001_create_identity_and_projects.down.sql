-- 000001_create_identity_and_projects.down.sql
-- Reverses the multi-tenant bootstrap schema.

BEGIN;

DROP TABLE IF EXISTS project_members;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS project_member_role;
DROP TYPE IF EXISTS project_type;
DROP TYPE IF EXISTS user_status;

COMMIT;
