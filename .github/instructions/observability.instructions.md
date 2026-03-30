---
applyTo: "**/*.go"
---

# Observability Instructions

## Rule: Structured Logging Only

**Description**: Logs must be structured, leveled, and context-aware.

**When it applies**: Generating or modifying logging statements.

**Copilot MUST**:
- Use structured logging patterns consistent with `logrus` usage in this repository.
- Attach `context.Context` using `logrus.WithContext(ctx)` in request/message paths.
- Include stable identifiers in fields (for example transaction IDs, client key, operation).
- Use appropriate levels (`Debug`, `Info`, `Warn`, `Error`).

**Copilot MUST NOT**:
- Use `fmt.Printf`/ad-hoc print logging for runtime events.
- Log large raw payloads without filtering.
- Use `Fatal`/`Panic` in library/business code paths.

**Example input → expected Copilot output**:
- Input: "Log enrichment processing failures."
- Expected output: in `internal/services/enrichment/service.go`, log with `WithContext(ctx).WithError(err)` and relevant fields before returning the error.

---

## Rule: Context Propagation for Telemetry

**Description**: Preserve request and trace continuity end-to-end.

**When it applies**: Adding service, repository, or client calls.

**Copilot MUST**:
- Pass incoming `ctx` through all internal and external calls.
- Start spans for meaningful I/O boundaries where tracing already exists.
- Record and set span status on errors in traced repository/client operations.

**Copilot MUST NOT**:
- Drop context while crossing layers.
- Create disconnected telemetry context in request flow.
- Add spans for trivial local-only operations.

**Example input → expected Copilot output**:
- Input: "Add query instrumentation in events repository."
- Expected output: follow `internal/repositories/events_repository.go` pattern with tracer span start/end and `span.RecordError(err)` on failure.

---

## Rule: Log Level Semantics

**Description**: Log severity must match operational significance.

**When it applies**: Choosing level for any new log line.

**Copilot MUST**:
- Use `Info` for normal lifecycle milestones.
- Use `Warn` for recoverable abnormal conditions.
- Use `Error` for failed operations requiring investigation.
- Reserve `Debug` for diagnosable details.

**Copilot MUST NOT**:
- Emit expected duplicate/idempotency outcomes as `Error`.
- Spam hot paths with `Info`/`Debug` logs lacking value.
- Hide failures by logging `Debug` only.

**Example input → expected Copilot output**:
- Input: "Message already processed; add log."
- Expected output: in `internal/services/enrichment/service.go`, keep this as `Info`, then return `nil`.

---

## Rule: Sensitive Data Redaction

**Description**: Observability output must never leak secrets or regulated data.

**When it applies**: Logging request payloads, errors, config values, or credentials.

**Copilot MUST**:
- Log only safe identifiers and operational metadata.
- Redact or omit secret tokens, credentials, and full sensitive payloads.
- Keep error logs useful while sanitized for external visibility.

**Copilot MUST NOT**:
- Log passwords, API keys, auth tokens, or raw cardholder data.
- Dump full message structs without filtering.
- Embed sensitive fields in free-text log messages.

**Example input → expected Copilot output**:
- Input: "Add debug log for inbound interchange message."
- Expected output: log safe message identifiers and routing metadata, not full payload object.