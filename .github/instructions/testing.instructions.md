---
applyTo: "**/*_test.go"
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
- Input: "Add tests for enrichment service Process."
- Expected output: scenarios named like `GivenNilMessage WhenProcess ThenReturnsError`, each with `// Given/When/Then` comments and explicit Arrange/Act/Assert blocks.

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
- Input: "Add tests for enrichment service Process."
- Expected output: test nil message, already processed, operation factory failure, and successful operation in `internal/services/enrichment/service_test.go`.

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
- Expected output: a table-driven test under `internal/services/debuilder/internal/.../*_test.go` with named scenarios.

---

## Rule: Dependency Mocking Strategy

**Description**: Mock external boundaries and keep unit tests focused.

**When it applies**: Testing code that depends on repositories, clients, producers, or config providers.

**Copilot MUST**:
- Mock interfaces, not concrete implementations.
- Verify key collaborator interactions where behavior requires it.
- Keep mock expectations aligned with observable behavior.
- Place mock setup in **Arrange** and interaction verification in **Assert** (AAA), with the scenario described via `// Given/When/Then`.

**Copilot MUST NOT**:
- Over-mock internal value objects.
- Assert implementation details unrelated to outcomes.
- Introduce brittle expectation chains that block refactoring.

**Example input → expected Copilot output**:
- Input: "Unit test enrichment operation publish failure."
- Expected output: mock producer dependency, force publish error, assert returned error and no extra side effects.

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
- Input: "Reuse Mastercard payload in multiple tests."
- Expected output: add fixture in `data/` and load it from helper in test file.

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
- Expected output: provide deterministic unit-level test or clearly isolate integration setup so `make testing` remains stable.