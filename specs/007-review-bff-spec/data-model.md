# Data Model: Review and Align Spec 006

## Entities

### 1. Source Specification (Spec 006)

- Purpose: Primary document under review and alignment.
- Attributes:
  - `feature_id`: `006-bff-http-separation`
  - `section_set`: present section inventory
  - `requirement_set`: FR list with testability status
  - `impact_declarations`: architecture and memory impact statements
- Relationships:
  - Compared against the Template Baseline.
  - Generates Alignment Findings.

### 2. Template Baseline

- Purpose: Defines mandatory section structure and content expectations.
- Attributes:
  - `template_path`: `.specify/templates/spec-template.md`
  - `mandatory_sections`: ordered section definitions
  - `governance_requirements`: memory and instruction impact obligations
- Relationships:
  - Evaluates Source Specification compliance.
  - Informs Checklist Result.

### 3. Memory Impact Record

- Purpose: Captures required memory-flow sync decisions for the reviewed feature.
- Attributes:
  - `affected_services`
  - `required_memory_files`
  - `no_change_rationales`
  - `sync_status`
- Relationships:
  - Referenced by Source Specification.
  - Cross-checked against memory files in `.specify/memory/`.

### 4. Alignment Finding

- Purpose: Represents one gap or confirmation discovered during review.
- Attributes:
  - `finding_type`: missing_section, placeholder_content, ambiguous_requirement, memory_mismatch, instruction_mismatch
  - `severity`: blocking, non_blocking
  - `target_artifact`
  - `resolution_action`
- Relationships:
  - Produced by comparing Source Specification to Template Baseline.
  - Drives Checklist Result and readiness decision.

### 5. Readiness Checklist Result

- Purpose: Formal pass/fail record for clarify/plan readiness.
- Attributes:
  - `content_quality_status`
  - `requirement_completeness_status`
  - `feature_readiness_status`
  - `timestamp`
- Relationships:
  - Aggregates Alignment Findings.
  - Determines Handoff Status.

## Relationship Summary

- Template Baseline evaluates Source Specification and produces Alignment Findings.
- Memory Impact Record constrains and verifies allowed synchronization targets.
- Alignment Findings are resolved into updated artifacts and then reflected in Readiness Checklist Result.
- Readiness Checklist Result determines whether the feature is handoff-ready for planning execution.

## Lifecycle

1. Load Source Specification and Template Baseline.
2. Compare sections, requirements quality, and mandatory impact declarations.
3. Record Alignment Findings and classify blocking versus non-blocking.
4. Apply direct-impact updates to Source Specification and required memory files.
5. Re-run checklist validation and publish final Readiness Checklist Result.
