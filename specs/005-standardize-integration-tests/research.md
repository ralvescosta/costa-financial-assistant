# Research: Standardize Integration Test System

## Decision 1: Canonical integration test folder layout
- Decision: Place service-owned tests under `backend/tests/integration/<service>/` and multi-service workflows under `backend/tests/integration/cross_service/`.
- Rationale: This preserves domain ownership, matches backend service boundaries, and removes ambiguity for cross-service scenarios.
- Alternatives considered:
  - User-story-only folders (`us1`, `us2`): rejected because story IDs drift over time and obscure ownership.
  - Transport-first folders (`http`, `grpc`, `rmq`): rejected because single behavior often crosses transport and domain boundaries.

## Decision 2: Canonical file naming
- Decision: Require behavior-based snake_case filenames, e.g., `create_bill_success_test.go`.
- Rationale: Behavior-oriented names are stable, searchable, and align with BDD readability requirements.
- Alternatives considered:
  - `usX_*.go` naming: rejected due to weak durability after backlog reshaping.
  - Endpoint-only naming (`post_bills_create_test.go`): rejected because it overfits transport details instead of behavior outcomes.

## Decision 3: Required BDD structure in Go integration tests
- Decision: Use table-driven `t.Run` scenarios with explicit `given`, `when`, and `then` fields or sections.
- Rationale: Preserves idiomatic Go testing while enforcing readable behavior contracts in CI output.
- Alternatives considered:
  - Ginkgo/Gomega `Describe/It`: rejected for this project because it adds a second test paradigm and migration overhead.
  - Free-form comments only: rejected because compliance is hard to review consistently.

## Decision 4: Approved integration test libraries
- Decision: Standardize on `testing`, `testify`, and `testcontainers-go`.
- Rationale: `testing` is native and deterministic, `testify` improves assertions and diagnostics, and `testcontainers-go` enables reproducible ephemeral dependencies.
- Alternatives considered:
  - Stdlib only: rejected because assertion ergonomics and setup utilities become inconsistent across contributors.
  - Fully open tooling: rejected because standardization goals require bounded choices.

## Decision 5: Ephemeral database lifecycle pattern
- Decision: Keep suite-level lifecycle in `backend/tests/integration/testmain_test.go`; provision isolated DB, run migrations, execute suite, and tear down deterministically.
- Rationale: Aligns with existing project rule for integration tests and prevents test state leakage.
- Alternatives considered:
  - Shared long-lived test DB: rejected due to flakiness and data contamination risk.
  - Per-test DB instance for every case: rejected as overly expensive for this suite scale.

## Decision 6: Migration strategy for existing tests
- Decision: Migrate in two passes: (1) structural relocation and rename with mapping table, (2) BDD/AAA normalization and helper extraction.
- Rationale: Separates mechanical moves from semantic refactors, reducing regression risk and easing code review.
- Alternatives considered:
  - Single massive rewrite: rejected due to low reviewability and higher break risk.
  - Only rules, no migration: rejected because current suite inconsistency would remain.

## Decision 7: Governance updates required for future compliance
- Decision: Update `.specify/memory/constitution.md`, `.github/instructions/testing.instructions.md`, and `.github/instructions/ai-behavior.instructions.md` to encode the new integration testing standard.
- Rationale: Durable compliance requires enforceable policy at constitution and AI instruction layers.
- Alternatives considered:
  - Add guidance only in feature docs: rejected because future features may bypass non-authoritative references.

## Decision 8: Compliance checks in review and task execution
- Decision: Define explicit review-time checks: folder placement, filename convention, BDD scenario naming/structure, approved stack usage, and deterministic cleanup.
- Rationale: Standardization fails without objective enforcement criteria.
- Alternatives considered:
  - Subjective review guidance: rejected due to inconsistent interpretation.
