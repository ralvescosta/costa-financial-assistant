# Data Model: Restore BFF gRPC Gateway Boundary

## Entity: CyclePreference

- **Purpose**: Project-scoped preferred payment-day configuration owned by the Payments service.
- **Fields**:
  - `id` (string): persistent identifier.
  - `project_id` (string): tenant scope key.
  - `preferred_day_of_month` (int): allowed range `1..28`.
  - `updated_by` (string): authenticated user who last changed the preference.
  - `updated_at` (timestamp): last update time.
- **Validation rules**:
  - `project_id` is required for every read/write.
  - `preferred_day_of_month` must be between `1` and `28`.
- **Ownership**: `backend/internals/payments/services/payment_cycle_service.go`

## Entity: MonthlyTimelinePoint

- **Purpose**: Month-level expenditure trend row returned through the new payments gRPC analytics contract.
- **Fields**:
  - `month` (string, `YYYY-MM-DD`): first day of the calendar month.
  - `total_amount` (decimal string): summed amount due for the month.
  - `bill_count` (int): number of bills included in the month.
- **Ownership**: `backend/internals/payments/repositories/history_repository.go`

## Entity: CategoryBreakdownPoint

- **Purpose**: One category’s spend total within a month for history analytics.
- **Fields**:
  - `month` (string)
  - `bill_type_name` (string)
  - `total_amount` (decimal string)
  - `bill_count` (int)
- **Ownership**: `backend/internals/payments/repositories/history_repository.go`

## Entity: MonthlyCompliancePoint

- **Purpose**: Monthly on-time payment compliance projection.
- **Fields**:
  - `month` (string)
  - `total_bills` (int)
  - `paid_on_time` (int)
  - `overdue` (int)
  - `compliance_rate` (decimal string)
- **Ownership**: `backend/internals/payments/repositories/history_repository.go`

## Entity: ReconciliationSummaryEntry

- **Purpose**: Project-scoped projection combining one transaction line with the linked bill information, if any.
- **Fields**:
  - `transaction_line_id` (string)
  - `transaction_date` (string)
  - `description` (string)
  - `amount` (decimal string)
  - `direction` (string)
  - `reconciliation_status` (enum: `unmatched | matched_auto | matched_manual | ambiguous`)
  - `linked_bill_id` (optional string)
  - `linked_bill_due_date` (optional string)
  - `linked_bill_amount` (optional string)
  - `link_type` (optional enum: `auto | manual`)
- **Ownership**: `backend/internals/payments/services/reconciliation_service.go`

## Entity: ReconciliationLink

- **Purpose**: Materialized manual or automatic relationship between a statement transaction line and a bill record.
- **Fields**:
  - `id` (string)
  - `project_id` (string)
  - `transaction_line_id` (string)
  - `bill_record_id` (string)
  - `link_type` (enum: `auto | manual`)
  - `linked_by` (optional string)
  - `created_at` (timestamp)
- **Ownership**: `backend/internals/payments/services/reconciliation_service.go`

## Entity: BFFRouteOwnershipRecord

- **Purpose**: Planning/verification record that maps a BFF route to its owning downstream service and target RPC.
- **Fields**:
  - `route_path` (string)
  - `http_method` (string)
  - `bff_controller` (string)
  - `bff_service_method` (string)
  - `owning_service` (enum: `bills | payments | files | onboarding | identity`)
  - `target_rpc` (string)
  - `verification_tests` (list of paths)
- **Invariant**: every supported BFF business route must resolve to exactly one owning service contract path and zero direct DB/repository dependencies.

## Entity: DirectAccessViolation

- **Purpose**: Record of a prohibited BFF dependency that must be removed during implementation.
- **Fields**:
  - `bff_file_path` (string)
  - `forbidden_dependency` (string)
  - `owning_service` (string)
  - `replacement_rpc` (string)
  - `status` (enum: `detected | rpc_defined | migrated | verified`)
- **Invariant**: no `DirectAccessViolation` may remain in `detected` or `rpc_defined` state at feature completion for supported flows.

## Relationships

- `BFFRouteOwnershipRecord` binds each HTTP route to one owning downstream service and one target RPC.
- `CyclePreference`, `MonthlyTimelinePoint`, `CategoryBreakdownPoint`, `MonthlyCompliancePoint`, `ReconciliationSummaryEntry`, and `ReconciliationLink` are all project-scoped data owned by the Payments service.
- `DirectAccessViolation` transitions to resolved status only when the new gRPC path is implemented and verified.

## State Transitions

1. **Cycle preference read**: `HTTP GET /payment-cycle/preferred-day` → BFF auth/validation → `payments.v1.GetCyclePreference` → Payments service → repository → PostgreSQL.
2. **Cycle preference write**: `HTTP PUT /payment-cycle/preferred-day` → BFF auth/validation → `payments.v1.SetCyclePreference` → Payments service → repository → PostgreSQL.
3. **History analytics**: `HTTP GET /history/*` → BFF auth → payments history RPC → Payments service → history repository → PostgreSQL aggregates.
4. **Manual reconciliation**: `HTTP POST /reconciliation/links` → BFF auth/guard → `payments.v1.CreateManualLink` → Payments service → reconciliation repository → PostgreSQL.
5. **Violation remediation lifecycle**: `detected` → `rpc_defined` → `migrated` → `verified` → `docs_synced`.
