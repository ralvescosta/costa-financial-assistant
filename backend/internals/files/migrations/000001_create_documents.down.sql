-- 000001_create_documents.down.sql
-- Reverses the documents table creation.

DROP TABLE IF EXISTS documents;
DROP TYPE IF EXISTS analysis_status;
DROP TYPE IF EXISTS document_kind;
