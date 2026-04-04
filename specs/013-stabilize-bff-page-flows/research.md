# Research: Stabilize Broken BFF Page Flows

## Scope Inventory

| In-scope page | Frontend entry points | BFF routes | Downstream owner |
|---|---|---|---|
| `documents` | `frontend/src/pages/UploadPage.tsx`, `frontend/src/hooks/useDocuments.ts`, `frontend/src/hooks/useUploadDocument.ts`, `frontend/src/services/documentsApi.ts` | `GET /api/v1/documents`, `POST /api/v1/documents/upload`, `POST /api/v1/documents/{documentId}/classify` | `files.v1` |
| `payments` | `frontend/src/pages/PaymentDashboardPage.tsx`, `frontend/src/hooks/usePaymentDashboard.ts`, `frontend/src/services/paymentsApi.ts` | `GET /api/v1/bills/payment-dashboard`, `POST /api/v1/bills/{billId}/mark-paid`, `GET/POST /api/v1/payment-cycle/preferred-day` | `bills.v1` + `payments.v1` |
| `analyses` | `frontend/src/pages/HistoryDashboardPage.tsx`, `frontend/src/pages/ReconciliationPage.tsx`, `frontend/src/hooks/useHistoryDashboard.ts`, `frontend/src/hooks/useReconciliationSummary.ts`, `frontend/src/services/historyApi.ts`, `frontend/src/services/reconciliationApi.ts` | `GET /api/v1/history/timeline`, `GET /api/v1/history/categories`, `GET /api/v1/reconciliation/summary`, `POST /api/v1/reconciliation/links` | `payments.v1` |
| `settings` | `frontend/src/pages/SettingsPage.tsx`, `frontend/src/hooks/useBankAccounts.ts`, `frontend/src/services/bankAccountsApi.ts` | `GET/POST/DELETE /api/v1/bank-accounts` | `files.v1` |

## Key Findings

- Existing BFF route-registration smoke tests confirm route wiring, but they do **not** validate the full page-load behavior for the four broken screens.
- Existing deeper integration coverage is concentrated in `backend/tests/integration/payments/` and `backend/tests/integration/cross_service/`, especially for payment dashboard, history, and reconciliation.
- Seeded demo data is currently fragmented and mostly created inline in tests, which makes it easy for the default user to have an authenticated session but inconsistent or incomplete page data.
- Likely breakage clusters already visible in the codebase include auth/project context propagation, empty-state/error translation mismatches, and data gaps across `files`, `bills`, and `payments`-owned records.

---

## Decision 1: Use the actual page-load request inventory as the stabilization boundary

**Decision**: Plan and validate the feature around the concrete frontend request inventory for `documents`, `payments`, `analyses`, and `settings`, rather than around a generic backend audit.

**Rationale**: The user-facing problem is page breakage, so the safest definition of done is: the default user opens each page and the exact page-load requests succeed or degrade to a supported empty/access state.

**Alternatives considered**:
- A broad backend-wide audit without page focus — rejected because it risks fixing unrelated areas while leaving the user-visible failures unresolved.
- A frontend-only workaround — rejected because the spec explicitly requires validating and fixing the backend flow.

## Decision 2: Treat `analyses` as the combination of history and reconciliation flows

**Decision**: The planning scope treats the analyses experience as the combined `history` and `reconciliation` route groups already exposed by the BFF.

**Rationale**: The current frontend codebase splits analysis behavior across `HistoryDashboardPage.tsx` and `ReconciliationPage.tsx`; both are backed by payments-owned read models and both can break the perceived “analyses” screen experience.

**Alternatives considered**:
- Treating only history as “analyses” — rejected because reconciliation is part of the same analytical workflow surfaced to the user.
- Treating reconciliation as out of scope — rejected because the current feature request explicitly wants the screen-level flow stabilized end-to-end.

## Decision 3: Require a populated default-user seed across all four pages

**Decision**: The canonical acceptance seed for the default user must populate `documents`, `payments`, `analyses`, and `settings` with representative data.

**Rationale**: Empty-state-only verification is not enough for this feature; meaningful data is required to understand the product flow, reproduce issues, and verify fixes reliably.

**Alternatives considered**:
- Populate only the “critical” pages — rejected by clarification because the desired outcome is a better overall system understanding and demo experience.
- Keep most pages empty-state based — rejected because it would mask data-linkage or mapper issues.

## Decision 4: Add regression coverage at both BFF and cross-service levels

**Decision**: New regression coverage should combine BFF-facing integration tests for page bootstrap behavior with cross-service tests where payments/history/reconciliation behavior spans multiple service boundaries.

**Rationale**: Route-registration smoke tests alone do not prove that page requests succeed with real downstream responses, while unit tests alone do not protect end-to-end behavior.

**Alternatives considered**:
- Rely only on unit tests in `backend/internals/bff/services/*_test.go` — rejected because mock-only coverage would not catch seeded-data or contract-integration failures.
- Add only frontend hook tests — rejected because the spec requires validating the BFF-to-service path.

## Decision 5: Fix failures at the owning backend boundary, not in the frontend

**Decision**: Failures must be corrected at the owning backend/service or mapping boundary while preserving the BFF as a thin auth/orchestration gateway.

**Rationale**: The architecture instructions forbid pushing domain ownership into the BFF or hiding service defects behind frontend workarounds.

**Alternatives considered**:
- Patch the frontend to suppress or ignore backend failures — rejected because it would leave the root cause in place.
- Add direct BFF data access shortcuts — rejected because it violates the gateway boundary and existing repo rules.

## Decision 6: Prioritize the most likely failure hotspots first

**Decision**: Investigation and implementation should prioritize these root-cause clusters first: auth/session and project context, default-user seed consistency, downstream contract/mapping gaps, and AppError/empty-state translation behavior.

**Rationale**: These areas align with the current code paths and with the existing coverage gaps already visible in the workspace.

**Alternatives considered**:
- Start by rewriting entire page modules or route structures — rejected because the likely failures are integration and seed-related, not structural redesign requirements.

## Observed Hotspots to Validate During Implementation

- `backend/internals/bff/transport/http/routes/history_routes.go` documents history routes as authentication-only; confirm that project scoping is still enforced correctly downstream.
- `backend/internals/bff/services/*` already translate downstream errors; validate that controller-level translation does not collapse specific access or not-found outcomes into generic 500s.
- `backend/tests/integration/bff/*_routes_registration_test.go` covers route presence only; it does not verify page bootstrap data for the default user.
- No shared default-user seed contract currently exists for the four broken pages; this feature should establish one for local/test/demo usage.
