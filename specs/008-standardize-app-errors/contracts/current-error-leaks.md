# Current Error Leak Points: Baseline Audit

**Baseline Date**: April 3, 2026  
**Auditor**: Implementation Phase 1  
**Status**: Pending detailed audit (placeholder for T001 execution results)

## Scope

This audit documents known backend error-leak points where raw dependency errors (database, gRPC, network) currently cross layer boundaries without translation to standardized `AppError` types.

## Known Leak Categories (To Be Discovered)

### Category: Database Errors
- **Location**: `backend/internals/*/repositories/*.go`
- **Issue**: Raw `sql.Error` instances may leak from repository methods to service layer
- **Impact**: Inconsistent error handling and potential information leakage
- **Example**: Error messages containing SQL syntax or internal schema details

### Category: gRPC Errors
- **Location**: `backend/internals/*/services/*.go` (where gRPC clients are used)
- **Issue**: Raw `grpc.error` status messages leak from service-to-service calls
- **Impact**: Inconsistent error handling and leakage of non-sanitized error messages
- **Example**: gRPC status messages containing internal service implementation details

### Category: Transport Layer Errors
- **Location**: `backend/internals/*/transport/grpc/server.go`
- **Issue**: Raw errors sometimes propagate into gRPC responses without catalog mapping
- **Impact**: Bypass of sanitization and error classification
- **Example**: HTTP/gRPC responses exposing database error messages

### Category: Async/Event Errors
- **Location**: `backend/internals/*/transport/rmq/*.go`
- **Issue**: Consumer error handlers may propagate raw errors into log statements
- **Impact**: Inconsistent error handling in event-driven paths
- **Example**: RabbitMQ consumer errors with unsanitized context

## Remediation Task

This baseline will be populated by **T001 audit findings** and used to:
1. Track which error-leak points have been remediated
2. Validate implementation completeness in Phase 3-5
3. Guide Phase 6 verification testing

---

## Audit Progress (To Be Filled)

- [ ] Repository layer audit complete
- [ ] Service layer audit complete
- [ ] Transport layer audit complete
- [ ] Async/RMQ layer audit complete
- [ ] Detailed fix plan created for Phase 1-5 remediation
