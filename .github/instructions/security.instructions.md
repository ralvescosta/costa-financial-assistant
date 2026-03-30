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
- Input: "Handle enrichment queue messages."
- Expected output: in `internal/transport/rmq/consumers/enrichment.go`, unmarshal, validate, log context safely, return domain error on failure.

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