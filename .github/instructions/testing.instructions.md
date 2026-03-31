---
applyTo: "**/*_test.go,**/*.test.ts,**/*.spec.ts"
---

# Testing Instructions

## Rule: BDD Scenarios + AAA Blocks (Mandatory)

**Description**: Unit tests must be designed from feature intent (behavioral contract), expressed as BDD scenarios, and implemented with explicit AAA blocks.

**When it applies**: Creating or modifying any unit tests (`*_test.go`).

**Copilot MUST**:
- Start by identifying the feature behavior / business rule / interface contract being validated (not just enumerating `if/else` branches).
- Write scenario names using BDD semantics (e.g., `Given... When... Then...`) so intent is visible in CI output.
- For each `t.Run(...)` scenario, include a required comment block:
  - `// Given ...`
  - `// When ...`
  - `// Then ...`
- For each `t.Run(...)` scenario, implement the AAA approach with explicit sections:
  - **Arrange (Arranje)**: build inputs + fakes/mocks
  - **Act**: call the unit under test
  - **Assert (Asert)**: verify outputs + observable side effects

**Copilot MUST NOT**:
- Generate tests by only mirroring the current control flow (“covers the if/else”) without stating the business scenario and expected outcome.
- Hide Arrange/Act/Assert steps in implicit setup that makes the scenario hard to read.

**Example input → expected Copilot output**:
- Input: "Add tests for document service upload."
- Expected output: scenarios named like `GivenNilDocument WhenUpload ThenReturnsError`, each with `// Given/When/Then` comments and explicit Arrange/Act/Assert blocks in `backend/internals/files/services/upload_service_test.go`.

---

## Rule: Test Business Paths and Failure Paths

**Description**: Tests must verify both successful and failing behavior for business logic.

**When it applies**: Creating or modifying tests for services, operations, and repositories.

**Copilot MUST**:
- Cover happy path, validation failures, dependency failures, and idempotency/duplicate behavior.
- Assert returned errors and critical side effects.
- Keep tests deterministic and isolated.
- Use BDD scenario naming and required Given/When/Then + AAA blocks (see rule above).

**Copilot MUST NOT**:
- Write happy-path-only tests for non-trivial logic.
- Depend on network or unstable external systems.
- Hide flaky behavior with loose assertions.
- Write tests that only mirror branches without domain intent.

**Example input → expected Copilot output**:
- Input: "Add tests for document service upload."
- Expected output: test nil document, duplicate file hash, storage failure, and successful upload in `backend/internals/files/services/upload_service_test.go`.

---

## Rule: Table-Driven Tests by Default

**Description**: Similar test scenarios should use table-driven style.

**When it applies**: Multiple scenarios for the same function/method.

**Copilot MUST**:
- Use a case table with explicit names.
- Run each case with `t.Run(...)`.
- Use BDD scenario names (e.g., `Given... When... Then...`) for each case.
- Include `// Given/When/Then` comments and explicit AAA blocks inside each scenario.

**Copilot MUST NOT**:
- Duplicate nearly identical test functions.
- Mix many unrelated scenarios into one unreadable case.
- Hide expectations in implicit setup.

**Example input → expected Copilot output**:
- Input: "Test DE builder subfield parser with many inputs."
- Expected output: a table-driven test under `backend/internals/<service>/services/` with named BDD scenarios.

---

## Rule: Mocking with uber/mock

**Description**: Backend unit tests MUST use `go.uber.org/mock` for interface mocking.

**When it applies**: Testing services or consumers that depend on repositories, gRPC clients, or other interfaces.

**Copilot MUST**:
- Generate mocks with `mockgen` targeting the interface definition file.
- Place generated mock files alongside the interface or in a `mocks/` subdirectory within the same package.
- Mock interfaces, not concrete implementations.
- Verify key collaborator interactions where behavior requires it.
- Place mock setup in **Arrange** and interaction verification in **Assert** (AAA).

**Copilot MUST NOT**:
- Over-mock internal value objects.
- Assert implementation details unrelated to observable outcomes.
- Introduce brittle expectation chains that block refactoring.

**Example input → expected Copilot output**:
- Input: "Unit test document service storage failure."
- Expected output: mock the storage client interface with `gomock`, force a storage error, assert returned error and no document record is persisted.

---

## Rule: Test Data Placement

**Description**: Shared fixtures must live in stable fixture directories.

**When it applies**: Adding reusable input/output payloads for tests.

**Copilot MUST**:
- Place reusable fixtures in `data/`.
- Keep fixture names domain-specific and version-stable.
- Generate simple inline test data programmatically when fixtures add no value.

**Copilot MUST NOT**:
- Copy large JSON payloads into test function bodies.
- Use production secrets or sensitive records in fixtures.
- Scatter duplicate fixture files across unrelated folders.

**Example input → expected Copilot output**:
- Input: "Reuse sample bill PDF payload in multiple tests."
- Expected output: add fixture in `backend/tests/data/` and load it from a test helper function.

---

## Rule: Integration Tests with Ephemeral DB Lifecycle

**Description**: Backend integration tests that require a database must provision, migrate, test, and destroy an isolated DB instance within `TestMain`.

**When it applies**: Creating integration test files under `backend/tests/integration/`.

**Copilot MUST**:
- Implement `TestMain(m *testing.M)` in `backend/tests/integration/testmain_test.go` to manage the full DB lifecycle: provision → `migrate up` → `m.Run()` → teardown.
- Use a separate ephemeral database (e.g., Docker-started container or in-process test DB) that is not shared with other test suites.
- Run all pending migrations against the ephemeral DB before any test function executes.
- Ensure teardown runs even when tests fail (`defer`).

**Copilot MUST NOT**:
- Share a persistent test database across test suite runs.
- Leave orphaned test containers or database state after the suite completes.
- Depend on an already-running production or development database.

**Example input → expected Copilot output**:
- Input: "Add BFF integration test for upload classify flow."
- Expected output: test in `backend/tests/integration/us1_upload_classify_test.go` relies on the ephemeral DB started in `testmain_test.go`; no local DB dependency assumed.

---

## Rule: CI Test Command Compatibility

**Description**: New tests must run under repository CI commands.

**When it applies**: Creating tests or test helpers.

**Copilot MUST**:
- Keep tests compatible with `go test ./...`.
- Keep package boundaries and imports go-test friendly.
- Avoid assumptions that only work in local IDE execution.

**Copilot MUST NOT**:
- Require manual pre-steps not represented in CI.
- Depend on local machine state/files not included in repo.
- Introduce tests that require non-deterministic timing.

**Example input → expected Copilot output**:
- Input: "Add integration-like behavior test."
- Expected output: provide a deterministic unit-level test or clearly isolate integration setup via `TestMain` ephemeral DB lifecycle so `make test/integration/<service>` remains stable.

---

## Rule: Frontend Hook Tests with Vitest (BDD + Triple-A)

**Description**: Frontend tests are hook-only, written with Vitest, and MUST follow BDD scenario naming with Triple-A structure.

**When it applies**: Creating or modifying any `*.test.ts` / `*.spec.ts` frontend files.

**Copilot MUST**:
- Place all test files alongside hooks in `frontend/src/hooks/` with a `.test.ts` suffix.
- Use Vitest (`describe`, `it`, `expect`) for all assertion and test runner needs.
- Name test cases using BDD semantics: `given <precondition>, when <action>, then <outcome>`.
- Include explicit `// Arrange`, `// Act`, `// Assert` comment sections within each `it(...)` block.
- Mock API calls via `@tanstack/react-query` testing utilities or `vi.fn()` — never make real HTTP calls in tests.

**Copilot MUST NOT**:
- Write tests for React page or component render trees (no component/snapshot tests).
- Test implementation details of internal hook state not observable via the hook's return value.
- Import server-side code into frontend tests.

**Example input → expected Copilot output**:
- Input: "Add hook test for upload document flow."
- Expected output: test file in `frontend/src/hooks/useUploadDocument.test.ts` with `describe` blocks using BDD names, `// Arrange / Act / Assert` sections, and mocked API mutation.