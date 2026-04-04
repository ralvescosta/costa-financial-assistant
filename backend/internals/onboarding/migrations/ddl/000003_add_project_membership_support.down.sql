DROP INDEX IF EXISTS idx_project_members_user_id;
DROP INDEX IF EXISTS idx_project_members_project_id;
DROP TABLE IF EXISTS project_members;

DROP INDEX IF EXISTS idx_projects_owner;

ALTER TABLE projects
    DROP COLUMN IF EXISTS type,
    DROP COLUMN IF EXISTS owner_id;

DROP TYPE IF EXISTS project_member_role;
DROP TYPE IF EXISTS project_type;
