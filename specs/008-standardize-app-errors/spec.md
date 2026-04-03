# Feature Specification: Standardize Backend App Errors

**Feature Branch**: `008-standardize-app-errors`  
**Created**: 2026-04-03  
**Status**: Ready for Implementation  
**Input**: User description: "the backend project was created without having a default error definition, i have created in the backend/pkg/errors a struct called AppError and this AppError must be used in ALL THE BACKEND to delivery error between in layer to other layer. Normaly when we use a library for database or grpc we will have an errors that this library will bring to us, we canot pass this error to the others layers we must ALWAYS log this error using zap.Error(e) and pass to the others layers our AppError. To do that we i haved created a few errors in the consts.go file which you can use as example to create others errors for each case, so e must have ALL THE POSSIBLE ERRORS in this file pkg/errors/consts.go which we will use and re use in the backend, you must analyse that if the error could be a retrybale error or not, rigth now we cannot woried about to create the retry strategy, but we must start to use the rigth error in the rigth context because when we implement the retry strategy we already have the rigth error for each case."

## Clarifications

### Session 2026-04-03

- Q: What does "all possible errors" mean for the centralized error catalog scope? -> A: Cover all currently known backend error categories across all services, plus a mandatory generic unknown fallback.
- Q: Which retryability policy should be used when there is no explicit previous rule? -> A: Use transient-failure classification as retryable and deterministic/business-validation failures as non-retryable.
- Q: What logging requirement should apply before translating dependency-native errors? -> A: Log dependency-native errors once at the boundary where they are translated, with structured context and sanitized propagation.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Consistent Error Contract Across Layers (Priority: P1)

As a backend developer, I want all backend modules to return a shared application error contract when crossing layer boundaries so services and controllers can handle failures predictably.

**Why this priority**: This is the foundation for stable behavior across all backend services and prevents leaking infrastructure/library internals into business and transport layers.

**Independent Test**: Can be fully tested by invoking representative repository, service, and transport error paths and confirming only shared application errors are returned across layer boundaries.

**Acceptance Scenarios**:

1. **Given** a dependency failure inside a repository, **When** the error is returned to the service layer, **Then** the repository returns a shared application error and does not expose raw dependency errors.
2. **Given** a service failure returned to a controller or handler, **When** the error crosses the boundary, **Then** the service returns the shared application error contract with expected business-safe messaging.

---

### User Story 2 - Dependency Error Logging and Sanitization (Priority: P2)

As a backend operator, I want low-level dependency errors logged with structured context before translation so root causes remain diagnosable without exposing internals externally.

**Why this priority**: Operational visibility is required for support and incident response, while avoiding leakage of library-specific or infrastructure-sensitive details.

**Independent Test**: Can be tested by triggering representative database and gRPC failures and verifying low-level error details are logged internally while only shared application errors are propagated.

**Acceptance Scenarios**:

1. **Given** a database client error, **When** the operation fails, **Then** the original failure is logged with structured fields and the propagated error is the shared application error.
2. **Given** a gRPC client/server error, **When** the operation fails, **Then** the original failure is logged with structured fields and the propagated error is the shared application error.

---

### User Story 3 - Retryability Classification for Future Policies (Priority: P3)

As a platform maintainer, I want each reusable application error definition to include retryability intent so future retry strategies can be enabled without reclassifying existing failures.

**Why this priority**: Correct retryability classification now reduces future migration risk and avoids broad rework when retry orchestration is introduced.

**Independent Test**: Can be tested by reviewing the centralized error catalog and validating that each known error case has a defined retryability classification.

**Acceptance Scenarios**:

1. **Given** a known transient failure case, **When** the corresponding application error is defined, **Then** the error is marked retryable.
2. **Given** a known non-transient failure case, **When** the corresponding application error is defined, **Then** the error is marked non-retryable.

### Edge Cases

- What happens when a raw dependency error is `nil` or missing details while creating an application error wrapper?
- How does the system behave when a failure category is encountered but no predefined application error constant exists yet?
- How are chained/wrapped dependency errors handled so internal diagnostics are preserved while external propagation stays standardized?
- What happens when different services currently classify the same failure type inconsistently?

## Architecture & Memory Diagram Flow Impact *(mandatory)*

- Affected services: bff, bills, files, identity, onboarding, payments, shared backend packages.
- Requires architecture diagram update (`.specify/memory/architecture-diagram.md`): No. This feature standardizes error semantics and propagation behavior but does not alter service topology, communication channels, or ownership boundaries.
- Required service-flow file updates in `.specify/memory/`:
  - [x] `.specify/memory/bff-flows.md` (error propagation expectations across BFF layers)
  - [x] `.specify/memory/files-service-flows.md` (repository/service error translation expectations)
  - [x] `.specify/memory/bills-service-flows.md` (repository/service error translation expectations)
  - [x] `.specify/memory/identity-service-flows.md` (identity service error handling expectations)
  - [x] `.specify/memory/onboarding-service-flows.md` (onboarding service error handling expectations)
  - [ ] Other impacted memory file(s): none identified.
- If none are impacted, include explicit no-impact rationale:
  Service topology is unchanged; only error-handling semantics and cross-layer contracts are standardized.

## Instruction Impact *(mandatory for refactor/reorganization)*

- Is this feature a refactor/reorganization?: Yes (cross-cutting backend error-handling pattern standardization).
- If Yes, impacted instruction files under `.github/instructions/`:
  - [x] `.github/instructions/observability.instructions.md`
  - [x] `.github/instructions/security.instructions.md`
  - [x] `.github/instructions/golang.instructions.md`
  - [x] `.github/instructions/ai-behavior.instructions.md`
  - [x] `.github/instructions/architecture.instructions.md`
- If Yes, impacted workflow templates under `.specify/templates/`:
  - [ ] none required.
- Pattern-preservation statement: Backend contributors must always log dependency-native errors internally and propagate only the shared application error contract between layers, with reusable centralized error definitions and explicit retryability classification.
- If backend behavior/integration flows are in scope, include explicit reference to canonical integration-test standard:
  `backend/tests/integration/<service>/` or `backend/tests/integration/cross_service/`, behavior-based snake_case filenames, and table-driven BDD Given/When/Then + AAA.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The backend system MUST use the shared application error type as the only propagated error contract across backend layer boundaries.
- **FR-002**: The backend system MUST maintain a centralized reusable application error catalog that covers all currently known backend failure categories across all services and includes a mandatory generic unknown fallback entry.
- **FR-003**: The backend system MUST classify each reusable application error as retryable or non-retryable using a shared policy: transient infrastructure failures are retryable, deterministic/business-validation failures are non-retryable.
- **FR-004**: The backend system MUST log dependency-native errors exactly once at the translation boundary using structured context before translating them into application errors.
- **FR-005**: The backend system MUST prevent propagation of raw dependency/library error messages outside the layer where they originate.
- **FR-006**: The backend system MUST define translation rules from dependency/library failures to the centralized application error catalog for each backend module.
- **FR-007**: The backend system MUST provide deterministic behavior for unknown or unmapped failures by translating them to a safe generic application error.
- **FR-008**: The backend system MUST preserve internal diagnosability by retaining original failures via standard wrapping compatible with `errors.Is`/`errors.As` while exposing only safe application-error messaging to upper layers.
- **FR-009**: The backend system MUST apply the same error translation and classification policy across synchronous flows (HTTP and gRPC request paths); asynchronous flows (message consumer/producer paths) are covered for files service in MVP and expanded to additional services in future phases.
- **FR-010**: The backend system MUST include coverage that verifies application error propagation, retryability classification, and non-leakage of dependency-native errors.

### Key Entities *(include if feature involves data)*

- **Application Error**: Shared backend error contract containing safe message, retryability flag, and optional wrapped original error for internal diagnosis.
- **Error Catalog Entry**: Reusable predefined error definition mapped to a known business or technical failure category, including retryability intent.
- **Error Translation Rule**: Policy that maps dependency-native failures to catalog entries for each layer boundary.
- **Error Event Log Record**: Structured operational log event capturing failure context and original low-level error metadata for troubleshooting.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of backend layer-boundary failure paths propagate the shared application error contract rather than dependency-native errors.
- **SC-002**: 100% of currently known backend failure categories are cataloged with explicit retryability classification and one mandatory unknown-fallback category is present.
- **SC-003**: In sampled failure traces for each backend service, 100% of dependency-native failures are present in structured logs and absent from externally propagated error messages.
- **SC-004**: CI verification reports zero dependency-native error payload leakage across covered layer-boundary tests.

## Assumptions

- Existing services already have structured logging available in operational paths.
- Existing backend error constants represent only an initial baseline and will be expanded to cover additional known failure categories.
- This feature standardizes contracts and behavior first; automated retry execution policies are out of scope for this iteration.
- Existing backend tests and quality gates will be extended as needed to validate error propagation and classification behavior.
