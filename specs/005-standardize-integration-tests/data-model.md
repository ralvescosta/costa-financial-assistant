# Data Model: Standardize Integration Test System

## Entity: IntegrationTestStandard
- Description: Canonical ruleset for backend integration tests.
- Fields:
  - `id` (string): Stable identifier for the standard version (e.g., `integration-test-standard-v1`).
  - `scope` (enum): `backend_integration`.
  - `directory_layout` (object): Canonical placement rules.
  - `filename_pattern` (string): Regex-compatible naming policy for test files.
  - `bdd_structure` (object): Required Given/When/Then representation in table-driven tests.
  - `approved_libraries` (string[]): Allowed test libraries and runtime helpers.
  - `compliance_checks` (string[]): Required review-time checks.
  - `effective_date` (date): Date when the standard becomes mandatory.

## Entity: TestSuiteSegment
- Description: Logical partition of integration tests by ownership.
- Fields:
  - `segment_key` (string): Service name or `cross_service`.
  - `path` (string): Directory path under `backend/tests/integration/`.
  - `ownership_type` (enum): `service_owned` or `cross_service`.
  - `responsible_team` (string): Owning domain/team label.
- Relationships:
  - One `IntegrationTestStandard` defines many `TestSuiteSegment` records.

## Entity: IntegrationScenario
- Description: Behavior scenario represented as a `t.Run` case.
- Fields:
  - `scenario_name` (string): Human-readable BDD scenario title.
  - `given` (string): Preconditions/context.
  - `when` (string): Triggering action.
  - `then` (string): Expected outcome.
  - `tags` (string[]): Optional labels (`security`, `idempotency`, `contract`, etc.).
  - `cleanup_strategy` (string): Fixture/data cleanup responsibility.
- Validation Rules:
  - `scenario_name`, `given`, `when`, `then` are required.
  - Scenario must be executable in isolation with deterministic setup/teardown.
- Relationships:
  - Many `IntegrationScenario` belong to one `TestSuiteSegment`.

## Entity: MigrationMapping
- Description: Traceability map from legacy integration tests to standardized form.
- Fields:
  - `legacy_path` (string): Original file location.
  - `legacy_name` (string): Original filename.
  - `new_path` (string): New canonical directory path.
  - `new_name` (string): New behavior-based snake_case filename.
  - `status` (enum): `planned`, `moved`, `verified`.
  - `coverage_note` (string): Coverage parity statement.
- Validation Rules:
  - Every in-scope legacy file must have one mapping record.
  - `status=verified` requires successful test execution and review sign-off.

## Entity: ComplianceRule
- Description: Enforceable rule used during PR review and task acceptance.
- Fields:
  - `rule_id` (string): Stable identifier (e.g., `IT-001`).
  - `description` (string): Human-readable rule statement.
  - `severity` (enum): `must` or `should`.
  - `verification_method` (enum): `automated`, `manual`, `hybrid`.
  - `evidence` (string): Expected proof (file diff, test output, checklist item).
- Relationships:
  - One `IntegrationTestStandard` defines many `ComplianceRule` entries.

## State Transitions

### MigrationMapping.status
- `planned` -> `moved` -> `verified`
- Transition constraints:
  - `planned` -> `moved`: file physically renamed/moved to canonical structure.
  - `moved` -> `verified`: behavior parity validated and CI tests pass.

### IntegrationTestStandard lifecycle
- `draft` -> `approved` -> `enforced`
- Transition constraints:
  - `draft` -> `approved`: constitution + instruction updates prepared.
  - `approved` -> `enforced`: governance merged and referenced in feature workflows.
