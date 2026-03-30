# Phase 1 Data Model - Financial Bill Organizer

## Core tenancy and identity entities

## User
- Purpose: Account identity that owns projects and/or collaborates on projects.
- Fields:
  - `id` (UUID, PK)
  - `email` (text, unique)
  - `display_name` (text)
  - `status` (enum: `active`, `inactive`)
  - `created_at`, `updated_at` (timestamp)

## Project
- Purpose: Tenant boundary for all financial data.
- Fields:
  - `id` (UUID, PK)
  - `owner_id` (UUID, FK -> `users.id`)
  - `name` (text)
  - `type` (enum: `personal`, `conjugal`, `shared`)
  - `created_at`, `updated_at` (timestamp)
- Relationships:
  - one-to-many with `project_members`
  - one-to-many with all domain records via `project_id`

## ProjectMember
- Purpose: Collaboration and role-based permissions within a project.
- Fields:
  - `id` (UUID, PK)
  - `project_id` (UUID, FK -> `projects.id`)
  - `user_id` (UUID, FK -> `users.id`)
  - `role` (enum: `read_only`, `update`, `write`)
  - `invited_by` (UUID, FK -> `users.id`, nullable)
  - `created_at`, `updated_at` (timestamp)
- Constraints:
  - unique (`project_id`, `user_id`)

## Domain entities

## Document
- Purpose: Uploaded PDF metadata and processing lifecycle.
- Fields:
  - `id` (UUID, PK)
  - `project_id` (UUID, FK -> `projects.id`, indexed)
  - `uploaded_by` (UUID, FK -> `users.id`)
  - `kind` (enum: `bill`, `statement`)
  - `storage_provider` (text)
  - `storage_key` (text)
  - `file_name` (text)
  - `file_hash` (text)
  - `analysis_status` (enum: `pending`, `processing`, `analysed`, `analysis_failed`)
  - `failure_reason` (text, nullable)
  - `uploaded_at`, `updated_at` (timestamp)
- Constraints:
  - unique (`project_id`, `file_hash`) for project-scoped duplicate detection.

## BillType
- Purpose: Bill categorization labels per project.
- Fields:
  - `id` (UUID, PK)
  - `project_id` (UUID, FK -> `projects.id`, indexed)
  - `name` (text)
  - `created_by` (UUID, FK -> `users.id`)
  - `created_at`, `updated_at` (timestamp)
- Constraints:
  - unique (`project_id`, `name`)

## BankAccount
- Purpose: Non-sensitive account label mapping for statement attribution.
- Fields:
  - `id` (UUID, PK)
  - `project_id` (UUID, FK -> `projects.id`, indexed)
  - `label` (text)
  - `created_by` (UUID, FK -> `users.id`)
  - `created_at`, `updated_at` (timestamp)
- Constraints:
  - unique (`project_id`, `label`)

## BillRecord
- Purpose: Structured bill extraction and payment lifecycle.
- Fields:
  - `id` (UUID, PK)
  - `project_id` (UUID, FK -> `projects.id`, indexed)
  - `document_id` (UUID, FK -> `documents.id`, unique)
  - `bill_type_id` (UUID, FK -> `bill_types.id`)
  - `due_date` (date)
  - `amount_due` (numeric(14,2))
  - `pix_payload` (text, nullable)
  - `pix_qr_image_ref` (text, nullable)
  - `barcode` (text, nullable)
  - `payment_status` (enum: `unpaid`, `paid`, `overdue`)
  - `paid_at` (timestamp, nullable)
  - `marked_paid_by` (UUID, FK -> `users.id`, nullable)
  - `created_at`, `updated_at` (timestamp)

## StatementRecord
- Purpose: Statement extraction root entity.
- Fields:
  - `id` (UUID, PK)
  - `project_id` (UUID, FK -> `projects.id`, indexed)
  - `document_id` (UUID, FK -> `documents.id`, unique)
  - `bank_account_id` (UUID, FK -> `bank_accounts.id`)
  - `period_start`, `period_end` (date)
  - `created_at`, `updated_at` (timestamp)

## TransactionLine
- Purpose: Individual statement entries used for reconciliation.
- Fields:
  - `id` (UUID, PK)
  - `project_id` (UUID, FK -> `projects.id`, indexed)
  - `statement_id` (UUID, FK -> `statement_records.id`, indexed)
  - `transaction_date` (date)
  - `description` (text)
  - `amount` (numeric(14,2))
  - `direction` (enum: `credit`, `debit`)
  - `reconciliation_status` (enum: `unmatched`, `matched_auto`, `matched_manual`, `ambiguous`)
  - `created_at`, `updated_at` (timestamp)

## ReconciliationLink
- Purpose: Materialized relationship between transactions and bills.
- Fields:
  - `id` (UUID, PK)
  - `project_id` (UUID, FK -> `projects.id`, indexed)
  - `transaction_line_id` (UUID, FK -> `transaction_lines.id`)
  - `bill_record_id` (UUID, FK -> `bill_records.id`)
  - `link_type` (enum: `auto`, `manual`)
  - `linked_by` (UUID, FK -> `users.id`, nullable for auto)
  - `created_at` (timestamp)
- Constraints:
  - unique (`transaction_line_id`, `bill_record_id`)

## PaymentCyclePreference
- Purpose: Preferred payment-day setting per project.
- Fields:
  - `id` (UUID, PK)
  - `project_id` (UUID, FK -> `projects.id`, unique)
  - `preferred_day_of_month` (smallint, check 1..28)
  - `updated_by` (UUID, FK -> `users.id`)
  - `updated_at` (timestamp)

## Operational entities

## AnalysisJob
- Purpose: Asynchronous processing state and retries.
- Fields:
  - `id` (UUID, PK)
  - `project_id` (UUID, FK -> `projects.id`, indexed)
  - `document_id` (UUID, FK -> `documents.id`)
  - `job_type` (enum: `extract_bill`, `extract_statement`, `reconcile_statement`)
  - `status` (enum: `queued`, `running`, `succeeded`, `failed`, `dead_lettered`)
  - `attempt_count` (int)
  - `last_error` (text, nullable)
  - `created_at`, `updated_at` (timestamp)

## IdempotencyKey
- Purpose: Idempotent protection for mutating operations.
- Fields:
  - `id` (UUID, PK)
  - `project_id` (UUID, FK -> `projects.id`, indexed)
  - `operation` (text)
  - `idempotency_key` (text)
  - `response_hash` (text, nullable)
  - `expires_at` (timestamp)
  - `created_at` (timestamp)
- Constraints:
  - unique (`project_id`, `operation`, `idempotency_key`)

## State transitions

## Document.analysis_status
- `pending` -> `processing` -> `analysed`
- `pending` -> `processing` -> `analysis_failed`
- Retriable failure: `analysis_failed` -> `processing`

## BillRecord.payment_status
- `unpaid` -> `paid`
- `unpaid` -> `overdue` (time-driven)
- `overdue` -> `paid`

## TransactionLine.reconciliation_status
- `unmatched` -> `matched_auto`
- `unmatched` -> `ambiguous`
- `ambiguous` -> `matched_manual`
- `unmatched` -> `matched_manual`

## Indexing and query-critical keys
- `(project_id, analysis_status, uploaded_at DESC)` on `documents`
- `(project_id, due_date, payment_status)` on `bill_records`
- `(project_id, bank_account_id, period_start, period_end)` on `statement_records`
- `(project_id, statement_id, transaction_date)` on `transaction_lines`
- `(project_id, role)` on `project_members`

All domain indexes above must be created through migrations with `IF NOT EXISTS`.
