# Research: Standardize Backend App Errors

## Decision 1: Centralized Error Catalog Coverage Scope

- Decision: The centralized catalog in `backend/pkgs/errors/consts.go` will cover all currently known backend failure categories across all services, plus one mandatory unknown-fallback error.
- Rationale: This satisfies the clarified requirement for broad standardization while keeping implementation finite and testable.
- Alternatives considered:
  - Minimal catalog with incremental growth only: rejected because it leaves immediate inconsistency unresolved.
  - Service-by-service staged coverage: rejected for this feature because cross-layer consistency is P1.

## Decision 2: Retryability Classification Policy

- Decision: Retryability is classified by failure nature: transient infrastructure failures are retryable; deterministic/business-validation failures are non-retryable.
- Rationale: This policy is deterministic and aligns with future retry strategy enablement without requiring retry engine implementation now.
- Alternatives considered:
  - Mark all unknowns as retryable: rejected because it can amplify non-recoverable failure loops.
  - Mark all errors non-retryable: rejected because it prevents resilient behavior for transient outages.

## Decision 3: Translation Boundary Logging Rule

- Decision: Dependency-native errors are logged once at the translation boundary with structured context and `zap.Error(err)` before propagation as `AppError`.
- Rationale: This preserves root-cause observability and avoids duplicate log noise across layers.
- Alternatives considered:
  - Log in every layer: rejected due to duplicated noise and inconsistent context.
  - Never log at translation boundary: rejected because diagnostics become incomplete.

## Decision 4: Unknown Failure Handling

- Decision: Any unmapped dependency failure is translated to a generic safe `AppError` while retaining wrapped internal error context via `Unwrap()`.
- Rationale: This enforces non-leakage and deterministic behavior under unforeseen failures.
- Alternatives considered:
  - Propagate raw errors on unknown cases: rejected because it violates contract and sanitization rules.
  - Drop internal cause entirely: rejected because it impairs diagnostics.

## Decision 5: Cross-Service Enforcement Strategy

- Decision: Enforce the same translation contract in synchronous and asynchronous paths across all backend services (`bff`, `bills`, `files`, `identity`, `onboarding`, `payments`).
- Rationale: A shared package-level standard must be uniformly applied to avoid fragmented behavior.
- Alternatives considered:
  - Apply only to HTTP/gRPC handlers first: rejected because background consumers would remain inconsistent.
  - Apply only to data repositories: rejected because service and transport boundaries would still leak raw dependency errors.

## Decision 6: Validation Approach

- Decision: Validation will include unit tests for translation and retryability semantics, plus integration verification for canonical backend integration test placement and naming conventions where behavior coverage is added.
- Rationale: This aligns with repository testing instructions while preserving deterministic CI behavior.
- Alternatives considered:
  - Manual validation only: rejected because it is not repeatable.
  - Integration-only validation: rejected because unit-level translation rules need faster deterministic feedback.
