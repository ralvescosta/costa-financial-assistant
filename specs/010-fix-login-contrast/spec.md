# Feature Specification: Dark Mode Login Button Contrast

**Feature Branch**: `010-fix-login-contrast`  
**Created**: 2026-04-04  
**Status**: Draft  
**Input**: User description: "The login page sign-in button is hard to read in dark mode on initial load and should stay readable when hovered and anywhere the same dark background color is reused."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Readable sign-in on first load (Priority: P1)

As a user opening the login page in dark mode, I can immediately read the `Sign in` button label and clearly identify it as the primary action without needing to hover or focus it first.

**Why this priority**: This is the entry point to the application. If the main login action is hard to read, users may hesitate or fail to complete authentication.

**Independent Test**: Open the login page in dark mode and verify that the `Sign in` label is clearly readable on first render before any interaction occurs.

**Acceptance Scenarios**:

1. **Given** the application opens directly in dark mode, **When** the login page first renders, **Then** the `Sign in` button label is clearly readable against its background.
2. **Given** the login page is visible in dark mode, **When** the user looks at the form without interacting, **Then** the primary sign-in action is visually distinct from surrounding elements.

---

### User Story 2 - Stable contrast during interaction (Priority: P2)

As a user interacting with the login form, I want the `Sign in` button to remain readable while hovering, focusing, clicking, or waiting for submission so the interface feels reliable and accessible.

**Why this priority**: Users interact with the button through mouse, keyboard, and loading states. Readability must remain intact across all of them to avoid confusion.

**Independent Test**: Hover over the button, tab to it with the keyboard, click it, and observe any loading or disabled state to confirm the label stays legible throughout.

**Acceptance Scenarios**:

1. **Given** the `Sign in` button is shown in dark mode, **When** the user hovers over it, **Then** the text and background remain easy to distinguish.
2. **Given** the user navigates with the keyboard, **When** the button receives focus, **Then** the focus state remains visible without reducing text readability.
3. **Given** the sign-in action is temporarily unavailable or processing, **When** the button becomes disabled or loading, **Then** the label remains legible while the state still looks inactive or busy.

---

### User Story 3 - Consistent dark-theme action styling (Priority: P3)

As a user moving through the app in dark mode, I want other buttons that reuse the same dark background treatment to remain readable as well, so the interface feels consistent and avoids repeat contrast issues.

**Why this priority**: The reported issue likely comes from a shared style. Fixing it consistently prevents the same defect from appearing elsewhere.

**Independent Test**: Review any other dark-mode primary actions that use the same visual treatment and confirm they remain readable in their default and interactive states.

**Acceptance Scenarios**:

1. **Given** another screen uses the same dark-theme primary action style, **When** that screen renders in dark mode, **Then** the button label is also clearly readable.
2. **Given** the user switches between light and dark themes, **When** buttons re-render, **Then** the dark-theme version preserves readable contrast without affecting light-theme clarity.

### Edge Cases

- The page opens directly in dark mode on first load with no prior interaction.
- The user switches themes after the page has already rendered.
- Browser hover, focus, autofill, or pressed-state styling should not reduce button readability.
- Disabled or loading buttons should still communicate their state without making the label disappear into the background.
- Shared buttons using the same dark background color on other screens should inherit the same readable treatment.

## Architecture & Memory Diagram Flow Impact *(mandatory)*

- **Affected services**: Frontend login presentation only; no backend behavior changes are required in `bff`, `identity`, `files`, `bills`, `payments`, `onboarding`, or `migrations`.
- **Requires architecture diagram update (`.specify/memory/architecture-diagram.md`)**: No. This issue changes visual presentation only and does not alter service responsibilities, system flow, or cross-service communication.
- **Required service-flow file updates in `.specify/memory/`**:
  - [ ] `.specify/memory/bff-flows.md`
  - [ ] `.specify/memory/files-service-flows.md`
  - [ ] `.specify/memory/bills-service-flows.md`
  - [ ] `.specify/memory/identity-service-flows.md`
  - [ ] `.specify/memory/onboarding-service-flows.md`
  - [ ] Other impacted memory file(s): None
- **No-impact rationale**: The feature corrects a shared dark-mode presentation issue and does not change request flow, authentication responsibilities, or backend integration behavior.

## Instruction Impact *(mandatory for refactor/reorganization)*

- **Is this feature a refactor/reorganization?**: No
- **Impacted instruction files under `.github/instructions/`**: None
- **Impacted workflow templates under `.specify/templates/`**: None
- **Pattern-preservation statement**: The fix should preserve the existing design-token and theming approach while standardizing one readable dark-mode primary button treatment for all affected uses of the shared background color.
- **Backend behavior / integration flow note**: Not in scope; no backend behavior or integration-test flow changes are required.
- **BFF transport/service boundary note**: Not touched by this feature.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST present a clearly readable `Sign in` button label on the login page when dark mode is active on initial render, with the label maintaining at least WCAG AA contrast (4.5:1 for normal text) against its background.
- **FR-002**: The system MUST preserve readable contrast for the `Sign in` button across default, hover, focus, pressed, loading, and disabled states in dark mode, with each state meeting the approved contrast threshold.
- **FR-003**: The system MUST keep disabled or loading primary actions visually distinguishable from enabled actions while preserving label legibility.
- **FR-004**: The system MUST apply the same readable dark-mode treatment to any solid primary action that reuses the same background color or shared style, including at minimum the buttons in `frontend/src/pages/LoginPage.tsx`, `frontend/src/app/ErrorBoundary.tsx`, `frontend/src/components/DraftRestoreModal.tsx`, and `frontend/src/components/ProjectSwitcher.tsx`.
- **FR-005**: The system MUST avoid brief low-contrast flashes for the sign-in action when the page loads, refreshes, or switches between light and dark themes.
- **FR-006**: The visual correction MUST remain consistent with the existing login-card and application theme so the button still looks like the intended primary action.
- **FR-007**: The update MUST not reduce readability, clarity, or perceived clickability of the same action in light mode, and light-mode contrast must remain equal to or better than the current approved appearance.
- **FR-008**: The system MUST use one consistent approved dark-theme primary-button appearance for all affected screens to prevent the same contrast defect from recurring.

### Key Entities *(include if feature involves data)*

- **Theme Mode**: The active visual mode, such as light or dark, that determines the required contrast behavior for UI controls.
- **Primary Action Button**: The main call-to-action element a user relies on to proceed, including its label, emphasis, and visual states.
- **Interaction State**: The current condition of the button, such as default, hover, focus, pressed, loading, or disabled.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In dark mode, the `Sign in` action meets WCAG AA contrast (≥ 4.5:1 for button text) on initial render in 100% of validation runs.
- **SC-002**: 100% of reviewed sign-in button states in dark mode (default, hover, focus, pressed, loading, and disabled) meet the same approved contrast threshold during validation.
- **SC-003**: In 10 consecutive theme-switch and full-page refresh validation runs, the sign-in action never renders with a sub-threshold contrast pair on first paint.
- **SC-004**: No known contrast-related defects remain on the identified shared primary-action surfaces after validation in both light and dark themes.

## Assumptions

- The existing login behavior, wording, and layout remain unchanged; this feature focuses on visual readability and consistency only.
- The dark-theme issue comes from a reusable button or background treatment that can be corrected once and reused consistently.
- The login screen is the first priority, but any other dark-mode action reusing the same background color should receive the same readability fix.
- In scope for this feature: all solid primary buttons using the shared dark-mode primary-action token contract, currently identified in `frontend/src/pages/LoginPage.tsx`, `frontend/src/app/ErrorBoundary.tsx`, `frontend/src/components/DraftRestoreModal.tsx`, and `frontend/src/components/ProjectSwitcher.tsx`.
- Out of scope unless discovered to reuse the same contract: sidebar text links, secondary/ghost buttons, and unrelated typography or layout redesign.
- A broader redesign of the authentication page, branding, or form structure is out of scope for this feature.
