package services_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"

	bffinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/interfaces"
	"github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services"
	filesv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/files/v1"
)

// ─── mock: FilesClient ────────────────────────────────────────────────────────

type mockFilesClient struct{ mock.Mock }

func (m *mockFilesClient) UploadDocument(ctx context.Context, in *filesv1.UploadDocumentRequest, opts ...grpc.CallOption) (*filesv1.UploadDocumentResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filesv1.UploadDocumentResponse), args.Error(1)
}

func (m *mockFilesClient) ClassifyDocument(ctx context.Context, in *filesv1.ClassifyDocumentRequest, opts ...grpc.CallOption) (*filesv1.ClassifyDocumentResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filesv1.ClassifyDocumentResponse), args.Error(1)
}

func (m *mockFilesClient) GetDocument(ctx context.Context, in *filesv1.GetDocumentRequest, opts ...grpc.CallOption) (*filesv1.GetDocumentResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filesv1.GetDocumentResponse), args.Error(1)
}

func (m *mockFilesClient) ListDocuments(ctx context.Context, in *filesv1.ListDocumentsRequest, opts ...grpc.CallOption) (*filesv1.ListDocumentsResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filesv1.ListDocumentsResponse), args.Error(1)
}

func (m *mockFilesClient) ListBankAccounts(ctx context.Context, in *filesv1.ListBankAccountsRequest, opts ...grpc.CallOption) (*filesv1.ListBankAccountsResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filesv1.ListBankAccountsResponse), args.Error(1)
}

func (m *mockFilesClient) CreateBankAccount(ctx context.Context, in *filesv1.CreateBankAccountRequest, opts ...grpc.CallOption) (*filesv1.CreateBankAccountResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filesv1.CreateBankAccountResponse), args.Error(1)
}

func (m *mockFilesClient) DeleteBankAccount(ctx context.Context, in *filesv1.DeleteBankAccountRequest, opts ...grpc.CallOption) (*filesv1.DeleteBankAccountResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filesv1.DeleteBankAccountResponse), args.Error(1)
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func newDocumentsService(t *testing.T, client bffinterfaces.FilesClient) bffinterfaces.DocumentsService {
	t.Helper()
	return services.NewDocumentsService(zaptest.NewLogger(t), client)
}

// ─── UploadDocument ───────────────────────────────────────────────────────────

func TestDocumentsService_UploadDocument_Success(t *testing.T) {
	// Arrange
	client := &mockFilesClient{}
	svc := newDocumentsService(t, client)
	ctx := context.Background()

	client.On("UploadDocument", ctx, mock.AnythingOfType("*filesv1.UploadDocumentRequest")).Return(
		&filesv1.UploadDocumentResponse{
			Document: &filesv1.Document{
				Id:        "doc-1",
				ProjectId: "proj-1",
				Kind:      filesv1.DocumentKind_DOCUMENT_KIND_BILL,
				FileName:  "bill.pdf",
			},
		}, nil)

	// Act
	result, err := svc.UploadDocument(ctx, "proj-1", "user-1", "bill.pdf", []byte("content"))

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "doc-1", result.ID)
	assert.Equal(t, "bill", result.Kind)
	client.AssertExpectations(t)
}

func TestDocumentsService_UploadDocument_ClientError(t *testing.T) {
	// Arrange
	client := &mockFilesClient{}
	svc := newDocumentsService(t, client)
	ctx := context.Background()
	wantErr := errors.New("grpc unavailable")

	client.On("UploadDocument", ctx, mock.Anything).Return(nil, wantErr)

	// Act
	result, err := svc.UploadDocument(ctx, "proj-1", "user-1", "bill.pdf", []byte("x"))

	// Assert
	assert.Nil(t, result)
	assert.ErrorIs(t, err, wantErr)
}

// ─── ClassifyDocument ─────────────────────────────────────────────────────────

func TestDocumentsService_ClassifyDocument_Success(t *testing.T) {
	// Arrange
	client := &mockFilesClient{}
	svc := newDocumentsService(t, client)
	ctx := context.Background()

	client.On("ClassifyDocument", ctx, mock.AnythingOfType("*filesv1.ClassifyDocumentRequest")).Return(
		&filesv1.ClassifyDocumentResponse{
			Document: &filesv1.Document{
				Id:   "doc-1",
				Kind: filesv1.DocumentKind_DOCUMENT_KIND_STATEMENT,
			},
		}, nil)

	// Act
	result, err := svc.ClassifyDocument(ctx, "proj-1", "doc-1", "statement")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "statement", result.Kind)
}

// ─── ListDocuments ────────────────────────────────────────────────────────────

func TestDocumentsService_ListDocuments_ReturnsPage(t *testing.T) {
	// Arrange
	client := &mockFilesClient{}
	svc := newDocumentsService(t, client)
	ctx := context.Background()

	client.On("ListDocuments", ctx, mock.AnythingOfType("*filesv1.ListDocumentsRequest")).Return(
		&filesv1.ListDocumentsResponse{
			Documents: []*filesv1.Document{
				{Id: "d1", ProjectId: "proj-1"},
				{Id: "d2", ProjectId: "proj-1"},
			},
		}, nil)

	// Act
	result, err := svc.ListDocuments(ctx, "proj-1", 10, "")

	// Assert
	require.NoError(t, err)
	assert.Len(t, result.Items, 2)
}

func TestDocumentsService_ListDocuments_DefaultsPageSize(t *testing.T) {
	// Arrange
	client := &mockFilesClient{}
	svc := newDocumentsService(t, client)
	ctx := context.Background()

	var capturedReq *filesv1.ListDocumentsRequest
	client.On("ListDocuments", ctx, mock.MatchedBy(func(req *filesv1.ListDocumentsRequest) bool {
		capturedReq = req
		return true
	})).Return(&filesv1.ListDocumentsResponse{}, nil)

	// Act
	_, err := svc.ListDocuments(ctx, "proj-1", 0, "")

	// Assert
	require.NoError(t, err)
	assert.EqualValues(t, 25, capturedReq.Pagination.PageSize)
}

// ─── GetDocument ──────────────────────────────────────────────────────────────

func TestDocumentsService_GetDocument_Success(t *testing.T) {
	// Arrange
	client := &mockFilesClient{}
	svc := newDocumentsService(t, client)
	ctx := context.Background()

	client.On("GetDocument", ctx, mock.AnythingOfType("*filesv1.GetDocumentRequest")).Return(
		&filesv1.GetDocumentResponse{
			Document: &filesv1.Document{Id: "doc-1"},
		}, nil)

	// Act
	result, err := svc.GetDocument(ctx, "proj-1", "doc-1")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "doc-1", result.ID)
}

func TestDocumentsHistoryServiceBoundaryContracts(t *testing.T) {
	t.Parallel()

	t.Run("GivenDocumentsServiceWhenBoundaryImportsAreCheckedThenTransportViewsAreNotImported", func(t *testing.T) {
		// Given
		servicePath := "documents_service.go"

		// Arrange
		content, err := os.ReadFile(servicePath)
		require.NoError(t, err)
		text := string(content)

		// Act
		hasViewsImport := strings.Contains(text, "transport/http/views")
		hasContractsImport := strings.Contains(text, "services/contracts")

		// Then
		assert.False(t, hasViewsImport)
		assert.True(t, hasContractsImport)
	})

	t.Run("GivenHistoryServiceWhenBoundaryImportsAreCheckedThenTransportViewsAreNotImported", func(t *testing.T) {
		// Given
		servicePath := "history_service.go"

		// Arrange
		content, err := os.ReadFile(servicePath)
		require.NoError(t, err)
		text := string(content)

		// Act
		hasViewsImport := strings.Contains(text, "transport/http/views")
		hasContractsImport := strings.Contains(text, "services/contracts")

		// Then
		assert.False(t, hasViewsImport)
		assert.True(t, hasContractsImport)
	})
}
