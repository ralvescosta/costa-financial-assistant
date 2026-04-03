---
applyTo: "**/*.go"
---

# Coding Conventions Instructions

## Rule: Intent-Revealing Naming

**Description**: Names must communicate business meaning and purpose directly.

**When it applies**: Naming packages, files, types, methods, variables, constants.

**Copilot MUST**:
- Use domain terms that match repository language.
- Use clear verb-based function names for actions.
- Keep exported/unexported naming consistent with Go conventions.

**Copilot MUST NOT**:
- Use vague names (`data`, `value`, `temp`, `helper`) for business concepts.
- Use obscure abbreviations that hide intent.
- Rename stable domain concepts without explicit request.

**Example input → expected Copilot output**:
- Input: "Add helper for duplicate upload detection."
- Expected output: choose explicit naming such as `isDuplicateUpload`/`detectDuplicate`, consistent with domain terms in `backend/internals/files/repositories/document_repository.go` (e.g., `FindByProjectAndHash`).

---

## Rule: Small, Focused Functions

**Description**: Functions should do one thing with clear control flow.

**When it applies**: Writing or modifying function bodies.

**Copilot MUST**:
- Keep functions focused on a single responsibility.
- Prefer early returns over deep nesting.
- Extract cohesive helper functions when logic becomes hard to scan.

**Copilot MUST NOT**:
- Mix parsing, validation, orchestration, and persistence in one method.
- Create hidden side effects unrelated to function name.
- Add complex branching that obscures error paths.

**Example input → expected Copilot output**:
- Input: "Extend consumer with extra validation and processing."
- Expected output: keep input validation in the RMQ consumer handler and delegate domain logic to `backend/internals/files/services/document_service.go`.

---

## Rule: Explicit Constants and Configuration Keys

**Description**: Avoid magic values and implicit configuration behavior.

**When it applies**: Introducing repeat literals, config keys, retry limits, timeouts.

**Copilot MUST**:
- Use named constants for repeated literals.
- Keep configuration values externally driven where applicable.
- Prefer repository-wide config access patterns already used in startup/runtime code.

**Copilot MUST NOT**:
- Scatter hardcoded values across files.
- Encode environment-specific constants in business logic.
- Duplicate config key strings without reason.

**Example input → expected Copilot output**:
- Input: "Add new max retry behavior for consumer."
- Expected output: read configured value via existing config mechanism, not a hardcoded inline number.

---

## Rule: Clear Error Messaging

**Description**: Errors and log messages must be actionable and stable.

**When it applies**: Returning errors or writing log messages.

**Copilot MUST**:
- Describe operation context succinctly.
- Preserve underlying error context when appropriate.
- Keep externally visible error text sanitized.

**Copilot MUST NOT**:
- Return generic messages without context.
- Leak sensitive/internal implementation details.
- Produce contradictory wording across similar flows.

**Example input → expected Copilot output**:
- Input: "Handle config retrieval error in analysis consumer."
- Expected output: include operation context in internal log and return an existing domain-safe error for unknown configuration keys.

---

## Rule: Deterministic Code Style

**Description**: Keep code generation consistent with existing repository style.

**When it applies**: Any source modification.

**Copilot MUST**:
- Follow existing file/package structure.
- Keep code gofmt/goimports compatible.
- Preserve current patterns unless user asks for refactor.

**Copilot MUST NOT**:
- Reformat unrelated blocks.
- Introduce new stylistic patterns without clear benefit.
- Rewrite working sections outside requested scope.

**Example input → expected Copilot output**:
- Input: "Change one operation behavior."
- Expected output: targeted edit in that operation file only; no broad style churn in unrelated directories.

---

## Rule: Value-Semantics Exception Traceability

**Description**: Any intentional value-semantic boundary in backend code must be explicitly documented and reviewable.

**When it applies**: Introducing or preserving value-semantic parameters/returns at cross-layer boundaries.

**Copilot MUST**:
- Add concise rationale comments when preserving value semantics at boundary points.
- Record approved exceptions in the feature contract log (for example `specs/<feature>/contracts/pointer-exceptions.md`).
- Keep naming and function signatures aligned with the documented policy decision.

**Copilot MUST NOT**:
- Keep undocumented value-semantics exceptions in boundary contracts.
- Mix pointer and value semantics arbitrarily in related boundary methods without rationale.