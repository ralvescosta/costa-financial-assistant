---
applyTo: "**/*.go"
---

# Go Language Instructions

## Rule: Blocking Quality Gates

**Description**: Generated Go code must pass CI quality checks.

**When it applies**: Writing or modifying any Go file.

**Copilot MUST**:
- Produce code expected to pass `golangci-lint`.
- Produce code expected to pass `go vet` and `go test ./...`.
- Keep code `gofmt` and `goimports` compliant.

**Copilot MUST NOT**:
- Suggest knowingly lint-breaking code.
- Ignore warnings by default.
- Add `//nolint` without explicit, documented justification.

**Example input → expected Copilot output**:
- Input: "Quick fix this by ignoring lint."
- Expected output: provide a lint-compliant implementation instead of adding blanket ignore directives.

---

## Rule: Go Security Enforcement

**Description**: Security findings are blocking and must be addressed deterministically.

**When it applies**: Handling I/O, crypto, file/network access, or dependency updates.

**Copilot MUST**:
- Generate code expected to pass `gosec`.
- Prefer safe standard library and maintained dependencies.
- Validate risky input paths and return controlled errors.

**Copilot MUST NOT**:
- Introduce insecure defaults to speed delivery.
- Disable security rules without rationale in code review.
- Recommend vulnerable dependency usage.

**Example input → expected Copilot output**:
- Input: "Disable gosec for this SQL path."
- Expected output: use parameterized query patterns from repositories and keep the security check enabled.

---

## Rule: Idiomatic Error and Context Handling

**Description**: Use explicit error handling and context propagation in all I/O paths.

**When it applies**: Functions performing external calls, DB operations, messaging, or orchestration.

**Copilot MUST**:
- Accept `context.Context` as first parameter when appropriate.
- Propagate incoming context through repository/client/service calls.
- Return wrapped/contextual errors or domain errors consistently.
- Log relevant context before returning operational failures.

**Copilot MUST NOT**:
- Ignore returned errors.
- Use `panic` for recoverable flows.
- Create detached contexts (`context.Background()`) inside request paths.

**Example input → expected Copilot output**:
- Input: "Add repository call in service processing."
- Expected output: forward `ctx` from `Process(...)` in `internal/services/enrichment/service.go` into repository method calls and handle returned errors explicitly.

---

## Rule: Keep Interfaces Consumer-Driven

**Description**: Define narrow interfaces around consumer needs.

**When it applies**: Adding interfaces or changing dependency contracts.

**Copilot MUST**:
- Keep interfaces small and purpose-specific.
- Place service/repository contracts under `internal/services/interfaces/...` when they are shared.
- Use constructor injection against interfaces.

**Copilot MUST NOT**:
- Create broad "do everything" interfaces.
- Bind upper layers to concrete implementations.
- Add methods that are not used by consumers.

**Example input → expected Copilot output**:
- Input: "Add a new service dependency to enrichment consumer."
- Expected output: wire the dependency through interface-based constructor signatures in `internal/transport/rmq/consumers/enrichment.go` and `cmd/container.go`.

---

## Rule: Import and Formatting Discipline

**Description**: Keep imports deterministic and formatting tool-compatible.

**When it applies**: Any import or formatting change.

**Copilot MUST**:
- Group imports as standard library, third-party, internal.
- Keep consistent formatting with `goimports`.
- Avoid unnecessary aliasing unless collision or clarity requires it.

**Copilot MUST NOT**:
- Mix internal and external imports in random order.
- Leave unused imports.
- Introduce formatting style that requires manual cleanup.

**Example input → expected Copilot output**:
- Input: "Add new package imports for container wiring."
- Expected output: follow grouped import pattern used in `cmd/container.go` and keep only used imports.