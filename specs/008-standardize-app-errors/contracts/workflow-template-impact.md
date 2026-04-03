# Workflow Template Impact: AppError Standardization

## Decision

Workflow templates are impacted and now require explicit AppError propagation checks when backend work crosses service/repository/transport boundaries.

## Required Template Guidance

- Backend implementation tasks must include:
  - translation boundary update tasks,
  - non-leakage tests,
  - retryability/fallback verification,
  - governance memory and instruction synchronization tasks.
- Completion criteria must require:
  - successful cross-service boundary tests,
  - no raw dependency error leakage,
  - updated memory and instructions.

## Re-validation Summary

Validated that this feature's task plan includes and now completes:

1. Layer-boundary implementation tasks (`T013`-`T043`).
2. Contract/integration validation tasks (`T011`, `T012`, `T037`, `T038`, `T061`, `T063`, `T064`).
3. Governance sync tasks (`T047`-`T060`, `T065`, `T066`).

Validation gate note: `backend/scripts/validate_integration_test_conventions.sh` now enforces non-regression using `backend/scripts/integration_convention_known_failures.txt`, while `STRICT_CONVENTION_VALIDATION=1` keeps full-repository remediation measurable as follow-up work.

Validation evidence: `specs/008-standardize-app-errors/contracts/integration-convention-validation.md`.

## Impact Scope

- Specs: backend features that alter error paths.
- Plans: must include layer-boundary AppError propagation strategy.
- Tasks: must include explicit tests and governance updates.

## Conclusion

No template format change is required, but template usage must enforce AppError-first boundary work as a mandatory acceptance criterion for backend features.
