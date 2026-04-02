# Tasks: Review and Align Spec 006

**Input**: Design documents from /specs/007-review-bff-spec/
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Prepare the review workspace and baseline artifacts.

- [x] T001 Confirm active template baseline in .specify/templates/spec-template.md
- [x] T002 Create alignment gap log in specs/007-review-bff-spec/research.md
- [x] T003 [P] Capture review contract scope guard in specs/007-review-bff-spec/contracts/spec-review-alignment-contract.md

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Establish shared rules that block all user-story execution until complete.

- [x] T004 Define direct-impact update boundaries in specs/007-review-bff-spec/plan.md
- [x] T005 [P] Define memory-sync decision rules in specs/007-review-bff-spec/research.md
- [x] T006 [P] Define readiness gate criteria in specs/007-review-bff-spec/checklists/requirements.md
- [x] T007 Define artifact lifecycle and handoff sequence in specs/007-review-bff-spec/quickstart.md

**Checkpoint**: Foundation ready. User stories can proceed.

---

## Phase 3: User Story 1 - Update Spec 006 to Current Template (Priority: P1) 🎯 MVP

**Goal**: Bring spec 006 to full current-template structure and content quality.

**Independent Test**: Open specs/006-bff-http-separation/spec.md and confirm all mandatory sections exist in template order with no placeholder text.

### Implementation for User Story 1

- [x] T008 [US1] Compare section inventory and record gaps in specs/007-review-bff-spec/research.md
- [x] T009 [US1] Add or repair missing mandatory sections in specs/006-bff-http-separation/spec.md
- [x] T010 [US1] Rewrite user stories and acceptance scenarios for template compliance in specs/006-bff-http-separation/spec.md
- [x] T011 [US1] Remove placeholder and instructional filler text from specs/006-bff-http-separation/spec.md
- [x] T012 [US1] Validate heading order and section completeness in specs/006-bff-http-separation/spec.md

**Checkpoint**: User Story 1 is independently complete and reviewable.

---

## Phase 4: User Story 2 - Align Memory-Flow Impact Statements (Priority: P2)

**Goal**: Ensure spec 006 memory-impact declarations match current memory artifacts.

**Independent Test**: Confirm specs/006-bff-http-separation/spec.md explicitly states impacted memory files and no-impact rationale, and directly referenced memory files are updated only when mismatch exists.

### Implementation for User Story 2

- [x] T013 [US2] Reconcile Architecture and Memory Impact section in specs/006-bff-http-separation/spec.md
- [x] T014 [P] [US2] Verify BFF flow alignment and update .specify/memory/bff-flows.md if mismatch is found
- [x] T015 [US2] Add explicit no-impact rationales for non-impacted memory files in specs/006-bff-http-separation/spec.md
- [x] T016 [US2] Update cross-reference maintenance notes in .specify/memory/architecture-diagram-maintenance.md only if mismatch is found
- [x] T017 [US2] Record memory-sync outcomes and scope compliance in specs/007-review-bff-spec/contracts/spec-review-alignment-contract.md

**Checkpoint**: User Story 2 is independently complete and auditable.

---

## Phase 5: User Story 3 - Make 006 Ready for Clarify/Plan (Priority: P3)

**Goal**: Prove readiness with objective checklist outcomes and no blocking ambiguities.

**Independent Test**: Validate specs/007-review-bff-spec/checklists/requirements.md is fully passing and specs/006-bff-http-separation/spec.md has no unresolved clarification markers.

### Implementation for User Story 3

- [x] T018 [US3] Execute readiness validation and capture outcomes in specs/007-review-bff-spec/checklists/requirements.md
- [x] T019 [US3] Resolve ambiguous requirement wording in specs/006-bff-http-separation/spec.md
- [x] T020 [US3] Validate measurability of success criteria and update specs/006-bff-http-separation/spec.md
- [x] T021 [US3] Update handoff and verification steps in specs/007-review-bff-spec/quickstart.md
- [x] T022 [US3] Record final readiness decision in specs/007-review-bff-spec/plan.md

**Checkpoint**: User Story 3 is independently complete and planning-ready.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final consistency and documentation quality pass.

- [x] T023 [P] Run editorial consistency pass in specs/006-bff-http-separation/spec.md
- [x] T024 [P] Run editorial consistency pass in specs/007-review-bff-spec/spec.md
- [x] T025 Run quickstart validation walkthrough and update notes in specs/007-review-bff-spec/quickstart.md

---

## Phase 7: Mandatory Governance Sync (Blocking)

**Purpose**: Ensure memory and governance obligations are explicitly finalized.

- [x] T026 Update final impacted memory-file status in specs/007-review-bff-spec/plan.md
- [x] T027 If direct mismatch is proven, update governance behavior in .github/instructions/ai-behavior.instructions.md (N/A: no direct mismatch proven)
- [x] T028 If direct mismatch is proven, update memory/architecture rule text in .github/instructions/architecture.instructions.md (N/A: no direct mismatch proven)
- [x] T029 If direct mismatch is proven, update structure rule text in .github/instructions/project-structure.instructions.md (N/A: no direct mismatch proven)
- [x] T030 If direct mismatch is proven, update spec workflow template in .specify/templates/spec-template.md (N/A: no direct mismatch proven)
- [x] T031 Verify canonical backend integration-test standard applicability and record compliant/N-A outcome in specs/007-review-bff-spec/checklists/requirements.md

**Checkpoint**: Feature is not complete until this phase is complete.

---

## Dependencies & Execution Order

### Phase Dependencies

- Phase 1 has no dependencies.
- Phase 2 depends on Phase 1 and blocks all user stories.
- Phase 3 depends on Phase 2.
- Phase 4 depends on Phase 3 baseline updates in specs/006-bff-http-separation/spec.md.
- Phase 5 depends on Phases 3 and 4 outputs.
- Phase 6 depends on all user-story phases.
- Phase 7 depends on all prior phases and must complete before merge.

### User Story Dependencies

- US1 (P1) starts after Phase 2 and is the MVP slice.
- US2 (P2) depends on US1’s normalized spec structure.
- US3 (P3) depends on US1 + US2 outputs to validate readiness.

### Within Each User Story

- Gap analysis before edits.
- Section/impact edits before readiness validation.
- Readiness validation before final handoff decision.

---

## Parallel Opportunities

- T003 can run in parallel with T002.
- T005 and T006 can run in parallel.
- T014 can run in parallel with T015.
- T023 and T024 can run in parallel.
- T027, T028, T029, and T030 are parallelizable conditional sync tasks when triggered.

---

## Parallel Example: User Story 1

- Run T008 while preparing T009 edit scaffolding.
- Run T010 and T011 in parallel only if they touch distinct sections of specs/006-bff-http-separation/spec.md.

## Parallel Example: User Story 2

- Run T014 and T015 in parallel.
- Run T016 after T014 only if a maintenance mismatch is discovered.

## Parallel Example: User Story 3

- Run T019 and T020 in parallel.
- Run T021 after T018 outcome is known.

---

## Implementation Strategy

### MVP First (US1 Only)

1. Complete Phase 1 and Phase 2.
2. Complete Phase 3 (US1).
3. Validate template conformance in specs/006-bff-http-separation/spec.md.

### Incremental Delivery

1. Deliver US1 for template conformance.
2. Deliver US2 for memory-flow alignment.
3. Deliver US3 for readiness evidence.
4. Execute Phase 6 and Phase 7 before final merge.

### Parallel Team Strategy

1. One contributor handles spec section normalization (US1).
2. One contributor handles memory-flow verification and sync (US2).
3. One contributor handles checklist/readiness evidence (US3).
4. Merge streams at Phase 6 for final consistency and governance sync.
