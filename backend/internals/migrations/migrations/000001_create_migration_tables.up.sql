CREATE TABLE IF NOT EXISTS migrations_ddl (
    version BIGINT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    execution_time_ms BIGINT NOT NULL,
    success BOOLEAN NOT NULL DEFAULT TRUE,
    error_message TEXT,
    executed_by TEXT,
    checksum TEXT
);

CREATE INDEX IF NOT EXISTS idx_migrations_ddl_executed_at ON migrations_ddl (executed_at DESC);

CREATE TABLE IF NOT EXISTS migrations_dml (
    version BIGINT NOT NULL,
    name TEXT NOT NULL,
    environment TEXT NOT NULL,
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    execution_time_ms BIGINT NOT NULL,
    success BOOLEAN NOT NULL DEFAULT TRUE,
    error_message TEXT,
    executed_by TEXT,
    checksum TEXT,
    PRIMARY KEY (version, environment)
);

CREATE INDEX IF NOT EXISTS idx_migrations_dml_env_executed ON migrations_dml (environment, executed_at DESC);
