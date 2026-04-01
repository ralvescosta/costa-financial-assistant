# Feature Specification: BFF Route-Controller Segregation

**Feature Branch**: `[004-segregate-bff-routing]`  
**Created**: 2026-04-01  
**Status**: Draft  
**Input**: User description: "refactor the bff http controllers to segregate the controllers with the routing, the bff controllers are doing the controller logic also declaring the routes, we must segregate this responsibility, we must have a route declaration separated than the controllers, each controller must be a struct which implements a default controller interface that could have default methods that we could or could not implement and the routes must be also a struct which implements a default interface and will receive the controller through dependency injection and will use the controllers into the routes. you also must ensure all the routes we have integration tests for them."

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

### User Story 1 - Separate Route Registration from Controller Behavior (Priority: P1)

As a backend maintainer, I want route declarations to be separated from controller behavior so that HTTP entrypoints are easier to read, change, and validate without mixing responsibilities.

**Why this priority**: This is the core architectural change requested and unlocks predictable maintenance for all BFF endpoints.

**Independent Test**: Can be fully tested by reviewing route setup modules and confirming controllers no longer own route declaration while existing endpoint behavior remains unchanged.

**Acceptance Scenarios**:

1. **Given** existing BFF endpoints, **When** route setup is inspected, **Then** route declarations are defined in dedicated route structures and not inside controller behavior structures.
2. **Given** the refactored codebase, **When** an existing endpoint is called, **Then** it preserves the same externally observable request/response behavior and authorization expectations.

---

### User Story 2 - Standardize Controller and Route Contracts (Priority: P2)

As a backend maintainer, I want controllers and routes to follow default contracts so that new endpoints can be introduced consistently with lower onboarding cost and fewer integration mistakes.

**Why this priority**: Contract consistency reduces divergence between modules and improves development speed for future features.

**Independent Test**: Can be independently tested by introducing or updating one endpoint module and confirming it can use the standard controller and route contracts without custom wiring patterns.

**Acceptance Scenarios**:

1. **Given** a controller module, **When** it is implemented, **Then** it can conform to a shared controller contract while only implementing behavior relevant to that module.
2. **Given** a route module, **When** dependencies are wired, **Then** it receives controller dependencies through dependency injection and uses those controllers to handle endpoint registration and execution flow.

---

### User Story 3 - Ensure Route-Level Integration Coverage (Priority: P3)

As a quality owner, I want integration tests to cover all BFF routes so that refactoring does not silently break endpoint exposure, middleware flow, or route accessibility.

**Why this priority**: Full route coverage protects against regressions introduced by routing reorganization.

**Independent Test**: Can be independently tested by running integration tests and verifying every declared BFF route is represented by at least one integration scenario.

**Acceptance Scenarios**:

1. **Given** the full list of declared BFF routes, **When** integration tests are executed, **Then** each route has at least one passing integration test validating expected accessibility and response behavior.
2. **Given** a new BFF route is added after this feature, **When** quality checks run, **Then** missing integration coverage for that route is detectable and considered incomplete work.

---

### Edge Cases

- A controller supports only a subset of optional default contract methods; unimplemented optional behavior must not block route registration for supported actions.
- A route is declared but not wired through dependency injection; startup or validation must fail clearly rather than silently exposing partial routing.
- A route exists in declarations but has no corresponding integration test; the quality process must identify the gap before feature acceptance.
- Existing middleware (authentication, authorization, project guard) must remain consistently applied after routing is moved out of controllers.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The BFF module MUST separate route declaration responsibilities from controller behavior responsibilities for all HTTP endpoints in scope.
- **FR-002**: The system MUST define narrow controller capability contracts consumed by route modules, allowing each controller to implement only the behaviors required by its resource routes.
- **FR-003**: The system MUST provide a shared route contract that standardizes how route groups register endpoints, receive middleware collaborators, and consume controller dependencies.
- **FR-004**: Route structures MUST receive controller dependencies through the existing dependency injection flow rather than creating controller instances locally.
- **FR-005**: Existing endpoint paths, HTTP methods, and externally visible response behavior MUST remain functionally equivalent after segregation, unless explicitly documented as a deliberate change.
- **FR-006**: Existing authentication, authorization, and project-isolation protections MUST remain enforced on the same route behaviors after refactoring.
- **FR-007**: Every declared BFF route MUST be covered by at least one integration test that verifies successful routing and expected outcome for the route's intended access pattern.
- **FR-008**: The route-to-test mapping MUST be maintainable, allowing maintainers to identify which integration test(s) validate each declared route.
- **FR-009**: Feature completion MUST require passing integration tests for all route coverage defined by this specification.
- **FR-010**: Project governance documents and coding instructions MUST be updated so the dedicated route-module pattern and route-coverage expectations become the required standard for future BFF changes.

### Key Entities *(include if feature involves data)*

- **Controller Capability Contract**: A narrow behavioral contract owned by the route layer and implemented by a controller for the specific handlers required by that route module. Shared reusable behavior may be provided through a base controller struct.
- **Route Contract**: A standard registration contract for route modules. Defines how a route module receives dependencies and declares endpoint mappings.
- **Controller Module**: A concrete controller unit that executes request handling behavior and conforms to the controller contract.
- **Route Module**: A concrete routing unit that declares endpoint mappings and delegates handling to injected controller modules.
- **Route Coverage Record**: A maintainable mapping artifact that links each declared route to one or more integration tests validating that route.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of BFF routes in scope are declared via dedicated route modules rather than inside controller behavior modules.
- **SC-002**: 100% of controller modules in scope conform to the default controller contract and are wired through dependency injection.
- **SC-003**: 100% of declared BFF routes in scope have at least one passing integration test.
- **SC-004**: No regression in existing route accessibility and expected response outcomes is detected across the full integration test suite for BFF routes in scope.
- **SC-005**: During code review, maintainers can identify route declarations and controller behavior locations for any in-scope endpoint in under 2 minutes.
- **SC-006**: The constitution and applicable instruction files reflect the route-module ownership model and route-coverage expectations before implementation is considered complete.

## Assumptions


- The refactor targets existing BFF HTTP routes and does not introduce new business capabilities beyond structural segregation and coverage completion.
- Existing dependency injection and middleware mechanisms remain the baseline and are reused rather than replaced.
- Integration route coverage is evaluated for routes that are active and supported in the BFF module at feature completion time.
- Existing API contracts consumed by clients remain stable unless a separate approved change request explicitly expands this scope.
