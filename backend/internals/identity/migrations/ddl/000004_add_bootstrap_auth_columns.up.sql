ALTER TABLE users
    ADD COLUMN IF NOT EXISTS username TEXT,
    ADD COLUMN IF NOT EXISTS password_hash TEXT;

UPDATE users
SET username = COALESCE(NULLIF(username, ''), SPLIT_PART(email, '@', 1))
WHERE email IS NOT NULL
  AND (username IS NULL OR username = '');

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_lower_username
    ON users ((LOWER(username)))
    WHERE username IS NOT NULL AND username <> '';
