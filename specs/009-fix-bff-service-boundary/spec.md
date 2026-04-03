# Feature Specification: Enforce Service Boundary Contracts

**Feature Branch**: `[009-fix-bff-service-boundary]`  
**Created**: 2026-04-03  
**Status**: Draft  
**Input**: User description: "Refactor the BFF separation of concerns to remove HTTP view leakage into services, require HTTP-to-service mapping contracts, and define backend-wide pointer-based struct passing/return conventions with instruction and memory updates after implementation"

## Clarifications

### Session 2026-04-03

- Q: Which rule should define when a backend struct must be passed/returned by pointer? → A: Use pointer when struct has reference-like fields or size > 3 machine words.
- Q: What is the required rollout scope for removing HTTP-view leakage in BFF services? → A: Apply to all active BFF routes/services in this feature.
- Q: Which memory locations are mandatory to update at implementation completion? → A: Update both .specify/memory/* and /memories/repo/*.

## User Scenarios & Testing *(mandatory)*

<!--
  IMPORTANT: User stories should be PRIORITIZED as user journeys ordered by importance.
  Each user story/journey must be INDEPENDENTLY TESTABLE - meaning if you implement just ONE of them,
  you should still have a viable MVP (Minimum Viable Product) that delivers value.
  
  Assign priorities (P1, P2, P3, etc.) to each story, where P1 is the most critical.
  Think of each story as a standalone slice of functionality that can be:
  - Developed independently
  - Tested independently
  - Deployed independently
  - Demonstrated to users independently
-->

### User Story 1 - Enforce BFF Layer Contract (Priority: P1)

As a backend maintainer, I need the BFF HTTP layer and BFF service layer to use strictly separated contracts so that service logic remains independent from HTTP transport details.

**Why this priority**: This is the core architectural correction. If this boundary remains violated, layering rules are unreliable and future features will continue to reintroduce coupling.

**Independent Test**: Can be fully tested by inspecting all BFF services and confirming they do not import or depend on HTTP request/response view models while all route/controller behavior, status semantics, and response contract meaning remain functionally equivalent.

**Acceptance Scenarios**:

1. **Given** an existing BFF service method, **When** code dependencies are analyzed, **Then** no HTTP transport request/response view type is referenced in service method inputs, outputs, or internal orchestration.
2. **Given** an HTTP endpoint mapped to a BFF service, **When** a request is processed, **Then** the HTTP layer maps transport-facing structures into service-facing contracts before service invocation and maps service results back into HTTP responses.
3. **Given** an endpoint behavior contract (status code + response shape), **When** controller/service boundary refactors are applied, **Then** endpoint status semantics and response contract meaning remain unchanged for supported cases.
4. **Given** a feature where proto domain messages are insufficient for service logic, **When** new service contracts are introduced, **Then** those contracts are defined in the service boundary and not in HTTP transport packages.

---

### User Story 2 - Reduce Struct Copy Overhead (Priority: P2)

As a backend maintainer, I need consistent pointer-based struct passing and returning conventions across backend layers so that unnecessary value copying is reduced and mutation semantics are explicit.

**Why this priority**: This improves performance predictability and consistency across services while preventing accidental large-value copying across call chains.

**Independent Test**: Can be fully tested by static review of service/repository/controller contracts and representative flows, confirming pointer semantics are used according to the defined convention and no behavioral regressions are introduced.

**Acceptance Scenarios**:

1. **Given** backend function contracts that pass or return domain structs, **When** conventions are applied, **Then** pointer semantics are used for non-trivial structs crossing function boundaries unless an explicit immutability/value-copy exception is documented.
2. **Given** existing backend modules, **When** refactoring is complete, **Then** contract signatures follow a single documented pointer usage policy that can be applied consistently to new features.

---

### User Story 3 - Preserve Refactor Rules in Project Guidance (Priority: P3)

As a team lead, I need repository instructions and memory artifacts updated after implementation so that future contributors preserve the same layer-separation and pointer-convention standards.

**Why this priority**: Long-term consistency requires explicit and discoverable guidance, otherwise the same architectural leaks can reappear.

**Independent Test**: Can be fully tested by reviewing updated instruction documents and repository memory artifacts to confirm they encode enforceable, deterministic rules aligned with the implemented refactor.

**Acceptance Scenarios**:

1. **Given** the refactor is implemented, **When** instruction files are reviewed, **Then** they explicitly require service-layer independence from HTTP views and define pointer contract expectations for backend function boundaries.
2. **Given** future feature planning artifacts are created, **When** memory files are consulted, **Then** they include updated flow and boundary guidance reflecting this refactor.

---

### Edge Cases

<!--
  ACTION REQUIRED: The content in this section represents placeholders.
  Fill them out with the right edge cases.
-->

- How is behavior preserved when a service currently depends on HTTP-only field names or optional transport-specific shapes that do not map one-to-one to existing proto messages?
- How are shared structs handled when multiple services consume the same data but currently use divergent value vs pointer signatures?
- How are nil-pointer and zero-value safety expectations validated after introducing pointer-based contracts?
- How are backward-compatible API responses ensured when mapper logic replaces direct view usage in services?

## Architecture & Memory Diagram Flow Impact *(mandatory)*

<!--
  ACTION REQUIRED: Determine whether this feature changes service flows,
  responsibilities, or cross-service communication. This section is mandatory.
-->

- Affected services: bff (mandatory), files, bills, identity, onboarding, payments (pointer convention standardization scope)
- Requires architecture diagram update (`.specify/memory/architecture-diagram.md`): Yes. Layer boundaries and mapping responsibilities in BFF are being corrected and must be reflected in architectural flow documentation.
- Required service-flow file updates in `.specify/memory/`:
  - [x] `.specify/memory/bff-flows.md` (BFF request flow and mapper responsibilities)
  - [x] `.specify/memory/files-service-flows.md` (pointer contract conventions for service/repository boundaries)
  - [x] `.specify/memory/bills-service-flows.md` (pointer contract conventions for service/repository boundaries)
  - [x] `.specify/memory/identity-service-flows.md` (pointer contract conventions for service/repository boundaries)
  - [x] `.specify/memory/onboarding-service-flows.md` (pointer contract conventions for service/repository boundaries)
  - [x] Other impacted memory file(s): `.specify/memory/payments-service-flows.md`
- Required repository memory updates in `/memories/repo/`:
  - [x] Add or update backend boundary-convention note(s) capturing BFF transport/service contract separation.
  - [x] Add or update pointer-policy convention note(s) for backend function signature guidance.
- If none are impacted, include explicit no-impact rationale: Not applicable because multiple service flows and cross-layer responsibilities are explicitly in scope.

## Instruction Impact *(mandatory for refactor/reorganization)*

<!--
  ACTION REQUIRED: Fill this section whenever the feature refactors or reorganizes
  project/service structure, module boundaries, layering, or implementation patterns.
-->

- Is this feature a refactor/reorganization?: Yes
- If Yes, impacted instruction files under `.github/instructions/`:
  - [x] `.github/instructions/architecture.instructions.md`
  - [x] `.github/instructions/project-structure.instructions.md`
  - [x] `.github/instructions/golang.instructions.md`
  - [x] `.github/instructions/coding-conventions.instructions.md`
  - [x] `.github/instructions/testing.instructions.md`
  - [x] `.github/instructions/ai-behavior.instructions.md`
- If Yes, impacted workflow templates under `.specify/templates/`:
  - [ ] `.specify/templates/spec-template.md`
  - [ ] `.specify/templates/plan-template.md`
- Pattern-preservation statement: Repository instructions will codify that BFF transport contracts are isolated from service contracts, and backend function signatures crossing layer boundaries use documented pointer conventions by default. Repository memory flow artifacts will be updated to reinforce this policy during future feature planning and implementation.
- If backend behavior/integration flows are in scope, include explicit reference to canonical integration-test standard: `backend/tests/integration/<service>/` or `backend/tests/integration/cross_service/`, behavior-based snake_case filenames, and table-driven BDD Given/When/Then + AAA.

## Requirements *(mandatory)*

<!--
  ACTION REQUIRED: The content in this section represents placeholders.
  Fill them out with the right functional requirements.
-->

### Functional Requirements

- **FR-001**: The BFF service layer for all active BFF routes/services MUST NOT depend on HTTP transport request/response view models for method inputs, outputs, or internal orchestration.
- **FR-002**: The HTTP transport layer MUST perform explicit mapping between HTTP views and service-layer contracts (proto domain messages and/or service-owned structs) before invoking service methods.
- **FR-003**: When proto domain messages do not satisfy service input or output needs, the service layer MUST define and own supplemental service contracts independent of HTTP transport packages.
- **FR-004**: The refactor MUST preserve existing externally observable API behavior for supported endpoints, including status semantics and response contract meaning.
- **FR-004a**: Verification of endpoint behavior preservation MUST include assertions for status semantics and response shape/fields, not only route registration or reachability.
- **FR-005**: Backend modules MUST adopt a single documented convention for pointer-based struct passing and returning across function boundaries.
- **FR-005a**: A struct MUST be passed/returned by pointer when it has reference-like fields (slice, map, chan, func, interface, pointer-containing composites) or when its value size exceeds 3 machine words.
- **FR-006**: Any intentional exceptions to pointer-based conventions MUST be explicitly documented with rationale (for example, immutability or tiny value-object semantics).
- **FR-007**: Refactored boundary code MUST avoid introducing nil-dereference regressions by defining and enforcing clear nil-handling expectations at mapper and service boundaries.
- **FR-007a**: Refactored mapper and service boundaries MUST include explicit tests for nil and empty boundary inputs/outputs where applicable.
- **FR-008**: Project instruction files MUST be updated at implementation completion to codify the layer-boundary and pointer-convention rules.
- **FR-009**: Both `.specify/memory/*` flow artifacts and `/memories/repo/*` repository memory notes MUST be updated at implementation completion to preserve these patterns for future features.
- **FR-010**: Integration and unit test coverage for all active BFF routes/services MUST demonstrate that route/controller behavior remains valid while service contracts are decoupled from HTTP views.

### Key Entities *(include if feature involves data)*

- **Transport View Contract**: HTTP-facing request/response shape used only at transport/controller level for protocol interaction.
- **Service Contract**: Service-owned input/output structure used for business orchestration, independent of HTTP concerns.
- **Proto Domain Message**: Inter-service or domain payload shared through proto contracts, eligible for service boundary usage when semantically sufficient.
- **Boundary Mapper**: Mapping responsibility in transport layer that transforms transport view contracts into service contracts and service results into transport responses.
- **Pointer Convention Policy**: Repository-wide contract definition: pass/return by pointer when a struct has reference-like fields or its size exceeds 3 machine words; value semantics are allowed only for documented exceptions.

## Success Criteria *(mandatory)*

<!--
  ACTION REQUIRED: Define measurable success criteria.
  These must be technology-agnostic and measurable.
-->

### Measurable Outcomes

- **SC-001**: 100% of BFF service modules for all active BFF routes/services show zero direct dependency on HTTP transport view contracts.
- **SC-002**: 100% of active BFF HTTP endpoints use explicit, test-validated mapping between transport views and service contracts.
- **SC-003**: 100% of scoped backend contract signatures for non-trivial structs conform to the defined pointer convention or include an explicit documented exception.
- **SC-004**: 0 critical behavior regressions are detected in route-level integration tests and scoped service unit tests after refactor completion.
- **SC-004**: 0 critical behavior regressions are detected in route-level integration tests and scoped service unit tests after refactor completion, including status semantics and response shape assertions.
- **SC-005**: All identified instruction and memory artifacts listed in this spec are updated and reviewed in the same implementation feature scope.
- **SC-006**: 100% of mapper/service boundary nil-safety tests added in scope pass with no nil-dereference regressions.

## Assumptions

<!--
  ACTION REQUIRED: The content in this section represents placeholders.
  Fill them out with the right assumptions based on reasonable defaults
  chosen when the feature description did not specify certain details.
-->

- Existing endpoint behavior and consumer expectations are the baseline contract and must remain stable unless explicitly approved in a separate feature.
- Pointer-convention standardization uses this threshold: pointer semantics for structs with reference-like fields or value size > 3 machine words; value semantics remain allowed only as documented exceptions.
- Teams accept incremental migration within the feature scope, but each touched module must be internally consistent with the new boundary rules.
- Current integration testing structure remains the verification baseline, and any new/updated tests follow existing repository standards.
