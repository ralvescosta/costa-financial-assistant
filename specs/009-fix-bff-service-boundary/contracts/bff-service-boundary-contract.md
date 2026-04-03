# Contract: BFF Service Boundary and Pointer Conventions

## Purpose

Define enforceable invariants for:
- BFF transport/service separation,
- mapper ownership and contract transformation,
- backend pointer-signature policy for struct boundaries.

## Scope

- Primary code scope:
  - `backend/internals/bff/transport/http/views/`
  - `backend/internals/bff/transport/http/controllers/`
  - `backend/internals/bff/transport/http/routes/`
  - `backend/internals/bff/services/`
  - `backend/internals/bff/interfaces/`
- Policy propagation scope:
  - `backend/internals/files/`
  - `backend/internals/bills/`
  - `backend/internals/identity/`
  - `backend/internals/onboarding/`
  - `backend/internals/payments/`

## Contract Invariants

1. Service boundary invariant:
   - BFF services MUST NOT import or depend on HTTP view request/response types.
2. Mapper ownership invariant:
   - HTTP transport layer MUST map transport views to service contracts before service invocation and map service outputs back to response views.
3. Contract shape invariant:
   - Service methods use proto domain messages when sufficient; otherwise use service-owned contracts, never controller/view-owned contracts.
4. Behavior preservation invariant:
   - Refactor MUST preserve externally observable endpoint behavior (status semantics and response contract meaning).
5. Pointer policy invariant:
   - Struct boundaries MUST use pointers when struct contains reference-like fields OR struct size exceeds 3 machine words.
6. Exception invariant:
   - Value-semantics exceptions are valid only with explicit justification.
7. Coverage invariant:
   - Applies to all active BFF routes/services in this feature.
8. Governance-sync invariant:
   - Completion requires updates in both `.specify/memory/*` and `/memories/repo/*`, plus impacted `.github/instructions/*.instructions.md` files.

## Acceptance Mapping

- FR-001, FR-002, FR-003 -> Invariants 1, 2, 3
- FR-004 -> Invariant 4
- FR-005, FR-005a, FR-006 -> Invariants 5, 6
- FR-010 -> Invariant 7
- FR-008, FR-009 -> Invariant 8

## Verification Checklist

- [ ] No BFF service method signature references HTTP view contracts.
- [ ] Controller/mapper layer performs all transport <-> service transformations.
- [ ] Service contracts are proto-backed or service-owned; no transport-owned leakage.
- [ ] Endpoint behavior remains equivalent for active routes.
- [ ] Pointer threshold rule is applied at modified backend boundaries.
- [ ] All value-semantic exceptions are explicitly documented.
- [ ] Integration tests follow canonical placement and naming (`backend/tests/integration/<service>/` or `backend/tests/integration/cross_service/`, snake_case, BDD + AAA).
- [ ] Instruction and memory synchronization artifacts are updated in the same feature cycle.
