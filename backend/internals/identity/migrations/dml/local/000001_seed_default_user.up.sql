INSERT INTO users (id, project_id, email, role, created_at, updated_at)
VALUES ('00000000-0000-0000-0000-000000000011', '11111111-1111-1111-1111-111111111111', 'local@example.com', 'write', NOW(), NOW())
ON CONFLICT DO NOTHING;
