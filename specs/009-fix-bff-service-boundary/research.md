# Research: Enforce Service Boundary Contracts

## Decision 1: BFF Contract Ownership Boundary

- Decision: HTTP request/response contracts remain exclusively in `transport/http/views`, while BFF services accept/return only transport-agnostic contracts.
- Rationale: This eliminates layer leakage and preserves clean architecture where HTTP details do not contaminate service orchestration.
- Alternatives considered:
  - Keep views in services for convenience: rejected because it violates the target separation of concerns.
  - Move all contracts to controllers package: rejected because controllers are adapters and must not own reusable service contracts.

## Decision 2: Mapper Responsibility Placement

- Decision: Mapping between HTTP views and service contracts is owned by the HTTP transport layer (controller + mapper helper) and never by BFF services.
- Rationale: Transport ownership keeps protocol concerns local and prevents business logic from encoding HTTP semantics.
- Alternatives considered:
  - Map inside services: rejected because service code would require transport-coupled assumptions.
  - Map in route registration modules: rejected because route modules should focus on operation registration and wiring.

## Decision 3: Service Contract Shape Selection

- Decision: Use existing proto domain messages as service contracts when semantically sufficient; otherwise introduce service-owned contracts in BFF service boundary packages.
- Rationale: Reusing proto messages minimizes duplication, while service-owned contracts cover orchestration needs not represented in proto domain models.
- Alternatives considered:
  - Always use proto messages only: rejected because not all orchestration payloads map cleanly to existing proto models.
  - Always create service-owned structs: rejected because it duplicates already-valid domain contracts.

## Decision 4: Pointer Signature Threshold

- Decision: Structs crossing backend function boundaries must use pointer semantics when they contain reference-like fields or when value size exceeds 3 machine words.
- Rationale: This threshold is measurable, consistent with the clarified spec, and reduces avoidable copy overhead while preserving explicit exceptions.
- Alternatives considered:
  - Pointer for all structs: rejected as too rigid and unnecessarily pointer-heavy.
  - Team discretion per module: rejected due to inconsistency risk and governance drift.

## Decision 5: Exception Handling for Value Semantics

- Decision: Value semantics are allowed only for documented exceptions (tiny immutable value objects and deliberate copy-by-value safety cases).
- Rationale: This keeps the default deterministic while preserving justified edge-case flexibility.
- Alternatives considered:
  - No exceptions allowed: rejected because it can produce unnecessary indirection for simple immutable values.
  - Implicit exceptions without documentation: rejected because it weakens enforceability.

## Decision 6: Rollout and Verification Scope

- Decision: Refactor applies to all active BFF routes/services in this feature and includes route-level integration verification plus service-unit coverage on changed boundaries.
- Rationale: Full-scope rollout avoids leaving known leaks behind and makes architecture enforcement immediate.
- Alternatives considered:
  - Only touched modules: rejected because it preserves existing contract leakage in untouched routes.
  - Pilot module only: rejected because it creates additional migration churn and delayed consistency.

## Decision 7: Governance and Memory Synchronization

- Decision: Implementation completion must update both `.specify/memory/*` flow artifacts and `/memories/repo/*` convention notes, plus impacted `.github/instructions/*.instructions.md` files.
- Rationale: Feature intent explicitly requires future-proofing through both planning memory and repository memory channels.
- Alternatives considered:
  - Update only instructions: rejected because memory-driven workflows would still lack the new boundary guidance.
  - Update only memory artifacts: rejected because instruction enforcement would remain incomplete.
