# Contract: Backend Integration Test Conventions

## 1. Purpose
This contract defines the required conventions for all backend integration tests in this repository.

## 2. Scope
- Applies to: `backend/tests/integration/**`
- Applies to services: `bff`, `bills`, `files`, `identity`, `onboarding`, `payments`
- Applies to cross-service workflows: `backend/tests/integration/cross_service/**`

## 3. Directory Contract
- Service-owned behavior tests MUST be located at:
  - `backend/tests/integration/<service>/`
- Multi-service behavior tests MUST be located at:
  - `backend/tests/integration/cross_service/`
- Files that do not belong to a service or cross-service bucket are non-compliant.

## 4. File Naming Contract
- Pattern: behavior-based snake_case ending in `_test.go`
- Canonical regex: `^[a-z0-9]+(_[a-z0-9]+)*_test\.go$`
- Examples:
  - `create_bill_success_test.go`
  - `reject_invalid_token_test.go`
  - `enforce_project_isolation_test.go`
- Forbidden patterns:
  - Story-ID-first names (e.g., `us1_upload_classify_test.go`)
  - Generic or non-behavior names (e.g., `integration_test.go`, `misc_test.go`)

## 5. Scenario Structure Contract (BDD)
- Integration tests MUST use table-driven `t.Run` scenarios.
- Each scenario MUST include explicit `given`, `when`, and `then` fields or sections.
- Scenario names MUST describe behavior outcome in plain language.
- Each scenario body MUST follow explicit AAA organization:
  - Arrange
  - Act
  - Assert

## 6. Library Contract
- Approved stack:
  - `testing` (stdlib)
  - `github.com/stretchr/testify`
  - `github.com/testcontainers/testcontainers-go`
- Additional libraries require explicit architecture/testing approval.

## 7. Lifecycle Contract
- Database-dependent integration suites MUST use ephemeral DB lifecycle via `TestMain`.
- Required flow:
  1. Provision isolated DB/container
  2. Apply migrations
  3. Run suite
  4. Teardown resources
- Tests MUST be deterministic and must not rely on pre-existing local state.

## 8. Migration Contract for Legacy Tests
- Every legacy file must be mapped from old path/name to new canonical path/name.
- Migration must preserve behavior coverage.
- Mapping status must be auditable (`planned`, `moved`, `verified`).

## 9. Compliance Contract (PR Review)
A change is compliant only if all checks pass:
1. File placement follows service or cross-service structure.
2. Filename matches behavior-based snake_case convention.
3. Scenarios use table-driven `t.Run` with explicit `given/when/then`.
4. AAA organization is readable in each scenario.
5. Approved library stack is used.
6. DB setup/cleanup is deterministic and isolated.

## 10. Versioning
- Contract version: `v1`
- Any breaking change to placement/naming/structure rules requires constitution and instruction updates in the same change set.

## 11. Canonical Maintainer Checklist

- [ ] Test file is under `backend/tests/integration/<service>/` or `backend/tests/integration/cross_service/`
- [ ] Filename follows `^[a-z0-9]+(_[a-z0-9]+)*_test\.go$`
- [ ] No legacy `us*_*.go` naming remains
- [ ] Scenarios are table-driven with explicit `given/when/then`
- [ ] Scenario bodies are readable with explicit `Arrange`, `Act`, `Assert`
- [ ] Database lifecycle uses deterministic `TestMain` + migration setup + teardown
- [ ] Mapping artifacts are updated (`migration-mapping.md`, `migration-baseline.md`)
