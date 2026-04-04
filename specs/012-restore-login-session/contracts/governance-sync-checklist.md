# Governance Sync Checklist

These sync items are part of the completion gate for `012-restore-login-session`.

## Memory sync

- [ ] Update `.specify/memory/architecture-diagram.md` to show the seeded owner login path and the downstream `Session` + `Pagination` propagation.
- [ ] Update `.specify/memory/bff-flows.md` with the restored login/refresh flow and BFF-side default pagination behavior.
- [ ] Update `.specify/memory/identity-service-flows.md` with the persistent bootstrap credential and token issuance path.
- [ ] Update `.specify/memory/onboarding-service-flows.md` with the seeded owner membership and project-permission path.
- [ ] Update any impacted `files`, `bills`, and `payments` service-flow files for session-carrying protected requests.
- [ ] Update `/memories/repo/bff-service-boundary-conventions.md` if implementation tightens or clarifies the `Session`/`Pagination` rule.

## Instruction sync

- [ ] Re-check `.github/instructions/architecture.instructions.md` so the BFF gateway and `common.v1.Session` rule stays explicit.
- [ ] Re-check `.github/instructions/ai-behavior.instructions.md` if the bootstrap-login and session-propagation convention needs clearer deterministic wording.
- [ ] Re-check `.github/instructions/testing.instructions.md` so login restoration and session-propagation regression coverage stays encoded in the canonical integration-test guidance.

## Verification sync

- [ ] Keep BFF route registration and reachability tests in `backend/tests/integration/bff/` with snake_case names and BDD + AAA structure.
- [ ] Keep cross-service auth/session tests in `backend/tests/integration/cross_service/`.
- [ ] Add or refresh identity-focused integration coverage in `backend/tests/integration/identity/`.
- [ ] Record fresh evidence for `make proto/generate`, `cd backend && go test ./...`, `make test/integration`, and `make frontend/test` before closing the feature.
