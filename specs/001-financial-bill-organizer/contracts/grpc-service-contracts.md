# gRPC Service Contracts

This feature extends existing service boundaries and introduces/consumes the following
contracts under `backend/protos/<module>/v1/{messages.proto,grpc.proto}`.

## files.v1
- Purpose: file ingestion and metadata persistence for uploaded PDFs.
- Core RPCs:
  - `UploadDocument`
  - `GetDocument`
  - `ListDocuments`
  - `ClassifyDocument`
- Notes:
  - `messages.proto` defines document domain types and analysis status enums.
  - `grpc.proto` defines request/response wrappers and service methods.

## bills.v1
- Purpose: bill extraction persistence, payment dashboard, and payment marking.
- Core RPCs:
  - `GetPaymentDashboard`
  - `MarkBillPaid`
  - `GetBill`
  - `ListBills`
- Notes:
  - Must enforce role permissions and project scoping in all read/write methods.

## onboarding.v1
- Purpose: user/project bootstrap and collaboration lifecycle.
- Core RPCs:
  - `CreateProject`
  - `InviteProjectMember`
  - `UpdateProjectMemberRole`
  - `ListProjectMembers`
- Notes:
  - Role enum values: `READ_ONLY`, `UPDATE`, `WRITE`.
  - In Phase-1, seed-based default project and members are mandatory.

## identity.v1
- Purpose: token issuance and identity verification source.
- Core RPCs:
  - `IssueBootstrapToken`
  - `ValidateToken`
  - `GetJwksMetadata`
- Notes:
  - JWT signing only in identity service.
  - Public JWKS also exposed via HTTP endpoint for validators.

## common.v1 shared messages
- Purpose: reusable structures imported by multiple modules.
- Shared message candidates:
  - `Pagination`
  - `Money`
  - `ProjectContext`
  - `ErrorEnvelope`
  - `AuditMetadata`

## Backward-compatibility rules
- No field number reuse/removal in `v1` protos.
- Breaking changes require a new major folder (`v2`).
- Generated files under `backend/protos/generated/` must be updated on every contract change.
