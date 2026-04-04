# Feature Specification: Restore BFF gRPC Gateway Boundary

**Feature Branch**: `[011-fix-bff-grpc-boundary]`  
**Created**: 2026-04-04  
**Status**: Draft  
**Input**: User description: "The BFF was misdesigned and must stop accessing domain data directly. It must handle authentication, authorization, and frontend-facing contract composition while always calling downstream services through gRPC for business data and actions. Existing direct-access patterns in payments and possibly other areas must be corrected, documented, and verified as buildable/runnable." 

## Clarifications

### Session 2026-04-04

- Q: How should affected BFF flows be handled when the owning downstream gRPC contract does not yet exist? → A: This feature must add or extend the required downstream gRPC contracts and fully migrate the affected flows; temporary dependency-error fallbacks are transition-only and are not an acceptable final state for supported paths.
- Q: What exactly is in scope for “supported paths” and “scoped backend services”? → A: The mandatory implementation and verification scope for this feature is `bff`, `payments`, and `bills`, covering payment-cycle preference, history timeline, reconciliation, and the bills-backed dashboard/mark-paid interactions exercised by those screens. `files`, `onboarding`, and `identity` are audit-in-scope and become implementation-in-scope only if the BFF audit uncovers direct-access violations in touched routes.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Keep the BFF as a secure gateway (Priority: P1)

As a frontend and API consumer, I need the BFF to authenticate and authorize requests and then gather screen data only through downstream service contracts so that responses stay consistent, secure, and aligned with service ownership.

**Why this priority**: This is the architectural correction at the center of the request. If the BFF continues to bypass domain services, the system will keep leaking responsibilities and reintroducing security and consistency risks.

**Independent Test**: This can be fully tested by exercising representative BFF endpoints and confirming that business data is returned through downstream service calls while authentication, authorization, and response composition still work as before.

**Acceptance Scenarios**:

1. **Given** a BFF endpoint that serves payment, history, document, or project data, **When** a request is processed, **Then** the BFF authenticates and authorizes the caller and obtains business data from downstream services instead of direct database or repository access.
2. **Given** a frontend screen that needs data from more than one service, **When** the BFF builds the response, **Then** it combines downstream service results into one frontend-facing contract without taking ownership of the underlying domain persistence.
3. **Given** a downstream service failure, **When** the BFF cannot complete the request, **Then** it returns a sanitized failure response and MUST NOT fall back to direct database access.

---

### User Story 2 - Remove direct BFF domain-data ownership (Priority: P1)

As a backend maintainer, I need all direct BFF dependencies on domain repositories, SQL access, or domain services that hide direct database access to be removed or replaced so that each domain service remains the single owner of its own data and business rules.

**Why this priority**: This prevents broken boundaries, duplicated logic, and future regressions where business logic is split between the BFF and domain services.

**Independent Test**: This can be tested through dependency review and targeted regression tests showing that BFF modules no longer depend on domain repositories or direct database access paths for business data flows.

**Acceptance Scenarios**:

1. **Given** a BFF module under `backend/internals/bff/`, **When** its dependencies are reviewed, **Then** it has no direct dependency on domain repositories, SQL clients, or database-backed domain services for business data retrieval or mutation.
2. **Given** a business capability currently surfaced through the BFF, **When** the boundary fix is applied, **Then** the authoritative downstream service owns the query or mutation and the BFF acts only as the authenticated/authorized orchestration layer.
3. **Given** a new BFF endpoint added after this feature, **When** it is implemented, **Then** the project instructions and memory artifacts make the allowed pattern unambiguous and repeatable.

---

### User Story 3 - Preserve reliability after the boundary correction (Priority: P2)

As a team lead, I need the updated BFF pattern, service-flow diagrams, and build verification captured in repository guidance so that future changes keep the same gateway model and the scoped services continue to compile and boot correctly.

**Why this priority**: The correction is only durable if the rules are documented and the affected services remain in a healthy, runnable state.

**Independent Test**: This can be tested by reviewing the updated instructions and memory files and by running the repository’s compile/startup verification commands for the affected backend services.

**Acceptance Scenarios**:

1. **Given** the feature work is complete, **When** contributors read the updated instructions and memory docs, **Then** they see that BFF is responsible for authentication, authorization, and response composition only, while domain data access happens through downstream services.
2. **Given** the affected backend services are built and started with the repository’s existing commands, **When** verification is run, **Then** the scoped services compile successfully and start without immediate startup failures caused by the boundary cleanup.

### Edge Cases

- If a legacy BFF flow currently relies on a repository that has no equivalent downstream RPC yet, this feature must add or extend that downstream contract rather than preserve direct access; any safe dependency-error fallback is transition-only.
- How does the system behave when a single screen needs data from multiple services with one of those services temporarily unavailable?
- How are auth-only infrastructure concerns handled so the BFF can still validate tokens and enforce project membership without becoming a domain-data owner?
- What happens when a downstream service response is missing optional data that the BFF previously filled from direct storage access?

## Architecture & Memory Diagram Flow Impact *(mandatory)*

- Affected services: `bff` (mandatory), `payments` (known hotspot), and `bills` for the already-owned dashboard/mark-paid interactions tied to the supported payment screens. `files`, `onboarding`, and `identity` remain audit targets and must be remediated within this feature if the BFF audit uncovers direct-access violations in touched routes.
- Requires architecture diagram update (`.specify/memory/architecture-diagram.md`): **Yes**. The current system overview must explicitly remove any BFF → PostgreSQL domain-data path and show BFF → gRPC service → data store ownership instead.
- Required service-flow file updates in `.specify/memory/`:
  - [x] `.specify/memory/bff-flows.md` (core BFF gateway rule and updated request flows)
  - [x] `.specify/memory/payments-service-flows.md` (payments-owned data access must stay behind the service boundary)
  - [x] Other impacted memory file(s): `.specify/memory/architecture-diagram.md`, and any additional downstream flow file that is touched once direct-access patterns are found during implementation.
- If none are impacted, include explicit no-impact rationale: Not applicable; this feature directly changes the documented ownership of BFF request flow and removes an invalid data-access path from the architecture diagrams.

## Instruction Impact *(mandatory for refactor/reorganization)*

- Is this feature a refactor/reorganization?: **Yes**
- If Yes, impacted instruction files under `.github/instructions/`:
  - [x] `.github/instructions/architecture.instructions.md`
  - [x] `.github/instructions/project-structure.instructions.md`
  - [x] `.github/instructions/ai-behavior.instructions.md`
- If Yes, impacted workflow templates under `.specify/templates/`:
  - [ ] None required for this feature
- Pattern-preservation statement: Repository instructions and memory artifacts will be updated so future contributors consistently treat the BFF as an authentication/authorization and response-composition gateway only. Domain data ownership and business mutations must stay behind downstream service boundaries, reached through service contracts rather than direct storage access.
- If backend behavior/integration flows are in scope, include explicit reference to canonical integration-test standard: `backend/tests/integration/<service>/` or `backend/tests/integration/cross_service/`, behavior-based snake_case filenames, and table-driven BDD Given/When/Then + AAA.
- If BFF transport/service boundaries are touched, explicitly state:
  - service-owned contracts path (`backend/internals/bff/services/contracts/`)
  - transport-owned views path (`backend/internals/bff/transport/http/views/`)
  - mapper boundary path (`backend/internals/bff/transport/http/controllers/mappers/`)
  - pointer-policy exception contract tracking path (`specs/011-fix-bff-grpc-boundary/contracts/pointer-exceptions.md`)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The BFF MUST remain responsible for request authentication, authorization, tenant/project validation, and frontend-facing response composition.
- **FR-002**: The BFF MUST NOT access domain databases, domain repositories, or domain-owned persistence abstractions directly for business reads or writes.
- **FR-003**: All BFF business data retrieval and mutation flows MUST go through the appropriate downstream service contract for the owning domain.
- **FR-004**: Payment-related BFF flows, including known history, cycle-preference, and reconciliation paths, MUST be reviewed and corrected by adding or extending the required downstream payments gRPC contracts so the BFF no longer reaches payments-owned data through direct in-process domain access in the final delivered state.
- **FR-005**: Any additional BFF direct-access pattern found by the required audit in other domains during implementation MUST be triaged explicitly and, for any touched or supported path, removed within this feature; if the owning service lacks a needed contract, that contract MUST be added or extended rather than leaving the BFF on a permanent degraded fallback.
- **FR-006**: The boundary correction MUST preserve the externally observable behavior of supported frontend endpoints, including successful authentication/authorization outcomes and response meaning for existing screens.
- **FR-007**: If a downstream service becomes unavailable during runtime, the BFF MUST fail safely with a sanitized error and MUST NOT bypass the service boundary to query storage directly; temporary safe fallbacks during migration are allowed only until the required downstream contract is in place.
- **FR-008**: Updated repository instructions MUST state that the BFF is a gateway/orchestration layer and that domain data ownership belongs to the downstream services.
- **FR-009**: Updated memory artifacts and architecture diagrams MUST remove the direct BFF database-access flow and show the corrected BFF → service → data-store pattern.
- **FR-010**: Verification for this feature MUST include build/startup checks for `bff`, `payments`, and `bills`, plus any additional service touched by audit findings, demonstrating that they still compile and can be launched with the repository’s supported commands after the boundary correction and documentation updates.
- **FR-011**: Regression coverage for touched BFF flows MUST verify real endpoint behavior or service behavior rather than only mock-only wiring assertions.
- **FR-012**: Regression coverage for the migrated payment-cycle, history, and reconciliation routes MUST include unauthorized, forbidden, and project-membership failure scenarios so the BFF’s authentication and authorization responsibilities remain explicitly verified after the gRPC migration.

### Key Entities *(include if feature involves data)*

- **BFF Gateway Flow**: The authenticated and authorized request path that receives frontend calls, invokes one or more downstream services, and returns a frontend-friendly response.
- **Downstream Service Contract**: The authoritative boundary used by the BFF to request or mutate domain data owned by `payments`, `bills`, `files`, `onboarding`, or `identity`.
- **Direct Access Violation**: Any BFF dependency on a domain repository, SQL client, or storage-backed domain abstraction that bypasses the owning service boundary.
- **Frontend Composition Response**: A screen-oriented response assembled by the BFF from one or more downstream service results without taking over persistence ownership.
- **Authorization Guard**: The BFF-owned authentication and project/role enforcement step that remains valid even after domain-data access is removed from the gateway.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of reviewed BFF modules in scope show zero direct dependency on domain repositories or direct database access for business data flows.
- **SC-002**: 100% of scoped BFF payment/history/dashboard/cycle/reconciliation flows obtain domain data through the owning downstream service contract in the final deliverable, with no remaining direct-access fallback for supported paths.
- **SC-003**: 0 approved frontend endpoint regressions are introduced for the supported flows covered by this refactor.
- **SC-004**: All instruction and memory artifacts listed in this specification are updated in the same feature scope and reflect the corrected gateway model.
- **SC-005**: The scoped backend services complete compile verification successfully and their startup commands reach a healthy boot state without immediate boundary-related failures.

## Assumptions

- The BFF may still own authentication, authorization, request validation, JWKS/cache usage, and frontend response composition, but it does not own domain persistence.
- Existing downstream services are the correct long-term owners of payment, billing, document, onboarding, and identity-related domain data.
- Any missing downstream contract needed to replace a direct-access path is in scope for this feature and can be added or extended as part of implementation without changing the intended user-facing behavior.
- Build and startup verification will use the repository’s existing Go service commands and available local configuration, not a newly introduced runtime path.
- If some legacy violations cannot be fully removed in one pass, they will be explicitly documented and prioritized rather than silently left in place.

