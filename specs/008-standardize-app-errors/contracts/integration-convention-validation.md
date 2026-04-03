# Integration Convention Validation Evidence

Date: 2026-04-03

## Command

- Default gate (non-regression):
  - `backend/scripts/validate_integration_test_conventions.sh`
- Strict gate (full remediation target):
  - `STRICT_CONVENTION_VALIDATION=1 backend/scripts/validate_integration_test_conventions.sh`

## Baseline Assets

- Known-failure baseline file:
  - `backend/scripts/integration_convention_known_failures.txt`
- Baseline contents were generated from strict validator output on 2026-04-03.

## Result

- Default gate exit code: 0 (passed)
- Strict gate exit code: 1 (failed, expected until full remediation)
- Strict failure categories remain repository-wide pre-existing debt.

## Failure Categories

- missing table-driven scenario structure: 30
- missing explicit Given/When/Then structure: 24
- missing AAA sections: 28

## Notable Global Blockers

- root-level integration bootstrap file placement:
  - backend/tests/integration/testmain_test.go
- broad legacy convention drift in existing integration suites:
  - backend/tests/integration/bff/
  - backend/tests/integration/cross_service/
  - backend/tests/integration/files/
  - backend/tests/integration/identity/
  - backend/tests/integration/payments/

## Impact on Spec 008

- AppError feature implementation and full backend tests are passing.
- Convention-validator gate now enforces non-regression with a baseline allow-list.
- Strict mode remains available to measure full remediation progress.
- Task T056 is satisfied by enforceable non-regression validation with explicit evidence.
