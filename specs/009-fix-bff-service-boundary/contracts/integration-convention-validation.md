# Integration Convention Validation

## Command

```bash
backend/scripts/validate_integration_test_conventions.sh
```

## Result

- Status: PASS
- Summary: `integration test convention validation passed`
- Notes: Executed during Phase 6 after boundary and governance sync tasks.

## Canonical Smoke Test Verification

- Target file: `backend/tests/integration/bff/bff_route_registration_smoke_test.go`
- Placement: `backend/tests/integration/bff/` (canonical)
- Naming convention: `bff_route_registration_smoke_test.go` (behavior-based snake_case)
- BDD + AAA comments in scenarios: Verified manually in Phase 7
