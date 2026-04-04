# Data Model: Dark Mode Login Button Contrast

## Overview

This feature does not introduce backend data storage. Its "data model" is the UI state and design-token contract that determine whether a primary action remains readable in dark mode.

## Entities

### 1. ThemeMode

| Field | Type | Description |
|------|------|-------------|
| `mode` | `light | dark` | The active UI theme applied to the application |
| `source` | `system | saved-preference` | Where the theme choice originated |
| `appliesBeforeRender` | `boolean` | Whether the theme class/tokens are ready before the first paint |

**Validation rules**:
- `mode` must always resolve to a valid supported theme.
- The selected mode must map to a complete token set.

---

### 2. PrimaryActionStyleContract

| Field | Type | Description |
|------|------|-------------|
| `role` | `string` | Semantic role for the action, e.g. primary solid button |
| `backgroundToken` | `string` | Token used for the default button background |
| `foregroundToken` | `string` | Token used for the button label/icon color |
| `hoverBackgroundToken` | `string` | Token used when hovered |
| `focusRingToken` | `string` | Token used for focus treatment |
| `disabledBackgroundToken` | `string` | Token or rule for disabled/loading appearance |
| `contrastExpectation` | `string` | Human-readable readability requirement for the state pair |

**Validation rules**:
- The background and foreground tokens must form a readable pair in both supported themes.
- Hover, focus, loading, and disabled states must remain readable and visually distinct.
- The contract must be reused consistently by all solid primary actions that share this styling.

---

### 3. InteractionState

| State | Meaning | Required Outcome |
|------|---------|------------------|
| `default` | Initial visible state on first render | Label is readable immediately |
| `hover` | Pointer interaction | Contrast stays readable and the control still feels interactive |
| `focus` | Keyboard/screen-reader navigation state | Focus indication is visible without obscuring the label |
| `pressed` | Click/tap feedback | The button stays visually distinct and readable |
| `loading` | Async submission in progress | Status text remains legible while showing busy state |
| `disabled` | Temporarily unavailable | State looks inactive but the label remains readable |

---

### 4. AffectedActionSurface

| Surface | Current Responsibility | Impact |
|---------|------------------------|--------|
| `frontend/src/pages/LoginPage.tsx` | Primary authentication submit action | Highest priority validation target |
| `frontend/src/app/ErrorBoundary.tsx` | Recovery / retry action | Must inherit the same safe solid-button contract |
| `frontend/src/components/DraftRestoreModal.tsx` | Restore draft confirmation action | Must remain readable in dark mode |
| `frontend/src/components/ProjectSwitcher.tsx` | Invite/submit action | Must not reuse the broken background/foreground pairing |

## State Relationships

1. `ThemeMode` selects the active semantic token mapping.
2. `PrimaryActionStyleContract` resolves the correct colors for the current theme.
3. `InteractionState` modifies the contract while preserving readability.
4. `AffectedActionSurface` consumes the contract to render visible buttons consistently.

## Non-Goals

- No new backend entity, API payload, or persistence model is introduced.
- Login behavior, routing, and authentication payload formats remain unchanged.