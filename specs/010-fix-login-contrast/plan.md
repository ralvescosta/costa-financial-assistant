# Implementation Plan: Dark Mode Login Button Contrast

**Branch**: `010-fix-login-contrast` | **Date**: 2026-04-04 | **Spec**: `/specs/010-fix-login-contrast/spec.md`
**Input**: Feature specification from `/specs/010-fix-login-contrast/spec.md`

**Note**: This plan covers Phase 0 research and Phase 1 design outputs for a frontend-only contrast regression affecting the login page and other shared primary actions in dark mode.

## Summary

Fix the dark-mode sign-in button contrast regression by correcting the shared theme-token contract used for solid primary actions, applying the approved contrast-safe treatment to all affected buttons, and adding frontend regression checks so the label stays readable on first render and through hover, focus, loading, and disabled states.

## Technical Context

**Language/Version**: TypeScript 5.8.x, React 18.3.x  
**Primary Dependencies**: Vite 6.3.x, Tailwind CSS 3.4.x, `react-router-dom` 6.30.x, existing theme tokens in `frontend/src/styles/tokens.ts`  
**Storage**: N/A for domain data; existing client-side theme preference persistence remains unchanged  
**Testing**: Vitest 3.2.x, Testing Library, JSDOM-based frontend page/integration tests in `frontend/src/pages/` and `frontend/src/app/`  
**Target Platform**: Modern desktop and mobile web browsers running the frontend bundle  
**Project Type**: Web application (`frontend/` + backend services already in place)  
**Performance Goals**: Preserve current login responsiveness and eliminate perceptible low-contrast flash for the primary sign-in action on initial dark-mode render  
**Constraints**: Must use the existing design-token system; must not hardcode one-off color fixes; must not change backend auth flow; must preserve light-mode behavior and keyboard accessibility  
**Scale/Scope**: Limited to shared frontend button styling affecting `LoginPage` and any other component reusing the same solid primary-action background

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Gate Status

- [x] `spec.md` includes `Architecture & Memory Diagram Flow Impact` with explicit no-impact rationale.
- [x] End-of-execution tasks do **not** require `.specify/memory/*.md` updates because no service flow or architecture path changes are in scope.
- [x] Feature is **not** a refactor/reorganization, so no `.github/instructions/*.instructions.md` updates are required.
- [x] Workflow behavior does not change, so no `.specify/templates/*.md` updates are required.
- [x] Backend integration behavior is not in scope; canonical backend integration-test placement rules are unaffected.
- [x] BFF boundaries are not modified; no service-contract or mapper-boundary work is required.

### Constitution Alignment

- **Frontend hook-centric and component-first**: PASS — the work stays in presentation/theme files and existing page/component styling.
- **Design token system**: PASS — the fix will be centralized in `frontend/src/styles/tokens.ts` and `frontend/src/styles/index.css`, then consumed by affected UI surfaces.
- **Architecture boundaries**: PASS — no backend, BFF, or transport boundary changes.

## Project Structure

### Documentation (this feature)

```text
specs/010-fix-login-contrast/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── primary-action-theme-contract.md
└── tasks.md              # Created later by /speckit.tasks
```

### Source Code (repository root)

```text
frontend/
├── src/pages/
│   ├── LoginPage.tsx
│   ├── LoginPage.test.tsx
│   ├── LoginPage.integration.test.tsx
│   └── LoginPage.a11y.test.tsx
├── src/components/
│   ├── DraftRestoreModal.tsx
│   ├── ProjectSwitcher.tsx
│   └── Sidebar.tsx
├── src/app/
│   └── ErrorBoundary.tsx
└── src/styles/
    ├── tokens.ts
    └── index.css
```

**Structure Decision**: Keep the implementation entirely within the existing `frontend/` module. The visual fix will be made at the shared theme-token/styling layer and then applied to the affected pages and components without changing backend or BFF behavior.

## Phase 0: Research Output

Research is documented in `/specs/010-fix-login-contrast/research.md` and resolves the technical approach as follows:

1. Treat the defect as a **shared token mismatch**, not a login-page-only bug.
2. Separate **solid primary action** styling from lighter **accent text/focus** usage so dark mode can keep both readable.
3. Keep first-paint readability by updating CSS variable mappings that load before React render.
4. Protect the fix with targeted frontend regression coverage for the login page and shared primary-action surfaces.

## Phase 1: Design & Contracts Output

1. **Data model**: `/specs/010-fix-login-contrast/data-model.md`
   - Defines the theme mode, primary button style contract, interaction states, and affected UI surfaces.
2. **UI contract**: `/specs/010-fix-login-contrast/contracts/primary-action-theme-contract.md`
   - Documents the approved semantic-token pairing and expected behavior for default, hover, focus, loading, and disabled states.
3. **Quickstart**: `/specs/010-fix-login-contrast/quickstart.md`
   - Provides manual validation and local verification steps for the dark-mode fix.

## Implementation Strategy (Phase 2 Preview)

1. Audit every solid primary-action usage currently bound to `var(--color-primary)` in dark mode.
2. Introduce or refine semantic/component tokens so button background, foreground, hover, and disabled states have explicit contrast-safe roles.
3. Update `LoginPage.tsx` and the other shared-action surfaces (`ErrorBoundary.tsx`, `DraftRestoreModal.tsx`, `ProjectSwitcher.tsx`) to consume the new contract instead of relying on a fragile background + `text-white` combination.
4. Preserve `Sidebar.tsx` accent text/focus behavior independently so navigation styling is not regressed by the button fix.
5. Add or update frontend tests to verify initial dark-mode readability, theme-switch/page-refresh stability, stable interaction states, and light-mode non-regression.
6. Verify with `npm run test`, `npm run lint`, and `npm run typecheck` in `frontend/` after implementation.

## Mandatory End-of-Execution Sync

**Memory Diagram Sync (required)**:
- Impacted memory files: None
- Update tasks required in `tasks.md`: No
- If `No`, rationale: This feature is limited to frontend theming and shared button presentation; it does not alter any service flow, architecture diagram, or backend memory artifact.

**Instruction Sync (required for refactor/reorganization)**:
- Refactor/reorganization in scope: No

**Completion gate**:
Implementation is not complete until the shared dark-mode primary-action contract is applied to all affected surfaces and the relevant frontend validation suite passes.

## Complexity Tracking

No constitution violations or exceptional complexity require justification.
