---
applyTo: "**/*.go"
---

# Observability Instructions

## Rule: Structured Logging Only

**Description**: Logs must be structured, leveled, and context-aware.

**When it applies**: Generating or modifying logging statements.

**Copilot MUST**:
- Use `go.uber.org/zap` as the sole logging library for all backend services.
- Inject a `*zap.Logger` (or `zap.SugaredLogger`) via constructor so it is available as a struct field.
- Attach stable identifiers as named fields: `zap.String("document_id", id)`, `zap.String("project_id", projectID)`, `zap.Error(err)`.
- Use appropriate levels (`Debug`, `Info`, `Warn`, `Error`).

**Copilot MUST NOT**:
- Use `fmt.Printf`, `log.Println`, or `logrus` for runtime events.
- Log large raw payloads without field-level filtering.
- Use `Fatal`/`Panic` in library or business logic paths.

**Example input → expected Copilot output**:
- Input: "Log document upload failure."
- Expected output: in `backend/internals/files/services/document_service.go`, log with `s.logger.Error("upload failed", zap.String("document_id", id), zap.Error(err))` before returning the error.

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
- Input: "Add query instrumentation in document repository."
- Expected output: follow `backend/internals/files/repositories/document_repository.go` pattern with `tracer.Start(ctx, "document.findByHash")` span start/end and `span.RecordError(err)` on failure.

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
- Input: "Document already classified; add log."
- Expected output: in `backend/internals/files/services/document_service.go`, log this at `Info` level with `zap.String("document_id", id)`, then return `nil` or a domain-safe result.

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
- Input: "Add debug log for inbound analysis job message."
- Expected output: log safe identifiers (document_id, job_id, project_id) and routing metadata only — never log raw PDF binary, Pix QR payloads, or extracted financial amounts.

---

## Rule: One-Boundary Dependency Error Logging

**Description**: Native dependency failures must be logged exactly once at the translation boundary before propagating sanitized `AppError` contracts.

**When it applies**: Repository/service/transport/async boundary handling.

**Copilot MUST**:
- Log native errors with `zap.Error(err)` and stable identifiers exactly once at the boundary where translation occurs.
- Propagate `AppError` to outer layers after logging and translation.
- Avoid duplicate error logs for the same failure path in immediate caller/callee boundaries.

**Copilot MUST NOT**:
- Return raw dependency error strings across service boundaries.
- Emit repeated error logs for a single boundary failure.
- Log translated/sanitized errors instead of the native error at the boundary.