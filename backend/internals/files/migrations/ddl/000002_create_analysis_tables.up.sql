CREATE TYPE job_type AS ENUM ('extract_bill', 'extract_statement', 'reconcile_statement');
CREATE TYPE job_status AS ENUM ('queued', 'running', 'succeeded', 'failed', 'dead_lettered');
CREATE TYPE payment_status AS ENUM ('unpaid', 'paid', 'overdue');
CREATE TYPE transaction_direction AS ENUM ('credit', 'debit');
CREATE TYPE reconciliation_status AS ENUM ('unmatched', 'matched_auto', 'matched_manual', 'ambiguous');

CREATE TABLE IF NOT EXISTS analysis_jobs (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id    UUID        NOT NULL,
    document_id   UUID        NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    job_type      job_type    NOT NULL,
    status        job_status  NOT NULL DEFAULT 'queued',
    attempt_count INT         NOT NULL DEFAULT 0,
    last_error    TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_analysis_jobs_project_id ON analysis_jobs (project_id);
CREATE INDEX IF NOT EXISTS idx_analysis_jobs_document_id ON analysis_jobs (document_id);
CREATE INDEX IF NOT EXISTS idx_analysis_jobs_status ON analysis_jobs (status);

CREATE TABLE IF NOT EXISTS bill_records (
    id               UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id       UUID           NOT NULL,
    document_id      UUID           NOT NULL UNIQUE REFERENCES documents(id) ON DELETE CASCADE,
    due_date         DATE           NOT NULL,
    amount_due       NUMERIC(14,2)  NOT NULL,
    pix_payload      TEXT,
    pix_qr_image_ref TEXT,
    barcode          TEXT,
    payment_status   payment_status NOT NULL DEFAULT 'unpaid',
    paid_at          TIMESTAMPTZ,
    marked_paid_by   UUID,
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bill_records_project_id ON bill_records (project_id);
CREATE INDEX IF NOT EXISTS idx_bill_records_document_id ON bill_records (document_id);
CREATE INDEX IF NOT EXISTS idx_bill_records_due_date ON bill_records (due_date);
CREATE INDEX IF NOT EXISTS idx_bill_records_payment_status ON bill_records (payment_status);

CREATE TABLE IF NOT EXISTS statement_records (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID        NOT NULL,
    document_id     UUID        NOT NULL UNIQUE REFERENCES documents(id) ON DELETE CASCADE,
    bank_account_id UUID,
    period_start    DATE        NOT NULL,
    period_end      DATE        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_statement_records_project_id ON statement_records (project_id);
CREATE INDEX IF NOT EXISTS idx_statement_records_document_id ON statement_records (document_id);

CREATE TABLE IF NOT EXISTS transaction_lines (
    id                    UUID                  PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id            UUID                  NOT NULL,
    statement_id          UUID                  NOT NULL REFERENCES statement_records(id) ON DELETE CASCADE,
    transaction_date      DATE                  NOT NULL,
    description           TEXT                  NOT NULL,
    amount                NUMERIC(14,2)         NOT NULL,
    direction             transaction_direction NOT NULL,
    reconciliation_status reconciliation_status NOT NULL DEFAULT 'unmatched',
    created_at            TIMESTAMPTZ           NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transaction_lines_project_id ON transaction_lines (project_id);
CREATE INDEX IF NOT EXISTS idx_transaction_lines_statement_id ON transaction_lines (statement_id);
CREATE INDEX IF NOT EXISTS idx_transaction_lines_reconciliation_status ON transaction_lines (reconciliation_status);