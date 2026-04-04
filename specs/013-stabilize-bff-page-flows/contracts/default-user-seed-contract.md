# Contract: Default User Seed Contract

## Purpose

Define the minimum populated seed data required for the canonical dev/test default user so the four broken pages (`documents`, `payments`, `analyses`, and `settings`) can be reviewed with meaningful data rather than mostly empty states.

## Scope

- **Environment**: Local, test, and demo-only environments
- **Out of scope**: Production seeding
- **Rule**: The seeded experience must be deterministic, idempotent, and project-scoped

## Required Default User Baseline

| Seed area | Required outcome |
|---|---|
| Identity/session | The default user can authenticate successfully and obtain the expected active project context |
| Project membership | The default user belongs to the project used by the four in-scope screens and has the permissions needed to access them |
| Documents | The project includes representative uploaded document records |
| Payments | The project includes representative bill/payment records and a stored payment-cycle preference |
| Analyses | The project includes enough payment/reconciliation history to render timeline, category, and summary widgets |
| Settings | The project includes at least one bank-account label for statement attribution |

## Minimum Representative Seed Set

### 1. Documents page

The default user should see:
- at least one bill-like document
- at least one statement-like document
- valid document metadata (`file_name`, `kind`, `analysis_status`, uploader)

### 2. Payments page

The default user should see:
- at least one unpaid upcoming bill
- at least one overdue bill
- at least one already-paid example when the dashboard/history needs contrast
- a valid preferred payment-cycle day

### 3. Analyses page

The default user should see:
- a non-empty timeline window
- category breakdown data
- reconciliation summary data that demonstrates at least two states from `matched_auto`, `matched_manual`, `ambiguous`, and `unmatched`

### 4. Settings page

The default user should see:
- at least one existing bank-account label
- the ability to create and delete a label without breaking the page flow

## Acceptance Rule

The canonical acceptance seed is **not** “no errors with empty data.” It is “no errors with representative populated data across all four pages.”

Empty states remain valid only for explicit fallback or secondary validation scenarios.

## Consistency Rules

- All seeded records must belong to the same active project used by the default user session.
- Cross-domain references must be coherent: documents, bills, payment history, reconciliation entries, and settings labels should point to the same seeded project context.
- Seed routines should be idempotent so repeated local/test setup does not create unstable or duplicated demo states.
