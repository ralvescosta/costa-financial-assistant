# Payments gRPC Service Contract

## Purpose

Define the new inter-service contract the BFF must use for payments-owned capabilities so the gateway no longer reaches Payments domain data through in-process interfaces, repositories, or direct database access.

## Ownership split

| Capability | Owning service | Evidence in current codebase | Contract outcome |
|---|---|---|---|
| Payment dashboard | `bills` | Existing `bills.v1.BillsService/GetPaymentDashboard` is already consumed by the BFF | **No change** |
| Mark bill paid | `bills` | Existing `bills.v1.BillsService/MarkBillPaid` owns `BillRecord` payment status mutation | **No change** |
| Preferred payment day | `payments` | `backend/internals/payments/services/payment_cycle_service.go` | **Add `payments.v1` RPC** |
| History analytics | `payments` | `backend/internals/payments/repositories/history_repository.go` | **Add `payments.v1` RPCs** |
| Reconciliation summary/link | `payments` | `backend/internals/payments/services/reconciliation_service.go` | **Add `payments.v1` RPCs** |

## Required payments RPC surface

The exact naming can be refined during implementation, but the contract must cover the following behaviors:

| RPC | Request essentials | Response essentials | Backing domain implementation |
|---|---|---|---|
| `GetCyclePreference` | `common.v1.ProjectContext` | `project_id`, `preferred_day_of_month`, `updated_at` | `PaymentCycleService.GetCyclePreference` |
| `SetCyclePreference` | `common.v1.ProjectContext`, `preferred_day_of_month`, audit metadata | persisted cycle preference | `PaymentCycleService.UpsertCyclePreference` |
| `GetHistoryTimeline` | `common.v1.ProjectContext`, `months` | repeated monthly timeline points | `HistoryRepository.GetTimeline` |
| `GetHistoryCategoryBreakdown` | `common.v1.ProjectContext`, `months` | repeated category breakdown points | `HistoryRepository.GetCategoryBreakdown` |
| `GetHistoryCompliance` | `common.v1.ProjectContext`, `months` | repeated compliance points | `HistoryRepository.GetComplianceMetrics` |
| `GetReconciliationSummary` | `common.v1.ProjectContext`, `period_start`, `period_end` | project-scoped reconciliation summary entries | `ReconciliationService.GetSummary` |
| `CreateManualLink` | `common.v1.ProjectContext`, `transaction_line_id`, `bill_record_id`, audit metadata | created reconciliation link | `ReconciliationService.CreateManualLink` |

## Contract rules

1. Domain messages live in `backend/protos/payments/v1/messages.proto`.
2. RPC wrapper messages and the service definition live in `backend/protos/payments/v1/grpc.proto`.
3. Every request must be project-scoped and authenticated via `common.v1.ProjectContext`.
4. The BFF must consume these RPCs through generated clients only; it must not fall back to payments repositories or in-process payments interfaces once the contract exists.
5. All boundary failures must translate to `AppError`-aligned responses/logging; raw database or dependency errors must not leak.

## Acceptance mapping back to BFF

| BFF file | Target downstream contract |
|---|---|
| `backend/internals/bff/services/payments_service.go` | `bills.v1` for dashboard/mark-paid + `payments.v1` for cycle preference |
| `backend/internals/bff/services/history_service.go` | `payments.v1` history analytics RPCs |
| `backend/internals/bff/services/reconciliation_service.go` | `payments.v1` reconciliation RPCs |
