# Spec Review Alignment Contract

## Purpose

Define the expected contract for reviewing and aligning `specs/006-bff-http-separation/spec.md` to the current template and memory-governance state.

## Inputs

- Source spec: `specs/006-bff-http-separation/spec.md`
- Current template: `.specify/templates/spec-template.md`
- Memory references:
  - `.specify/memory/bff-flows.md`
  - `.specify/memory/architecture-diagram.md`
  - `.specify/memory/architecture-diagram-maintenance.md`
- Scope constraint: direct-impact updates only (Option B)

## Scope Guard (T003)

Allowed edits in this feature cycle:
- `specs/006-bff-http-separation/spec.md`
- `.specify/memory/bff-flows.md` (only when mismatch is proven)
- `.specify/memory/architecture-diagram-maintenance.md` (only when mismatch is proven)
- `.specify/memory/architecture-diagram.md` (only when cross-service impact is proven)

Disallowed edits in this feature cycle:
- Broad instruction corpus refactors not triggered by a direct mismatch
- Broad template changes not triggered by a direct mismatch
- Runtime backend/frontend source code changes

## Validation Rules

1. Mandatory template sections must be present in updated spec 006.
2. No placeholder instructional text may remain in updated spec 006.
3. Requirements must be testable and unambiguous.
4. Success criteria must be measurable.
5. Memory impact section must explicitly list impacted files and no-impact rationales.
6. If instruction/template mismatch is identified, only direct-impact corrections are allowed under this feature scope.

## Output Contract

### Required Outputs

- Updated `specs/006-bff-http-separation/spec.md` when misalignment exists.
- Updated directly referenced memory file(s) only when mismatch exists.
- Checklist evidence showing pass status.
- Handoff recommendation for `/speckit.plan` readiness.

### Prohibited Outputs

- Broad refactor of unrelated instruction files.
- Broad updates to template corpus not directly required by 006 alignment.
- Runtime code changes outside documentation scope.

## Completion Conditions

- All validation rules pass.
- Scope constraints are respected.
- Readiness decision is explicit and evidence-backed.

## Execution Outcome Record (T017)

- Scope guard compliance: PASS.
- Direct mismatch requiring memory-file update: YES (`.specify/memory/bff-flows.md` wording alignment note added).
- Direct mismatch requiring instruction/template updates: NO.
- Readiness evidence attached: `specs/007-review-bff-spec/checklists/requirements.md`.
