# Data Model: Restore Seeded Login & Session Propagation

## Entity: BootstrapUser

- **Purpose**: The default owner account used to enter the application when registration is intentionally unavailable.
- **Fields**:
  - `id` (UUIDv7): canonical identity value reused in `common.v1.Session`
  - `username` (string): fixed bootstrap username `ralvescosta`
  - `email` (string): non-empty email used in the shared session envelope
  - `password_hash` (string): stored credential hash, never returned to clients
  - `status` (enum): `active | locked | disabled`
  - `created_at` / `updated_at` (timestamp)
- **Validation rules**:
  - `username` must be unique.
  - The stored secret must be a hash; the plain password `mudar@1234` is seed input only.
  - The user must exist before bootstrap login can succeed.
- **Ownership**: Identity service persistence + seed/bootstrap path.

## Entity: CredentialState

- **Purpose**: Tracks login verification and temporary lockout state for the bootstrap identity.
- **Fields**:
  - `user_id` (UUIDv7)
  - `password_hash` (string)
  - `failed_attempt_count` (int)
  - `lockout_until` (optional timestamp)
  - `last_login_at` (optional timestamp)
  - `last_password_rotation_at` (optional timestamp)
- **Invariant**: If lockout is active, the BFF login flow must return a deterministic auth error rather than a partial session.

## Entity: OwnerProjectMembership

- **Purpose**: Project-scoped relationship that gives the bootstrap user owner-equivalent access to protected routes.
- **Fields**:
  - `id` (string)
  - `project_id` (string)
  - `user_id` (UUIDv7)
  - `role` (enum): `write` in the current project-role contract
  - `invited_by` (string): `system-seed` or equivalent bootstrap marker
  - `created_at` / `updated_at` (timestamp)
- **Invariant**: The bootstrap user must belong to at least one project before the authenticated result is considered valid.
- **Ownership**: Onboarding service membership model.

## Entity: Session

- **Purpose**: Canonical gRPC caller-identity envelope shared across services.
- **Fields**:
  - `id` (UUIDv7)
  - `email` (string)
  - `username` (string)
- **Ownership**: `backend/protos/common/v1/messages.proto`
- **Invariant**: `Session` supplements but never replaces `common.v1.ProjectContext`.

## Entity: AuthenticatedResult

- **Purpose**: Login/refresh response envelope consumed by `frontend/src/hooks/useAuthContext.tsx`.
- **Fields**:
  - `expires_in` (int seconds)
  - `refresh_at` (int seconds until refresh)
  - `csrf_token` (string)
  - `user` (`id`, `username`, `email`)
  - `active_project` (`id`, `name`, `role`)
- **Invariant**: The frontend persists only non-sensitive metadata; tokens remain transport-managed by the BFF.

## Entity: PaginationDefaults

- **Purpose**: Deterministic first-page behavior when the frontend omits query params.
- **Fields**:
  - `resource_key` (string): e.g. `documents`, `project_members`, `payment_dashboard`
  - `default_page_size` (int)
  - `page_token_default` (string, usually `""`)
  - `source` (enum): `frontend_query | bff_fallback`
- **Target default for the verified scope**:
  - `default_page_size = 20`
  - `page_token_default = ""`
  - Existing route implementations that still default to `25` must be normalized or explicitly documented during implementation.
- **Invariant**: The BFF always forwards a populated `common.v1.Pagination` object on list/select gRPC requests.

## Entity: ProtectedRouteAccessCheck

- **Purpose**: Verification record proving the seeded owner can traverse protected flows safely.
- **Fields**:
  - `route_id` (string)
  - `service_name` (string)
  - `required_project_role` (string)
  - `session_present` (bool)
  - `pagination_present` (bool, for list/select flows)
  - `expected_outcome` (enum): `allowed | denied | setup_error`
- **Use**: Regression planning for BFF and cross-service tests.

## Relationships

- `BootstrapUser` has one `CredentialState`.
- `BootstrapUser` must have at least one `OwnerProjectMembership`.
- `Session` is derived from the authenticated `BootstrapUser` and travels alongside `ProjectContext`.
- `AuthenticatedResult` exposes session-adjacent user/project data to the frontend after successful login.
- `PaginationDefaults` applies to each `ProtectedRouteAccessCheck` that invokes a list/select RPC.

## State Transitions

1. **Seeded** → bootstrap user, hashed credential, and owner membership exist.
2. **Login attempt** → credentials validate, or the flow returns a lockout/setup/auth error.
3. **Authenticated** → the BFF issues the authenticated result and session-carrying downstream requests can begin.
4. **Protected request** → BFF forwards `Session` + `ProjectContext` (+ `Pagination` when needed) to the owning service.
5. **Expired/unauthorized** → the system denies access safely and clears or refuses ambiguous session state.
