INSERT INTO users (
    id,
    project_id,
    email,
    role,
    username,
    password_hash,
    created_at,
    updated_at
)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000010',
    'ralvescosta@local.dev',
    'write',
    'ralvescosta',
    '$2a$10$AjPfEDzY4NI/NhnKuN9UEu6X6J6zRUNO2e79dfh3E1VbdkIpYHzcy',
    NOW(),
    NOW()
)
ON CONFLICT (id) DO UPDATE
SET project_id = EXCLUDED.project_id,
    email = EXCLUDED.email,
    role = EXCLUDED.role,
    username = EXCLUDED.username,
    password_hash = EXCLUDED.password_hash,
    updated_at = NOW();
