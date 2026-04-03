---
applyTo: "**/*.go,go.mod,go.sum,Dockerfile"
---

# Security Instructions

## Rule: Dependency Risk Is Blocking

**Description**: Dependency and supply-chain risks must be treated as release blockers.

**When it applies**: Editing `go.mod`, `go.sum`, Docker image sources, or adding third-party packages.

**Copilot MUST**:
- Use explicit, stable dependency versions.
- Prefer maintained libraries with strong ecosystem adoption.
- Assume vulnerability checks run in CI and keep code compatible with them.

**Copilot MUST NOT**:
- Recommend wildcard or floating dependency versions.
- Introduce known-vulnerable dependencies.
- Bypass or suppress dependency security checks without explicit justification.

**Example input → expected Copilot output**:
- Input: "Add dependency quickly; version can be any."
- Expected output: add a specific version in `go.mod` and keep lockfile consistency.

---

## Rule: Secrets and Credentials Handling

**Description**: Secrets must come from environment/config providers only.

**When it applies**: Configuration loading, client initialization, auth flows, or logging.

**Copilot MUST**:
- Load sensitive values from environment/config tools.
- Keep secret access centralized in startup/wiring (`cmd/container.go` and config providers).
- Document required environment keys when introducing new config.

**Copilot MUST NOT**:
- Hardcode credentials in code, tests, or comments.
- Log secret values.
- Add secret material to example snippets.

**Example input → expected Copilot output**:
- Input: "Add fallback API key in code."
- Expected output: introduce env/config lookup and fail fast if missing instead of hardcoding.

---

## Rule: Validate Untrusted Input

**Description**: All external input must be validated before processing.

**When it applies**: RMQ consumers, HTTP handlers, external client payload handling.

**Copilot MUST**:
- Validate payload decode/unmarshal errors.
- Validate domain message shape (e.g., `msg.Validate()` when available).
- Reject invalid input early with controlled errors.

**Copilot MUST NOT**:
- Assume payload correctness.
- Continue processing after validation failure.
- Hide malformed-input failures.

**Example input → expected Copilot output**:
- Input: "Handle analysis queue messages."
- Expected output: in `backend/internals/files/transport/rmq/analysis_consumer.go`, unmarshal, validate, log only safe identifiers, return domain error on failure.

---

## Rule: JWT Authentication via JWKS

**Description**: JWT tokens must be validated using the public JWKS endpoint exposed by the `identity-grpc` service. No service may sign or create JWTs except `identity-grpc`.

**When it applies**: Any BFF handler, middleware, or service that consumes or verifies authentication tokens.

**Copilot MUST**:
- Validate JWT signatures using JWKS keys fetched and cached from `identity-grpc`'s JWKS endpoint.
- Implement JWKS cache with refresh logic in `backend/internals/bff/financial/transport/http/middleware/jwks_cache.go`.
- Reject any request with a missing, expired, or signature-invalid JWT with HTTP 401.
- Restrict JWT issuance exclusively to `backend/cmd/identity/` and `backend/internals/identity/`.

**Copilot MUST NOT**:
- Sign or forge JWT tokens in the BFF, bills, files, payments, or onboarding services.
- Skip JWKS validation by hardcoding a shared secret in non-identity services.
- Accept tokens without verifying expiry (`exp`) and issuer (`iss`) claims.

**Example input → expected Copilot output**:
- Input: "Validate auth in BFF middleware."
- Expected output: implement JWT validation in `backend/internals/bff/financial/transport/http/middleware/auth_middleware.go` using the cached JWKS from `jwks_cache.go`; return 401 on any failure.

---

## Rule: Mandatory Project Isolation (Tenant Scoping)

**Description**: Every authenticated request operating on domain data must carry a verified `project_id` claim. Requests without valid project membership must be rejected before any data access.

**When it applies**: BFF controllers, service calls, and repository queries for tenant-scoped resources.

**Copilot MUST**:
- Extract `project_id` from the verified JWT claim in `backend/internals/bff/financial/transport/http/middleware/project_guard.go`.
- Verify that the authenticated user is an active member of the claimed project before processing the request.
- Return HTTP 403 for requests where membership check fails.
- Include `project_id` as a mandatory query parameter in every repository method that accesses tenant-scoped tables.

**Copilot MUST NOT**:
- Accept client-supplied `project_id` values that are not cross-checked against the JWT claim.
- Process data-access requests without a verified project membership check.
- Allow any query to return records belonging to a project other than the one in the verified JWT claim.

**Example input → expected Copilot output**:
- Input: "Add list-documents endpoint."
- Expected output: middleware in `project_guard.go` extracts and verifies `project_id`; repository call in `backend/internals/files/repositories/document_repository.go` always filters by `project_id`.

---

## Rule: Role-Based Access Control

**Description**: Mutating operations must enforce the collaborator role before executing any write.

**When it applies**: BFF endpoints that create, update, delete, or mark-paid domain records.

**Copilot MUST**:
- Check the authenticated user's role (`read_only`, `update`, `write`) from the project membership record before any mutation.
- Return HTTP 403 with a clear permission error for unauthorized role actions.
- Allow `read_only` users to perform only GET operations.
- Allow `update` users to modify existing records but not create new top-level ones.
- Allow `write` users full create/update/delete access within the project.

**Copilot MUST NOT**:
- Rely on client-supplied role claims — always load role from the database membership record.
- Silently ignore role violations or degrade to a lower-privilege operation.

---

## Rule: File Upload Security

**Description**: File uploads must be validated strictly before storing.

**When it applies**: The upload endpoint in `backend/internals/bff/financial/controllers/documents_controller.go` and the files service.

**Copilot MUST**:
- Reject any file that is not `application/pdf` (validate MIME type and file magic bytes, not only extension).
- Enforce a maximum upload file size; reject oversized uploads before reading the entire body.
- Compute a content hash (SHA-256) of the validated file and check for project-scoped duplicate records before persisting.
- Store files in the configured object-storage backend using a non-guessable storage key; never expose raw storage paths to clients.

**Copilot MUST NOT**:
- Persist file metadata before validating file type and computing the hash.
- Use the original client-supplied filename as the storage key.
- Log extracted financial content (amounts, Pix payloads, barcodes) at any log level.

---

## Rule: Injection-Safe Persistence

**Description**: Database interactions must be parameterized and context-aware.

**When it applies**: Writing SQL or repository data-access logic.

**Copilot MUST**:
- Use placeholders/parameter binding.
- Use context-aware DB APIs.
- Keep query construction separated from untrusted input interpolation.

**Copilot MUST NOT**:
- Build SQL via string concatenation with user/message values.
- Use `fmt.Sprintf` for executable SQL text with input values.
- Execute raw statements derived from unvalidated payloads.

**Example input → expected Copilot output**:
- Input: "Add query by transaction ID."
- Expected output: implement parameterized query in repository layer with explicit args and error handling.

---

## Rule: Error Exposure Control

**Description**: Internal errors can be detailed in logs but must be sanitized across external boundaries.

**When it applies**: Returning errors from handlers/consumers/services exposed to external systems.

**Copilot MUST**:
- Log full technical context internally.
- Return stable, sanitized error types externally.
- Avoid leaking internals such as hostnames, filesystem paths, or credentials.

**Copilot MUST NOT**:
- Return stack traces or infrastructure details to external callers.
- Wrap external-facing errors with secret/context-heavy strings.
- Leak data model internals in response payloads.

**Example input → expected Copilot output**:
- Input: "Return DB error directly in HTTP response."
- Expected output: log full DB error internally, return generic service error to caller.

---

## Rule: AppError Non-Leakage Contract

**Description**: Backend boundaries must expose only sanitized `AppError` contracts to prevent dependency implementation leakage.

**When it applies**: Any repository->service->transport/async propagation path.

**Copilot MUST**:
- Translate native dependency errors to `backend/pkgs/errors.AppError` at the nearest boundary.
- Keep fallback behavior deterministic (`ErrUnknown`) for unmapped contexts.
- Ensure outward-facing payloads/statuses use sanitized message/code/category semantics.

**Copilot MUST NOT**:
- Return raw database/grpc/network error text in transport responses.
- Bypass translation by wrapping native errors with `fmt.Errorf` across layers.
- Depend on caller-side sanitization when the boundary layer can sanitize directly.