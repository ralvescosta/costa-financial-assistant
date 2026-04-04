# Data Model: Stabilize Broken BFF Page Flows

## Overview

This feature does not introduce a brand-new domain. It stabilizes the existing request and data relationships needed for the default user to load the `documents`, `payments`, `analyses`, and `settings` screens successfully through the BFF.

## Entities

### 1. AuthenticatedDefaultUser

| Field | Description | Notes |
|---|---|---|
| `user_id` | Canonical authenticated user identifier | Must propagate through auth/session handling and any downstream session envelope already used by the BFF |
| `username` | Default dev/test user name | Used for local/demo validation |
| `email` | Default dev/test user email | Must match the seeded identity record |
| `active_project_id` | Project scope for the current session | Required for all tenant-scoped page flows |
| `role_membership` | User’s project membership and access rights | Must allow the four in-scope pages to load |

### 2. ProjectMembership

| Field | Description | Notes |
|---|---|---|
| `project_id` | Tenant/project boundary identifier | Every in-scope request remains project-scoped |
| `user_id` | Owner/member relation back to the default user | Must be consistent with identity/onboarding seed state |
| `permissions` | Access rights for protected BFF routes | Must include access needed for documents, payments, analyses, and settings |

### 3. DocumentRecord

| Field | Description | Notes |
|---|---|---|
| `document_id` | Unique uploaded document identifier | Files-owned |
| `project_id` | Project scope | Required for document listing |
| `file_name` | User-visible name | Used on the documents screen |
| `kind` | Document type classification | Example values include invoice/statement |
| `analysis_status` | Current analysis/classification state | Drives document readiness on the page |
| `uploaded_by` | User who uploaded the document | Should resolve to the default user in the seed path |

### 4. PaymentDashboardSnapshot

| Field | Description | Notes |
|---|---|---|
| `bill_id` | Bill record shown on the dashboard | Bills-owned |
| `project_id` | Project scope | Must match the default user’s active project |
| `amount_due` | Monetary amount owed | Displayed on the payments page |
| `due_date` | Expected due date | Supports overdue/upcoming status |
| `payment_status` | Current payment state | Must support examples such as unpaid, overdue, paid |
| `cycle_preference_day` | Preferred payment day | Payments-owned preference surfaced on the same page |

### 5. HistoryAnalyticsSnapshot

| Field | Description | Notes |
|---|---|---|
| `period_start` / `period_end` | Requested analysis window | Used by the history/analyses page |
| `timeline_points` | Time-series breakdown of spending/payment activity | Payments-owned read model |
| `category_breakdown` | Category-level totals and percentages | Used for analytical summaries |
| `compliance_metrics` | Optional compliance or quality indicators | May be secondary but must not break the page |

### 6. ReconciliationEntry

| Field | Description | Notes |
|---|---|---|
| `transaction_line_id` | Statement transaction identifier | Payments/reconciliation-owned |
| `bill_record_id` | Linked bill identifier | Used for manual or automatic matching |
| `reconciliation_status` | Current match state | Expected states include `unmatched`, `matched_auto`, `matched_manual`, `ambiguous` |
| `link_type` | Automatic or manual linkage type | Drives reconciliation UI messaging |

### 7. BankAccountLabel

| Field | Description | Notes |
|---|---|---|
| `bank_account_id` | Unique settings entry identifier | Files-owned via bank-account endpoints |
| `project_id` | Project scope | Must be tied to the default user’s active project |
| `label` | User-visible bank account label | Rendered on the settings screen |
| `created_by` | User who created the label | Should resolve to the default user in the seed path |

## Relationships

- `AuthenticatedDefaultUser` **must have** an active `ProjectMembership` before any protected BFF route can succeed.
- A `ProjectMembership` scopes access to `DocumentRecord`, `PaymentDashboardSnapshot`, `HistoryAnalyticsSnapshot`, `ReconciliationEntry`, and `BankAccountLabel`.
- `DocumentRecord` can feed downstream bill/payment workflows when the uploaded artifact is a bill or statement.
- `PaymentDashboardSnapshot`, `HistoryAnalyticsSnapshot`, and `ReconciliationEntry` all depend on consistent `bills` + `payments` data for the same project.
- `BankAccountLabel` provides configuration context used by statement attribution and should coexist with the seeded documents/transactions for the same project.

## Validation Rules

- Every in-scope request remains authenticated and project-scoped.
- The canonical acceptance seed must populate all four pages with representative data for the default user.
- Empty states are allowed only for explicit fallback or secondary validation scenarios; they are not the default acceptance path for this feature.
- The BFF may compose responses for the frontend, but business data ownership remains with the downstream services.

## State Transitions to Preserve

### Document lifecycle
`uploaded` → `analyzing/classifying` → `classified/ready`

### Bill/payment lifecycle
`unpaid` → `overdue` or `paid`

### Reconciliation lifecycle
`unmatched` → `matched_auto` / `matched_manual` or `ambiguous`

## Planning Implications

- If a page still breaks after auth/session succeeds, the next checks should be seeded data integrity, downstream contract responses, and mapper/error translation behavior.
- The seed path must be deterministic and idempotent so the same default user experience can be reproduced repeatedly in local and integration environments.
