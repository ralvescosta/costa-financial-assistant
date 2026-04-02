INSERT INTO documents (id, project_id, filename, file_hash, analysis_status, created_at, updated_at)
VALUES ('00000000-0000-0000-0000-000000000001', '11111111-1111-1111-1111-111111111111', 'sample.pdf', 'hash-local', 'pending', NOW(), NOW())
ON CONFLICT DO NOTHING;
