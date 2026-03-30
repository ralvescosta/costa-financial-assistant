# BFF OpenAPI Contract (Echo + Huma)

## API style
- Transport: HTTP/JSON
- Contract source: Huma operation declarations
- Published document: `/openapi.json`
- Auth context: JWT (validated via identity-grpc JWKS)
- Tenant context: active `project_id` claim required for project-scoped routes

## Route groups and operations

## Upload and classification
- `POST /api/v1/documents/upload`
  - OperationID: `upload-document`
  - Summary: Upload PDF and create pending analysis document record
  - Tag: `documents`
- `POST /api/v1/documents/{documentId}/classify`
  - OperationID: `classify-document`
  - Summary: Set document type and attribution metadata (bill/statement)
  - Tag: `documents`
- `GET /api/v1/documents`
  - OperationID: `list-documents`
  - Summary: List project-scoped documents with status filters
  - Tag: `documents`
- `GET /api/v1/documents/{documentId}`
  - OperationID: `get-document`
  - Summary: Fetch document details and extraction fields
  - Tag: `documents`

## Bill payment workflow
- `GET /api/v1/bills/payment-dashboard`
  - OperationID: `get-payment-dashboard`
  - Summary: List outstanding bills for current cycle
  - Tag: `payments`
- `POST /api/v1/bills/{billId}/mark-paid`
  - OperationID: `mark-bill-paid`
  - Summary: Mark bill as paid (idempotent)
  - Tag: `payments`

## Reconciliation workflow
- `GET /api/v1/reconciliation/summary`
  - OperationID: `get-reconciliation-summary`
  - Summary: Return reconciled/unconfirmed overview by period
  - Tag: `reconciliation`
- `POST /api/v1/reconciliation/links`
  - OperationID: `create-reconciliation-link`
  - Summary: Create manual link between statement transaction and bill
  - Tag: `reconciliation`

## Financial history
- `GET /api/v1/history/timeline`
  - OperationID: `get-history-timeline`
  - Summary: Return monthly expenditure timeline
  - Tag: `history`
- `GET /api/v1/history/months/{yyyyMm}/categories`
  - OperationID: `get-month-category-breakdown`
  - Summary: Return monthly category totals
  - Tag: `history`
- `GET /api/v1/history/compliance`
  - OperationID: `get-payment-compliance`
  - Summary: Return on-time vs overdue rates by month
  - Tag: `history`

## Configuration and labels
- `GET /api/v1/bank-accounts`
  - OperationID: `list-bank-accounts`
  - Summary: List project-scoped bank account labels
  - Tag: `settings`
- `POST /api/v1/bank-accounts`
  - OperationID: `create-bank-account`
  - Summary: Create bank account label
  - Tag: `settings`
- `DELETE /api/v1/bank-accounts/{bankAccountId}`
  - OperationID: `delete-bank-account`
  - Summary: Delete bank account label with attribution guard
  - Tag: `settings`
- `PUT /api/v1/payment-cycle/preferred-day`
  - OperationID: `set-preferred-payment-day`
  - Summary: Set preferred payment day for active project
  - Tag: `settings`

## Project collaboration
- `GET /api/v1/projects/current`
  - OperationID: `get-current-project`
  - Summary: Fetch active project metadata and permissions
  - Tag: `projects`
- `POST /api/v1/projects/{projectId}/members/invite`
  - OperationID: `invite-project-member`
  - Summary: Invite project member with role
  - Tag: `projects`
- `PUT /api/v1/projects/{projectId}/members/{memberId}/role`
  - OperationID: `update-project-member-role`
  - Summary: Update collaborator role
  - Tag: `projects`

## Standard response/error model
- Success responses:
  - 200/201 with typed response body
- Validation error:
  - 400 with field-level validation details
- Unauthorized:
  - 401 when JWT is missing/invalid/expired
- Forbidden:
  - 403 when role does not allow operation
- Not found:
  - 404 for missing project-scoped resources
- Conflict:
  - 409 for duplicate upload within same project or idempotency conflicts

## Huma operation requirements
Every operation must define:
- `OperationID`
- `Summary`
- `Description`
- `Tags`
- Input validation metadata (`required`, `pattern`, `minimum`/`maximum`, etc.)
