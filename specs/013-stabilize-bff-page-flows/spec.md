# Feature Specification: Stabilize Broken BFF Page Flows

**Feature Branch**: `[013-stabilize-bff-page-flows]`  
**Created**: 2026-04-04  
**Status**: Draft  
**Input**: User description: "The frontend pages for documents, payments, analyses, and settings are failing because several requests to the BFF break. Investigate every frontend request used by those screens, validate why the backend flow is failing, add integration coverage that exercises the full BFF-to-service path, and provide representative default-user mock seed data so the application can be viewed and used without request errors." 

## Clarifications

### Session 2026-04-04

- Q: For the default user seed data, which outcome should the spec require for the four broken pages? → A: The canonical seeded experience must provide representative populated data on all four in-scope pages (`documents`, `payments`, `analyses`, and `settings`), with empty states reserved for explicit secondary validation scenarios rather than the default acceptance path.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Open key screens without backend errors (Priority: P1)

As an authenticated user, I want the `documents`, `payments`, `analyses`, and `settings` pages to load their expected data without request failures so that I can use the application normally.

**Why this priority**: These are visible, core product flows that are currently broken. Until they load reliably, the frontend experience is blocked.

**Independent Test**: This can be fully tested by signing in as the default user, opening each in-scope page, and confirming that the page renders representative populated data on all four screens, with empty states appearing only in explicitly covered fallback scenarios rather than as the default outcome.

**Acceptance Scenarios**:

1. **Given** the default user has a valid session and the canonical dev/test seed is loaded, **When** they open the `documents` page, **Then** all required page requests complete successfully and the screen renders representative populated content rather than a blocker-level backend error.
2. **Given** the default user has a valid session and the canonical dev/test seed is loaded, **When** they open the `payments`, `analyses`, or `settings` pages, **Then** each page completes its required BFF requests and renders representative populated content for that screen.
3. **Given** a page depends on optional or legitimately absent secondary data, **When** one response returns no business data, **Then** the page shows a safe empty-state experience instead of a generic request error, but this is treated as fallback behavior rather than the default acceptance path.

---

### User Story 2 - Validate the full frontend-to-service flow (Priority: P1)

As a backend maintainer, I want each in-scope frontend request mapped to and validated through its BFF and downstream-service path so that the real cause of the breakage is identified and prevented from returning.

**Why this priority**: Fixing symptoms on the frontend is not enough; the team needs reliable proof that the full backend flow works from request entry to service-owned response.

**Independent Test**: This can be independently tested by running end-to-end integration coverage for the in-scope BFF routes and verifying the expected outcomes for success, auth failure, authorization failure, and dependency-related failure handling.

**Acceptance Scenarios**:

1. **Given** a documented list of frontend requests used by the in-scope pages, **When** each request is exercised through the BFF, **Then** the owning downstream flow is validated and any failure is traceable to a specific backend responsibility.
2. **Given** a protected BFF endpoint for one of the in-scope pages, **When** the caller is unauthenticated or lacks project access, **Then** the request fails with the intended access response rather than an internal server error.
3. **Given** the owning downstream service is available, **When** the BFF endpoint is called, **Then** the response contract remains stable and usable by the frontend page.

---

### User Story 3 - Use realistic default-user data for investigation and demos (Priority: P2)

As a developer or tester, I want representative default-user seed data across the affected areas so that I can understand the product behavior, reproduce the page flows, and verify fixes without manual record creation.

**Why this priority**: Meaningful, connected data is required to debug multi-page failures and to confirm that the system behaves correctly once the request path is stabilized.

**Independent Test**: This can be independently tested by preparing the default-user environment and confirming that the in-scope pages render realistic example states without extra manual setup.

**Acceptance Scenarios**:

1. **Given** a fresh local or test environment, **When** the approved default-user seed data is loaded, **Then** the default user can reach all four in-scope pages with representative populated content for documents, payments, analyses, and settings.
2. **Given** a page depends on optional or secondary data that is legitimately absent, **When** the seeded environment is used, **Then** the absence is represented intentionally as a valid empty state and not as broken linkage between services.

### Edge Cases

- The default user is authenticated but is missing project membership or a required context value for one of the protected pages.
- One page issues several backend requests and only one fails; the page must degrade predictably instead of collapsing the entire screen.
- A downstream service returns no records for a valid user; the BFF must return a supported empty-state response rather than a generic failure.
- Seeded records exist but are inconsistent across domains, causing one page to render while another fails; the feature must detect and prevent this mismatch.
- A previously working page regresses after a backend contract change; the validation suite must expose the regression before the frontend is considered healthy.

## Architecture & Memory Diagram Flow Impact *(mandatory)*

- Affected services: `bff` (entry point for all in-scope requests), plus the downstream service owners behind `documents`, `payments`, `analyses`, and `settings`, expected to include `files`, `payments`, `bills`, and `identity`. `onboarding` is audit-in-scope if one of the page flows depends on it.
- Requires architecture diagram update (`.specify/memory/architecture-diagram.md`): **Yes**. The page-to-BFF-to-service request map for the broken screens needs to be reflected in repository memory so future debugging follows the same verified flow ownership.
- Required service-flow file updates in `.specify/memory/`:
  - [x] `.specify/memory/bff-flows.md` (inventory of page requests, auth/guard expectations, and BFF orchestration boundaries)
  - [x] `.specify/memory/files-service-flows.md` (documents page dependency verification)
  - [x] `.specify/memory/bills-service-flows.md` (analysis/payment summary and bill-backed flows if used by the screens)
  - [x] `.specify/memory/identity-service-flows.md` (settings/profile/session context expectations)
  - [ ] `.specify/memory/onboarding-service-flows.md` (only if the request audit confirms a dependency)
  - [x] Other impacted memory file(s): `.specify/memory/architecture-diagram.md`, `.specify/memory/payments-service-flows.md`
- If none are impacted, include explicit no-impact rationale: Not applicable; this feature directly investigates and validates cross-service request flow for several frontend pages.

## Instruction Impact *(mandatory for refactor/reorganization)*

- Is this feature a refactor/reorganization?: **No** — this is a stabilization and verification feature, not a planned structural reorganization. It must, however, preserve existing architectural ownership and testing conventions while fixing the broken flows.
- If Yes, impacted instruction files under `.github/instructions/`:
  - [ ] None required by default; update only if the final investigation establishes a new mandatory seed or validation convention
- If Yes, impacted workflow templates under `.specify/templates/`:
  - [ ] None expected
- Pattern-preservation statement: The feature must preserve the established gateway pattern where the BFF authenticates, authorizes, validates, and composes screen responses while downstream services remain the owners of business data and behavior. Any stabilization work must strengthen this pattern rather than bypass it.
- If backend behavior/integration flows are in scope, include explicit reference to canonical integration-test standard:
  `backend/tests/integration/<service>/` or `backend/tests/integration/cross_service/`, behavior-based snake_case filenames, and table-driven BDD Given/When/Then + AAA.
- If BFF transport/service boundaries are touched, explicitly state:
  - service-owned contracts path (`backend/internals/bff/services/contracts/`)
  - transport-owned views path (`backend/internals/bff/transport/http/views/`)
  - mapper boundary path (`backend/internals/bff/transport/http/controllers/mappers/`)
  - pointer-policy exception contract tracking path (`specs/013-stabilize-bff-page-flows/contracts/pointer-exceptions.md`)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST allow an authenticated default user to open the `documents`, `payments`, `analyses`, and `settings` pages without blocking backend request failures in the supported environment.
- **FR-002**: The feature MUST produce a complete inventory of the frontend requests issued by the in-scope pages and the BFF endpoints responsible for serving them.
- **FR-003**: Each in-scope BFF request path MUST be validated across the full flow from frontend call entry through the owning downstream service response.
- **FR-004**: Any backend defect uncovered in the in-scope page flows MUST be corrected at the owning backend boundary so that the frontend contract becomes stable again without bypassing established service ownership.
- **FR-005**: The BFF MUST return a supported success response or a supported empty-state response for legitimate no-data scenarios and MUST NOT expose a generic internal failure for expected user states.
- **FR-006**: Validation coverage for the in-scope flows MUST include successful access, unauthenticated access, unauthorized or wrong-project access, and known dependency-failure handling.
- **FR-007**: The feature MUST provide representative, populated default-user seed data for the domains needed to exercise all four in-scope screens: `documents`, `payments`, `analyses`, and `settings`.
- **FR-008**: Seeded records for the default user MUST remain internally consistent across related domains so that all four pages render meaningful connected states and one screen does not appear healthy while another fails due to broken data linkage.
- **FR-009**: The feature MUST define or update end-to-end integration coverage that exercises the in-scope BFF routes through their real service dependencies rather than only mock-level wiring.
- **FR-010**: The integration coverage for this feature MUST follow the canonical backend integration-test standard under `backend/tests/integration/<service>/` or `backend/tests/integration/cross_service/`, using behavior-based snake_case filenames and table-driven BDD Given/When/Then + AAA scenarios.
- **FR-011**: The investigation outcome MUST make it clear which backend responsibility caused each broken page request and what validated behavior now protects that flow from regression.
- **FR-012**: The stabilized flows MUST preserve the BFF’s responsibilities for authentication, authorization, request validation, and frontend response composition while keeping business data ownership in the downstream services.

### Key Entities *(include if feature involves data)*

- **Screen Request Flow**: The set of backend requests a frontend page issues during initial load or core interaction, including the page’s dependency ordering and expected outcomes.
- **BFF Screen Endpoint**: A frontend-facing request contract served by the BFF for one of the in-scope pages.
- **Owning Downstream Service**: The backend service that remains responsible for the business data and rules behind a specific page flow.
- **Default User Seed Set**: The connected, representative data prepared for the default user so the affected pages can be exercised meaningfully.
- **Supported Empty State**: A valid page response where the user has no records yet, but the application still renders a stable, understandable experience.
- **Regression Validation Suite**: The end-to-end checks that prove the in-scope pages continue to work after fixes or future backend changes.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of the in-scope screens (`documents`, `payments`, `analyses`, and `settings`) can be opened by the default user in the supported environment without blocker-level backend request errors.
- **SC-002**: 100% of identified frontend-to-BFF requests for the in-scope screens have explicit validation coverage and recorded expected outcomes.
- **SC-003**: 100% of blocker-level backend failures found during the in-scope request audit are mapped to an owning backend responsibility and resolved or intentionally downgraded to a supported empty/access state before the feature is accepted.
- **SC-004**: A new developer or tester can prepare the default-user environment in under 10 minutes and immediately review all four in-scope screens with representative populated data.
- **SC-005**: After the feature is completed, no in-scope page is considered release-ready if it still depends on an unvalidated BFF request path.

## Assumptions

- The existing login and session model remains the entry point for the default user, and auth/session restoration itself is not redefined by this feature.
- Scope is limited to the page flows required to make `documents`, `payments`, `analyses`, and `settings` load and behave correctly for the default user; unrelated screens are outside the initial acceptance scope unless the investigation proves a shared blocker.
- The BFF continues to act as a thin gateway/orchestration layer and does not take over direct ownership of downstream business data while stabilizing these flows.
- Representative seed data is intended for local, test, and non-production validation environments only.
- Empty states remain valid for explicit fallback or secondary validation scenarios, but the canonical acceptance seed for the default user must populate all four in-scope pages with representative data.
