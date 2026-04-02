# Quickstart: Plan Execution for Review and Align Spec 006

## Goal

Produce a planning-ready, template-conformant 006 specification with accurate memory-impact declarations and explicit readiness evidence.

## Prerequisites

- Repository on branch `007-review-bff-spec`
- Speckit scripts available in `.specify/scripts/bash/`
- Current template and memory files present under `.specify/templates/` and `.specify/memory/`

## Execution Steps

1. Verify planning context paths:

```bash
.specify/scripts/bash/check-prerequisites.sh --json --paths-only
```

2. Compare `specs/006-bff-http-separation/spec.md` to `.specify/templates/spec-template.md` and list alignment findings.

3. Apply direct-impact updates only:
- Update `specs/006-bff-http-separation/spec.md` for section and quality conformance.
- Update `.specify/memory/bff-flows.md` only if impact statement mismatch is confirmed.
- Update `.specify/memory/architecture-diagram-maintenance.md` only if cross-reference maintenance mismatch is confirmed.

4. Validate checklist readiness:
- Confirm all items in `specs/007-review-bff-spec/checklists/requirements.md` remain passing.
- Confirm no unresolved clarification markers remain.

5. Regenerate planning context for AI agent:

```bash
.specify/scripts/bash/update-agent-context.sh copilot
```

## Artifact Lifecycle and Handoff Sequence (T007, T021)

1. Establish baseline and gap log (`research.md`).
2. Apply direct-impact edits to `specs/006-bff-http-separation/spec.md`.
3. Apply conditional memory sync updates only when mismatch is proven.
4. Re-run readiness checklist and capture evidence.
5. Record final readiness decision in `plan.md`.
6. Handoff package for next workflow stage:
	- `specs/006-bff-http-separation/spec.md` (updated)
	- `specs/007-review-bff-spec/checklists/requirements.md` (pass evidence)
	- `specs/007-review-bff-spec/contracts/spec-review-alignment-contract.md` (scope compliance record)

## Expected Output

- `specs/007-review-bff-spec/plan.md` completed.
- `specs/007-review-bff-spec/research.md` completed.
- `specs/007-review-bff-spec/data-model.md` completed.
- `specs/007-review-bff-spec/contracts/spec-review-alignment-contract.md` completed.
- `specs/007-review-bff-spec/quickstart.md` completed.
- Agent context update script executed successfully.

## Validation Walkthrough Notes (T025)

- Walkthrough executed against current branch artifacts.
- Checklist status: PASS.
- Clarification markers in updated 006 spec: none.
- Scope control validation: PASS (direct-impact only).
