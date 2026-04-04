INSERT INTO projects (
    id,
    owner_user_id,
    owner_id,
    name,
    type,
    created_at,
    updated_at
)
VALUES (
    '00000000-0000-0000-0000-000000000010',
    '00000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000001',
    'Costa Financial Assistant',
    'personal',
    NOW(),
    NOW()
)
ON CONFLICT (id) DO UPDATE
SET owner_user_id = EXCLUDED.owner_user_id,
    owner_id = EXCLUDED.owner_id,
    name = EXCLUDED.name,
    type = EXCLUDED.type,
    updated_at = NOW();

INSERT INTO project_members (
    project_id,
    user_id,
    role,
    invited_by,
    created_at,
    updated_at
)
VALUES (
    '00000000-0000-0000-0000-000000000010',
    '00000000-0000-0000-0000-000000000001',
    'write',
    '00000000-0000-0000-0000-000000000001',
    NOW(),
    NOW()
)
ON CONFLICT (project_id, user_id) DO UPDATE
SET role = EXCLUDED.role,
    invited_by = EXCLUDED.invited_by,
    updated_at = NOW();
