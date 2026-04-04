# Tasks: Dark Mode Login Button Contrast

**Input**: Design documents from `/specs/010-fix-login-contrast/`  
**Prerequisites**: `plan.md` (required), `spec.md` (required), `research.md`, `data-model.md`, `contracts/primary-action-theme-contract.md`  
**Tests**: Included — the plan explicitly requires frontend regression coverage for the dark-mode contrast fix  
**Organization**: Tasks are grouped by user story so each increment can be implemented and validated independently.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel when it touches different files and has no dependency on unfinished work
- **[Story]**: User story label (`US1`, `US2`, `US3`)
- Every task includes the exact file path to change or validate

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm the affected surfaces and validation flow before changing the shared theme contract.

- [X] T001 Review the current solid primary-action usages in `frontend/src/styles/tokens.ts`, `frontend/src/styles/index.css`, `frontend/src/pages/LoginPage.tsx`, `frontend/src/app/ErrorBoundary.tsx`, `frontend/src/components/DraftRestoreModal.tsx`, and `frontend/src/components/ProjectSwitcher.tsx`
- [X] T002 [P] Finalize the manual dark-mode validation checklist in `specs/010-fix-login-contrast/quickstart.md` for first render, hover, focus, loading, and disabled states

**Checkpoint**: Affected files and acceptance checks are confirmed.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Establish the shared token contract that all affected dark-mode primary buttons will consume.

**⚠️ CRITICAL**: No user story work should start until this phase is complete.

- [X] T003 Update the semantic token definitions for solid primary-action background, foreground, hover, and disabled states in `frontend/src/styles/tokens.ts`
- [X] T004 Update the CSS variable bindings for the new light/dark primary-action contract in `frontend/src/styles/index.css`
- [X] T005 Adjust the accent/focus token consumers in `frontend/src/components/Sidebar.tsx` so the primary-action token split does not regress navigation styling

**Checkpoint**: The shared theme contract is ready for story-specific rollout.

---

## Phase 3: User Story 1 - Readable sign-in on first load (Priority: P1) 🎯 MVP

**Goal**: Make the login page `Sign in` action readable immediately on initial dark-mode render.

**Independent Test**: Open the login page in dark mode and verify the `Sign in` label is clearly readable before hover or focus.

### Tests for User Story 1

> **NOTE**: Add these regression checks first and confirm they fail before the implementation change.

- [X] T006 [P] [US1] Add an initial dark-mode render regression test for the sign-in button in `frontend/src/pages/LoginPage.test.tsx`
- [X] T007 [P] [US1] Add a login-page dark-mode first-render scenario in `frontend/src/pages/LoginPage.integration.test.tsx`

### Implementation for User Story 1

- [X] T008 [US1] Update the primary submit button in `frontend/src/pages/LoginPage.tsx` to consume the shared solid-action token contract on first render
- [X] T009 [US1] Adjust the sign-in button label and spinner treatment in `frontend/src/pages/LoginPage.tsx` so loading text remains readable in dark mode

**Checkpoint**: User Story 1 is independently functional and readable on initial dark-mode load.

---

## Phase 4: User Story 2 - Stable contrast during interaction (Priority: P2)

**Goal**: Keep the sign-in button readable through hover, focus, pressed, loading, and disabled states.

**Independent Test**: Hover over the sign-in button, tab to it, and trigger loading/disabled states to confirm the label remains legible throughout.

### Tests for User Story 2

- [X] T010 [P] [US2] Extend interaction-state regression coverage for hover, loading, and disabled button behavior in `frontend/src/pages/LoginPage.test.tsx`
- [X] T011 [P] [US2] Extend keyboard-focus accessibility coverage for the sign-in control in `frontend/src/pages/LoginPage.a11y.test.tsx`
- [X] T011A [P] [US2] Add a theme-toggle and full-page refresh regression scenario in `frontend/src/pages/LoginPage.integration.test.tsx` to verify the approved dark-mode token pair is present on first paint

### Implementation for User Story 2

- [X] T012 [US2] Update the hover, focus, and pressed state classes for the sign-in button in `frontend/src/pages/LoginPage.tsx` to use the new contrast-safe state tokens
- [X] T013 [US2] Refine the dark-mode disabled and loading token values in `frontend/src/styles/tokens.ts` and `frontend/src/styles/index.css` so inactive actions stay readable without appearing enabled

**Checkpoint**: User Story 2 is independently functional and readable across all interaction states.

---

## Phase 5: User Story 3 - Consistent dark-theme action styling (Priority: P3)

**Goal**: Apply the same contrast-safe dark-mode treatment to every other shared solid primary action that reuses the broken styling.

**Independent Test**: Review each affected dark-mode primary button outside the login page and verify it remains readable in default and interactive states.

### Tests for User Story 3

- [X] T014 [P] [US3] Add shared-button regression coverage for the recovery action in `frontend/src/app/ErrorBoundary.test.tsx`
- [X] T015 [P] [US3] Add shared-button regression coverage for `frontend/src/components/DraftRestoreModal.tsx` and `frontend/src/components/ProjectSwitcher.tsx` in `frontend/src/components/DraftRestoreModal.test.tsx` and `frontend/src/components/ProjectSwitcher.test.tsx`

### Implementation for User Story 3

- [X] T016 [P] [US3] Update the solid action styling in `frontend/src/app/ErrorBoundary.tsx` and `frontend/src/components/DraftRestoreModal.tsx` to consume the shared primary-action token contract
- [X] T017 [P] [US3] Update the solid action styling in `frontend/src/components/ProjectSwitcher.tsx` to consume the same dark-mode primary-action token contract
- [X] T018 [US3] Sweep the remaining shared-action surfaces in `frontend/src/pages/LoginPage.tsx`, `frontend/src/app/ErrorBoundary.tsx`, `frontend/src/components/DraftRestoreModal.tsx`, and `frontend/src/components/ProjectSwitcher.tsx` for any leftover low-contrast background/foreground pairings

**Checkpoint**: All known shared dark-theme primary actions are consistent and readable.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Validate the full fix, guard against regressions, and confirm the quickstart flow still matches the real implementation.

- [X] T019 [P] Run the frontend validation suite from `frontend/package.json` (`npm run test`, `npm run lint`, and `npm run typecheck`) and resolve any regressions in the affected files
- [X] T019A [P] Add explicit light-mode non-regression assertions for the shared primary-action contract in `frontend/src/pages/LoginPage.test.tsx`, `frontend/src/app/ErrorBoundary.test.tsx`, and `frontend/src/components/ProjectSwitcher.test.tsx`
- [X] T020 [P] Execute the manual dark-mode validation steps in `specs/010-fix-login-contrast/quickstart.md` and update any final acceptance notes there

---

## Phase 7: Mandatory Governance Sync (Blocking)

**Purpose**: Ensure the feature’s governance requirements stay aligned before merge.

- [X] T021 Confirm that no `.specify/memory/*.md` updates are required for this frontend-only styling fix and keep the no-impact rationale current in `specs/010-fix-login-contrast/plan.md`
- [X] T022 Confirm that no `.github/instructions/*.instructions.md` or `.specify/templates/*.md` updates are required for this non-reorganization change and keep that decision current in `specs/010-fix-login-contrast/plan.md`

**Checkpoint**: The feature is not complete until this governance sync phase is closed.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: Can start immediately
- **Phase 2 (Foundational)**: Depends on Phase 1 and blocks all story work
- **Phase 3 (US1)**: Depends on Phase 2 and is the recommended MVP slice
- **Phase 4 (US2)**: Depends on Phase 2; can proceed after or alongside US1 once the token contract exists
- **Phase 5 (US3)**: Depends on Phase 2; can proceed in parallel with US2 after the shared contract is defined
- **Phase 6 (Polish)**: Depends on the target user stories being complete
- **Phase 7 (Governance Sync)**: Depends on all implementation and validation work and must complete before merge

### User Story Dependencies

- **US1 (P1)**: No dependency on other user stories after the foundational token work is complete
- **US2 (P2)**: Builds on the same shared token contract as US1 but is independently testable on the login page
- **US3 (P3)**: Reuses the shared token contract to fix other affected buttons and is independently testable outside the login page

### Within Each User Story

- Add the regression tests first and confirm they fail
- Roll out the minimum code change to satisfy that story’s acceptance criteria
- Validate the story independently before moving to the next priority

### Parallel Opportunities

- `T002` can run in parallel with `T001`
- After `T003` and `T004` are complete, `T006` and `T007` can run in parallel for US1
- `T010` and `T011` can run in parallel for US2
- `T014` and `T015` can run in parallel for US3 tests
- `T016` and `T017` can run in parallel because they touch different files

---

## Parallel Example: User Story 1

```bash
# Run the US1 regression checks in parallel:
Task: "Add an initial dark-mode render regression test in frontend/src/pages/LoginPage.test.tsx"
Task: "Add a login-page dark-mode first-render scenario in frontend/src/pages/LoginPage.integration.test.tsx"
```

## Parallel Example: User Story 3

```bash
# Update the shared action surfaces in parallel once the token contract is ready:
Task: "Update frontend/src/app/ErrorBoundary.tsx and frontend/src/components/DraftRestoreModal.tsx"
Task: "Update frontend/src/components/ProjectSwitcher.tsx"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete **Phase 1: Setup**
2. Complete **Phase 2: Foundational**
3. Complete **Phase 3: User Story 1**
4. **Stop and validate** the login page in dark mode before proceeding further

### Incremental Delivery

1. Land the shared token contract and login-page readability fix first
2. Add interaction-state hardening for hover/focus/loading/disabled states
3. Roll the same safe contract out to every other affected solid primary action
4. Finish with validation and governance sync before merge

### Parallel Team Strategy

With multiple developers:

1. One developer owns the token contract in `frontend/src/styles/tokens.ts` and `frontend/src/styles/index.css`
2. One developer owns the login-page regression tests and `frontend/src/pages/LoginPage.tsx`
3. One developer owns the shared-action follow-up work in `frontend/src/app/ErrorBoundary.tsx`, `frontend/src/components/DraftRestoreModal.tsx`, and `frontend/src/components/ProjectSwitcher.tsx`

---

## Notes

- `[P]` tasks touch different files and are safe to parallelize
- Each user story remains independently testable from the spec acceptance scenarios
- Keep the fix centralized in the design-token system; avoid one-off hardcoded dark-mode colors
- Verify light-mode behavior still looks correct after the dark-mode adjustments
- Do not consider the feature complete until the validation suite and governance sync phase are both finished