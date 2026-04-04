# Feature Specification: Restore Seeded Login & Session Propagation

**Feature Branch**: `[012-restore-login-session]`  
**Created**: 2026-04-04  
**Status**: Draft  
**Input**: User description: "Login is not working, the system is missing the persistent user setup needed for sign-in, there must be no registration screen, a seeded default user `ralvescosta` / `mudar@1234` must exist as the project owner with all project permissions, all protected routes must work from the authenticated result, and all gRPC requests must carry a shared `Session` plus defaulted `Pagination` for list/select flows." 

## Clarifications

### Session 2026-04-04

- Q: What route scope counts as “all the other routes will work”? → A: The verified scope is **all currently exposed authenticated BFF routes/screens**, not just a small MVP subset or a single happy-path route per service.
- Q: What default pagination values should the BFF send when the frontend omits query params? → A: The BFF must forward deterministic first-page defaults of `page_size = 20` and `page_token = ""` unless a route-specific smaller clamp is already documented.
- Q: What email should back the seeded owner and shared `Session.email` field? → A: The bootstrap owner uses the deterministic email `ralvescosta@local.dev` so `common.v1.Session.email` is always populated during local/demo sign-in.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Sign in with the seeded owner account (Priority: P1)

As a user opening the application in a fresh environment, I need a working sign-in flow that uses a pre-seeded owner account so that I can access the platform immediately without needing a registration screen or manual database edits.

**Why this priority**: If sign-in is broken, the rest of the product is effectively blocked. Restoring a reliable bootstrap login is the minimum path to make the application usable again.

**Independent Test**: This can be fully tested by starting from a fresh database, running the seed path, signing in with the default credential, and confirming the application returns an authenticated result without any manual data repair.

**Acceptance Scenarios**:

1. **Given** a fresh environment with migrations and seed data applied, **When** the user signs in with `ralvescosta` and `mudar@1234`, **Then** the system authenticates successfully and returns a usable authenticated result.
2. **Given** the application does not provide self-registration in this phase, **When** a new user opens the login screen, **Then** the expected bootstrap path is to use the seeded owner account rather than a register screen.
3. **Given** the seeded account or its backing storage is missing, **When** sign-in is attempted, **Then** the system fails clearly and points to the missing bootstrap setup instead of leaving the user in a partial session.

---

### User Story 2 - Use the authenticated session across protected routes (Priority: P1)

As the seeded project owner, I need the authenticated result from sign-in to be recognized consistently across the BFF and downstream services so that all currently exposed authenticated BFF routes/screens continue to work with the same signed-in identity.

**Why this priority**: A login screen is only useful if the authenticated user can actually navigate and use protected functionality afterward. Broken session propagation would keep the app unusable even if login itself succeeds.

**Independent Test**: This can be tested by signing in as the default owner user and then exercising representative protected routes from each major area, confirming the authenticated result is accepted throughout the flow.

**Acceptance Scenarios**:

1. **Given** the default owner user has signed in successfully, **When** the user opens authenticated pages and actions, **Then** the BFF and downstream services treat the request as coming from the same logged-in user.
2. **Given** the default user is seeded as the project owner, **When** owner-only or project-scoped routes are called, **Then** those routes succeed without extra manual role assignment.
3. **Given** the session is missing, invalid, or inconsistent with project membership, **When** a protected route is called, **Then** access is denied safely and the user is not shown unauthorized data.

---

### User Story 3 - Keep list and search flows usable with default pagination (Priority: P2)

As a signed-in user, I need list and search screens to return predictable first-page results even when the frontend does not send pagination parameters so that dashboards and list views remain usable by default.

**Why this priority**: Many authenticated routes depend on list-style responses. If pagination is optional at the UI layer but missing downstream, those routes can fail or behave inconsistently.

**Independent Test**: This can be fully tested by calling representative list endpoints with and without pagination query parameters and confirming the BFF always forwards a populated pagination request to downstream services.

**Acceptance Scenarios**:

1. **Given** the frontend omits pagination query parameters on a list page, **When** the BFF forwards the request downstream, **Then** it sends deterministic default pagination values of `page_size = 20` and `page_token = ""` instead of an empty pagination object.
2. **Given** the frontend supplies explicit pagination values, **When** a list or search route is called, **Then** the BFF forwards those values consistently to the downstream gRPC request.
3. **Given** a list or search operation returns multiple records, **When** the response is generated, **Then** the caller receives a predictable first page and the information needed to continue paging.

### Edge Cases

- What happens when the seed migration runs but the user row, credential hash, or project-owner membership is only partially created?
- How does the system respond when the seeded username is correct but the stored password or login metadata is stale or invalid?
- What happens when one protected route is updated to accept the authenticated session but another downstream request still omits the required `Session` contract?
- How are list/select routes handled when the frontend sends invalid pagination values or no pagination values at all?

## Architecture & Memory Diagram Flow Impact *(mandatory)*

- Affected services: `bff`, `identity`, `onboarding`, and `migrations` are directly affected by login restoration, seeding, and token/session propagation. `bills`, `files`, and `payments` are also in scope because their gRPC request contracts and protected flows must accept the propagated session and pagination defaults.
- Requires architecture diagram update (`.specify/memory/architecture-diagram.md`): **Yes**. The flow documentation must show the seeded owner user, successful sign-in, issuance of the authenticated session, and the BFF forwarding `Session` and `Pagination` consistently to downstream gRPC services.
- Required service-flow file updates in `.specify/memory/`:
  - [x] `.specify/memory/bff-flows.md` (login result propagation and default pagination forwarding)
  - [x] `.specify/memory/files-service-flows.md` (authenticated list/select request handling)
  - [x] `.specify/memory/bills-service-flows.md` (owner-access payment list/dashboard request flow)
  - [x] `.specify/memory/identity-service-flows.md` (bootstrap login/session issuance)
  - [x] `.specify/memory/onboarding-service-flows.md` (seeded owner membership and project permissions)
  - [x] Other impacted memory file(s): `.specify/memory/architecture-diagram.md`, `.specify/memory/payments-service-flows.md`
- If none are impacted, include explicit no-impact rationale: Not applicable; this feature changes authentication bootstrapping and the cross-service request contract expected by multiple protected flows.

## Instruction Impact *(mandatory for refactor/reorganization)*

- Is this feature a refactor/reorganization?: **Yes**
- If Yes, impacted instruction files under `.github/instructions/`:
  - [x] `.github/instructions/architecture.instructions.md`
  - [x] `.github/instructions/ai-behavior.instructions.md`
  - [x] `.github/instructions/testing.instructions.md`
- If Yes, impacted workflow templates under `.specify/templates/`:
  - [ ] None required for this feature
- Pattern-preservation statement: Future contributors must preserve one deterministic bootstrap-login pattern: a seeded owner user exists for the non-registration flow, `common.v1.Session` travels with every gRPC request, and `common.v1.Pagination` travels with every list/select request. The BFF remains the place that reads query parameters, applies default pagination values when absent, and forwards the authenticated session context downstream.
- If backend behavior/integration flows are in scope, include explicit reference to canonical integration-test standard: `backend/tests/integration/<service>/` or `backend/tests/integration/cross_service/`, behavior-based snake_case filenames, and table-driven BDD Given/When/Then + AAA.
- If BFF transport/service boundaries are touched, explicitly state:
  - service-owned contracts path (`backend/internals/bff/services/contracts/`)
  - transport-owned views path (`backend/internals/bff/transport/http/views/`)
  - mapper boundary path (`backend/internals/bff/transport/http/controllers/mappers/`)
  - pointer-policy exception contract tracking path (`specs/012-restore-login-session/contracts/pointer-exceptions.md`)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST provide a working sign-in path without a registration screen, using seed-backed bootstrap data so a fresh environment can be accessed immediately.
- **FR-002**: The bootstrap seed data MUST create a default user with username `ralvescosta`, deterministic email `ralvescosta@local.dev`, and password `mudar@1234` for the intended local/demo bootstrap flow of this feature.
- **FR-003**: The seeded default user MUST be attached to at least one project as the project owner and MUST receive the full set of project permissions needed to access all currently exposed authenticated BFF routes/screens in the verified scope.
- **FR-004**: The system MUST create or correct the persistent user storage and supporting seed records required for authentication so login does not depend on manual table edits.
- **FR-005**: A successful sign-in MUST return an authenticated result that the frontend can use immediately on subsequent protected requests.
- **FR-006**: All currently exposed authenticated BFF routes/screens in the verified scope MUST honor the authenticated result of the default owner user and continue to work without requiring extra manual membership or permission setup.
- **FR-007**: `common.v1.Session` MUST be introduced and maintained as the canonical logged-in user contract in `proto/common`, containing `id`, `email`, and `username`.
- **FR-008**: The `Session.id` value MUST represent the logged-in user with UUIDv7 format and remain consistent across downstream requests created from the same authenticated interaction.
- **FR-009**: Every gRPC request message in scope for authenticated application behavior MUST carry the `Session` so downstream services can rely on the same caller identity.
- **FR-010**: Every list or select gRPC request that can return more than one value MUST carry `common.v1.Pagination`.
- **FR-011**: The BFF MUST accept pagination from query parameters and, when the frontend omits pagination, MUST apply deterministic default values of `page_size = 20` and `page_token = ""` and always forward a populated pagination object to the downstream gRPC request.
- **FR-012**: The login and downstream route flow MUST fail safely and clearly when credentials are wrong, bootstrap data is missing, or the authenticated user is not a valid project member; the system MUST NOT leave the caller in a partial or ambiguous session state.
- **FR-013**: Verification for this feature MUST include end-to-end evidence for login success, invalid-login failure, owner-route access, and at least one representative list/select route using default pagination behavior.
- **FR-014**: Repository instructions and memory artifacts MUST be updated in the same feature scope so future work preserves the seeded-login, `Session`, and default-pagination conventions deterministically.

### Key Entities *(include if feature involves data)*

- **Bootstrap User**: The pre-seeded account used to enter the application when no registration screen exists in this phase.
- **Project Owner Membership**: The project-scoped relationship that grants the bootstrap user full owner permissions across protected routes.
- **Session**: The authenticated identity envelope shared across the system, carrying the logged-in user ID, email, and username for downstream requests.
- **Pagination Contract**: The standard paging input used on list/select requests so callers always receive deterministic multi-record behavior.
- **Authenticated Result**: The post-login state that allows the frontend and BFF to continue calling protected routes as the same verified user.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A fresh local/demo environment can be authenticated successfully with the seeded owner account on the first attempt, without manual database edits.
- **SC-002**: 100% of protected routes in the verified scope accept the seeded owner’s authenticated result and do not fail because of missing bootstrap identity, project membership, or permissions.
- **SC-003**: 100% of gRPC request contracts in the verified scope carry `Session`, and 100% of list/select request contracts in scope carry `Pagination`.
- **SC-004**: In 100% of tested list or search flows, the first page of results is returned predictably even when the frontend omits pagination query parameters.
- **SC-005**: Invalid credentials or missing seed data always produce a clear failure outcome and never leave the caller in a partially authenticated state.

## Assumptions

- The bootstrap credential is intended for local, development, or controlled demo bootstrapping; stricter secret-handling or credential rotation for production can be layered on later without reintroducing a registration screen for this phase.
- Existing token issuance and validation behavior will be extended rather than replaced, with the authenticated result enriched by the shared `Session` contract.
- The application will continue enforcing project isolation and permissions through the existing project membership model even after the default owner seed is added.
- The BFF remains the gateway for authentication, authorization, request validation, and downstream orchestration; the new requirement is consistent propagation of authenticated session and pagination context.
- Representative login and route verification will be added using the repository’s existing integration and regression test patterns rather than ad-hoc manual-only validation.
