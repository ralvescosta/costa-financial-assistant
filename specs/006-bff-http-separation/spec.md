# Feature Specification: BFF HTTP Boundary Separation

**Feature Branch**: `[006-bff-http-separation]`  
**Created**: 2026-04-02  
**Status**: Draft  
**Input**: User description: "i would like to review the spec 004-segregate-bff-routing, so i need you to also add a few other requiriments, this spec is to refactor the bff controller to also have a separation of concerc about the routes, but also i would like to add other sepration of concerns. The controllers must handle only the HTTP stuf the controller must not do any busninss logic, grpc client execution, the controller must only grab the request, validate the conteact, execute the service and based on the service response format a response for the http request. Also we must have a new layer in the transport/http called views this layer of views will be our http contracts structs, the structs that represent the jsons that will be transmited over the HTTP layer, so ALL the structs which represent the HTTP contract must live in this folder transport/http/views, and we MUST always have a golang go-playground/validator tags in each strucut field to allow us in the controller to perform a contract check"

## Clarifications

### Session 2026-04-02

- Q: Which BFF routes are in scope for this architectural standard? → A: All active BFF HTTP routes.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Keep Controllers Focused on HTTP Work (Priority: P1)

As a backend maintainer, I want BFF controllers to be limited to HTTP-specific responsibilities so that request handling stays predictable, reviewable, and free from business orchestration.

**Why this priority**: This is the main separation-of-concerns change. Without it, route separation alone still leaves controllers overloaded and difficult to maintain.

**Independent Test**: Can be fully tested by reviewing one refactored controller and confirming it only receives HTTP input, validates the request and context, invokes the service layer, and formats the HTTP response without embedding business rules or downstream client execution.

**Acceptance Scenarios**:

1. **Given** an incoming HTTP request for an existing BFF endpoint, **When** the controller handles it, **Then** it only parses request data, validates boundary inputs and context, calls the appropriate service, and translates the result into the HTTP response.
2. **Given** a controller that previously contained business decisions or downstream execution flow, **When** the refactor is complete, **Then** those responsibilities are no longer owned by the controller.

---

### User Story 2 - Centralize HTTP Contracts in a Dedicated View Layer (Priority: P2)

As a backend maintainer, I want all HTTP request and response contracts to live in a dedicated view layer so that transport payloads are easy to find, validate, and evolve consistently.

**Why this priority**: Stable and discoverable HTTP contracts reduce ambiguity between routes, controllers, and services, and make boundary validation consistent.

**Independent Test**: Can be independently tested by inspecting one feature module and confirming that every JSON request or response structure transmitted through the HTTP layer is defined in the view layer and carries field-level validation metadata.

**Acceptance Scenarios**:

1. **Given** an HTTP endpoint that accepts or returns JSON, **When** its boundary contracts are inspected, **Then** the transport-facing structs are defined in the dedicated `transport/http/views` layer rather than inside controllers or services.
2. **Given** a request contract used by a controller, **When** the controller performs boundary validation, **Then** the contract exposes validator tags for every field that requires validation.

---

### User Story 3 - Preserve Route Clarity and Coverage During Refactor (Priority: P3)

As a quality owner, I want route declarations, controllers, and HTTP contracts to remain clearly separated and fully covered by integration tests so that the refactor does not introduce routing regressions or hidden coupling.

**Why this priority**: Structural refactors are only safe if route behavior remains intact and test coverage makes missing or partially migrated endpoints visible.

**Independent Test**: Can be independently tested by inspecting the route modules for all active BFF endpoints and running integration tests that cover each declared route.

**Acceptance Scenarios**:

1. **Given** the full set of active BFF routes, **When** route modules are inspected, **Then** route registration remains separate from controller behavior and each route delegates to the appropriate controller capability.
2. **Given** the active BFF endpoints, **When** integration tests are executed, **Then** every declared route has at least one passing scenario that verifies expected accessibility and outcome.

### Edge Cases

- A controller receives a request whose boundary contract is missing required validation metadata; the gap must be detectable during review and treated as incomplete contract definition.
- A service returns a result that does not map cleanly to the HTTP contract; the controller must still remain limited to response translation rather than absorbing business decision logic.
- A route is moved into a dedicated route module but continues to rely on request or response structs defined outside the view layer; the refactor must treat that as incomplete separation.
- Existing authentication, authorization, and project-scoping behavior must remain intact after route registration, controller behavior, and HTTP contracts are separated.
- A route is declared but has no integration coverage after migration; the refactor must treat the route as not ready for completion.

## Architecture & Memory Diagram Flow Impact *(mandatory)*

- Affected services: bff
- Requires architecture diagram update (`.specify/memory/architecture-diagram.md`): No. This feature changes BFF transport-layer responsibility boundaries but does not introduce new cross-service communication topology.
- Required service-flow file updates in `.specify/memory/`:
	- [x] `.specify/memory/bff-flows.md` (if BFF impacted)
	- [ ] `.specify/memory/files-service-flows.md` (if Files impacted)
	- [ ] `.specify/memory/bills-service-flows.md` (if Bills impacted)
	- [ ] `.specify/memory/identity-service-flows.md` (if Identity impacted)
	- [ ] `.specify/memory/onboarding-service-flows.md` (if Onboarding impacted)
	- [ ] Other impacted memory file(s): `.specify/memory/architecture-diagram-maintenance.md` may receive cross-reference clarification only if maintenance guidance drift is discovered.
- If none are impacted, include explicit no-impact rationale:
	Service-flow impact is isolated to BFF transport ownership and does not change files, bills, identity, or onboarding flow semantics.

## Instruction Impact *(mandatory for refactor/reorganization)*

- Is this feature a refactor/reorganization?: Yes.
- If Yes, impacted instruction files under `.github/instructions/`:
	- [x] `.github/instructions/architecture.instructions.md`
	- [x] `.github/instructions/project-structure.instructions.md`
	- [x] `.github/instructions/testing.instructions.md`
- If Yes, impacted workflow templates under `.specify/templates/`:
	- [x] `.specify/templates/spec-template.md`
	- [x] `.specify/templates/tasks-template.md`
- Pattern-preservation statement: Route modules remain the only location for Huma registration, controllers remain HTTP-only adapters, service layer owns orchestration, and all HTTP contracts stay in `transport/http/views` so future changes follow the same deterministic boundary model.
- If backend behavior/integration flows are in scope, include explicit reference to canonical integration-test standard:
	Backend integration tests remain under `backend/tests/integration/<service>/` or `backend/tests/integration/cross_service/`, using behavior-based snake_case filenames and table-driven BDD Given/When/Then + AAA.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The BFF module MUST separate route declaration responsibilities from controller behavior responsibilities for all active BFF HTTP endpoints.
- **FR-002**: BFF controllers MUST handle only HTTP boundary work: receiving the request, validating transport input and request context, invoking the service layer, and formatting the HTTP response.
- **FR-003**: BFF controllers MUST NOT own business rules, business decision making, or direct downstream client execution flow.
- **FR-004**: Route modules MUST remain responsible for endpoint declaration and MUST delegate execution to injected controller capabilities.
- **FR-005**: The refactor MUST preserve the existing externally visible endpoint behavior for paths, methods, authorization expectations, and response semantics unless a deliberate change is explicitly approved and documented in the same feature cycle.
- **FR-006**: The transport HTTP layer MUST include a dedicated `transport/http/views` layer that owns every struct representing request or response JSON transmitted through the HTTP boundary.
- **FR-007**: All HTTP request and response contracts for active BFF endpoints MUST be defined in the dedicated views layer rather than inside controllers, services, or route modules.
- **FR-008**: Every field in an HTTP contract that requires validation MUST declare validator tags so controllers can perform consistent boundary checks before invoking services.
- **FR-009**: Controllers MUST use the dedicated HTTP contract structs as the input and output models for transport-layer processing.
- **FR-010**: Existing authentication, authorization, and project-isolation protections MUST remain enforced after the separation of routes, controllers, and HTTP contracts.
- **FR-011**: Every declared active BFF route MUST be covered by at least one integration test that verifies expected route accessibility and outcome.
- **FR-012**: The route, controller, and HTTP contract ownership model MUST be documented clearly enough that maintainers can determine where to place new routing, controller, and contract changes without ambiguity.

### Key Entities *(include if feature involves data)*

- **Route Module**: The transport-layer unit that declares HTTP endpoints and delegates each endpoint to an injected controller capability.
- **Controller Capability**: The HTTP-facing behavior contract that accepts validated transport input, coordinates with the service layer, and returns data ready to be formatted into an HTTP response.
- **HTTP View Contract**: The request or response structure owned by the `transport/http/views` layer that represents JSON transmitted at the HTTP boundary and exposes validator tags for controller-level contract checks.
- **Service Response Model**: The business-layer result returned to controllers for translation into HTTP responses without moving business rules into the transport layer.
- **Route Coverage Record**: The maintainable mapping that shows which integration test scenarios validate each declared BFF route.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of active BFF routes are declared outside controller behavior modules.
- **SC-002**: 100% of controllers serving active BFF routes are limited to HTTP boundary responsibilities and contain no business orchestration responsibilities.
- **SC-003**: 100% of HTTP JSON contracts used by active BFF endpoints are defined in the dedicated HTTP view layer.
- **SC-004**: 100% of required HTTP contract fields for active BFF endpoints expose validator tags needed for controller-side contract checks.
- **SC-005**: 100% of declared active BFF routes have at least one passing integration test.
- **SC-006**: During review, a maintainer can identify where a new route declaration, controller behavior change, and HTTP contract change belong in under 2 minutes.
- **SC-007**: No regression is detected in existing route accessibility, authorization behavior, and expected response outcomes across the in-scope BFF integration suite.
- **SC-008**: Route coverage matrix and integration results provide explicit evidence for every active operation before feature completion.

## Assumptions

- This feature extends the earlier route-segregation effort rather than replacing its core intent.
- Existing service-layer abstractions remain the place where business orchestration and downstream execution occur.
- The refactor applies to all BFF HTTP endpoints that are active and supported at feature completion time.
- Existing client-facing API behavior remains stable unless a separate approved change expands the public contract.
- Integration coverage expectations apply to the full set of active BFF routes included in this feature scope.
