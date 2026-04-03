# Service Coverage Matrix: AppError Standardization

Feature: `008-standardize-app-errors`  
Updated: 2026-04-03  
Status: Phase 6 validation complete

## Legend

- ✅ Complete
- 🟡 Partial
- N/A Not applicable

## Services x Layers

| Service | Repository Layer | Service Layer | Transport Layer | Async Layer | Tests | Status |
|---|---|---|---|---|---|---|
| BFF | N/A | ✅ | ✅ (HTTP adapter sanitization via services) | N/A | ✅ | ✅ |
| Files | ✅ | ✅ | ✅ | ✅ (consumer + publisher sanitization helpers) | ✅ | ✅ |
| Bills | ✅ | ✅ | ✅ | N/A | ✅ | ✅ |
| Identity | N/A | ✅ | ✅ | N/A | ✅ | ✅ |
| Onboarding | ✅ | ✅ | ✅ | N/A | ✅ | ✅ |
| Payments | ✅ | ✅ | ✅ (via BFF orchestration and payments services) | N/A | ✅ | ✅ |

## Story Task Coverage Summary

| Story | Focus | Coverage |
|---|---|---|
| US1 | AppError-only cross-layer propagation | ✅ Completed (`T011`, `T012`, `T013`-`T025`, `T061`, `T062`) |
| US2 | One-boundary structured error logging | ✅ Completed (`T026`-`T036`) |
| US3 | Retryability + unknown fallback determinism | ✅ Completed (`T037`-`T043`, `T063`) |

## Validation Evidence

- Unit/service/repository/transport suites were re-run for all touched packages from `backend/go.mod` scope.
- Cross-service integration tests include:
  - `app_error_propagation_test.go`
  - `app_error_async_publisher_propagation_test.go`
  - `app_error_unknown_fallback_test.go`
  - `app_error_boundary_logging_test.go`

## Residual Notes

- Transport mappings now consistently convert `AppError` category/retryability to protocol-safe status codes/messages.
- Unknown or unmapped layer translations deterministically resolve to `ErrUnknown` fallback.
