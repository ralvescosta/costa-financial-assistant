# Implementation Progress Matrix: AppError Standardization

**Feature**: `008-standardize-app-errors`  
**Date**: April 3, 2026  
**Status**: Phase 1 Baseline (Pre-Implementation)

## Implementation Coverage Matrix

This matrix tracks error translation implementation progress across backend services and layer boundaries.

### Legend
- ✅ Complete (AppError contract enforced, tests passing)
- ⚙️ In Progress (implementation underway)
- ❌ Not Started
- 🟡 Partial (some parts done, others pending)
- N/A Not applicable for this service

## Services x Layers Implementation Status

| Service | Repository Layer | Service Layer | Transport Layer | Async Layer | Tests | Status |
|---------|------------------|---------------|-----------------|-------------|-------|--------|
| **BFF** | N/A | ❌ Not Started | ❌ Not Started | N/A | ❌ Not Started | ❌ |
| **Files** | ❌ Not Started | ❌ Not Started | ❌ Not Started | ❌ Not Started | ❌ Not Started | ❌ |
| **Bills** | ❌ Not Started | ❌ Not Started | ❌ Not Started | N/A | ❌ Not Started | ❌ |
| **Identity** | N/A | ❌ Not Started | ❌ Not Started | N/A | ❌ Not Started | ❌ |
| **Onboarding** | ❌ Not Started | ❌ Not Started | ❌ Not Started | N/A | ❌ Not Started | ❌ |
| **Payments** | ❌ Not Started | ❌ Not Started | ❌ Not Started | N/A | ❌ Not Started | ❌ |

---

## Phase Progress

### Phase 1: Setup (T001-T005)

**Status**: ⚙️ In Progress

| Task | File | Status | Est. Completion |
|------|------|--------|-----------------|
| T001 | `specs/008-standardize-app-errors/contracts/current-error-leaks.md` | ✅ Created | 04-03-2026 |
| T002 | `backend/pkgs/errors/error.go` | ✅ Expanded | 04-03-2026 |
| T003 | `backend/pkgs/errors/translate.go` | ✅ Created | 04-03-2026 |
| T004 | `backend/pkgs/errors/consts.go` | ✅ Expanded | 04-03-2026 |
| T005 | `backend/pkgs/errors/error_test.go` | ✅ Created | 04-03-2026 |

### Phase 2: Foundational (T006-T010)

**Status**: 🟡 Partial (T006 in progress)

| Task | File | Status | Est. Completion |
|------|------|--------|-----------------|
| T006 | `backend/pkgs/errors/mapping.go` | ✅ Created | 04-03-2026 |
| T007 | `backend/pkgs/errors/native_classifiers.go` | ✅ Created | 04-03-2026 |
| T008 | `backend/tests/integration/helpers/assert_app_error.go` | ✅ Created | 04-03-2026 |
| T009 | `specs/008-standardize-app-errors/contracts/service-adoption-checklist.md` | ✅ Created | 04-03-2026 |
| T010 | `specs/008-standardize-app-errors/contracts/service-coverage-matrix.md` | ✅ Created | 04-03-2026 |

### Phase 3: US1 (T011-T025, T062)

**Status**: ❌ Not Started

| Task | File | Status | Est. Completion |
|------|------|--------|-----------------|
| T011 | `backend/internals/files/repositories/error_translation_test.go` | ❌ Not Started | |
| T012 | `backend/tests/integration/cross_service/app_error_propagation_test.go` | ❌ Not Started | |
| T013-T025 | Service implementations | ❌ Not Started | |
| T061 | `backend/tests/integration/cross_service/app_error_async_publisher_propagation_test.go` | ❌ Not Started | |
| T062 | Async producer sanitization | ❌ Not Started | |

### Phase 4: US2 (T026-T036)

**Status**: ❌ Not Started

All logging implementation tasks pending Phase 3 completion.

### Phase 5: US3 (T037-T043, T063)

**Status**: ❌ Not Started

All retryability classification tasks pending Phase 2 completion.

### Phase 6: Polish (T044-T046, T064, T067)

**Status**: ❌ Not Started

Documentation and CI validation pending all implementation phases.

### Phase 7: Governance Sync (T047-T066)

**Status**: ❌ Not Started

Memory and instruction sync pending Phase 6 completion.

---

## Error Catalog Coverage

### Validation Errors
- ✅ `ErrValidationError` - Created
- ✅ `ErrInvalidRequest` - Created

### Auth Errors
- ✅ `ErrUnauthorized` - Created
- ✅ `ErrForbidden` - Created

### Not Found Errors
- ✅ `ErrResourceNotFound` - Created
- ✅ `ErrProjectNotFound` - Created

### Conflict Errors
- ✅ `ErrConflict` - Created
- ✅ `ErrResourceAlreadyExists` - Created

### Database Errors
- ✅ `ErrDatabaseError` - Created
- ✅ `ErrDatabaseConnection` - Created
- ✅ `ErrDatabaseTimeout` - Created

### gRPC Errors
- ✅ `ErrGRPCError` - Created
- ✅ `ErrGRPCUnavailable` - Created

### Network Errors
- ✅ `ErrNetworkError` - Created
- ✅ `ErrNetworkTimeout` - Created

### Unknown (Fallback)
- ✅ `ErrUnknown` - Created (mandatory)
- ✅ `ErrInternal` - Created

---

## Regression & Validation

### Shared Package Tests

| Test Suite | Status | Coverage |
|-----------|--------|----------|
| `backend/pkgs/errors/error_test.go` | ✅ Created | ~15 test cases |
| `backend/pkgs/errors/translate_test.go` | ❌ Not Started | Pending |
| `backend/pkgs/errors/mapping_test.go` | ❌ Not Started | Pending |
| `backend/pkgs/errors/native_classifiers_test.go` | ❌ Not Started | Pending |

### Integration Tests

| Test Suite | Status | Scope |
|-----------|--------|-------|
| `backend/tests/integration/cross_service/app_error_propagation_test.go` | ❌ Not Started | Multi-service error flow |
| `backend/tests/integration/cross_service/app_error_boundary_logging_test.go` | ❌ Not Started | Boundary log verification |
| `backend/tests/integration/cross_service/app_error_unknown_fallback_test.go` | ❌ Not Started | Fallback behavior |
| Service-specific tests | ❌ Not Started | Per-service validation |

---

## Overall Completion

### Metrics

- **Total Tasks**: 67 (T001-T067, with some reordered)
- **Tasks Complete**: 10 (T001-T010)
- **Completion %**: 15%
- **Phase Completion**:
  - Phase 1: 100% ✅
  - Phase 2: 100% ✅
  - Phase 3: 0% ❌
  - Phase 4: 0% ❌
  - Phase 5: 0% ❌
  - Phase 6: 0% ❌
  - Phase 7: 0% ❌

### Blockers

- ⏳ Awaiting Phase 3 user story implementation (test-driven development)
- ⏳ Awaiting Phase 2 finalization before Phase 3 can start

### Next Steps

1. Execute Phase 3 test tasks (T011, T012, T061)
2. Execute Phase 3 implementation tasks (T013-T025, T062)
3. Validate MVP compliance
4. Proceed to Phase 4-5 incrementally

---

## Historical Changes

| Date | Phase | Status | Summary |
|------|-------|--------|---------|
| 2026-04-03 | Phase 1-2 | ✅ Complete | Shared error package and foundational infrastructure complete. All Phase 1-2 tasks delivered. |
