# Tasks: Frontend Authentication & Navigation System

**Input**: Design documents from `/specs/002-frontend-auth-navigation/`  
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/auth-bff.openapi.yaml  
**Tests**: Included (feature specifies comprehensive integration and accessibility tests)  
**Organization**: Tasks grouped by user story (US1, US2, US3) to enable independent implementation and testing

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and frontend structure verification

- [x] T001 Verify React 18.3.x + TypeScript 5.8.x setup in frontend/package.json
- [x] T002 [P] Verify react-router-dom 6.30.x installed in frontend/package.json
- [x] T003 [P] Verify @tanstack/react-query 5.76.x installed in frontend/package.json
- [x] T004 [P] Verify zod 3.24.x installed for schema validation in frontend/package.json
- [x] T005 [P] Verify Vitest 3.2.x and Testing Library installed for tests
- [x] T006 [P] Verify MSW (Mock Service Worker) installed for API mocking in tests
- [x] T007 Create frontend/src/types/auth.ts for TypeScript auth interfaces (AuthenticationContext, User, Session)
- [x] T008 Create frontend/src/services/api.client.ts with `credentials: include` HTTP-only cookie support

**Checkpoint**: Frontend environment configured and auth types ready

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core authentication and navigation infrastructure that MUST be complete before user story implementation

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T009 [P] Create frontend/src/hooks/useAuthContext.ts hook with auth state management (login, logout, refreshAccessToken, setActiveProject)
- [x] T010 [P] Create frontend/src/hooks/useAuthRefresh.ts hook with 75% of token lifetime refresh scheduling logic
- [x] T011 [P] Create frontend/src/types/auth-response.schema.ts with zod schema for BFF login/refresh response validation
- [x] T012 [P] Create frontend/src/types/lockout.schema.ts with zod schema for lockout metadata and countdown calculations
- [x] T013 Create frontend/src/hooks/useTokenRefreshInterceptor.ts hook to intercept 401 responses, refresh token, and retry original request
- [x] T014 Create frontend/src/hooks/useDraftStateRestore.ts hook to manage short-lived draft state restoration after re-login (with TTL enforcement)
- [x] T015 [P] Create frontend/src/components/Sidebar.tsx component skeleton with menu item list and active highlighting
- [x] T016 [P] Create frontend/src/components/HamburgerMenu.tsx component skeleton for mobile/tablet navigation toggle
- [x] T017 [P] Create frontend/src/components/SkeletonPlaceholder.tsx reusable skeleton component for login and protected-page loading states
- [x] T018 Create frontend/src/app/AppLayout.tsx layout component to wire Sidebar + HamburgerMenu + main content area with responsive behavior
- [x] T019 [P] Create frontend/src/types/navigation.ts for NavigationState, BreadcrumbItem, and MenuItem interfaces
- [x] T020 Update frontend/src/styles/tokens.ts to include auth/login-specific design token overrides (accessible form styling, error states, lockout countdown colors)
- [x] T021 Create frontend/src/hooks/useResponsiveNavigation.ts hook to detect viewport size and toggle sidebar visibility (desktop 1024px+, tablet 768-1023px, mobile <768px)

**Checkpoint**: Auth lifecycle management, token interceptors, navigation components, and layout infrastructure ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Default User Auto-Filled Login (Priority: P0) 🎯 MVP

**Goal**: Deliver a secure, responsive login screen with pre-filled default development credentials, skeleton loading feedback, and lockout-aware error handling

**Independent Test**: Launch frontend, verify login screen displays with default credentials pre-filled, attempt login with invalid credentials to verify error message and lockout handling, submit valid login and confirm successful navigation to dashboard with token stored in HTTP-only cookies

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T022 [P] [US1] Create frontend/src/hooks/useAuthSession.test.ts contract test for BFF POST /api/auth/login with valid/invalid credentials response envelope validation
- [x] T023 [P] [US1] Create frontend/src/pages/LoginPage.test.tsx unit test suite with MSW mocks: default pre-fill, skeleton display within 300ms, error handling, lockout countdown
- [x] T024 [US1] Create frontend/src/pages/LoginPage.integration.test.tsx integration test: user lands on login → sees auto-filled credentials → submits → lands on dashboard under 5s
- [x] T025 [US1] Create frontend/src/pages/LoginPage.a11y.test.tsx accessibility test: keyboard-only navigation (Tab, Enter) without mouse

### Implementation for User Story 1

- [x] T026 [P] [US1] Create frontend/src/pages/LoginPage.tsx component with auto-filled defaults from env vars, skeleton placeholder, error display, lockout countdown
- [x] T027 [P] [US1] Create frontend/src/hooks/useAuthSession.ts hook with login(username, password) mutation, BFF POST /api/auth/login, response validation, error/lockout handling
- [x] T028 [P] [US1] Create frontend/src/components/ErrorMessage.tsx component with clear error display and lockout countdown timer
- [x] T029 [P] [US1] Create frontend/src/types/auth-response.schema.ts with zod schema for login/refresh response and lockout metadata
- [x] T030 [P] [US1] Add loading state indicators to LoginPage: skeleton placeholder appears within 300ms, login button shows disabled + spinner during submission
- [x] T031 [P] [US1] Create frontend/src/config/auth.config.ts for VITE_DEFAULT_USERNAME/VITE_DEFAULT_PASSWORD from environment with production safety check
- [x] T032 [US1] Add keyboard accessibility to LoginPage: Tab order (username → password → button), Enter submits, focus management
- [x] T033 [US1] Update frontend/src/styles/tokens.ts to include login-specific design tokens: colorInputError, colorLockoutWarning, colorLoadingSkeleton

**Checkpoint**: User Story 1 complete - login screen fully functional, all tests pass, default credentials work, lockout protection active, skeleton loading visible, error messages clear, keyboard accessible

---

## Phase 4: User Story 2 - Sidebar Navigation with Screen Routes (Priority: P1)

**Goal**: Implement persistent sidebar navigation with responsive mobile hamburger menu, active route highlighting, and full screen routing support

**Independent Test**: Log in successfully, verify sidebar displays with all menu items, click each navigation item and confirm correct screen loads without page reload and active item highlights, toggle hamburger menu on mobile viewport and confirm sidebar visibility

### Tests for User Story 2

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T034 [P] [US2] Create frontend/src/components/Sidebar.test.tsx unit test: all menu items render, current route highlighted, click navigates, keyboard-accessible
- [x] T035 [P] [US2] Create frontend/src/components/HamburgerMenu.test.tsx unit test: hidden on desktop, visible on mobile/tablet, toggle sidebar visibility
- [x] T036 [US2] Create frontend/src/app/AppLayout.integration.test.tsx integration test: log in → sidebar displays → click each route → page loads → URL updates → active highlighting accurate
- [x] T037 [US2] Create frontend/src/app/AppLayout.a11y.test.tsx accessibility test: menu keyboard-traversable, active item announced by screen readers, hamburger has aria-label

### Implementation for User Story 2

- [x] T038 [P] [US2] Create frontend/src/components/Sidebar.tsx component with menu items (Dashboard, Documents, Bills, Payments, Analytics, Settings), active highlighting, icons, keyboard nav
- [x] T039 [P] [US2] Create frontend/src/components/HamburgerMenu.tsx component with three-line icon, aria-label, aria-pressed, desktop hidden, mobile/tablet visible
- [x] T040 [US2] Update frontend/src/app/AppLayout.tsx to integrate Sidebar + HamburgerMenu with responsive grid layout: desktop (200px fixed sidebar + flex content), mobile (full width toggle)
- [x] T041 [P] [US2] Create frontend/src/app/router.config.ts with route definitions: Dashboard (/dashboard), Documents, Bills, Payments, Analytics, Settings with path/element/icon/label
- [x] T042 [P] [US2] Create placeholder page components: DashboardPage, DocumentsPage, BillsPage, PaymentsPage, AnalyticsPage, SettingsPage in frontend/src/pages/
- [x] T043 [US2] Update frontend/src/main.tsx to wire BrowserRouter with configured routes from router.config.ts and AppLayout as root container
- [x] T044 [US2] Implement responsive behavior via frontend/src/hooks/useResponsiveNavigation.ts: track viewport, breakpoints (≥1024px desktop, 768-1023px tablet, <768px mobile)
- [x] T045 [P] [US2] Create frontend/src/types/navigation.ts with MenuItem, NavigationConfig interfaces
- [x] T046 [P] [US2] Add navigation design tokens to frontend/src/styles/tokens.ts: colorSidebarBackground, colorMenuItemText, colorMenuItemActive, colorHamburgerIcon, sidebarWidth, hamburgerButtonSize
- [x] T047 [US2] Auto-hide sidebar on mobile after navigation: detect mobile in Sidebar, close sidebar after link click via parent callback

**Checkpoint**: User Story 2 complete - sidebar navigation fully functional on all viewports, responsive hamburger menu, route switching without page reload, active highlighting accurate, accessibility tests pass

---

## Phase 5: User Story 3 - Token Refresh and Session Persistence (Priority: P1)

**Goal**: Implement automatic token refresh at 75% of lifetime, session persistence across page reloads, automatic retry on 401, and draft state restoration after forced re-login

**Independent Test**: Log in, wait for token to approach refresh time, verify no interruption and token refreshes silently, make authenticated API call to verify new token works, refresh browser and confirm logged-in state persists

### Tests for User Story 3

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T048 [P] [US3] Create frontend/src/hooks/useAuthRefresh.test.ts unit test: refresh scheduled at 75% of lifetime, refresh request sent to BFF, new token stored, timer reset, silent operation
- [x] T049 [P] [US3] Create frontend/src/hooks/useTokenRefreshInterceptor.test.ts unit test: 401 intercepted, token refresh attempted, original request retried, failed refresh redirects
- [x] T050 [P] [US3] Create frontend/src/hooks/useDraftStateRestore.test.ts unit test: draft saved to localStorage with TTL, restored after re-login, one-time restore enforced
- [x] T051 [P] [US3] Create frontend/src/services/auth.api.test.ts test: BFF refresh endpoint POST /api/auth/refresh accepts refresh_token, returns new token + expires_in
- [x] T052 [US3] Create frontend/src/app/AppLayout.integration.test.tsx test for token refresh lifecycle: login → wait 75% of lifetime → verify silent refresh → API call succeeds
- [x] T053 [US3] Create LoginPage integration test: login → edit document → force token expiry → verify draft saved → re-login → verify restore prompt → restore → continue editing

### Implementation for User Story 3

- [x] T054 [US3] Update frontend/src/hooks/useAuthRefresh.ts: calculate refreshTime = expiresIn * 0.75, schedule timer with setInterval, POST to /api/auth/refresh, handle response, reset timer on unmount
- [x] T055 [US3] Update frontend/src/services/api.client.ts to add refreshAccessToken(refreshToken) function: POST /api/auth/refresh, credentials: include, zod validation, return AuthResponse
- [x] T056 [US3] Create frontend/src/hooks/useTokenRefreshInterceptor.ts: intercept all API requests, on 401 attempt refresh, retry original request, prevent retry loop (max 1 attempt)
- [x] T057 [US3] Create frontend/src/hooks/usePersistentSession.ts: serialize auth context to localStorage on login, restore session on app mount, check expiry, clear on logout
- [x] T058 [US3] Create frontend/src/hooks/useDraftStateRestore.ts: saveDraftState(key, data, ttl), getDraftState(key), clearDraftState(key) with TTL enforcement and one-time restore
- [x] T059 [P] [US3] Create frontend/src/components/DraftRestoreModal.tsx: modal with "Restore your work?" prompt, Restore/Discard buttons, auto-dismiss on TTL expiry
- [x] T060 [US3] Update LoginPage to integrate draft state restoration: check for available draft after login, show DraftRestoreModal, navigate with draft data on restore
- [x] T061 [P] [US3] Add HTTP-only cookie handling to frontend/src/services/api.client.ts: include credentials: 'include' in all requests, document that BFF sets HttpOnly/SameSite=Strict
- [x] T062 [US3] Create frontend/src/types/session.schema.ts with zod schemas: session metadata (userId, expiryTimestamp, refreshAtTimestamp), draft restore data (key, data, savedAt, ttl, used)
- [x] T063 [P] [US3] Update frontend/src/styles/tokens.ts with session tokens: colorSessionWarning, colorDraftRestoreModal

**Checkpoint**: User Story 3 complete - token refresh works silently at 75% of lifetime, session persists across page reloads, 401 triggers refresh + retry, draft state restoration available, all tests pass

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final integration, documentation, performance, and security hardening

- [x] T064 [P] Run full integration test suite: LoginPage, AppLayout, token refresh tests
- [x] T065 [P] Run accessibility test suite: LoginPage.a11y + AppLayout.a11y tests
- [x] T066 Create/update frontend/README.md: auth system architecture, env variables, login flow diagram, token refresh, draft restoration, running tests
- [x] T067 [P] Update specs/002-frontend-auth-navigation/quickstart.md: step-by-step testing (login, navigate, refresh, logout, restore draft)
- [x] T068 Create specs/002-frontend-auth-navigation/checklists/implementation.md with verification checklist covering all functional requirements
- [x] T069 [P] Security review: credentials not logged, refresh token HTTP-only, CSRF handling, no sensitive data in localStorage except draft metadata
- [x] T070 [P] Performance optimization: code splitting (lazy-load pages), memoize components, useCallback for handlers, Core Web Vitals, Lighthouse ≥90
- [x] T071 Add error boundary to frontend/src/app/root.tsx: catch render errors, display fallback with session-expired message
- [x] T072 Final validation against spec success criteria in frontend/src/VALIDATION.md: verify all SC-001 through SC-010

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - **BLOCKS all user stories**
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - Can proceed in parallel (separate developers) after foundational is done
  - Or sequentially in priority order (P0 → P1 → P1)
- **Polish (Phase 6)**: Depends on all desired user stories being complete (OR at least US1 for MVP validation)

### User Story Dependencies

- **User Story 1 (P0)**: Can start after Foundational (Phase 2) - MVP entry point, no dependencies on other stories
- **User Story 2 (P1)**: Can start after Foundational (Phase 2) - Independent of US1 (uses same auth context)
- **User Story 3 (P1)**: Can start after Foundational (Phase 2) - Independent of US1/US2 (extends auth context)

### Within Each User Story

- Tests MUST be written FIRST and FAIL before implementation begins
- Models/types (marked [P]) before hooks/services
- Hooks/services before components
- Components before integration/E2E tests
- Core implementation before Polish tasks

### Parallel Opportunities

**Phase 1 Setup**: All tasks marked [P] can run in parallel (T002-T006)

**Phase 2 Foundational**: All tasks marked [P] can run in parallel
- T009, T010 (auth hooks)
- T011, T012 (auth schemas)
- T015, T016, T017 (navigation components)
- T019 (navigation types)
- T020, T021 (design tokens and responsive hook)

**After Foundational completes**: All user stories can proceed in parallel
- Developer A: US1 (Phase 3) - Login page development
- Developer B: US2 (Phase 4) - Sidebar navigation development
- Developer C: US3 (Phase 5) - Token refresh and persistence development

**Within Each User Story**:

*US1 Tests*: T022, T023 parallel  
*US1 Implementation*: T026, T027, T028, T029, T030, T031 parallel → T032, T033 sequentially

*US2 Tests*: T034, T035 parallel  
*US2 Implementation*: T038, T039 parallel → T040, T041, T042, T043, T044, T045, T046, T047 sequentially

*US3 Tests*: T048, T049, T050, T051 parallel  
*US3 Implementation*: T054, T055, T056, T057, T058 parallel → T059, T060, T061, T062, T063 sequentially

---

## Implementation Strategy

### MVP First (User Story 1 Only)

Recommended for initial release:

1. ✅ Complete Phase 1: Setup (T001-T008)
2. ✅ Complete Phase 2: Foundational (T009-T021) — **CRITICAL GATE**
3. ✅ Complete Phase 3: User Story 1 (T022-T033) — **MVP FEATURE**
4. 🧪 Validate: Run login-to-dashboard flow end-to-end
5. 🚀 Deploy/Demo: Users can log in and see dashboard

### Incremental Delivery (MVP → Full Feature)

1. **Iteration 1**: Setup + Foundational + US1 → Login works ✓
2. **Iteration 2**: Add US2 → Navigation works ✓
3. **Iteration 3**: Add US3 → Token refresh + persistence works ✓
4. **Final**: Polish + validation → Feature complete ✓

### Suggested Team Approach (3+ developers)

1. **Day 1-2**: All team together on Setup + Foundational (Phase 1-2)
2. **Day 3-4**: Split into three parallel streams:
   - Developer A: US1 complete (Phase 3)
   - Developer B: US2 complete (Phase 4)
   - Developer C: US3 complete (Phase 5)
3. **Day 5**: All team converges on Polish + validation (Phase 6)

---

## Notes

- ✅ [P] tasks = different files, no dependencies - safe to parallelize
- ✅ [US1/US2/US3] label = directly contributing to specific user story
- ✅ Each user story independently completable and testable per acceptance scenarios in spec.md
- ✅ Verify ALL tests fail before implementation begins
- ✅ Commit after each task or logical group (every 2-3 tasks)
- ✅ Stop at any checkpoint to validate story independently before proceeding
- ✅ Skeleton placeholders must appear within 300ms (SC-009)
- ✅ Token refresh at exactly 75% of token lifetime
- ✅ HTTP-only cookies = no JavaScript access to tokens
- ✅ Draft state TTL = short window (suggested 30 min max)
- ✅ Lockout = 5 failures in 15 minutes (server-enforced)
- ✅ Keyboard accessibility = Tab + Enter at minimum
- ✅ Responsive breakpoints = Desktop ≥1024px, Tablet 768-1023px, Mobile <768px
- ⚠️ Do NOT test token refresh with real backend token expiry (mock with MSW)
- ⚠️ Do NOT hardcode credentials in source code (use env variables)
- ⚠️ Do NOT access HTTP-only cookies with `document.cookie` (browser prevents this)
- [x] T038 [P] [US3] Add refresh endpoint integration test for cookie-based session in `backend/tests/integration/auth_refresh_cookie_test.go`

### Implementation for User Story 3

- [x] T039 [US3] Implement refresh API call and session metadata mapping in `frontend/src/services/auth_service.ts`
- [x] T040 [US3] Implement `useAuthRefresh` hook with 75%-lifetime scheduling in `frontend/src/hooks/useAuthRefresh.ts`
- [x] T041 [US3] Implement 401 interception and single-flight retry policy in `frontend/src/services/http_client.ts`
- [x] T042 [US3] Implement short-TTL draft state persistence utility in `frontend/src/services/draft_restore_store.ts`
- [x] T043 [US3] Implement one-time draft restoration after re-login in `frontend/src/hooks/useDraftRestore.ts`
- [x] T044 [US3] Implement session-expired redirect flow preserving draft state in `frontend/src/app/auth_provider.tsx`
- [x] T045 [US3] Add protected-page bootstrap skeleton states in `frontend/src/components/layout/app_shell.tsx`
- [x] T046 [US3] Update BFF refresh controller response metadata (`expiresIn`, `refreshAt`) in `backend/internals/bff/financial/controllers/auth_controller.go`
- [x] T047 [US3] Update BFF auth refresh service behavior and cookie rotation in `backend/internals/bff/services/auth_service.go`

**Checkpoint**: User Story 3 is independently functional and testable.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Hardening, consistency, and end-to-end validation across all stories.

- [x] T048 [P] Update auth/navigation quickstart verification steps in `specs/002-frontend-auth-navigation/quickstart.md`
- [x] T049 [P] Add accessibility refinements for loading and navigation focus behavior in `frontend/src/components/navigation/sidebar.tsx`
- [x] T050 Validate responsive behavior at 320/375/768/1024/1280 widths in `frontend/src/components/layout/app_shell.tsx`
- [x] T051 Run full frontend test suite and fix regressions in `frontend/package.json`
- [x] T052 Run backend integration tests covering auth contract and lockout in `backend/tests/integration/`

---

## Dependencies & Execution Order

### Phase Dependencies

- Setup (Phase 1): starts immediately
- Foundational (Phase 2): depends on Setup completion; blocks user stories
- User Stories (Phases 3-5): depend on Foundational completion
- Polish (Phase 6): depends on desired user stories complete

### User Story Dependencies

- US1 (P0): starts after Phase 2; MVP
- US2 (P1): starts after Phase 2; independent from US3
- US3 (P1): starts after Phase 2; integrates with US1 auth base

### Within Each User Story

- Tests first (fail before implementation)
- Services/hooks before page composition wiring
- Frontend behavior and BFF contract alignment completed before story checkpoint

### Parallel Opportunities

- Phase 1: T003, T004 and T006 parallel
- Phase 2: T010, T011 and T013 parallel
- US1: T015, T016 and T017 parallel
- US2: T026, T027 and T028 parallel
- US3: T035, T036, T037 and T038 parallel
- Polish: T048 and T049 parallel

---

## Parallel Example: User Story 1

```bash
# Parallel test authoring
T015 frontend/src/services/auth_service.test.ts
T016 frontend/src/pages/login_page.test.tsx
T017 backend/tests/integration/auth_lockout_test.go

# Parallel implementation chunks after tests
T019 frontend/src/hooks/useLogin.ts
T023 frontend/src/components/auth/lockout_notice.tsx
T024 backend/internals/bff/financial/controllers/auth_controller.go
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 and Phase 2
2. Complete Phase 3 (US1)
3. Validate lockout + login success/failure + loading skeleton behavior
4. Demo/deploy MVP auth entrypoint

### Incremental Delivery

1. Foundation complete
2. Deliver US1 (login)
3. Deliver US2 (navigation shell)
4. Deliver US3 (session lifecycle)
5. Finish with polish and full regression run

### Parallel Team Strategy

1. Team completes Setup + Foundational
2. Developer A: US1 frontend/login UX + tests
3. Developer B: US2 sidebar/hamburger + accessibility
4. Developer C: US3 refresh/draft-restore + backend refresh contract alignment
