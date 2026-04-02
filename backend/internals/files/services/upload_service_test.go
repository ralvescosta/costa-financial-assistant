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

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/files/interfaces"
	"github.com/ralvescosta/costa-financial-assistant/backend/internals/files/repositories"
	"github.com/ralvescosta/costa-financial-assistant/backend/internals/files/services"
	filesv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/files/v1"
)

// ─── mock: DocumentRepository ─────────────────────────────────────────────────

type mockDocumentRepository struct {
	mock.Mock
}

func (m *mockDocumentRepository) Create(ctx context.Context, tx *sql.Tx, doc *filesv1.Document) (*filesv1.Document, error) {
	args := m.Called(ctx, tx, doc)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filesv1.Document), args.Error(1)
}

func (m *mockDocumentRepository) FindByProjectAndHash(ctx context.Context, projectID, hash string) (*filesv1.Document, error) {
	args := m.Called(ctx, projectID, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filesv1.Document), args.Error(1)
}

func (m *mockDocumentRepository) FindByProjectAndID(ctx context.Context, projectID, id string) (*filesv1.Document, error) {
	args := m.Called(ctx, projectID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filesv1.Document), args.Error(1)
}

func (m *mockDocumentRepository) UpdateKind(ctx context.Context, tx *sql.Tx, projectID, id string, kind filesv1.DocumentKind) (*filesv1.Document, error) {
	args := m.Called(ctx, tx, projectID, id, kind)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filesv1.Document), args.Error(1)
}

func (m *mockDocumentRepository) ListByProject(ctx context.Context, projectID string, pageSize int32, offsetToken string) ([]*filesv1.Document, error) {
	args := m.Called(ctx, projectID, pageSize, offsetToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*filesv1.Document), args.Error(1)
}

// ─── mock: UnitOfWork ─────────────────────────────────────────────────────────

type mockUnitOfWork struct {
	mock.Mock
}

func (m *mockUnitOfWork) Begin(ctx context.Context) (*sql.Tx, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sql.Tx), args.Error(1)
}

func (m *mockUnitOfWork) Commit(tx *sql.Tx) error {
	args := m.Called(tx)
	return args.Error(0)
}

func (m *mockUnitOfWork) Rollback(tx *sql.Tx) error {
	args := m.Called(tx)
	return args.Error(0)
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func newService(t *testing.T, repo interfaces.DocumentRepository, uow interfaces.UnitOfWork) services.DocumentServiceIface {
	t.Helper()
	logger := zaptest.NewLogger(t)
	return services.NewDocumentService(repo, uow, logger)
}

// validInput returns a well-formed UploadDocumentInput for use in Arrange blocks.
func validInput() *services.UploadDocumentInput {
	return &services.UploadDocumentInput{
		ProjectID:       "project-1",
		UploadedBy:      "user-1",
		FileName:        "document.pdf",
		FileHash:        "aabbcc112233",
		StorageProvider: "local",
		StorageKey:      "local/aabbcc112233",
	}
}

// ─── UploadDocument ───────────────────────────────────────────────────────────

func TestDocumentService_UploadDocument(t *testing.T) {
	t.Run("GivenNewDocument WhenUpload ThenDocumentCreated", func(t *testing.T) {
		// Given a project with no prior document at this hash
		// When the service uploads a new document
		// Then the document is persisted and returned with pending analysis status

		// Arrange
		repo := new(mockDocumentRepository)
		uow := new(mockUnitOfWork)
		svc := newService(t, repo, uow)
		ctx := context.Background()
		input := validInput()

		expected := &filesv1.Document{
			Id:             "doc-uuid-1",
			ProjectId:      input.ProjectID,
			UploadedBy:     input.UploadedBy,
			Kind:           filesv1.DocumentKind_DOCUMENT_KIND_UNSPECIFIED,
			FileName:       input.FileName,
			FileHash:       input.FileHash,
			AnalysisStatus: filesv1.AnalysisStatus_ANALYSIS_STATUS_PENDING,
		}

		repo.On("FindByProjectAndHash", ctx, input.ProjectID, input.FileHash).
			Return(nil, repositories.ErrDocumentNotFound)
		uow.On("Begin", ctx).Return((*sql.Tx)(nil), nil)
		repo.On("Create", ctx, (*sql.Tx)(nil), mock.AnythingOfType("*filesv1.Document")).
			Return(expected, nil)
		uow.On("Commit", (*sql.Tx)(nil)).Return(nil)
		uow.On("Rollback", (*sql.Tx)(nil)).Return(nil)

		// Act
		doc, err := svc.UploadDocument(ctx, input)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expected.Id, doc.Id)
		assert.Equal(t, filesv1.AnalysisStatus_ANALYSIS_STATUS_PENDING, doc.AnalysisStatus)
		assert.Equal(t, filesv1.DocumentKind_DOCUMENT_KIND_UNSPECIFIED, doc.Kind)
		repo.AssertExpectations(t)
		uow.AssertExpectations(t)
	})

	t.Run("GivenDuplicateHash WhenUpload ThenErrDuplicateDocument", func(t *testing.T) {
		// Given a project that already contains a document with the same file hash
		// When the service attempts to upload the same document again
		// Then ErrDuplicateDocument is returned without any write to storage or DB

		// Arrange
		repo := new(mockDocumentRepository)
		uow := new(mockUnitOfWork)
		svc := newService(t, repo, uow)
		ctx := context.Background()
		input := validInput()

		existing := &filesv1.Document{
			Id:        "existing-doc",
			ProjectId: input.ProjectID,
			FileHash:  input.FileHash,
		}
		repo.On("FindByProjectAndHash", ctx, input.ProjectID, input.FileHash).
			Return(existing, nil)

		// Act
		doc, err := svc.UploadDocument(ctx, input)

		// Assert
		require.Error(t, err)
		assert.Nil(t, doc)
		assert.True(t, errors.Is(err, repositories.ErrDuplicateDocument))
		uow.AssertNotCalled(t, "Begin")
		repo.AssertNotCalled(t, "Create")
	})

	t.Run("GivenHashCheckFails WhenUpload ThenWrappedError", func(t *testing.T) {
		// Given the repository returns an unexpected error during duplicate check
		// When the service tries to upload a document
		// Then the error is wrapped and propagated without attempting a write

		// Arrange
		repo := new(mockDocumentRepository)
		uow := new(mockUnitOfWork)
		svc := newService(t, repo, uow)
		ctx := context.Background()
		input := validInput()

		dbErr := errors.New("connection refused")
		repo.On("FindByProjectAndHash", ctx, input.ProjectID, input.FileHash).
			Return(nil, dbErr)

		// Act
		doc, err := svc.UploadDocument(ctx, input)

		// Assert
		require.Error(t, err)
		assert.Nil(t, doc)
		assert.ErrorContains(t, err, "document service: upload")
		uow.AssertNotCalled(t, "Begin")
	})

	t.Run("GivenBeginTxFails WhenUpload ThenWrappedError", func(t *testing.T) {
		// Given the database transaction cannot be started
		// When the service tries to upload a document
		// Then the error is wrapped and no write is attempted

		// Arrange
		repo := new(mockDocumentRepository)
		uow := new(mockUnitOfWork)
		svc := newService(t, repo, uow)
		ctx := context.Background()
		input := validInput()

		repo.On("FindByProjectAndHash", ctx, input.ProjectID, input.FileHash).
			Return(nil, repositories.ErrDocumentNotFound)
		uow.On("Begin", ctx).Return((*sql.Tx)(nil), errors.New("db unavailable"))

		// Act
		doc, err := svc.UploadDocument(ctx, input)

		// Assert
		require.Error(t, err)
		assert.Nil(t, doc)
		assert.ErrorContains(t, err, "begin tx")
		repo.AssertNotCalled(t, "Create")
	})

	t.Run("GivenCreateFails WhenUpload ThenRollbackAndError", func(t *testing.T) {
		// Given the repository Create() call fails
		// When the service uploads a document
		// Then the transaction is rolled back and an error is returned

		// Arrange
		repo := new(mockDocumentRepository)
		uow := new(mockUnitOfWork)
		svc := newService(t, repo, uow)
		ctx := context.Background()
		input := validInput()

		dbErr := errors.New("insert error")
		repo.On("FindByProjectAndHash", ctx, input.ProjectID, input.FileHash).
			Return(nil, repositories.ErrDocumentNotFound)
		uow.On("Begin", ctx).Return((*sql.Tx)(nil), nil)
		repo.On("Create", ctx, (*sql.Tx)(nil), mock.AnythingOfType("*filesv1.Document")).
			Return(nil, dbErr)
		uow.On("Rollback", (*sql.Tx)(nil)).Return(nil)

		// Act
		doc, err := svc.UploadDocument(ctx, input)

		// Assert
		require.Error(t, err)
		assert.Nil(t, doc)
		uow.AssertCalled(t, "Rollback", (*sql.Tx)(nil))
		uow.AssertNotCalled(t, "Commit")
	})
}

// ─── ClassifyDocument ─────────────────────────────────────────────────────────

func TestDocumentService_ClassifyDocument(t *testing.T) {
	t.Run("GivenExistingDocument WhenClassifyAsBill ThenKindUpdated", func(t *testing.T) {
		// Given a document exists in the project
		// When the user classifies it as a bill
		// Then the document kind is updated and the updated document is returned

		// Arrange
		repo := new(mockDocumentRepository)
		uow := new(mockUnitOfWork)
		svc := newService(t, repo, uow)
		ctx := context.Background()

		updated := &filesv1.Document{
			Id:        "doc-1",
			ProjectId: "project-1",
			Kind:      filesv1.DocumentKind_DOCUMENT_KIND_BILL,
		}
		uow.On("Begin", ctx).Return((*sql.Tx)(nil), nil)
		repo.On("UpdateKind", ctx, (*sql.Tx)(nil), "project-1", "doc-1", filesv1.DocumentKind_DOCUMENT_KIND_BILL).
			Return(updated, nil)
		uow.On("Commit", (*sql.Tx)(nil)).Return(nil)
		uow.On("Rollback", (*sql.Tx)(nil)).Return(nil)

		// Act
		doc, err := svc.ClassifyDocument(ctx, "project-1", "doc-1", filesv1.DocumentKind_DOCUMENT_KIND_BILL)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, filesv1.DocumentKind_DOCUMENT_KIND_BILL, doc.Kind)
		repo.AssertExpectations(t)
	})

	t.Run("GivenMissingDocument WhenClassify ThenErrDocumentNotFound", func(t *testing.T) {
		// Given no document exists with the given ID in the project
		// When the service tries to classify it
		// Then ErrDocumentNotFound is returned

		// Arrange
		repo := new(mockDocumentRepository)
		uow := new(mockUnitOfWork)
		svc := newService(t, repo, uow)
		ctx := context.Background()

		uow.On("Begin", ctx).Return((*sql.Tx)(nil), nil)
		repo.On("UpdateKind", ctx, (*sql.Tx)(nil), "project-1", "missing-doc", filesv1.DocumentKind_DOCUMENT_KIND_BILL).
			Return(nil, repositories.ErrDocumentNotFound)
		uow.On("Rollback", (*sql.Tx)(nil)).Return(nil)

		// Act
		doc, err := svc.ClassifyDocument(ctx, "project-1", "missing-doc", filesv1.DocumentKind_DOCUMENT_KIND_BILL)

		// Assert
		require.Error(t, err)
		assert.Nil(t, doc)
		assert.True(t, errors.Is(err, repositories.ErrDocumentNotFound))
	})
}

// ─── GetDocument ──────────────────────────────────────────────────────────────

func TestDocumentService_GetDocument(t *testing.T) {
	t.Run("GivenExistingDocument WhenGet ThenReturnsDocument", func(t *testing.T) {
		// Given a document exists in the project scope
		// When the service is asked for it by ID
		// Then the document is returned

		// Arrange
		repo := new(mockDocumentRepository)
		uow := new(mockUnitOfWork)
		svc := newService(t, repo, uow)
		ctx := context.Background()

		expected := &filesv1.Document{Id: "doc-1", ProjectId: "project-1"}
		repo.On("FindByProjectAndID", ctx, "project-1", "doc-1").Return(expected, nil)

		// Act
		doc, err := svc.GetDocument(ctx, "project-1", "doc-1")

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "doc-1", doc.Id)
	})

	t.Run("GivenMissingDocument WhenGet ThenErrDocumentNotFound", func(t *testing.T) {
		// Given no document with the given ID exists in the scope
		// When the service tries to retrieve it
		// Then ErrDocumentNotFound is returned

		// Arrange
		repo := new(mockDocumentRepository)
		uow := new(mockUnitOfWork)
		svc := newService(t, repo, uow)
		ctx := context.Background()

		repo.On("FindByProjectAndID", ctx, "project-1", "missing-doc").
			Return(nil, repositories.ErrDocumentNotFound)

		// Act
		doc, err := svc.GetDocument(ctx, "project-1", "missing-doc")

		// Assert
		require.Error(t, err)
		assert.Nil(t, doc)
		assert.True(t, errors.Is(err, repositories.ErrDocumentNotFound))
	})
}

// ─── ListDocuments ────────────────────────────────────────────────────────────

func TestDocumentService_ListDocuments(t *testing.T) {
	t.Run("GivenProjectDocuments WhenList ThenReturnsAll", func(t *testing.T) {
		// Given a project has multiple documents
		// When the service lists them with no page token
		// Then all documents in the project are returned

		// Arrange
		repo := new(mockDocumentRepository)
		uow := new(mockUnitOfWork)
		svc := newService(t, repo, uow)
		ctx := context.Background()

		docs := []*filesv1.Document{
			{Id: "doc-1", ProjectId: "project-1"},
			{Id: "doc-2", ProjectId: "project-1"},
		}
		repo.On("ListByProject", ctx, "project-1", int32(10), "").Return(docs, nil)

		// Act
		result, err := svc.ListDocuments(ctx, "project-1", 10, "")

		// Assert
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("GivenEmptyProject WhenList ThenReturnsEmptySlice", func(t *testing.T) {
		// Given a project has no documents
		// When the service lists documents
		// Then an empty slice (not nil) is returned without error

		// Arrange
		repo := new(mockDocumentRepository)
		uow := new(mockUnitOfWork)
		svc := newService(t, repo, uow)
		ctx := context.Background()

		repo.On("ListByProject", ctx, "project-1", int32(10), "").
			Return([]*filesv1.Document{}, nil)

		// Act
		result, err := svc.ListDocuments(ctx, "project-1", 10, "")

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("GivenRepositoryError WhenList ThenWrappedError", func(t *testing.T) {
		// Given the repository returns an unexpected error
		// When the service tries to list documents
		// Then the error is wrapped and returned

		// Arrange
		repo := new(mockDocumentRepository)
		uow := new(mockUnitOfWork)
		svc := newService(t, repo, uow)
		ctx := context.Background()

		repo.On("ListByProject", ctx, "project-1", int32(10), "").
			Return(nil, errors.New("db error"))

		// Act
		result, err := svc.ListDocuments(ctx, "project-1", 10, "")

		// Assert
		require.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorContains(t, err, "document service: list")
	})
}
