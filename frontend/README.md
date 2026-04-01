# Frontend — Costa Financial Assistant

React 18 + TypeScript SPA secured by BFF-managed HTTP-only cookies.

## Tech Stack

| Library | Version | Purpose |
|---|---|---|
| React | 18.3.x | UI framework |
| TypeScript | 5.8.x | Type safety |
| react-router-dom | 6.30.x | Client-side routing |
| @tanstack/react-query | 5.76.x | Server state |
| zod | 3.24.x | Schema validation |
| TailwindCSS | 3.4.x | Styling |
| Vitest + @testing-library/react | 3.2.x / 16.x | Tests |
| MSW | 2.7.x | API mocking |

## Auth System Architecture

```
main.tsx
  └── ErrorBoundary
        └── AuthProvider          ← manages auth lifecycle (useAuthContext.tsx)
              └── AppProviders   ← QueryClientProvider + ThemeProvider
                    └── AppRouter
                          ├── /login → LoginPage
                          └── ProtectedLayout (redirects to /login if unauthenticated)
                                └── AppLayout  ← Sidebar + HamburgerMenu + useAuthRefresh + useTokenRefreshInterceptor
                                      └── <Outlet> → page routes
```

### Token Lifecycle

1. Login → BFF sets `access_token` + `refresh_token` as **HTTP-only SameSite=Strict cookies**
2. Frontend stores only non-sensitive session metadata in `localStorage` under `cfa:session`
3. `useAuthRefresh` schedules a BFF refresh call at **75% of `expires_in`**
4. `useTokenRefreshInterceptor` patches `window.fetch`: on 401, refresh once and retry
5. On refresh failure → `logout()` → redirect to `/login`

### Draft State Restoration

When token refresh fails during an active session, draft UI state is saved to `localStorage` under `cfa:draft:<key>` with a 15-minute TTL. After re-login, `DraftRestoreModal` offers a one-time restore.

## Environment Variables

| Variable | Description | Required in prod |
|---|---|---|
| `VITE_DEFAULT_USERNAME` | Pre-fill login form (dev only) | No |
| `VITE_DEFAULT_PASSWORD` | Pre-fill login form (dev only) | No |
| `VITE_API_BASE_URL` | BFF base URL | Yes |

> In production builds `authConfig.defaultUsername` and `authConfig.defaultPassword` are `undefined`.

## Directory Layout

```
src/
  app/          # Router, providers, layout, error boundary, route config
  components/   # Reusable UI (Sidebar, HamburgerMenu, SkeletonPlaceholder, ErrorMessage, DraftRestoreModal)
  config/       # auth.config.ts — env var safe-access
  hooks/        # All server-state and business logic (hook-centric architecture)
  pages/        # Composition roots only (LoginPage, DashboardPage, BillsPage …)
  services/     # Raw API client (api.client.ts) and auth API calls
  styles/       # Design token system (tokens.ts + index.css)
  test/         # setup.ts — Vitest globals, jest-dom matchers, matchMedia mock
  types/        # Shared TypeScript types and zod schemas
```

## Running Locally

```bash
# Install dependencies
npm install

# Start dev server
npm run dev        # http://localhost:3000

# Run tests
npm test

# Run tests with coverage
npm run test -- --coverage
```

## Key Hooks

| Hook | Purpose |
|---|---|
| `useAuthContext` | Centralised auth state: `isAuthenticated`, `login()`, `logout()`, `refreshAccessToken()` |
| `useAuthSession` | Thin wrapper exposing `{ isAuthenticated, isLoading, error, lockoutUntil, login, logout }` |
| `useAuthRefresh` | Schedules proactive token refresh at 75% of lifetime |
| `useTokenRefreshInterceptor` | Patches `window.fetch` to retry once after 401 |
| `useResponsiveNavigation` | Breakpoint detection + sidebar open state |
| `useDraftStateRestore` | TTL-aware draft persistence + one-time restore |
| `usePersistentSession` | Reads `cfa:session` from localStorage |

## Security Notes

- No token values are stored in JavaScript-accessible storage — all tokens live in HTTP-only cookies managed by the BFF
- `cfa:session` in localStorage contains only: `userId`, `username`, `expiryTimestamp`, `refreshAtTimestamp`, `activeProjectId`
- `VITE_DEFAULT_*` credentials are nullified in production builds by `auth.config.ts`
- CSRF token (from BFF login response) is sent in `X-CSRF-Token` header for mutating requests
