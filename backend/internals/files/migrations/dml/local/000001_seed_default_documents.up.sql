INSERT INTO documents (id, project_id, uploaded_by, kind, storage_provider, storage_key, file_name, file_hash, analysis_status, uploaded_at, updated_at)
VALUES ('00000000-0000-0000-0000-000000000001', '11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222', 'unspecified', 'local', 'seed/sample.pdf', 'sample.pdf', 'hash-local', 'pending', NOW(), NOW())
ON CONFLICT DO NOTHING;
