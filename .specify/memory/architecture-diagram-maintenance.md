# Architecture Diagram Maintenance Process

This document defines how and when to keep `.specify/memory/architecture-diagram.md` synchronized with the evolving Costa Financial Assistant architecture.

## Overview

The architecture diagram serves as the **single source of truth** for system design and communication patterns. It must be updated during each iteration to reflect:
- New services or components added
- Changed communication patterns (gRPC, HTTP, events, etc.)
- New technology dependencies
- Data flow modifications
- Technology version updates

---

## When to Update

### Service-Flow Mapping Enforcement (MUST)

When a feature impacts a service, update the corresponding memory flow file in the
same execution cycle:

- `bff` → `.specify/memory/bff-flows.md`
- `files` → `.specify/memory/files-service-flows.md`
- `bills` → `.specify/memory/bills-service-flows.md`
- `identity` → `.specify/memory/identity-service-flows.md`
- `onboarding` → `.specify/memory/onboarding-service-flows.md`

For migration-heavy features:
- Update the affected service flow files above.
- Update `.specify/memory/architecture-diagram.md` whenever migration changes modify
  cross-service flow semantics, ownership boundaries, or integration behavior.

### Mandatory Update Triggers (MUST update immediately)

| Event | Impact | Section to Update |
|-------|--------|-------------------|
| New microservice created | System design | Add service box, connections, responsibilities table |
| Service protocol changed | Communication | Update connection arrows and communication matrix |
| New external dependency added | Infrastructure | Add to data layer, update tech stack table |
| Inter-service connection added/removed | Data flow | Update service connection arrows, data flow examples |
| Service removed or deprecated | System design | Remove service, update related flows |
| Technology version major bump | Infrastructure | Update tech stack version range |
| Database schema pattern change | Data model | Update data flow examples if multi-tenant queries affected |

### Recommended Update Triggers (SHOULD update by iteration end)

| Event | Impact | Section to Update |
|-------|--------|-------------------|
| Service responsibility clarification | Documentation | Update Service Responsibilities table description |
| New gRPC proto file added | Service contract | Update service protocol/methods if applicable |
| New middleware or pattern introduced | Communication | Document in communication matrix if cross-cutting |
| Performance optimization (caching, indexing) | Data layer | May affect Redis/cache discussion if architectural change |
| Integration test coverage expands | Test dependencies | Not typically in main diagram, but note in changelog |

For spec-review-only features (for example alignment workflows that change wording,
governance declarations, or ownership documentation without changing runtime
cross-service topology), `.specify/memory/architecture-diagram.md` SHOULD remain
unchanged and the feature MUST record an explicit no-topology-change rationale in
its spec and plan artifacts.

### Refactor/Reorganization Trigger (MUST)

If a feature refactors or reorganizes project or service structure, the same feature
execution MUST include instruction updates under `.github/instructions/` for every
impacted architectural/coding/testing pattern.
If Speckit workflow behavior changes, update impacted `.specify/templates/*.md` files
in the same feature cycle.

### Optional/Periodic Updates

| Event | Cadence | Section to Update |
|-------|---------|-------------------|
| Minor wording/clarity improvements | Per sprint | Maintenance of tables, descriptions |
| "Last Updated" metadata refresh | Per iteration/week | Footer section timestamps |
| Tech stack minor version bumps | Per quarter review | Tech stack table (only if noteworthy) |

---

## Update Process (Step-by-Step)

### Step 1: Identify Trigger
Before opening `.specify/memory/architecture-diagram.md`, confirm the change qualifies as a **Mandatory** or **Recommended** trigger.

### Step 2: Determine Scope
Identify which sections are affected:
- **Mermaid Graph** (visual): Service additions, removals, connection changes
- **Service Responsibilities Table**: Role clarifications, new services
- **Data Flow Examples**: New flows from feature development
- **Communication Matrix**: New inter-service paths
- **Technology Stack Table**: New or upgraded dependencies

### Step 3: Make Changes

**For Mermaid Graph updates:**
1. Add/remove service subgraph boxes within appropriate layers (Client, API Gateway, Core Business, Data, etc.)
2. Update connection arrows (solid for sync/gRPC, dashed for observability, colored for async)
3. Keep legend consistent: service name, protocol, brief purpose
4. Revalidate graph syntax in a Mermaid editor before commit

**For Table updates:**
1. Add/remove rows matching the service or connection
2. Keep column alignment and formatting consistent
3. Use consistent terminology (gRPC, HTTP REST, RabbitMQ, etc.)

**For Data Flow examples:**
1. Add new flows only if they introduce new architectural patterns (e.g., first event-driven flow)
2. Update existing flows if communication protocol changes
3. Always test the narrative against the current Mermaid graph

**For Communication Matrix:**
1. Add new rows for each new service-to-service connection
2. Mark method as gRPC, HTTP, async (RabbitMQ), or direct (PostgreSQL/Redis)
3. Tag with the feature/spec that introduced the connection

### Step 4: Update Metadata
- Increment `version` in footer using **semantic versioning**:
  - **MAJOR**: Service removed, protocol rewrite, major architecture shift
  - **MINOR**: New service added, new critical connection, significant new flow
  - **PATCH**: Clarifications, version bumps, wording improvements
- Set `Last Updated` to current date (ISO format YYYY-MM-DD)
- Add changelog entry (optional but recommended)

**Version Examples:**
- `1.0.0` → `1.1.0`: Added Payments service to architecture diagram
- `1.1.0` → `2.0.0`: Replaced gRPC communication with HTTP event streaming across all services
- `1.0.1` → `1.0.2`: Clarified Redis cache-aside pattern description

### Step 5: Validate
Before commit, verify:
- [ ] All new boxes/connections visible in Mermaid graph
- [ ] Service Responsibilities table is complete for new services
- [ ] Communication Matrix covers all new inter-service paths
- [ ] Data flow narratives match Mermaid graph layout
- [ ] No orphaned services or connections
- [ ] Version number makes sense
- [ ] Last Updated date is correct

### Step 6: Commit & Propagate

**Commit message template:**
```
docs: update architecture diagram v<VERSION> (<brief_reason>)

- Added/Updated <service/connection/pattern>
- Impact: <affected components>
- Trigger: <which spec/feature/PR introduced this>
```

**Example commits:**
```
docs: update architecture diagram v1.1.0 (add Payments service)
- Added Payments gRPC service to core business layer  
- Added RabbitMQ consumer connection for payment events
- Impact: BFF now routes payment queries to new service
- Trigger: spec 004-payment-tracking

docs: update architecture diagram v1.0.2 (clarify Redis usage)
- Expanded Redis cache-aside pattern description
- No connection changes
- Trigger: tech review suggestion
```

**Propagate to dependent templates (if major update):**
- Check if `.specify/templates/spec-template.md` needs updates (e.g., new service domains)
- Check if `.specify/templates/plan-template.md` needs architecture section updates
- Check if `README.md` architecture section aligns with diagram
- Update `.specify/memory/constitution.md` if architectural principles changed
- For refactor/reorganization work, update impacted instruction files in
  `.github/instructions/` and any affected `.specify/templates/*.md` workflow templates
  before merge

---

## Roles & Responsibilities

| Role | Responsibility |
|------|-----------------|
| **Feature Implementer (Developer)** | File issue/comment in PR when architecture changes are made → review trigger checklist|
| **Tech Lead / Architecture Reviewer** | Review architecture diagram updates in PRs, approve version bumps, validate scope |
| **DevOps / Infrastructure** | Update technology stack table when adding services, dependencies, or infrastructure |
| **Product/Spec Owner** | Ensure new specs include architecture impact statement → inform diagram updates |

---

## Integration into Development Workflow

### Before Spec Approval
- Spec author adds "Architecture Impact" section to `spec.md`:
  ```
  ### Architecture Impact
  - [x] New service? (Yes/No)
  - [x] New external dependency? (Yes/No)
  - [x] New gRPC service contract? (Yes/No)
  - Diagram sections affected: Data flow, Communication matrix
  ```
- Spec author MUST list impacted service flow files under `.specify/memory/` and provide
  explicit no-impact rationale if no update is required.

### During Sprint Planning
- Tech lead reviews all approved specs' architecture impacts
- Planning task: "Update architecture diagram per spec XYZ changes"
- Assign to implementer or architecture owner with explicit trigger checklist
- Planning MUST include explicit task(s) to update impacted service-flow memory files.
- For refactor/reorganization specs, planning MUST include explicit instruction update
  tasks under `.github/instructions/`.

### Before Merge (PR Checklist)
If PR touches service creation, communication, or infrastructure:
```markdown
- [ ] Architecture diagram updated?
  - [ ] Service/connection added/removed to Mermaid graph
  - [ ] Service Responsibilities or Communication Matrix updated
  - [ ] Version number incremented
  - [ ] Last Updated date set
  - [ ] Validated in Mermaid editor
```

### Post-Release  
- After each major release, conduct 30-min "Architecture Sync" meeting:
  - Review all merged PRs for architecture changes
  - Ensure diagram reflects deployed state
  - Identify any missed updates
  - Plan next iteration's diagram maintenance

---

## FAQ & Troubleshooting

### Q: A new gRPC service was added but it doesn't appear in the diagram. Now what?

**A:** File an issue or PR:
1. Add service to appropriate Mermaid subgraph (usually "Core Business Services")
2. Add connections from BFF and other services that call it
3. Add row to Service Responsibilities table
4. Add rows to Communication Matrix for each new connection
5. Add example data flow if it represents a new pattern
6. Bump version to MINOR
7. Update Last Updated date

### Q: Should I update the diagram if I rename a service internally but the external role stays the same?

**A:** If the *name* changes but the *protocol*, *dependencies*, or *role* don't:
- Yes, update the name in the Mermaid box and tables
- This is a PATCH version bump
- Example: rename `Bills` to `BillsAnalyzer` (same gRPC interface, same clients)

### Q: What if a service is being refactored (split into two) mid-sprint?

**A:** Create a separate PR/issue to update the diagram:
1. Make the change clearly: old service → two new services
2. Update all affected tables and flows
3. This is typically a MINOR version bump
4. Include architectural rationale in commit message
5. Plan follow-up: which PRs must reference this architectural change?

### Q: The diagram is getting too large. Should I split it into multiple files?

**A:** Per iteration growth guidelines:
- **Small codebase** (1-3 services): Single diagram OK at any size
- **Medium codebase** (4-6 services): Keep single diagram; start planning domain views
- **Large codebase** (7+ services): Consider sub-diagrams by domain (e.g., `architecture-diagram-payments.md`, `architecture-diagram-files.md`) but maintain a primary high-level overview
- **Current project**: 7 services → near split threshold; continue with single holistic diagram for now

### Q: Who decides version bumps (MAJOR vs MINOR vs PATCH)?

**A:** Hierarchy of authority:
1. **Automated**: If trigger in "Mandatory" section → follows rule (typically MINOR for new service, PATCH for rewording)
2. **Technical Lead**: Makes final call on MAJOR bumps (architectural paradigm shifts)
3. **Team consensus**: If ambiguous, discuss in PR review or async comment thread

---

## Appendix: Git Workflow Example

```bash
# Start new feature that affects architecture
git checkout -b feat/add-reconciliation-service

# ... implement Payments + Reconciliation changes ...

# Time to commit diagram
# 1. Edit architecture-diagram.md locally
# 2. Make changes per step-by-step process above
# 3. Verify with `cat .specify/memory/architecture-diagram.md`

git add .specify/memory/architecture-diagram.md
git commit -m "docs: update architecture diagram v1.1.0 (add Payments & reconciliation consumer)

- Added Payments gRPC service to core business layer
- Added RabbitMQ consumer for reconciliation workflows
- Updated Communication Matrix: Payments <-> Bills, Payments <-> PostgreSQL
- Updated Data Flow: new reconciliation pipeline example
- Impact: BFF routes /payments/reconcile to new service
- Trigger: spec 004-payment-tracking implementation"

# Push and create PR with diagram changes visible
git push origin feat/add-reconciliation-service

# In PR review, architecture reviewer spots the diagram update:
# ✅ Mermaid graph correct  
# ✅ Service Responsibilities complete
# ✅ Communication Matrix reflects new connections
# ✅ Data Flow example helpful
# ✅ Version bump 1.1.0 justified
# → PR approved with diagram changes
```

---

## Backlog Suggestions for Future Iterations

- [ ] Add CLI script to auto-validate Mermaid syntax before commit
- [ ] Create architecture diff tooling to highlight changes in diagram PRs
- [ ] Integrate with decision log (ADR - Architecture Decision Records) to link decisions to diagram versions
- [ ] Build dashboard to show architecture evolution over time (version history visualization)
- [ ] Automate service discovery to cross-check active services in codebase vs diagram

---

## Document Metadata

- **Created**: 2026-03-31
- **Last Updated**: 2026-03-31
- **Version**: 1.0.0
- **Status**: Active
- **Owner**: Architecture Team
- **Related**: `.specify/memory/architecture-diagram.md`, `README.md`, `.github/instructions/architecture.instructions.md`

