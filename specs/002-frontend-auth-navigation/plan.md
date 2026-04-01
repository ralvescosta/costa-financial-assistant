# Implementation Plan: Frontend Authentication & Navigation System

**Branch**: `002-frontend-auth-navigation` | **Date**: 2026-04-01 | **Spec**: `/specs/002-frontend-auth-navigation/spec.md`
**Input**: Feature specification from `/specs/002-frontend-auth-navigation/spec.md`

## Summary

Implement a secure, production-safe frontend authentication shell with default dev-only credential auto-fill, BFF-backed login/refresh using HTTP-only cookies, lockout-aware error handling, responsive sidebar + hamburger navigation, and skeleton-first loading UX. The implementation keeps all server-state and auth orchestration in hooks, keeps pages composition-only, and preserves short-lived draft state on refresh failure for one-time restore after re-login.

## Technical Context

**Language/Version**: TypeScript 5.8.x, React 18.3.x  
**Primary Dependencies**: react-router-dom 6.30.x, @tanstack/react-query 5.76.x, zod 3.24.x, TailwindCSS 3.4.x  
**Storage**: Browser HTTP-only cookies for auth/session (set by BFF), local client storage only for UI preferences and short-lived draft restore metadata  
**Testing**: Vitest 3.2.x, Testing Library, MSW for API mocking, existing integration tests in backend/tests/integration where applicable  
**Target Platform**: Web browsers (desktop/tablet/mobile), Vite dev server + production static bundle
**Project Type**: Web application (frontend + backend/BFF)  
**Performance Goals**: Login-to-dashboard under 5s on simulated 4G, skeleton visible within 300ms during loading, route transition without full reload  
**Constraints**: HTTP-only cookie auth with SameSite=Strict, token refresh at 75% of token lifetime, lockout after 5 failed attempts in 15 minutes, mobile-first UX  
**Scale/Scope**: Auth shell + navigation shell for all current frontend screens; no multi-project switcher in this feature

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-Phase 0 Gate Review

- **I. Modular Monorepo Architecture**: PASS
  - Changes are isolated to existing frontend module and BFF contract usage.
  - No cross-service direct coupling introduced.
- **II. SOLID / Clean Architecture**: PASS
  - Frontend logic remains hook-centric; no business logic in page bodies.
  - BFF remains Echo + Huma owned contract boundary.
- **III. Cloud Native & Containerization**: PASS (N/A for direct scope)
  - No infra/container topology changes required by this feature.
- **IV. Frontend Component-First, Hook Isolation & Design System**: PASS
  - Requires data fetching + auth lifecycle in `frontend/src/hooks/`.
  - Requires tokenized UI styles and mobile-first behavior.

### Post-Phase 1 Re-Check

- **I. Modular Monorepo Architecture**: PASS
- **II. SOLID / Clean Architecture**: PASS
- **III. Cloud Native & Containerization**: PASS
- **IV. Frontend Component-First, Hook Isolation & Design System**: PASS

No constitution violations requiring justification.

## Project Structure

### Documentation (this feature)

```text
specs/002-frontend-auth-navigation/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── auth-bff.openapi.yaml
└── tasks.md
```

### Source Code (repository root)

```text
backend/
├── cmd/bff/
├── internals/bff/
│   ├── financial/controllers/
│   ├── services/
│   └── transport/http/middleware/
└── tests/integration/

frontend/
├── src/app/
├── src/components/
├── src/hooks/
├── src/pages/
├── src/services/
├── src/styles/
└── src/types/
```

**Structure Decision**: Use the existing web-application split (`frontend/` + `backend/`) and implement frontend auth/navigation behavior through hooks and composition pages, consuming BFF contracts without moving domain logic into presentation components.

## Phase 0: Research Output

Research completed in `/specs/002-frontend-auth-navigation/research.md` with concrete decisions and alternatives for:

1. HTTP-only cookie auth model and CSRF posture
2. Token refresh timing at 75% lifetime
3. Responsive sidebar/hamburger behavior
4. Skeleton UX and accessibility expectations
5. Draft state preservation on refresh failure
6. Brute-force lockout behavior and frontend UX handling

All prior ambiguities from specification clarifications are resolved.

## Phase 1: Design & Contracts Output

1. **Data model**: `/specs/002-frontend-auth-navigation/data-model.md`
   - Authentication/session entities
   - Navigation entities and UI state
   - Draft restore model and lockout metadata
2. **Interface contracts**: `/specs/002-frontend-auth-navigation/contracts/auth-bff.openapi.yaml`
   - Login endpoint
   - Refresh endpoint
   - Lockout response envelope and metadata
3. **Quickstart**: `/specs/002-frontend-auth-navigation/quickstart.md`
   - Dev setup, test flows, troubleshooting aligned to cookie auth

## Implementation Strategy (Phase 2 Preview)

1. Implement auth API client with `credentials: include` defaults.
2. Create `useAuthSession` / `useAuthRefresh` hooks with 75% refresh scheduling.
3. Add lockout-aware login error mapping and countdown UX.
4. Implement sidebar + hamburger layout with mobile-first behavior.
5. Implement skeleton states for login and protected-page bootstrapping.
6. Add one-time draft restore path after forced re-login.
7. Add tests for lockout, refresh timing, cookie-based request behavior, and skeleton rendering.

## Complexity Tracking

No constitution violations. Complexity table not required.
