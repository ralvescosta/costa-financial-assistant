# Implementation Continuation Guide: Phase 3+ Tasks

**Last Updated**: April 3, 2026  
**Status**: Phase 1-2 Complete; Phase 3 Ready to Begin  
**Next Developer Entry Point**: Phase 3 (T011-T025, T062)

## Quick Start for Continuing Development

### Prerequisites Met ✅

1. **Shared Error Package** (`backend/pkgs/errors/`):
   - Error types and categories defined
   - Catalog with 16 standardized entries ready
   - Translation policy and classification helpers implemented
   - Unit tests passing

2. **Integration Test Helpers** (`backend/tests/integration/helpers/`):
   - `AppErrorAssertions` class ready for test development
   - Non-leakage contract validation functions available

3. **Documentation**:
   - Service adoption checklist created (guides per-service implementation)
   - Coverage matrix established (tracks progress)
   - Baseline error-leak audit template ready

### Environment Setup

```bash
# From project root
cd backend

# Verify error package compiles
go build ./pkgs/errors/...

# Run existing unit tests
go test -v ./pkgs/errors/...

# View error catalog
grep -n "var Err" pkgs/errors/consts.go
```

---

## Phase 3: US1 Implementation Guide (T011-T025, T062)

**Goal**: Enforce `AppError` as the ONLY error type crossing backend layer boundaries.

### Step-by-Step Approach

#### 1. Create Test Files First (T011, T012, T061) - Test-Driven Development

**T011**: Repository-layer error translation tests
```bash
# Create test file
touch backend/internals/files/repositories/error_translation_test.go

# Template: Test that repository methods translate DB errors to AppError
# Example test cases:
# - sql.ErrNoRows → ErrResourceNotFound
# - Connection errors → ErrDatabaseConnection  
# - Timeout errors → ErrDatabaseTimeout
# - Unknown DB error → ErrUnknown (fallback)
```

**T012**: Cross-service propagation integration test
```bash
# Create test file
touch backend/tests/integration/cross_service/app_error_propagation_test.go

# Test: Trigger a repository error in one service, verify:
# 1. Service method returns AppError (not raw DB error)
# 2. gRPC handler receives AppError (not raw error)
# 3. gRPC response is sanitized (no internal details)
```

**T061**: Async publisher propagation test
```bash
# Create test file
touch backend/tests/integration/cross_service/app_error_async_publisher_propagation_test.go

# Test: Verify RabbitMQ producer errors are translated
```

#### 2. Implement Phase 3 Tasks (T013-T025, T062) - Service by Service

**Pattern for each service**:

```go
// 1. Repository layer (T013-T017): Translate at DB boundary
func (r *DocumentRepository) GetDocument(ctx context.Context, docID string) (*Document, error) {
    row := r.db.QueryRowContext(ctx, "SELECT ... FROM documents WHERE id = $1", docID)
    var doc Document
    err := row.Scan(...)
    if err != nil {
        // KEY: Translate native error to AppError
        appErr := errors.TranslateError(err, "repository")
        // Log the native error ONCE (boundary logging)
        r.logger.Error("repository: database error in GetDocument", zap.Error(err))
        return nil, appErr
    }
    return &doc, nil
}

// 2. Service layer (T018-T020): Don't translate, just propagate
func (s *DocumentService) GetDocument(ctx context.Context, docID string) (*Document, error) {
    doc, err := s.repo.GetDocument(ctx, docID)
    if err != nil {
        // IMPORTANT: Don't re-translate AppError; propagate as-is
        return nil, err
    }
    return doc, nil
}

// 3. Transport layer (T021-T025): Map to response contract
func (srv *DocumentServer) GetDocument(ctx context.Context, req *pb.GetDocumentRequest) (*pb.DocumentResponse, error) {
    doc, err := srv.svc.GetDocument(ctx, req.DocId)
    if err != nil {
        // Map AppError to gRPC response (never expose raw error)
        if appErr := errors.AsAppError(err); appErr != nil {
            // Use Sanitize() or custom mapping to convert AppError to gRPC status
            return nil, status.Error(codes.Internal, appErr.Message)
        }
        return nil, status.Error(codes.Internal, "unknown error")
    }
    return &pb.DocumentResponse{...}, nil
}
```

#### 3. Service Implementation Order (Recommended)

1. **Files  service** (T013, T018, T021, T025, T062)
2. **Bills service** (T014, T019, T022)
3. **Onboarding service** (T015, T020, T023)
4. **Payments service** (T016, T024)
5. **BFF/Identity** (as applicable)

---

## Key Implementation Rules (Non-negotiable)

### Router 1: Single Translation Boundary Per Layer
- **Repository**: Translate DB errors → AppError, log native error once
- **Service**: Propagate AppError as-is, DO NOT re-translate
- **Transport**: MAP AppError to external contract (gRPC status, HTTP response)
- **Async**: Translate consumer errors → AppError

### Rule 2: Fallback to Unknown
Every error path MUST eventually map to a catalog entry. If no specific rule matches:
```go
appErr := errors.TranslateError(nativeErr, "repository")
if appErr == nil {
    appErr = errors.NewCatalogError(errors.ErrUnknown).WithError(nativeErr)
}
return nil, appErr
```

### Rule 3: Log Native Errors Only at Boundary
Native errors are logged ONCE using structured logging:
```go
// Good: Log native error at translation boundary
if err != nil {
    logger.Error("repository: database error", zap.Error(err))
    return nil, errors.TranslateError(err, "repository")
}

// Bad: Log translated error (no diagnostic value)
if err != nil {
    appErr := errors.TranslateError(err, "service")
    logger.Error("service error", zap.Error(appErr))  // Won't show native details!
    return nil, appErr
}
```

### Rule 4: Never Expose Native Errors in Responses
```go
// Bad: gRPC response includes native error details
if err != nil {
    return nil, status.Error(codes.Internal, err.Error())  // Could leak SQL syntax!
}

// Good: Use sanitized AppError message
if err != nil {
    if appErr := errors.AsAppError(err); appErr != nil {
        return nil, status.Error(codes.Internal, appErr.Message)
    }
    return nil, status.Error(codes.Internal, "unknown error")
}
```

---

## Testing Checklist

For each service:

- [ ] **Unit Tests**:
  - [ ] Repository translation test exists and passes
  - [ ] Service propagation test exists and passes
  - [ ] Transport mapping test exists and passes
  - [ ] Async consumer test exists and passes (if applicable)

- [ ] **Integration Tests**:
  - [ ] Cross-service error propagation works (no raw errors leak)
  - [ ] Error message is safe for client exposure
  - [ ] Retryable flag is set correctly
  - [ ] Unknown-fallback behavior works deterministically

- [ ] **Compliance**:
  - [ ] No raw DB errors cross repository boundary
  - [ ] No raw gRPC errors cross service boundary
  - [ ] No raw errors in transport responses
  - [ ] Boundary logs exist and don't contain sensitive data

---

## File Checklist: Phase 3 Deliverables

### Tests (T011, T012, T061)
- [ ] `backend/internals/files/repositories/error_translation_test.go`
- [ ] `backend/tests/integration/cross_service/app_error_propagation_test.go`
- [ ] `backend/tests/integration/cross_service/app_error_async_publisher_propagation_test.go`

### Implementation (T013-T025, T062)  

**Files**:
- [ ] `backend/internals/files/repositories/document_repository.go` (T013)
- [ ] `backend/internals/files/services/document_service.go` (T018)
- [ ] `backend/internals/files/transport/grpc/server.go` (T021, T025)
- [ ] `backend/internals/files/transport/rmq/analysis_consumer.go` (T062)

**Bills**:
- [ ] `backend/internals/bills/repositories/payment_repository.go` (T014)
- [ ] `backend/internals/bills/services/payment_service.go` (T019)
- [ ] `backend/internals/bills/transport/grpc/server.go` (T022)

**Onboarding**:
- [ ] `backend/internals/onboarding/repositories/project_members_repository.go` (T015)
- [ ] `backend/internals/onboarding/services/project_members_service.go` (T020)
- [ ] `backend/internals/onboarding/transport/grpc/server.go` (T023)

**Payments**:
- [ ] `backend/internals/payments/repositories/payment_cycle_repository.go` (T016)
- [ ] `backend/internals/payments/repositories/reconciliation_repository.go` (T017)
- [ ] `backend/internals/payments/services/payment_cycle_service.go` (T024, implies T035-T036 in US2)
- [ ] `backend/internals/payments/services/reconciliation_service.go` (T024, implies T035-T036 in US2)

---

## Common Pitfalls to Avoid

1. **Re-translating AppErrors**: Once translated, don't call `TranslateError` again on an `AppError`.
2. **Logging translated errors**: Log native errors at boundary, not AppErrors.
3. **Exposing SQL/gRPC details**: Always use `appErr.Message`, never `err.Error()` in responses.
4. **Forgetting the fallback**: Always ensure no error path returns nil without translation.
5. **Skipping tests**: Tests are MANDATORY before implementation (TDD).

---

## Verification Commands

```bash
# Build packages
cd backend && go build ./...

# Run all tests
go test ./...

# Run only error package tests
go test ./pkgs/errors/... -v

# Check for raw error type assertions (potential leaks)
grep -r "err\." internals/*/transport/grpc/server.go | grep Error

# Verify no duplicate imports
go mod tidy
```

---

## Useful References

- **Error Package API**: `backend/pkgs/errors/error.go` (types and constructors)
- **Catalog Entries**: `backend/pkgs/errors/consts.go` (all catalog definitions)
- **Translation Policy**: `backend/pkgs/errors/mapping.go` (layer-specific rules)
- **Classifiers**: `backend/pkgs/errors/native_classifiers.go` (DB/gRPC classification)
- **Test Helpers**: `backend/tests/integration/helpers/assert_app_error.go` (assertion utilities)
- **Service Checklist**: `specs/008-standardize-app-errors/contracts/service-adoption-checklist.md`
- **Coverage Matrix**: `specs/008-standardize-app-errors/contracts/service-coverage-matrix.md`

---

## Expected Outcomes

After Phase 3 completion:
1. ✅ MVP delivered with consistent error contract across MVP services
2. ✅ No raw dependency errors leak across layer boundaries
3. ✅ Error messages are safe for client exposure
4. ✅ All integration tests pass
5. ✅ Service coverage matrix updated with Phase 3 status

---

Generated: April 3, 2026  
Next Review: After Phase 3 completion
