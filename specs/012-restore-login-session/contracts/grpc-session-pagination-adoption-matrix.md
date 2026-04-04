# gRPC Session + Pagination Adoption Matrix

This matrix defines the contract obligations for `012-restore-login-session`.

## Rules

1. `common.v1.Session` is required on every authenticated gRPC request in scope.
2. `common.v1.ProjectContext` remains required for tenant isolation; `Session` supplements it and does not replace it.
3. Every list/select request that can return multiple items must carry a populated `common.v1.Pagination` object.
4. The BFF is responsible for reading HTTP query parameters and applying deterministic defaults when they are absent.

## Adoption matrix

| Service | Request messages in verified scope | `Session` required | `Pagination` required | Notes |
|---|---|---:|---:|---|
| `identity.v1` | login/bootstrap issuance and token validation RPCs | No for the unauthenticated bootstrap call itself | No | These RPCs establish the session; downstream protected calls use `Session` after login succeeds. |
| `onboarding.v1` | `CreateProjectRequest`, `InviteProjectMemberRequest`, `UpdateProjectMemberRoleRequest`, `ListProjectMembersRequest`, `GetProjectRequest` | Yes | `ListProjectMembersRequest`: Yes | Preserve `ProjectContext`; when the frontend omits query params, the BFF should forward `page_size = 20` and `page_token = ""`. |
| `files.v1` | `UploadDocumentRequest`, `ClassifyDocumentRequest`, `GetDocumentRequest`, `ListDocumentsRequest`, `CreateBankAccountRequest`, `ListBankAccountsRequest`, `DeleteBankAccountRequest` | Yes | `ListDocumentsRequest` and `ListBankAccountsRequest`: Yes | Normalize any current `25`-item BFF defaults to the verified-scope standard of `20` unless a smaller documented clamp is intentionally kept. |
| `bills.v1` | `GetPaymentDashboardRequest`, `MarkBillPaidRequest`, `GetBillRequest`, `ListBillsRequest` | Yes | `GetPaymentDashboardRequest` and `ListBillsRequest`: Yes | The current payment dashboard default already aligns with the target `20`-item first page. |
| `payments.v1` | `GetCyclePreferenceRequest`, `SetCyclePreferenceRequest`, `GetHistoryTimelineRequest`, `GetHistoryCategoryBreakdownRequest`, `GetHistoryComplianceRequest`, `GetReconciliationSummaryRequest`, `CreateManualLinkRequest` | Yes | `GetHistoryTimelineRequest`, `GetHistoryCategoryBreakdownRequest`, `GetHistoryComplianceRequest`, and `GetReconciliationSummaryRequest`: Yes | History and reconciliation requests already carry pagination in the verified scope; the BFF must still forward the default first page when the UI omits pagination. |

## Verification expectations

- Add or refresh unit tests proving the BFF always creates non-nil pagination with the documented verified-scope defaults when the UI omits query params (`25` for documents and project members, `20` for payment dashboard/settings flows unless a route-specific override is documented).
- Add or refresh integration tests proving protected BFF routes forward authenticated session context to downstream services.
- Fail the feature if any in-scope authenticated gRPC request still relies on caller identity that is not represented in `common.v1.Session`.
