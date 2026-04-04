# Research: Dark Mode Login Button Contrast

## Decision 1: Fix the issue at the shared token contract level

**Decision**: Treat the unreadable `Sign in` button as a shared design-token mismatch and correct it through the central theme contract instead of a one-off page patch.

**Rationale**:
- `frontend/src/styles/tokens.ts` and `frontend/src/styles/index.css` define `dark` theme `--color-primary` as a very light indigo value.
- Multiple components render solid action buttons using that background together with `text-white`, including `frontend/src/pages/LoginPage.tsx`, `frontend/src/app/ErrorBoundary.tsx`, `frontend/src/components/DraftRestoreModal.tsx`, and `frontend/src/components/ProjectSwitcher.tsx`.
- The hover state appears readable because `--color-primary-hover` switches to a darker indigo, which confirms the defect is caused by the default dark-mode background/foreground pairing.

**Alternatives considered**:
- **Patch only `LoginPage.tsx`**: Rejected because the same shared styling exists in other components and would reintroduce the problem elsewhere.
- **Change `--color-primary` globally without separating roles**: Rejected because `Sidebar.tsx` also uses the primary token for accent text and focus treatment, so a global swap could regress other UI states.

---

## Decision 2: Separate solid-action colors from accent text colors

**Decision**: Use an explicit contrast-safe token pair for solid primary actions so background and foreground colors are defined together for each theme and state.

**Rationale**:
- A single `colorPrimary` token is currently carrying too many responsibilities: solid button backgrounds, active text color, and focus accents.
- Dark mode needs different values for readable accent text versus readable solid action backgrounds.
- A dedicated primary-action contract prevents future regressions and aligns with the repo's design-token guidance.

**Alternatives considered**:
- **Keep hardcoded `text-white` and only darken the background on the login page**: Rejected because it would not address other shared buttons and would keep the design system fragile.
- **Use inline per-component overrides**: Rejected because it bypasses the centralized token system and makes maintenance harder.

---

## Decision 3: Preserve first-paint readability through CSS variable updates

**Decision**: Update the semantic token mapping in the CSS variable layer that loads before React renders instead of calculating colors at runtime in components.

**Rationale**:
- The bug is visible on initial page load, so the fix must be available before user interaction.
- Theme variables in `frontend/src/styles/index.css` are already applied at the root/theme-class level and are the right place to avoid low-contrast flashes.

**Alternatives considered**:
- **Compute button colors inside React on mount**: Rejected because it can still allow a brief incorrect first paint and adds unnecessary presentation logic to components.

---

## Decision 4: Protect the fix with focused regression coverage

**Decision**: Extend the existing frontend tests around `LoginPage` and shared UI surfaces to verify the dark-mode button contract and non-regression in light mode.

**Rationale**:
- This repo already contains `LoginPage.test.tsx`, `LoginPage.integration.test.tsx`, and `LoginPage.a11y.test.tsx`, which are natural places to add regression coverage.
- Automated checks reduce the chance of a future token change reintroducing unreadable primary buttons.

**Alternatives considered**:
- **Manual QA only**: Rejected because the issue is tied to a shared theme contract and can silently regress during later styling work.

---

## Resulting Direction

The implementation plan will:
1. refine the shared dark-mode primary-action token contract,
2. update all affected solid buttons to use that contract consistently, and
3. add frontend validation to ensure default, hover, focus, loading, and disabled states remain readable in dark mode without breaking light mode.