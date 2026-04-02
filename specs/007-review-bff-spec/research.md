# Research: Review and Align Spec 006

## Baseline Confirmation (T001)

- Baseline source confirmed: `.specify/templates/spec-template.md`.
- Reviewed target: `specs/006-bff-http-separation/spec.md`.
- Method: section-by-section comparison against mandatory template blocks.

## Alignment Gap Log (T002, T008)

| Gap ID | Area | Finding | Resolution |
|---|---|---|---|
| G-001 | Section completeness | `Architecture & Memory Diagram Flow Impact` was missing in 006 spec | Added full section with impacted files and explicit rationale |
| G-002 | Section completeness | `Instruction Impact` was missing in 006 spec | Added full section including refactor scope, impacted files, and pattern-preservation statement |
| G-003 | Scope clarity | No explicit no-impact rationale for non-impacted memory files | Added no-impact rationale text in memory impact section |
| G-004 | Readiness evidence | No explicit conformance summary for updated 006 structure | Added conformance summary and readiness validation notes |

## Memory Sync Decision Rules (T005)

1. Update `.specify/memory/bff-flows.md` only when 006 memory-impact statements diverge from current BFF flow ownership language.
2. Update `.specify/memory/architecture-diagram-maintenance.md` only when maintenance guidance needs a direct cross-reference clarification for this spec-review workflow.
3. Update `.specify/memory/architecture-diagram.md` only when cross-service flow topology changes are introduced; otherwise keep unchanged and record no-impact rationale.
4. Any instruction/template file update is allowed only when a direct mismatch is proven by 006 alignment work.

## Decision 1: Treat the active template as the strict baseline

- Decision: Use `.specify/templates/spec-template.md` as the authoritative structure baseline for reviewing `specs/006-bff-http-separation/spec.md`.
- Rationale: The template reflects current governance and memory-impact requirements and avoids subjective spec formatting decisions.
- Alternatives considered:
  - Baseline against prior spec artifacts only: rejected because older specs can contain legacy omissions.
  - Baseline against ad-hoc reviewer judgment: rejected because this weakens deterministic execution.

## Decision 2: Use direct-impact scope for updates (Option B)

- Decision: Limit updates to `specs/006-bff-http-separation/spec.md` and directly referenced memory-flow files only when alignment gaps are found.
- Rationale: This honors FR-010 and keeps the feature focused on producing a planning-ready 006 spec without broad governance churn.
- Alternatives considered:
  - Spec-only updates with no memory sync: rejected because it can leave memory drift unresolved.
  - Broad instruction/template normalization: rejected because it exceeds scoped intent.

## Decision 3: Memory-flow sync rule for this feature

- Decision: `bff-flows.md` is required sync when 006 impact statements diverge from current flow reality; `architecture-diagram.md` remains unchanged unless cross-service communication impact appears; `architecture-diagram-maintenance.md` is updated only for procedural cross-reference corrections.
- Rationale: This preserves accuracy while preventing unnecessary edits to stable architecture topology.
- Alternatives considered:
  - Always modify all memory files: rejected because it introduces noise and review burden.
  - Never modify memory files during spec review: rejected because it can leave explicit governance mismatch.

## Decision 4: Checklist-driven readiness gate

- Decision: Use `specs/007-review-bff-spec/checklists/requirements.md` as the blocking readiness gate before handoff to planning.
- Rationale: Readiness becomes observable and repeatable, preventing ambiguous handoff quality.
- Alternatives considered:
  - Narrative-only readiness statement: rejected because it is non-verifiable.
  - Tooling-only lint gate: rejected because this feature is primarily documentation quality alignment.

## Decision 5: Instruction/template updates are conditional and minimal

- Decision: Include explicit plan tasks for instruction/template sync checks, but perform edits only when direct mismatches are proven by the 006 review.
- Rationale: This satisfies constitution gate expectations for refactor/reorganization while respecting constrained scope.
- Alternatives considered:
  - Omit instruction/template sync tasks: rejected because constitution requires explicit handling.
  - Preemptively update instruction/template files: rejected because it violates FR-010 scope boundaries.

## Conformance Summary (T012)

- Mandatory template sections are present in updated 006 spec.
- Heading order follows template sequence.
- No placeholder instructional text remains in updated 006 spec.
- Memory and instruction impact declarations are explicit and scoped.
