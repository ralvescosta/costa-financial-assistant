# Auth Bootstrap Contract

## Goal

Restore the login-only bootstrap flow consumed by `frontend/src/hooks/useAuthContext.tsx` and `frontend/src/pages/LoginPage.tsx` without reintroducing self-registration.

## HTTP endpoints in scope

| Endpoint | Method | Owner | Purpose |
|---|---|---|---|
| `/api/auth/login` | `POST` | BFF → identity/onboarding | Verify username/password and issue the authenticated result |
| `/api/auth/refresh` | `POST` | BFF → identity | Refresh access-token / CSRF state |
| `/api/auth/logout` | `POST` | BFF | Clear the active session when implemented in this cycle |

## Login request

```json
{
  "username": "ralvescosta",
  "password": "mudar@1234"
}
```

Rules:
- `username` and `password` are required.
- The bootstrap user must exist before success is possible.
- No `/register` endpoint or registration screen is part of this feature.

## Successful response shape

This must remain compatible with `frontend/src/types/auth-response.schema.ts`:

```json
{
  "statusCode": 200,
  "data": {
    "expiresIn": 86400,
    "refreshAt": 82800,
    "csrfToken": "<token>",
    "user": {
      "id": "<uuidv7>",
      "username": "ralvescosta",
      "email": "ralvescosta@local.dev"
    },
    "activeProject": {
      "id": "<project-id>",
      "name": "<project-name>",
      "role": "write"
    }
  }
}
```

Transport notes:
- Access and refresh tokens remain transport-managed by the BFF (HTTP-only cookies or the existing secure mechanism).
- The frontend persists only non-sensitive session metadata.

## Failure responses

- `401` with code `INVALID_CREDENTIALS` for wrong username/password.
- `429` with code `AUTH_LOCKED` when lockout semantics are active.
- `401` or `403` for expired/invalid protected-route sessions.
- `503` or an AppError-mapped dependency/setup error when bootstrap seed data is missing or incomplete.

## Downstream orchestration

```text
LoginPage -> BFF auth route -> BFF auth service -> identity/onboarding gRPC -> seed-backed persistence
```

The BFF must not bypass downstream ownership by calling repositories or SQL directly for auth or membership data.

## Current rollout status

- `identity.v1.IdentityService` now exposes `AuthenticateUser` and `RefreshSession` for the seeded owner flow.
- The BFF now owns `POST /api/auth/login` and `POST /api/auth/refresh`, returning the validated frontend envelope while rotating the `cfa_session` HTTP-only cookie.
- Fresh verification evidence currently includes `cd backend && go test ./...` and the focused frontend auth-hook Vitest suite.
