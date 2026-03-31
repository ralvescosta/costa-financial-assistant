# Architecture Documentation Quick Reference

**Location**: `.specify/memory/`

This folder now contains architecture-related documentation for the Costa Financial Assistant project.

---

## Files Overview

### 1. 📊 `architecture-diagram.md` (THE DIAGRAM)
**What**: Visual + detailed system architecture documentation  
**Contains**:
- Mermaid graph showing all 7 services, data layers, external dependencies
- Service responsibilities table (what each service does)
- Data flow examples (document upload, payment dashboard, reconciliation)
- Communication matrix (who talks to whom, how)
- Technology stack

**When to read**: 
- Onboarding new team members
- Planning new features to understand existing architecture
- Checking service dependencies before refactoring

**When to update**: See maintenance process (below)

---

### 2. 📋 `architecture-diagram-maintenance.md` (THE PROCESS)
**What**: Rules for keeping the diagram up to date  
**Contains**:
- Trigger checklist (when MUST you update the diagram)
- Step-by-step update process
- Version bumping rules (MAJOR/MINOR/PATCH)
- Commit message templates
- Roles & responsibilities
- PR checklist for developers
- FAQ & troubleshooting

**When to read**:
- Before making architecture changes (add new service, new gRPC connection, etc.)
- When reviewing PRs that touch architecture
- When unsure if diagram needs updating

**When to use**: 
- During sprint planning (assign diagram update tasks)
- During PR reviews (check diagram update checklist)
- Each iteration end (verify architecture is synchronized)

---

## Quick Start: Using These Files

### Scenario 1: "I'm adding a new service"
1. Read: `architecture-diagram-maintenance.md` → **Mandatory Update Triggers** table
2. Your change matches: "New microservice created" → MUST update diagram
3. Read: `architecture-diagram-maintenance.md` → **Step-by-step process**
4. Edit: `architecture-diagram.md` following the steps
5. Commit with template from maintenance guide

### Scenario 2: "I'm implementing a feature with new gRPC calls"
1. Read: `architecture-diagram-maintenance.md` → **Mandatory Update Triggers**
2. Your change matches: "Inter-service connection added/removed" → MUST update diagram
3. Update diagram connections, communication matrix
4. Commit

### Scenario 3: "I'm reviewing a PR that touches architecture"
1. Check PR checklist in `architecture-diagram-maintenance.md`
2. Verify diagram sections were updated
3. Run through validation checklist before approving

### Scenario 4: "I'm deploying a new release"
1. Run "Architecture Sync" meeting (post-release process in maintenance guide)
2. Verify diagram matches deployed state
3. Plan next iteration's architecture tasks

---

## Integration Points

| Workflow | Reference | Action |
|----------|-----------|--------|
| **Sprint Planning** | `architecture-diagram-maintenance.md` | Review specs' "Architecture Impact" sections → assign diagram update tasks |
| **PR Review** | `architecture-diagram-maintenance.md` | Check "Before Merge" checklist for architecture changes |
| **Spec Writing** | `architecture-diagram-maintenance.md` | Include "Architecture Impact" in `spec.md` |
| **Release Planning** | `architecture-diagram-maintenance.md` | Schedule "Architecture Sync" post-release |
| **Onboarding** | `architecture-diagram.md` | Show new team members the visual diagram first |

---

## Key Update Triggers at a Glance

If ANY of these happen, update the diagram:

✅ **New service created or removed**  
✅ **New gRPC/HTTP connection between services**  
✅ **New external dependency** (PostgreSQL, Redis, S3, RabbitMQ upgrade, etc.)  
✅ **Service communication protocol changed** (e.g., HTTP → gRPC)  
✅ **Service removed or deprecated**  
✅ **Major data flow pattern introduced** (e.g., first event-driven flow)  

ℹ️ **Does NOT require immediate update:**
- Internal service refactoring (if external API unchanged)
- Non-architectural bug fixes
- Performance optimizations (unless fundamental architectural change)

---

## Mermaid Diagram Syntax Tips

The diagram uses standard Mermaid graph syntax:

```mermaid
graph TB
    subgraph "Layer Name"
        NodeID["Display Label"]
    end

    %% Connections
    NodeA -->|label| NodeB          %% Solid arrow (sync)
    NodeA -.->|label| NodeB         %% Dashed arrow (observability)
    NodeA -->|gRPC| NodeB           %% Labeled arrow

    style NodeA fill:#e1f5f7         %% Color styling
```

**For diagram edits**: Update between ` ```mermaid ` tags, then validate in [Mermaid Live Editor](https://mermaid.live).

---

## Common Mistakes to Avoid

❌ **Updating only the table but forgetting the Mermaid graph**  
→ Always sync both visual + text

❌ **Adding a service without updating Communication Matrix**  
→ Matrix should list ALL new service connections

❌ **Not updating version number**  
→ Helps track architecture changes over time

❌ **Forgetting to sync dependent files** (README.md, spec.md)  
→ If it's a MAJOR change, plan follow-up updates

❌ **Mixing service additions with unrelated clarity improvements**  
→ Separate concerns: one commit per logical change

---

## Suggested Cadence

| Frequency | Task |
|-----------|------|
| **Every spec** | Author adds "Architecture Impact" to spec.md |
| **Every sprint** | Tech lead reviews architecture changes in specs |
| **Every PR** | Reviewer checks diagram update checklist if architecture affected |
| **Every release** | Post-release Architecture Sync meeting (30 min) |
| **Every quarter** | Quarterly architecture review: is diagram still accurate? |

---

## File Locations in Repo

```
.specify/
├── memory/
│   ├── constitution.md                          (project governance)
│   ├── architecture-diagram.md                  ← UPDATE THIS (the diagram)
│   ├── architecture-diagram-maintenance.md      ← FOLLOW THIS (the process)
│   └── architecture-diagram-reference.md        ← YOU ARE HERE (quick guide)
├── templates/
│   ├── spec-template.md
│   ├── plan-template.md
│   ├── tasks-template.md
│   └── ...
└── ...
```

---

## Next Steps

1. **Read** [architecture-diagram.md](./architecture-diagram.md) to understand current system design
2. **Bookmark** [architecture-diagram-maintenance.md](./architecture-diagram-maintenance.md) for reference during development
3. **Add to team wiki/onboarding**: Link team members to this guide during project kick-off
4. **Create script** (optional): Auto-validate Mermaid syntax on or pre-commit

---

## Questions?

- **"When do I update the diagram?"** → See Mandatory & Recommended triggers in maintenance guide
- **"How much detail should the diagram show?"** → Current level (7 services, data layer, external deps) is appropriate; split into sub-diagrams if 7+ services grow substantially
- **"Can I simplify the diagram?"** → No; it's the SSOT. Instead, create sub-diagrams (e.g., payment domain details)
- **"Who approves architecture changes?"** → Tech lead has final say on diagram accuracy. All engineers can suggest updates via PR.

---

## Metadata

- **Created**: 2026-03-31
- **Part of**: Costa Financial Assistant Architecture Documentation  
- **Maintenance**: Owner: Architecture Team | Review: Quarterly
- **Related Docs**: [architecture.instructions.md](../../.github/instructions/architecture.instructions.md), [README.md](../../README.md)

