INSERT INTO transactions (id, project_id, amount_cents, status, created_at, updated_at)
VALUES ('00000000-0000-0000-0000-000000000033', '11111111-1111-1111-1111-111111111111', 3000, 'pending', NOW(), NOW())
ON CONFLICT DO NOTHING;
