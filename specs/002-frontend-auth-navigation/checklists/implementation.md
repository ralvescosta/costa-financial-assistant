# Implementation Checklist — 002-frontend-auth-navigation

Verification checklist for the Frontend Authentication & Navigation System implementation.

## Phase 1 — Setup

- [x] React 18.3.x + TypeScript 5.8.x confirmed in `frontend/package.json`
- [x] react-router-dom 6.30.x installed
- [x] @tanstack/react-query 5.76.x installed
- [x] zod 3.24.x installed
- [x] Vitest 3.2.x + @testing-library/react installed
- [x] MSW 2.7.x installed
- [x] `frontend/src/types/auth.ts` created with `AuthenticationContext`, `User`, `ActiveProject`, `Session` interfaces
- [x] `frontend/src/services/api.client.ts` created with `credentials: 'include'`

## Phase 2 — Foundational Infrastructure

- [x] `useAuthContext.tsx` with `AuthProvider` and `useAuthContext()` created
- [x] `useAuthRefresh.ts` schedules refresh at 75% of `expires_in`
- [x] `useTokenRefreshInterceptor.ts` patches `window.fetch`, retries once on 401
- [x] `useDraftStateRestore.ts` with TTL enforcement and one-time restore
- [x] `useResponsiveNavigation.ts` with desktop/tablet/mobile breakpoints
- [x] `auth-response.schema.ts` zod schemas for login/refresh envelopes
- [x] `lockout.schema.ts` zod schemas + `calcLockoutRemainingSeconds()` helper
- [x] `navigation.ts` interfaces: `MenuItem`, `NavigationState`, `BreadcrumbItem`
- [x] `session.schema.ts` schemas for `SessionMetadata` and `DraftRestoreData`
- [x] `SkeletonPlaceholder.tsx` — `role="status"` accessible skeleton
- [x] `HamburgerMenu.tsx` — `aria-label` + `aria-pressed`, `lg:hidden`
- [x] `Sidebar.tsx` — 6 nav items, `NavLink` with `aria-current="page"`, mobile backdrop
- [x] `AppLayout.tsx` — wires sidebar + hamburger + refresh hooks
- [x] `router.config.ts` — `NAVIGATION_ITEMS` and `ROUTES` constants
- [x] Design tokens extended in `tokens.ts` + `index.css` + `tailwind.config.js`

## Phase 3 — User Story 1 (Login)

- [x] `LoginPage.tsx` created with skeleton, error message, lockout guard
- [x] `useAuthSession.ts` wraps `useAuthContext` with login/logout surface
- [x] `ErrorMessage.tsx` with lockout countdown timer
- [x] `auth.config.ts` reads `VITE_DEFAULT_USERNAME/PASSWORD`, nullified in production
- [x] Login form — Tab order: username → password → button; Enter submits
- [x] Skeleton appears during submission; button disabled with spinner
- [x] Unit test: `useAuthSession.test.tsx` — 3 tests passing
- [x] Unit test: `LoginPage.test.tsx` — all tests passing
- [x] Integration test: `LoginPage.integration.test.tsx` — passes
- [x] a11y test: `LoginPage.a11y.test.tsx` — passes

## Phase 4 — User Story 2 (Navigation)

- [x] `Sidebar.tsx` renders all 6 menu items with `NavLink` + `aria-current="page"`
- [x] `HamburgerMenu.tsx` hidden at `lg:`, `aria-pressed` toggles correctly
- [x] `AppLayout.tsx` integrates responsive sidebar + content
- [x] `router.tsx` uses `ProtectedLayout` + `<Outlet>` pattern
- [x] `main.tsx` wraps with `AuthProvider` + `ErrorBoundary`
- [x] `DashboardPage.tsx` and `BillsPage.tsx` placeholder pages created
- [x] Unit test: `Sidebar.test.tsx` — 5 tests passing
- [x] Unit test: `HamburgerMenu.test.tsx` — 6 tests passing
- [x] Integration test: `AppLayout.integration.test.tsx` — passes
- [x] a11y test: `AppLayout.a11y.test.tsx` — passes

## Phase 5 — User Story 3 (Token Refresh & Session Persistence)

- [x] `useAuthRefresh.ts` sets timer at `refreshAtTimestamp`, cancels on unmount
- [x] `useTokenRefreshInterceptor.ts` patches fetch, max 1 retry, restores on unmount
- [x] `usePersistentSession.ts` reads `cfa:session` from localStorage
- [x] `useDraftStateRestore.ts` — `saveDraftState`, `getDraftState`, `clearDraftState`
- [x] `DraftRestoreModal.tsx` — one-time restore prompt, auto-dismisses on TTL expiry
- [x] `session.schema.ts` — `SessionMetadataSchema` + `DraftRestoreDataSchema`
- [x] `useAuthContext.tsx` persists session to `cfa:session` on login; clears on logout
- [x] Unit test: `useAuthRefresh.test.tsx` — 3 tests passing
- [x] Unit test: `useTokenRefreshInterceptor.test.tsx` — 3 tests passing
- [x] Unit test: `useDraftStateRestore.test.ts` — 6 tests passing
- [x] Unit test: `auth.api.test.ts` — 2 tests passing

## Phase 6 — Polish & Security

- [x] `ErrorBoundary.tsx` added to `main.tsx` as outermost wrapper
- [x] Full test suite run: 21 test files, 98 tests — all passing
- [x] Accessibility tests pass: keyboard navigation, aria-current, aria-label
- [x] `@testing-library/jest-dom` installed and configured in `test/setup.ts`
- [x] `window.matchMedia` mock in `test/setup.ts` for jsdom environment
- [x] `frontend/README.md` created with arch docs, env vars, hook reference
- [x] `quickstart.md` updated with step-by-step testing guide
- [x] Security: no tokens stored in JS-accessible storage
- [x] Security: default credentials nullified in production via `auth.config.ts`
- [x] Security: session metadata only (userId, expiry, refreshAt) stored in localStorage
- [x] Performance: page routes lazy-loaded via `React.lazy()` in `router.tsx`

## Test Results Summary

```
Test Files  21 passed (21)
     Tests  98 passed (98)
```
