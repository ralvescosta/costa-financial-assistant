package services_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/files/repositories"
	"github.com/ralvescosta/costa-financial-assistant/backend/internals/files/services"
	filesv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/files/v1"
)

// ─── mock: AnalysisJobRepository ─────────────────────────────────────────────

type mockAnalysisJobRepository struct {
	mock.Mock
}

func (m *mockAnalysisJobRepository) Create(ctx context.Context, tx *sql.Tx, job *filesv1.AnalysisJob) (*filesv1.AnalysisJob, error) {
	args := m.Called(ctx, tx, job)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filesv1.AnalysisJob), args.Error(1)
}

func (m *mockAnalysisJobRepository) FindByDocumentID(ctx context.Context, projectID, documentID string) (*filesv1.AnalysisJob, error) {
	args := m.Called(ctx, projectID, documentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filesv1.AnalysisJob), args.Error(1)
}

func (m *mockAnalysisJobRepository) UpdateStatus(ctx context.Context, tx *sql.Tx, jobID, status, lastError string, attemptCount int32) error {
	args := m.Called(ctx, tx, jobID, status, lastError, attemptCount)
	return args.Error(0)
}

func (m *mockAnalysisJobRepository) UpdateDocumentAnalysisStatus(ctx context.Context, tx *sql.Tx, projectID, documentID, analysisStatus, failureReason string) error {
	args := m.Called(ctx, tx, projectID, documentID, analysisStatus, failureReason)
	return args.Error(0)
}

// ─── mock: BillRecordRepository ──────────────────────────────────────────────

type mockBillRecordRepository struct {
	mock.Mock
}

func (m *mockBillRecordRepository) Create(ctx context.Context, tx *sql.Tx, record *filesv1.BillRecord) (*filesv1.BillRecord, error) {
	args := m.Called(ctx, tx, record)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filesv1.BillRecord), args.Error(1)
}

func (m *mockBillRecordRepository) FindByProjectAndDocumentID(ctx context.Context, projectID, documentID string) (*filesv1.BillRecord, error) {
	args := m.Called(ctx, projectID, documentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filesv1.BillRecord), args.Error(1)
}

// ─── mock: StatementRecordRepository ─────────────────────────────────────────

type mockStatementRecordRepository struct {
	mock.Mock
}

func (m *mockStatementRecordRepository) Create(ctx context.Context, tx *sql.Tx, record *filesv1.StatementRecord) (*filesv1.StatementRecord, error) {
	args := m.Called(ctx, tx, record)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filesv1.StatementRecord), args.Error(1)
}

func (m *mockStatementRecordRepository) FindByProjectAndDocumentID(ctx context.Context, projectID, documentID string) (*filesv1.StatementRecord, error) {
	args := m.Called(ctx, projectID, documentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filesv1.StatementRecord), args.Error(1)
}

// ─── mock: PDFExtractor ───────────────────────────────────────────────────────

type mockPDFExtractor struct {
	mock.Mock
}

func (m *mockPDFExtractor) ExtractBill(ctx context.Context, storageKey string) (*services.BillExtractionResult, error) {
	args := m.Called(ctx, storageKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.BillExtractionResult), args.Error(1)
}

func (m *mockPDFExtractor) ExtractStatement(ctx context.Context, storageKey string) (*services.StatementExtractionResult, error) {
	args := m.Called(ctx, storageKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.StatementExtractionResult), args.Error(1)
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func newExtractionService(
	t *testing.T,
	docRepo repositories.DocumentRepository,
	jobRepo repositories.AnalysisJobRepository,
	billRepo repositories.BillRecordRepository,
	stmtRepo repositories.StatementRecordRepository,
	uow repositories.UnitOfWork,
	extractor services.PDFExtractorIface,
) services.ExtractionServiceIface {
	t.Helper()
	logger := zaptest.NewLogger(t)
	return services.NewExtractionService(docRepo, jobRepo, billRepo, stmtRepo, uow, extractor, logger)
}

const (
	testProjectID  = "project-1"
	testDocumentID = "doc-1"
	testJobID      = "job-1"
	testStorageKey = "local/abc123"
)

func billDocument() *filesv1.Document {
	return &filesv1.Document{
		Id:             testDocumentID,
		ProjectId:      testProjectID,
		Kind:           filesv1.DocumentKind_DOCUMENT_KIND_BILL,
		StorageKey:     testStorageKey,
		AnalysisStatus: filesv1.AnalysisStatus_ANALYSIS_STATUS_PENDING,
	}
}

func statementDocument() *filesv1.Document {
	return &filesv1.Document{
		Id:             testDocumentID,
		ProjectId:      testProjectID,
		Kind:           filesv1.DocumentKind_DOCUMENT_KIND_STATEMENT,
		StorageKey:     testStorageKey,
		AnalysisStatus: filesv1.AnalysisStatus_ANALYSIS_STATUS_PENDING,
	}
}

func analysedBillDocument() *filesv1.Document {
	doc := billDocument()
	doc.AnalysisStatus = filesv1.AnalysisStatus_ANALYSIS_STATUS_ANALYSED
	return doc
}

func analysedStatementDocument() *filesv1.Document {
	doc := statementDocument()
	doc.AnalysisStatus = filesv1.AnalysisStatus_ANALYSIS_STATUS_ANALYSED
	return doc
}

// ─── ProcessDocument ─────────────────────────────────────────────────────────

func TestExtractionService_ProcessDocument(t *testing.T) {
	t.Run("GivenBillDocument WhenProcessDocument ThenBillRecordPersistedAndStatusAnalysed", func(t *testing.T) {
		// Given a classified bill document with a valid storage key
		// When ProcessDocument is called with DOCUMENT_KIND_BILL
		// Then bill extraction runs, BillRecord is persisted, and document reaches "analysed"

		// Arrange
		ctx := context.Background()
		docRepo := new(mockDocumentRepository)
		jobRepo := new(mockAnalysisJobRepository)
		billRepo := new(mockBillRecordRepository)
		stmtRepo := new(mockStatementRecordRepository)
		uow := new(mockUnitOfWork)
		extractor := new(mockPDFExtractor)

		svc := newExtractionService(t, docRepo, jobRepo, billRepo, stmtRepo, uow, extractor)

		extractResult := &services.BillExtractionResult{
			DueDate:   "2024-02-15",
			AmountDue: "1500.00",
		}

		// UoW for processing status transition
		uow.On("Begin", ctx).Return((*sql.Tx)(nil), nil).Once()
		jobRepo.On("UpdateDocumentAnalysisStatus", ctx, (*sql.Tx)(nil), testProjectID, testDocumentID, "processing", "").Return(nil).Once()
		jobRepo.On("UpdateStatus", ctx, (*sql.Tx)(nil), testJobID, "running", "", int32(1)).Return(nil).Once()
		uow.On("Commit", (*sql.Tx)(nil)).Return(nil).Once()
		uow.On("Rollback", (*sql.Tx)(nil)).Return(nil)

		// PDF extraction + BillRecord persistence
		docRepo.On("FindByProjectAndID", ctx, testProjectID, testDocumentID).Return(billDocument(), nil)
		extractor.On("ExtractBill", ctx, testStorageKey).Return(extractResult, nil)
		uow.On("Begin", ctx).Return((*sql.Tx)(nil), nil).Once()
		billRepo.On("Create", ctx, (*sql.Tx)(nil), mock.AnythingOfType("*filesv1.BillRecord")).
			Return(&filesv1.BillRecord{Id: "bill-1", DocumentId: testDocumentID}, nil)
		uow.On("Commit", (*sql.Tx)(nil)).Return(nil).Once()

		// Final status update
		uow.On("Begin", ctx).Return((*sql.Tx)(nil), nil).Once()
		jobRepo.On("UpdateDocumentAnalysisStatus", ctx, (*sql.Tx)(nil), testProjectID, testDocumentID, "analysed", "").Return(nil).Once()
		jobRepo.On("UpdateStatus", ctx, (*sql.Tx)(nil), testJobID, "succeeded", "", int32(1)).Return(nil).Once()
		uow.On("Commit", (*sql.Tx)(nil)).Return(nil).Once()

		// Act
		err := svc.ProcessDocument(ctx, testJobID, testProjectID, testDocumentID, filesv1.DocumentKind_DOCUMENT_KIND_BILL)

		// Assert
		require.NoError(t, err)
		billRepo.AssertCalled(t, "Create", ctx, (*sql.Tx)(nil), mock.AnythingOfType("*filesv1.BillRecord"))
		jobRepo.AssertCalled(t, "UpdateDocumentAnalysisStatus", ctx, (*sql.Tx)(nil), testProjectID, testDocumentID, "analysed", "")
	})

	t.Run("GivenStatementDocument WhenProcessDocument ThenStatementRecordPersistedAndStatusAnalysed", func(t *testing.T) {
		// Given a classified statement document
		// When ProcessDocument is called with DOCUMENT_KIND_STATEMENT
		// Then statement extraction runs and StatementRecord is persisted

		// Arrange
		ctx := context.Background()
		docRepo := new(mockDocumentRepository)
		jobRepo := new(mockAnalysisJobRepository)
		billRepo := new(mockBillRecordRepository)
		stmtRepo := new(mockStatementRecordRepository)
		uow := new(mockUnitOfWork)
		extractor := new(mockPDFExtractor)

		svc := newExtractionService(t, docRepo, jobRepo, billRepo, stmtRepo, uow, extractor)

		extractResult := &services.StatementExtractionResult{
			PeriodStart: "2024-01-01",
			PeriodEnd:   "2024-01-31",
		}

		// UoW for processing status
		uow.On("Begin", ctx).Return((*sql.Tx)(nil), nil).Once()
		jobRepo.On("UpdateDocumentAnalysisStatus", ctx, (*sql.Tx)(nil), testProjectID, testDocumentID, "processing", "").Return(nil)
		jobRepo.On("UpdateStatus", ctx, (*sql.Tx)(nil), testJobID, "running", "", int32(1)).Return(nil).Once()
		uow.On("Commit", (*sql.Tx)(nil)).Return(nil).Once()
		uow.On("Rollback", (*sql.Tx)(nil)).Return(nil)

		docRepo.On("FindByProjectAndID", ctx, testProjectID, testDocumentID).Return(statementDocument(), nil)
		extractor.On("ExtractStatement", ctx, testStorageKey).Return(extractResult, nil)
		uow.On("Begin", ctx).Return((*sql.Tx)(nil), nil).Once()
		stmtRepo.On("Create", ctx, (*sql.Tx)(nil), mock.AnythingOfType("*filesv1.StatementRecord")).
			Return(&filesv1.StatementRecord{Id: "stmt-1", DocumentId: testDocumentID}, nil)
		uow.On("Commit", (*sql.Tx)(nil)).Return(nil).Once()

		// Final status
		uow.On("Begin", ctx).Return((*sql.Tx)(nil), nil).Once()
		jobRepo.On("UpdateDocumentAnalysisStatus", ctx, (*sql.Tx)(nil), testProjectID, testDocumentID, "analysed", "").Return(nil).Once()
		jobRepo.On("UpdateStatus", ctx, (*sql.Tx)(nil), testJobID, "succeeded", "", int32(1)).Return(nil).Once()
		uow.On("Commit", (*sql.Tx)(nil)).Return(nil).Once()

		// Act
		err := svc.ProcessDocument(ctx, testJobID, testProjectID, testDocumentID, filesv1.DocumentKind_DOCUMENT_KIND_STATEMENT)

		// Assert
		require.NoError(t, err)
		stmtRepo.AssertCalled(t, "Create", ctx, (*sql.Tx)(nil), mock.AnythingOfType("*filesv1.StatementRecord"))
		jobRepo.AssertCalled(t, "UpdateDocumentAnalysisStatus", ctx, (*sql.Tx)(nil), testProjectID, testDocumentID, "analysed", "")
	})

	t.Run("GivenUnspecifiedKind WhenProcessDocument ThenStatusAnalysisFailed", func(t *testing.T) {
		// Given a document with unspecified kind
		// When ProcessDocument is invoked with DOCUMENT_KIND_UNSPECIFIED
		// Then an error is returned and the document is marked analysis_failed

		// Arrange
		ctx := context.Background()
		docRepo := new(mockDocumentRepository)
		jobRepo := new(mockAnalysisJobRepository)
		billRepo := new(mockBillRecordRepository)
		stmtRepo := new(mockStatementRecordRepository)
		uow := new(mockUnitOfWork)
		extractor := new(mockPDFExtractor)

		svc := newExtractionService(t, docRepo, jobRepo, billRepo, stmtRepo, uow, extractor)

		uow.On("Begin", ctx).Return((*sql.Tx)(nil), nil)
		jobRepo.On("UpdateDocumentAnalysisStatus", ctx, (*sql.Tx)(nil), testProjectID, testDocumentID, "processing", "").Return(nil)
		jobRepo.On("UpdateStatus", ctx, (*sql.Tx)(nil), testJobID, "running", "", int32(1)).Return(nil)
		uow.On("Commit", (*sql.Tx)(nil)).Return(nil)
		uow.On("Rollback", (*sql.Tx)(nil)).Return(nil)

		// Expect failure status after unsupported kind
		jobRepo.On("UpdateDocumentAnalysisStatus", ctx, (*sql.Tx)(nil), testProjectID, testDocumentID, "analysis_failed", mock.AnythingOfType("string")).Return(nil)
		jobRepo.On("UpdateStatus", ctx, (*sql.Tx)(nil), testJobID, "failed", mock.AnythingOfType("string"), int32(1)).Return(nil)

		// Act
		err := svc.ProcessDocument(ctx, testJobID, testProjectID, testDocumentID, filesv1.DocumentKind_DOCUMENT_KIND_UNSPECIFIED)

		// Assert
		require.Error(t, err, "processing an unsupported kind should return an error")
		jobRepo.AssertCalled(t, "UpdateDocumentAnalysisStatus", ctx, (*sql.Tx)(nil), testProjectID, testDocumentID, "analysis_failed", mock.AnythingOfType("string"))
		billRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
		stmtRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("GivenExtractorFails WhenProcessDocument ThenStatusAnalysisFailed", func(t *testing.T) {
		// Given the PDF extractor encounters an error during bill extraction
		// When ProcessDocument is called
		// Then the document is marked analysis_failed and the error is propagated

		// Arrange
		ctx := context.Background()
		docRepo := new(mockDocumentRepository)
		jobRepo := new(mockAnalysisJobRepository)
		billRepo := new(mockBillRecordRepository)
		stmtRepo := new(mockStatementRecordRepository)
		uow := new(mockUnitOfWork)
		extractor := new(mockPDFExtractor)

		svc := newExtractionService(t, docRepo, jobRepo, billRepo, stmtRepo, uow, extractor)

		extractErr := errors.New("malformed PDF")

		uow.On("Begin", ctx).Return((*sql.Tx)(nil), nil)
		jobRepo.On("UpdateDocumentAnalysisStatus", ctx, (*sql.Tx)(nil), testProjectID, testDocumentID, "processing", "").Return(nil)
		jobRepo.On("UpdateStatus", ctx, (*sql.Tx)(nil), testJobID, "running", "", int32(1)).Return(nil).Once()
		uow.On("Commit", (*sql.Tx)(nil)).Return(nil)
		uow.On("Rollback", (*sql.Tx)(nil)).Return(nil)

		docRepo.On("FindByProjectAndID", ctx, testProjectID, testDocumentID).Return(billDocument(), nil)
		extractor.On("ExtractBill", ctx, testStorageKey).Return(nil, extractErr)

		jobRepo.On("UpdateDocumentAnalysisStatus", ctx, (*sql.Tx)(nil), testProjectID, testDocumentID, "analysis_failed", mock.AnythingOfType("string")).Return(nil)
		jobRepo.On("UpdateStatus", ctx, (*sql.Tx)(nil), testJobID, "failed", mock.AnythingOfType("string"), int32(1)).Return(nil).Once()

		// Act
		err := svc.ProcessDocument(ctx, testJobID, testProjectID, testDocumentID, filesv1.DocumentKind_DOCUMENT_KIND_BILL)

		// Assert
		require.Error(t, err)
		jobRepo.AssertCalled(t, "UpdateDocumentAnalysisStatus", ctx, (*sql.Tx)(nil), testProjectID, testDocumentID, "analysis_failed", mock.AnythingOfType("string"))
		billRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("GivenTxBeginFails WhenProcessDocument ThenWrappedErrorReturned", func(t *testing.T) {
		// Given the database cannot begin a transaction
		// When ProcessDocument is called
		// Then a wrapped error is returned without any status update

		// Arrange
		ctx := context.Background()
		docRepo := new(mockDocumentRepository)
		jobRepo := new(mockAnalysisJobRepository)
		billRepo := new(mockBillRecordRepository)
		stmtRepo := new(mockStatementRecordRepository)
		uow := new(mockUnitOfWork)
		extractor := new(mockPDFExtractor)

		svc := newExtractionService(t, docRepo, jobRepo, billRepo, stmtRepo, uow, extractor)

		dbErr := errors.New("db unavailable")
		uow.On("Begin", ctx).Return((*sql.Tx)(nil), dbErr)
		uow.On("Rollback", (*sql.Tx)(nil)).Return(nil)

		// Act
		err := svc.ProcessDocument(ctx, testJobID, testProjectID, testDocumentID, filesv1.DocumentKind_DOCUMENT_KIND_BILL)

		// Assert
		require.Error(t, err)
		assert.ErrorContains(t, err, "begin tx")
		jobRepo.AssertNotCalled(t, "UpdateDocumentAnalysisStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})
}

// ─── GetDocumentDetail ────────────────────────────────────────────────────────

func TestExtractionService_GetDocumentDetail(t *testing.T) {
	t.Run("GivenPendingDocument WhenGetDocumentDetail ThenDocumentReturnedWithoutRecords", func(t *testing.T) {
		// Given a document in pending analysis state
		// When GetDocumentDetail is called
		// Then the document is returned without a BillRecord or StatementRecord

		// Arrange
		ctx := context.Background()
		docRepo := new(mockDocumentRepository)
		jobRepo := new(mockAnalysisJobRepository)
		billRepo := new(mockBillRecordRepository)
		stmtRepo := new(mockStatementRecordRepository)
		uow := new(mockUnitOfWork)
		extractor := new(mockPDFExtractor)

		svc := newExtractionService(t, docRepo, jobRepo, billRepo, stmtRepo, uow, extractor)

		docRepo.On("FindByProjectAndID", ctx, testProjectID, testDocumentID).Return(billDocument(), nil)

		// Act
		doc, bill, stmt, err := svc.GetDocumentDetail(ctx, testProjectID, testDocumentID)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, doc)
		assert.Nil(t, bill, "no BillRecord expected for pending document")
		assert.Nil(t, stmt, "no StatementRecord expected for pending document")
		billRepo.AssertNotCalled(t, "FindByProjectAndDocumentID", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("GivenAnalysedBillDocument WhenGetDocumentDetail ThenBillRecordReturned", func(t *testing.T) {
		// Given a document in analysed state and classified as a bill
		// When GetDocumentDetail is called
		// Then the document and its BillRecord are returned

		// Arrange
		ctx := context.Background()
		docRepo := new(mockDocumentRepository)
		jobRepo := new(mockAnalysisJobRepository)
		billRepo := new(mockBillRecordRepository)
		stmtRepo := new(mockStatementRecordRepository)
		uow := new(mockUnitOfWork)
		extractor := new(mockPDFExtractor)

		svc := newExtractionService(t, docRepo, jobRepo, billRepo, stmtRepo, uow, extractor)

		expectedBill := &filesv1.BillRecord{
			Id:         "bill-1",
			DocumentId: testDocumentID,
			DueDate:    "2024-02-15",
			AmountDue:  "1500.00",
		}

		docRepo.On("FindByProjectAndID", ctx, testProjectID, testDocumentID).Return(analysedBillDocument(), nil)
		billRepo.On("FindByProjectAndDocumentID", ctx, testProjectID, testDocumentID).Return(expectedBill, nil)

		// Act
		doc, bill, stmt, err := svc.GetDocumentDetail(ctx, testProjectID, testDocumentID)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, doc)
		assert.Equal(t, expectedBill.Id, bill.Id)
		assert.Equal(t, "2024-02-15", bill.DueDate)
		assert.Nil(t, stmt)
	})

	t.Run("GivenAnalysedStatementDocument WhenGetDocumentDetail ThenStatementRecordReturned", func(t *testing.T) {
		// Given a document in analysed state and classified as a statement
		// When GetDocumentDetail is called
		// Then the document and its StatementRecord are returned

		// Arrange
		ctx := context.Background()
		docRepo := new(mockDocumentRepository)
		jobRepo := new(mockAnalysisJobRepository)
		billRepo := new(mockBillRecordRepository)
		stmtRepo := new(mockStatementRecordRepository)
		uow := new(mockUnitOfWork)
		extractor := new(mockPDFExtractor)

		svc := newExtractionService(t, docRepo, jobRepo, billRepo, stmtRepo, uow, extractor)

		expectedStmt := &filesv1.StatementRecord{
			Id:          "stmt-1",
			DocumentId:  testDocumentID,
			PeriodStart: "2024-01-01",
			PeriodEnd:   "2024-01-31",
		}

		docRepo.On("FindByProjectAndID", ctx, testProjectID, testDocumentID).Return(analysedStatementDocument(), nil)
		stmtRepo.On("FindByProjectAndDocumentID", ctx, testProjectID, testDocumentID).Return(expectedStmt, nil)

		// Act
		doc, bill, stmt, err := svc.GetDocumentDetail(ctx, testProjectID, testDocumentID)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, doc)
		assert.Nil(t, bill)
		assert.Equal(t, expectedStmt.Id, stmt.Id)
		assert.Equal(t, "2024-01-01", stmt.PeriodStart)
	})

	t.Run("GivenDocumentNotFound WhenGetDocumentDetail ThenErrorPropagated", func(t *testing.T) {
		// Given no document exists with this ID in the project
		// When GetDocumentDetail is called
		// Then the not-found error is returned without any record lookup

		// Arrange
		ctx := context.Background()
		docRepo := new(mockDocumentRepository)
		jobRepo := new(mockAnalysisJobRepository)
		billRepo := new(mockBillRecordRepository)
		stmtRepo := new(mockStatementRecordRepository)
		uow := new(mockUnitOfWork)
		extractor := new(mockPDFExtractor)

		svc := newExtractionService(t, docRepo, jobRepo, billRepo, stmtRepo, uow, extractor)

		docRepo.On("FindByProjectAndID", ctx, testProjectID, testDocumentID).
			Return(nil, repositories.ErrDocumentNotFound)

		// Act
		doc, bill, stmt, err := svc.GetDocumentDetail(ctx, testProjectID, testDocumentID)

		// Assert
		require.Error(t, err)
		assert.Nil(t, doc)
		assert.Nil(t, bill)
		assert.Nil(t, stmt)
		billRepo.AssertNotCalled(t, "FindByProjectAndDocumentID", mock.Anything, mock.Anything, mock.Anything)
		stmtRepo.AssertNotCalled(t, "FindByProjectAndDocumentID", mock.Anything, mock.Anything, mock.Anything)
	})
}
