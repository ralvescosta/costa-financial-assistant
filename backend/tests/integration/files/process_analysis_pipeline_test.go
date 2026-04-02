//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	filesrepo "github.com/ralvescosta/costa-financial-assistant/backend/internals/files/repositories"
	filessvc "github.com/ralvescosta/costa-financial-assistant/backend/internals/files/services"
	commonv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/common/v1"
	filesv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/files/v1"
)

// TestUS2_AnalysisPipeline validates the full async extraction pipeline:
// upload → classify → process → analysed with persisted BillRecord.
func TestUS2_AnalysisPipeline(t *testing.T) {
	require.NoError(t, runMigrations(testDSN(), "file://../../../internals/files/migrations"))

	t.Cleanup(func() {
		ctx := context.Background()
		_, _ = testDB.ExecContext(ctx, "DELETE FROM bill_records")
		_, _ = testDB.ExecContext(ctx, "DELETE FROM statement_records")
		_, _ = testDB.ExecContext(ctx, "DELETE FROM transaction_lines")
		_, _ = testDB.ExecContext(ctx, "DELETE FROM analysis_jobs")
		_, _ = testDB.ExecContext(ctx, "DELETE FROM documents")
	})

	const (
		projectID  = "00000000-0000-0000-0000-000000000010"
		uploadedBy = "00000000-0000-0000-0000-000000000001"
	)

	logger := zaptest.NewLogger(t)
	client := newFilesClient(t, testDB)

	repo := filesrepo.NewDocumentRepository(testDB, logger)
	jobRepo := filesrepo.NewAnalysisJobRepository(testDB, logger)
	billRepo := filesrepo.NewBillRecordRepository(testDB, logger)
	stmtRepo := filesrepo.NewStatementRecordRepository(testDB, logger)
	uow := filesrepo.NewUnitOfWork(testDB)
	extractor := filessvc.NewStubPDFExtractor()
	extSvc := filessvc.NewExtractionService(repo, jobRepo, billRepo, stmtRepo, uow, extractor, logger)

	projectCtx := &commonv1.ProjectContext{ProjectId: projectID}
	audit := &commonv1.AuditMetadata{PerformedBy: uploadedBy}

	// ── Scenario 1: Bill document is processed and reaches "analysed" status ──

	t.Run("GivenBillDocument WhenProcessDocument ThenStatusAnalysedAndBillRecordPersisted", func(t *testing.T) {
		// Given: a document uploaded and classified as a bill
		uploadResp, err := client.UploadDocument(context.Background(), &filesv1.UploadDocumentRequest{
			Ctx:             projectCtx,
			FileName:        "invoice_2024_02.pdf",
			FileHash:        "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			StorageProvider: "local",
			StorageKey:      "local/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			Audit:           audit,
		})
		require.NoError(t, err)
		docID := uploadResp.Document.Id

		_, err = client.ClassifyDocument(context.Background(), &filesv1.ClassifyDocumentRequest{
			Ctx:        projectCtx,
			DocumentId: docID,
			Kind:       filesv1.DocumentKind_DOCUMENT_KIND_BILL,
			Audit:      audit,
		})
		require.NoError(t, err)

		// Arrange: create an analysis job for the bill document
		ctx := context.Background()
		tx, err := uow.Begin(ctx)
		require.NoError(t, err)
		job, err := jobRepo.Create(ctx, tx, &filesv1.AnalysisJob{
			ProjectId:  projectID,
			DocumentId: docID,
			JobType:    "extract_bill",
			Status:     "queued",
		})
		require.NoError(t, err)
		require.NoError(t, uow.Commit(tx))

		// When: the extraction pipeline processes the job
		err = extSvc.ProcessDocument(ctx, job.Id, projectID, docID, filesv1.DocumentKind_DOCUMENT_KIND_BILL)

		// Then: no error is returned and document is in analysed state
		require.NoError(t, err, "processing a bill document should succeed")

		getResp, err := client.GetDocument(context.Background(), &filesv1.GetDocumentRequest{
			Ctx:        projectCtx,
			DocumentId: docID,
		})
		require.NoError(t, err)
		assert.Equal(t, "analysed", analysisStatusName(getResp.Document.AnalysisStatus),
			"document should transition to analysed")
		assert.NotNil(t, getResp.BillRecord, "BillRecord should be populated after successful analysis")
		assert.Equal(t, docID, getResp.BillRecord.DocumentId, "BillRecord should reference the document")
	})

	// ── Scenario 2: Statement document is processed and reaches "analysed" status ──

	t.Run("GivenStatementDocument WhenProcessDocument ThenStatusAnalysedAndStatementRecordPersisted", func(t *testing.T) {
		// Given: a document uploaded and classified as a statement
		uploadResp, err := client.UploadDocument(context.Background(), &filesv1.UploadDocumentRequest{
			Ctx:             projectCtx,
			FileName:        "statement_2024_01.pdf",
			FileHash:        "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
			StorageProvider: "local",
			StorageKey:      "local/cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
			Audit:           audit,
		})
		require.NoError(t, err)
		docID := uploadResp.Document.Id

		_, err = client.ClassifyDocument(context.Background(), &filesv1.ClassifyDocumentRequest{
			Ctx:        projectCtx,
			DocumentId: docID,
			Kind:       filesv1.DocumentKind_DOCUMENT_KIND_STATEMENT,
			Audit:      audit,
		})
		require.NoError(t, err)

		// Arrange: create an analysis job for the statement document
		ctx := context.Background()
		tx, err := uow.Begin(ctx)
		require.NoError(t, err)
		job, err := jobRepo.Create(ctx, tx, &filesv1.AnalysisJob{
			ProjectId:  projectID,
			DocumentId: docID,
			JobType:    "extract_statement",
			Status:     "queued",
		})
		require.NoError(t, err)
		require.NoError(t, uow.Commit(tx))

		// When: the extraction pipeline processes the job
		err = extSvc.ProcessDocument(ctx, job.Id, projectID, docID, filesv1.DocumentKind_DOCUMENT_KIND_STATEMENT)

		// Then: document is in analysed state with a StatementRecord
		require.NoError(t, err, "processing a statement document should succeed")

		getResp, err := client.GetDocument(context.Background(), &filesv1.GetDocumentRequest{
			Ctx:        projectCtx,
			DocumentId: docID,
		})
		require.NoError(t, err)
		assert.Equal(t, "analysed", analysisStatusName(getResp.Document.AnalysisStatus))
		assert.NotNil(t, getResp.StatementRecord, "StatementRecord should be populated after successful analysis")
		assert.Equal(t, docID, getResp.StatementRecord.DocumentId)
	})

	// ── Scenario 3: Analysis fails for unsupported document kind ─────────────

	t.Run("GivenUnspecifiedKind WhenProcessDocument ThenStatusAnalysisFailed", func(t *testing.T) {
		// Given: a document uploaded and left unclassified (UNSPECIFIED kind)
		uploadResp, err := client.UploadDocument(context.Background(), &filesv1.UploadDocumentRequest{
			Ctx:             projectCtx,
			FileName:        "unknown_2024.pdf",
			FileHash:        "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
			StorageProvider: "local",
			StorageKey:      "local/dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
			Audit:           audit,
		})
		require.NoError(t, err)
		docID := uploadResp.Document.Id

		// Arrange: create an analysis job with an unspecified kind
		ctx := context.Background()
		tx, err := uow.Begin(ctx)
		require.NoError(t, err)
		job, err := jobRepo.Create(ctx, tx, &filesv1.AnalysisJob{
			ProjectId:  projectID,
			DocumentId: docID,
			JobType:    "extract_bill", // valid enum; kind check happens in ProcessDocument
			Status:     "queued",
		})
		require.NoError(t, err)
		require.NoError(t, uow.Commit(tx))

		// When: the extraction pipeline is invoked with unspecified kind
		err = extSvc.ProcessDocument(ctx, job.Id, projectID, docID, filesv1.DocumentKind_DOCUMENT_KIND_UNSPECIFIED)

		// Then: an error is returned and the document status is analysis_failed
		require.Error(t, err, "processing with unspecified kind should fail")

		getResp, err := client.GetDocument(context.Background(), &filesv1.GetDocumentRequest{
			Ctx:        projectCtx,
			DocumentId: docID,
		})
		require.NoError(t, err)
		assert.Equal(t, "analysis_failed", analysisStatusName(getResp.Document.AnalysisStatus),
			"document should be marked as analysis_failed")
	})
}
