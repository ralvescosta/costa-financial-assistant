---
applyTo: "**/*.md"
---

# Diagram Standards Instructions

## Rule: Mermaid-Only Feature Diagrams

**Description**: All feature diagrams in documentation must be authored only with Mermaid so they render natively on GitHub.

**When it applies**: Creating or editing any feature diagram in project markdown documentation.

**Copilot MUST**:
- Use Mermaid fenced code blocks: ```mermaid.
- Keep one logical diagram per Mermaid block.
- Prefer GitHub-supported Mermaid families (`flowchart`, `sequenceDiagram`, `classDiagram`, `stateDiagram-v2`, `erDiagram`, `journey`, `gantt`, `pie`).
- Keep feature diagrams in markdown docs under `specs/`, `.specify/`, or other repository docs.

**Copilot MUST NOT**:
- Use non-Mermaid diagram syntaxes (PlantUML, Graphviz DOT, D2, Excalidraw JSON, draw.io XML) inside markdown docs.
- Embed screenshots/images as the source of truth for architecture or feature flows when a Mermaid source can be provided.
- Mix multiple unrelated diagrams in one Mermaid block.

---

## Rule: GitHub-Friendly Mermaid Structure

**Description**: Mermaid blocks must follow markdown structure that GitHub reliably parses.

**When it applies**: Writing Mermaid diagrams in markdown.

**Copilot MUST**:
- Put a blank line before and after each Mermaid block.
- Ensure fences are balanced and not nested.
- Keep Mermaid blocks at top-level markdown indentation (not inside blockquotes, HTML tags, or nested list items).
- Use ASCII-safe node identifiers (e.g., `A`, `BFF`, `PAY_1`) and keep labels inside brackets.
- Keep edge labels simple and explicit to avoid parser ambiguity.

**Copilot MUST NOT**:
- Leave unclosed or malformed code fences.
- Rely on markdown constructs that can break rendering context around fences.
- Depend on renderer-specific extensions that are not available in GitHub Markdown.

---

## Rule: Diagram Refactors Must Preserve Semantics

**Description**: Refactors for rendering compatibility must not change feature behavior or architecture meaning.

**When it applies**: Converting existing text flows or non-rendering diagram blocks into Mermaid.

**Copilot MUST**:
- Preserve original flow order, decision points, and responsibilities.
- Keep section headings and narrative context unchanged unless explicitly requested.
- Prefer structural fixes (fence type, block boundaries, parser-safe formatting) over content rewrites.

**Copilot MUST NOT**:
- Introduce new actors, services, or protocol assumptions during formatting-only refactors.
- Remove meaningful steps from existing flows.

---

## Rule: Feature Diagram Definition of Done

**Description**: A feature diagram change is complete only when the markdown source is GitHub-renderable and maintainable.

**When it applies**: Before finalizing markdown changes with diagrams.

**Copilot MUST VERIFY**:
- Diagram blocks are Mermaid fences.
- No plain text pseudo-diagrams remain where diagrams are expected.
- Markdown remains valid and readable with consistent headings.
- Diagram source is editable text and serves as the documentation source of truth.

**Copilot MUST NOT**:
- Finalize a feature diagram change without a Mermaid source block in the same file.
