CREATE TYPE document_kind AS ENUM ('unspecified', 'bill', 'statement');
CREATE TYPE analysis_status AS ENUM ('pending', 'processing', 'analysed', 'analysis_failed');

CREATE TABLE IF NOT EXISTS documents (
    id                UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id        UUID          NOT NULL,
    uploaded_by       UUID          NOT NULL,
    kind              document_kind NOT NULL DEFAULT 'unspecified',
    storage_provider  TEXT          NOT NULL DEFAULT '',
    storage_key       TEXT          NOT NULL DEFAULT '',
    file_name         TEXT          NOT NULL,
    file_hash         TEXT          NOT NULL,
    analysis_status   analysis_status NOT NULL DEFAULT 'pending',
    failure_reason    TEXT,
    uploaded_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_documents_project_hash UNIQUE (project_id, file_hash)
);

CREATE INDEX IF NOT EXISTS idx_documents_project_id ON documents (project_id);
CREATE INDEX IF NOT EXISTS idx_documents_analysis_status ON documents (analysis_status);
