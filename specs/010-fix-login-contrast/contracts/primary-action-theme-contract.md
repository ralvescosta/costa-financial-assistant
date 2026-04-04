# UI Contract: Primary Action Theme Contract

## Purpose

Define the required appearance contract for solid primary-action buttons so they remain readable and visually consistent in both light and dark themes.

## Scope

This contract applies to any frontend surface that renders a solid primary action using the shared theme system, including:
- `frontend/src/pages/LoginPage.tsx`
- `frontend/src/app/ErrorBoundary.tsx`
- `frontend/src/components/DraftRestoreModal.tsx`
- `frontend/src/components/ProjectSwitcher.tsx`

## Contract Rules

### 1. Semantic Ownership

- The authoritative theme values must live in:
  - `frontend/src/styles/tokens.ts`
  - `frontend/src/styles/index.css`
- Components must consume the shared semantic/component tokens rather than hardcoded light/dark color literals.

### 2. Required State Behavior

| State | Contract Requirement |
|------|-----------------------|
| `default` | Button label is readable on first paint in both themes |
| `hover` | Hover feedback preserves readability and action emphasis |
| `focus` | Focus ring is visible without lowering text/background contrast |
| `pressed` | Active state remains readable and clearly interactive |
| `loading` | Busy indicator and label text remain legible |
| `disabled` | Disabled styling communicates inactivity while keeping the label readable |

### 3. Theme Expectations

- **Light theme**: existing primary-action clarity must be preserved.
- **Dark theme**: the background/foreground pair must be chosen as a readable set for solid buttons and must not rely on an incompatible fallback such as white text over a very light button fill.

### 4. Reuse Policy

- Any new solid primary button introduced later must reuse this contract.
- If a surface needs a special-case appearance, it must define a new explicit token role rather than silently overriding the shared contract inline.

## Validation Checklist

- The login page `Sign in` button is readable on first render in dark mode.
- Shared button states remain readable during hover, focus, loading, and disabled transitions.
- Light-mode appearance remains acceptable after the dark-mode correction.
- No affected component continues to rely on the broken background/foreground pairing.