INSERT INTO projects (id, owner_user_id, name, created_at, updated_at)
VALUES ('00000000-0000-0000-0000-000000000021', '00000000-0000-0000-0000-000000000011', 'Local Project', NOW(), NOW())
ON CONFLICT DO NOTHING;
