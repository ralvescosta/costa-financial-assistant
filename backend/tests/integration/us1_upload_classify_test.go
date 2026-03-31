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
	commonv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/common/v1"
	filesv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/files/v1"
)

// TestUS1_UploadAndClassifyDocument validates the full upload → classify → list flow.
//
// It starts an in-process gRPC server backed by the real PostgreSQL test database,
// sends a document upload request, classifies it, and verifies the project-scoped list
// returns the expected record in pending-analysis state, then classified state.
func TestUS1_UploadAndClassifyDocument(t *testing.T) {
	require.NoError(t, runMigrations(testDSN(), "file://../../internals/files/migrations"))

	t.Cleanup(func() {
		_, _ = testDB.ExecContext(context.Background(), "DELETE FROM documents")
	})

	client := newFilesClient(t, testDB)

	const (
		projectID  = "00000000-0000-0000-0000-000000000010" // bootstrap project from seed
		uploadedBy = "00000000-0000-0000-0000-000000000001" // bootstrap user from seed
		fileName   = "invoice_2024_01.pdf"
		fileHash   = "abc123def456abc123def456abc123def456abc123def456abc123def456abc1" // 64-char hex
	)

	projectCtx := &commonv1.ProjectContext{ProjectId: projectID}
	audit := &commonv1.AuditMetadata{PerformedBy: uploadedBy}

	// ── Step 1: Upload document ───────────────────────────────────────────────
	uploadResp, err := client.UploadDocument(context.Background(), &filesv1.UploadDocumentRequest{
		Ctx:             projectCtx,
		FileName:        fileName,
		FileHash:        fileHash,
		StorageProvider: "local",
		StorageKey:      "local/" + fileHash,
		Audit:           audit,
	})
	require.NoError(t, err, "upload should succeed")
	require.NotNil(t, uploadResp.Document)

	docID := uploadResp.Document.Id
	assert.NotEmpty(t, docID, "document ID must be populated")
	assert.Equal(t, "pending", analysisStatusName(uploadResp.Document.AnalysisStatus))
	assert.Equal(t, "unspecified", documentKindName(uploadResp.Document.Kind))
	assert.Equal(t, projectID, uploadResp.Document.ProjectId)

	// ── Step 2: Duplicate upload should be rejected ───────────────────────────
	_, err = client.UploadDocument(context.Background(), &filesv1.UploadDocumentRequest{
		Ctx:             projectCtx,
		FileName:        fileName,
		FileHash:        fileHash,
		StorageProvider: "local",
		StorageKey:      "local/" + fileHash,
		Audit:           audit,
	})
	require.Error(t, err, "duplicate upload must return an error")
	assert.Contains(t, err.Error(), "AlreadyExists", "duplicate should produce AlreadyExists gRPC error")

	// ── Step 3: Classify the document as a bill ───────────────────────────────
	classifyResp, err := client.ClassifyDocument(context.Background(), &filesv1.ClassifyDocumentRequest{
		Ctx:        projectCtx,
		DocumentId: docID,
		Kind:       filesv1.DocumentKind_DOCUMENT_KIND_BILL,
		Audit:      audit,
	})
	require.NoError(t, err, "classify should succeed")
	assert.Equal(t, "bill", documentKindName(classifyResp.Document.Kind))
	assert.Equal(t, docID, classifyResp.Document.Id)

	// ── Step 4: List documents — should show exactly one record ──────────────
	listResp, err := client.ListDocuments(context.Background(), &filesv1.ListDocumentsRequest{
		Ctx: projectCtx,
		Pagination: &commonv1.Pagination{
			PageSize: 10,
		},
	})
	require.NoError(t, err, "list should succeed")
	require.Len(t, listResp.Documents, 1, "expected exactly one document in project")
	assert.Equal(t, docID, listResp.Documents[0].Id)
	assert.Equal(t, "pending", analysisStatusName(listResp.Documents[0].AnalysisStatus))

	// ── Step 5: Get document by ID ────────────────────────────────────────────
	getResp, err := client.GetDocument(context.Background(), &filesv1.GetDocumentRequest{
		Ctx:        projectCtx,
		DocumentId: docID,
	})
	require.NoError(t, err, "get document should succeed")
	assert.Equal(t, "bill", documentKindName(getResp.Document.Kind))
	assert.Equal(t, fileName, getResp.Document.FileName)
}

// ─── helpers ──────────────────────────────────────────────────────────────────

// newFilesClient starts an in-process gRPC files server backed by testDB and
// returns a client connected to it.
func newFilesClient(t *testing.T, db *sql.DB) filesv1.FilesServiceClient {
	t.Helper()
	logger := zaptest.NewLogger(t)

	repo := filesrepo.NewDocumentRepository(db, logger)
	jobRepo := filesrepo.NewAnalysisJobRepository(db, logger)
	billRepo := filesrepo.NewBillRecordRepository(db, logger)
	stmtRepo := filesrepo.NewStatementRecordRepository(db, logger)
	uow := filesrepo.NewUnitOfWork(db)
	extractor := filessvc.NewStubPDFExtractor()

	svc := filessvc.NewDocumentService(repo, uow, logger)
	extSvc := filessvc.NewExtractionService(repo, jobRepo, billRepo, stmtRepo, uow, extractor, logger)
	srv := filesgrpc.NewServer(svc, extSvc, logger)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcSrv := grpc.NewServer()
	filesv1.RegisterFilesServiceServer(grpcSrv, srv)

	t.Cleanup(func() { grpcSrv.Stop() })
	go func() { _ = grpcSrv.Serve(lis) }()

	conn, err := grpc.NewClient(
		lis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	return filesv1.NewFilesServiceClient(conn)
}

func analysisStatusName(s filesv1.AnalysisStatus) string {
	switch s {
	case filesv1.AnalysisStatus_ANALYSIS_STATUS_PENDING:
		return "pending"
	case filesv1.AnalysisStatus_ANALYSIS_STATUS_PROCESSING:
		return "processing"
	case filesv1.AnalysisStatus_ANALYSIS_STATUS_ANALYSED:
		return "analysed"
	case filesv1.AnalysisStatus_ANALYSIS_STATUS_ANALYSIS_FAILED:
		return "analysis_failed"
	default:
		return "unspecified"
	}
}

func documentKindName(k filesv1.DocumentKind) string {
	switch k {
	case filesv1.DocumentKind_DOCUMENT_KIND_BILL:
		return "bill"
	case filesv1.DocumentKind_DOCUMENT_KIND_STATEMENT:
		return "statement"
	default:
		return "unspecified"
	}
}
