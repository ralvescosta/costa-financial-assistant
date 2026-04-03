# Service Adoption Checklist: Standardize Backend App Errors

**Feature**: `008-standardize-app-errors`  
**Date Created**: April 3, 2026  
**Status**: Template for service-by-service adoption tracking

This checklist ensures each backend service correctly implements the `AppError` standardization across all layer boundaries.

## Service: BFF (Backend For Frontend)

### Phase 3: US1 - Consistent Error Contract (Priority: P1)

- [ ] Repository layer (if applicable) implements error translation
  - [ ] `backend/internals/bff/repositories/` - all query methods translate errors
  - [ ] Tests: `backend/tests/integration/bff/error_translation_test.go`
- [ ] Service layer implements error propagation
  - [ ] `backend/internals/bff/services/` - all public methods propagate only `AppError`
  - [ ] Tests: `backend/tests/integration/bff/service_error_propagation_test.go`
- [ ] Transport layer implements error sanitization
  - [ ] `backend/internals/bff/transport/grpc/server.go` - gRPC handler error mapping
  - [ ] `backend/internals/bff/transport/http/` - HTTP endpoint error mapping (if applicable)
  - [ ] Tests: `backend/tests/integration/bff/transport_error_sanitization_test.go`

### Phase 4: US2 - Boundary Logging

- [ ] Service layer implements one-boundary logging
  - [ ] `zap.Error(err)` logged once before translation
  - [ ] Tests: `backend/internals/bff/services/error_logging_test.go`

### Phase 5: US3 - Retryability Classification

- [ ] All catalog errors for BFF paths have retryability classification
  - [ ] Catalog entries in `backend/pkgs/errors/consts.go` include BFF-used errors
  - [ ] Tests: `backend/tests/integration/cross_service/app_error_unknown_fallback_test.go`

---

## Service: Files

### Phase 3: US1 - Consistent Error Contract (Priority: P1)

- [ ] Repository layer implements error translation
  - [ ] `backend/internals/files/repositories/document_repository.go`
  - [ ] `backend/internals/files/repositories/` - all query methods translate errors
  - [ ] Tests: `backend/internals/files/repositories/error_translation_test.go`
- [ ] Service layer implements error propagation
  - [ ] `backend/internals/files/services/` - all public methods propagate only `AppError`
  - [ ] Tests: `backend/tests/integration/files/service_error_propagation_test.go`
- [ ] Transport layer implements error sanitization
  - [ ] `backend/internals/files/transport/grpc/server.go` - gRPC handler error mapping
  - [ ] Tests: `backend/tests/integration/files/transport_error_sanitization_test.go`
- [ ] Async layer implements error sanitization
  - [ ] `backend/internals/files/transport/rmq/analysis_consumer.go` - consumer error handling
  - [ ] Tests: `backend/tests/integration/files/consumer_error_sanitization_test.go`

### Phase 4: US2 - Boundary Logging

- [ ] Service layer implements one-boundary logging
  - [ ] `backend/internals/files/services/extraction_service.go`
  - [ ] `backend/internals/files/services/bank_account_service.go`
  - [ ] Tests: `backend/internals/files/services/error_logging_test.go`

### Phase 5: US3 - Retryability Classification

- [ ] Catalog includes retryability for all Files service errors

---

## Service: Bills

### Phase 3: US1 - Consistent Error Contract (Priority: P1)

- [ ] Repository layer implements error translation
  - [ ] `backend/internals/bills/repositories/payment_repository.go`
  - [ ] Tests: `backend/tests/integration/bills/error_translation_test.go`
- [ ] Service layer implements error propagation
  - [ ] `backend/internals/bills/services/payment_service.go`
  - [ ] Tests: `backend/tests/integration/bills/service_error_propagation_test.go`
- [ ] Transport layer implements error sanitization
  - [ ] `backend/internals/bills/transport/grpc/server.go`
  - [ ] Tests: `backend/tests/integration/bills/transport_error_sanitization_test.go`

### Phase 4: US2 - Boundary Logging

- [ ] Service layer implements one-boundary logging
  - [ ] Tests: `backend/internals/bills/services/error_logging_test.go`

---

## Service: Identity

### Phase 3: US1 - Consistent Error Contract (Priority: P1)

- [ ] Service layer implements error propagation
  - [ ] `backend/internals/identity/services/` - all public methods propagate only `AppError`
  - [ ] Tests: `backend/tests/integration/identity/service_error_propagation_test.go`
- [ ] Transport layer implements error sanitization
  - [ ] `backend/internals/identity/transport/grpc/server.go`
  - [ ] Tests: `backend/tests/integration/identity/transport_error_sanitization_test.go`

### Phase 4: US2 - Boundary Logging

- [ ] Service layer implements one-boundary logging
  - [ ] `backend/internals/identity/services/token_service.go`
  - [ ] Tests: `backend/internals/identity/services/error_logging_test.go`

---

## Service: Onboarding

### Phase 3: US1 - Consistent Error Contract (Priority: P1)

- [ ] Repository layer implements error translation
  - [ ] `backend/internals/onboarding/repositories/project_members_repository.go`
  - [ ] Tests: `backend/tests/integration/onboarding/error_translation_test.go`
- [ ] Service layer implements error propagation
  - [ ] `backend/internals/onboarding/services/project_members_service.go`
  - [ ] Tests: `backend/tests/integration/onboarding/service_error_propagation_test.go`
- [ ] Transport layer implements error sanitization
  - [ ] `backend/internals/onboarding/transport/grpc/server.go`
  - [ ] Tests: `backend/tests/integration/onboarding/transport_error_sanitization_test.go`

### Phase 4: US2 - Boundary Logging

- [ ] Service layer implements one-boundary logging
  - [ ] Tests: `backend/internals/onboarding/services/error_logging_test.go`

---

## Service: Payments

### Phase 3: US1 - Consistent Error Contract (Priority: P1)

- [ ] Repository layer implements error translation
  - [ ] `backend/internals/payments/repositories/payment_cycle_repository.go`
  - [ ] `backend/internals/payments/repositories/reconciliation_repository.go`
  - [ ] Tests: `backend/tests/integration/payments/error_translation_test.go`
- [ ] Service layer implements error propagation
  - [ ] `backend/internals/payments/services/payment_cycle_service.go`
  - [ ] `backend/internals/payments/services/reconciliation_service.go`
  - [ ] Tests: `backend/tests/integration/payments/service_error_propagation_test.go`
- [ ] Transport layer implements error sanitization
  - [ ] `backend/internals/payments/transport/grpc/server.go`
  - [ ] Tests: `backend/tests/integration/payments/transport_error_sanitization_test.go`

### Phase 4: US2 - Boundary Logging

- [ ] Service layer implements one-boundary logging
  - [ ] `backend/internals/payments/services/payment_cycle_service.go`
  - [ ] `backend/internals/payments/services/reconciliation_service.go`
  - [ ] Tests: `backend/internals/payments/services/error_logging_test.go`

### Phase 5: US3 - Retryability Classification

- [ ] Service layer implements retryability-aware translation
  - [ ] `backend/internals/payments/repositories/history_repository.go`
  - [ ] Tests: `backend/tests/integration/payments/retryability_test.go`

---

## Shared Infrastructure (Phase 2)

- [ ] Error catalog in `backend/pkgs/errors/consts.go` is complete
- [ ] Translation policy in `backend/pkgs/errors/mapping.go` is finalized
- [ ] Classification helpers in `backend/pkgs/errors/native_classifiers.go` work correctly
- [ ] Integration test helpers in `backend/tests/integration/helpers/assert_app_error.go` are ready
- [ ] Service adoption checklist (this file) is kept in sync

---

## Adoption Success Criteria

Each service is considered **adopted** when:
1. All layer boundaries propagate only `AppError` types
2. Raw dependency errors are never exposed to upper layers
3. One boundary log exists before each translation
4. Retryability semantics are captured for all errors
5. Integration tests validate all the above requirements
6. No syntax errors in updated code; test suite passes

---

## Tracking Status

| Service | Phase 1 | Phase 2 | Phase 3 | Phase 4 | Phase 5 | Overall |
|---------|---------|---------|---------|---------|---------|---------|
| BFF | N/A | N/A | - | - | - | - |
| Files | N/A | N/A | - | - | - | - |
| Bills | N/A | N/A | - | - | - | - |
| Identity | N/A | N/A | - | - | - | - |
| Onboarding | N/A | N/A | - | - | - | - |
| Payments | N/A | N/A | - | - | - | - |
| **Overall** | ❌ Started | ⏳ In Progress | - | - | - | - |

Checklist last updated: April 3, 2026
