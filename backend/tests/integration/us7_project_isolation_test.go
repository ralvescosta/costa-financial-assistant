//go:build integration

package integration

import (
	"context"
	"database/sql"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	filesrepo "github.com/ralvescosta/costa-financial-assistant/backend/internals/files/repositories"
	filessvc "github.com/ralvescosta/costa-financial-assistant/backend/internals/files/services"
	filesgrpc "github.com/ralvescosta/costa-financial-assistant/backend/internals/files/transport/grpc"
	onboardingrepo "github.com/ralvescosta/costa-financial-assistant/backend/internals/onboarding/repositories"
	onboardingsvc "github.com/ralvescosta/costa-financial-assistant/backend/internals/onboarding/services"
	onboardinggrpc "github.com/ralvescosta/costa-financial-assistant/backend/internals/onboarding/transport/grpc"
	commonv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/common/v1"
	filesv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/files/v1"
	onboardingv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/onboarding/v1"
)

// TestUS7_CrossProjectIsolation verifies that documents uploaded in project A are
// invisible when the request is made with a project B context.
func TestUS7_CrossProjectIsolation(t *testing.T) {
	require.NoError(t, runMigrations(testDSN(), "file://../../internals/files/migrations"))

	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(), "DELETE FROM documents")
		_, _ = testDB.ExecContext(context.Background(),
			"DELETE FROM project_members WHERE project_id NOT IN ('00000000-0000-0000-0000-000000000010')")
		_, _ = testDB.ExecContext(context.Background(),
			"DELETE FROM projects WHERE id NOT IN ('00000000-0000-0000-0000-000000000010')")
	})

	onboardingClient := newOnboardingClient(t, testDB)
	filesClient := newFilesClient(t, testDB)

	const (
		ownerID    = "00000000-0000-0000-0000-000000000001" // bootstrap user
		projectAID = "00000000-0000-0000-0000-000000000010" // bootstrap project
		fileHash   = "aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899"
	)

	// ── Step 1: Create a second project (project B) ───────────────────────────
	createResp, err := onboardingClient.CreateProject(context.Background(), &onboardingv1.CreateProjectRequest{
		Ctx:  &commonv1.ProjectContext{ProjectId: projectAID, UserId: ownerID, Role: "write"},
		Name: "Project B",
		Type: onboardingv1.ProjectType_PROJECT_TYPE_PERSONAL,
	})
	require.NoError(t, err, "create project B should succeed")
	projectBID := createResp.GetProject().GetId()
	require.NotEmpty(t, projectBID, "project B must have an ID")

	// ── Step 2: Upload a document in project A ────────────────────────────────
	uploadResp, err := filesClient.UploadDocument(context.Background(), &filesv1.UploadDocumentRequest{
		Ctx:             &commonv1.ProjectContext{ProjectId: projectAID},
		FileName:        "bill_project_a.pdf",
		FileHash:        fileHash,
		StorageProvider: "local",
		StorageKey:      "local/" + fileHash,
		Audit:           &commonv1.AuditMetadata{PerformedBy: ownerID},
	})
	require.NoError(t, err, "upload in project A should succeed")
	require.NotNil(t, uploadResp.Document)

	// ── Step 3: List documents from project A context → should see 1 doc ─────
	listA, err := filesClient.ListDocuments(context.Background(), &filesv1.ListDocumentsRequest{
		Ctx:        &commonv1.ProjectContext{ProjectId: projectAID},
		Pagination: &commonv1.Pagination{PageSize: 10},
	})
	require.NoError(t, err, "list in project A should succeed")
	assert.GreaterOrEqual(t, len(listA.Documents), 1, "project A should see at least the uploaded document")

	docIDs := make([]string, 0, len(listA.Documents))
	for _, d := range listA.Documents {
		docIDs = append(docIDs, d.Id)
	}
	assert.Contains(t, docIDs, uploadResp.Document.Id, "project A list must include uploaded doc ID")

	// ── Step 4: List documents from project B context → should see 0 docs ────
	listB, err := filesClient.ListDocuments(context.Background(), &filesv1.ListDocumentsRequest{
		Ctx:        &commonv1.ProjectContext{ProjectId: projectBID},
		Pagination: &commonv1.Pagination{PageSize: 10},
	})
	require.NoError(t, err, "list in project B should succeed")
	assert.Empty(t, listB.Documents, "project B must not see project A documents (strict isolation)")

	// ── Step 5: Direct get of project A doc from project B should fail ────────
	_, err = filesClient.GetDocument(context.Background(), &filesv1.GetDocumentRequest{
		Ctx:        &commonv1.ProjectContext{ProjectId: projectBID},
		DocumentId: uploadResp.Document.Id,
	})
	require.Error(t, err, "cross-project get must return an error")
	assert.Contains(t, err.Error(), "NotFound", "cross-project get must return NotFound")
}

// ─── helpers ──────────────────────────────────────────────────────────────────

// newOnboardingClient starts an in-process gRPC onboarding server backed by
// testDB and returns a client connected to it.
func newOnboardingClient(t *testing.T, db *sql.DB) onboardingv1.OnboardingServiceClient {
	t.Helper()
	logger := zaptest.NewLogger(t)

	repo := onboardingrepo.NewProjectMembersRepository(db, logger)
	svc := onboardingsvc.NewProjectMembersService(repo, logger)
	grpcSrv, _ := onboardinggrpc.NewServer(svc, logger)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	t.Cleanup(func() { grpcSrv.Stop() })
	go func() { _ = grpcSrv.Serve(lis) }()

	conn, err := grpc.NewClient(
		lis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	return onboardingv1.NewOnboardingServiceClient(conn)
}

// filesClientForProject creates a files gRPC client backed by the shared test DB.
// Re-uses the existing newFilesClient helper; aliased here for readability in
// multi-service tests.
func filesClientForProject(t *testing.T, db *sql.DB) filesv1.FilesServiceClient {
	t.Helper()
	logger := zaptest.NewLogger(t)

	repo := filesrepo.NewDocumentRepository(db, logger)
	jobRepo := filesrepo.NewAnalysisJobRepository(db, logger)
	billRepo := filesrepo.NewBillRecordRepository(db, logger)
	stmtRepo := filesrepo.NewStatementRecordRepository(db, logger)
	bankRepo := filesrepo.NewBankAccountRepository(db, logger)
	uow := filesrepo.NewUnitOfWork(db)
	extractor := filessvc.NewStubPDFExtractor()

	svc := filessvc.NewDocumentService(repo, uow, logger)
	extSvc := filessvc.NewExtractionService(repo, jobRepo, billRepo, stmtRepo, uow, extractor, logger)
	bankSvc := filessvc.NewBankAccountService(bankRepo, logger)
	srv := filesgrpc.NewServer(svc, extSvc, bankSvc, logger)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	filesv1.RegisterFilesServiceServer(grpcServer, srv)

	t.Cleanup(func() { grpcServer.Stop() })
	go func() { _ = grpcServer.Serve(lis) }()

	conn, err := grpc.NewClient(
		lis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	return filesv1.NewFilesServiceClient(conn)
}
