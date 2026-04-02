INSERT INTO bills (id, project_id, bill_type, status, amount_cents, created_at, updated_at)
VALUES ('00000000-0000-0000-0000-000000000001', '11111111-1111-1111-1111-111111111111', 'utility', 'pending', 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
