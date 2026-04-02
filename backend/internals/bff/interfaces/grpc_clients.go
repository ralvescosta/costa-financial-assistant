// Package interfaces defines the BFF-facing gRPC client interfaces for downstream service consumers.
// These narrow interfaces isolate the BFF from specific gRPC implementations and enable mocking in tests.
package interfaces

import (
	"context"

	"google.golang.org/grpc"

	filesv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/files/v1"
	identityv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/identity/v1"
	onboardingv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/onboarding/v1"
)

// FilesClient defines the subset of the files gRPC service consumed by BFF services.
// The generated filesv1.FilesServiceClient satisfies this interface.
type FilesClient interface {
	UploadDocument(ctx context.Context, in *filesv1.UploadDocumentRequest, opts ...grpc.CallOption) (*filesv1.UploadDocumentResponse, error)
	ClassifyDocument(ctx context.Context, in *filesv1.ClassifyDocumentRequest, opts ...grpc.CallOption) (*filesv1.ClassifyDocumentResponse, error)
	GetDocument(ctx context.Context, in *filesv1.GetDocumentRequest, opts ...grpc.CallOption) (*filesv1.GetDocumentResponse, error)
	ListDocuments(ctx context.Context, in *filesv1.ListDocumentsRequest, opts ...grpc.CallOption) (*filesv1.ListDocumentsResponse, error)
	ListBankAccounts(ctx context.Context, in *filesv1.ListBankAccountsRequest, opts ...grpc.CallOption) (*filesv1.ListBankAccountsResponse, error)
	CreateBankAccount(ctx context.Context, in *filesv1.CreateBankAccountRequest, opts ...grpc.CallOption) (*filesv1.CreateBankAccountResponse, error)
	DeleteBankAccount(ctx context.Context, in *filesv1.DeleteBankAccountRequest, opts ...grpc.CallOption) (*filesv1.DeleteBankAccountResponse, error)
}

// IdentityClient defines the subset of the identity gRPC service consumed by BFF middleware.
// The generated identityv1.IdentityServiceClient satisfies this interface.
type IdentityClient interface {
	ValidateToken(ctx context.Context, in *identityv1.ValidateTokenRequest, opts ...grpc.CallOption) (*identityv1.ValidateTokenResponse, error)
	GetJwksMetadata(ctx context.Context, in *identityv1.GetJwksMetadataRequest, opts ...grpc.CallOption) (*identityv1.GetJwksMetadataResponse, error)
}

// OnboardingClient defines the subset of the onboarding gRPC service consumed by BFF collaboration controllers.
// The generated onboardingv1.OnboardingServiceClient satisfies this interface.
type OnboardingClient interface {
	CreateProject(ctx context.Context, in *onboardingv1.CreateProjectRequest, opts ...grpc.CallOption) (*onboardingv1.CreateProjectResponse, error)
	InviteProjectMember(ctx context.Context, in *onboardingv1.InviteProjectMemberRequest, opts ...grpc.CallOption) (*onboardingv1.InviteProjectMemberResponse, error)
	UpdateProjectMemberRole(ctx context.Context, in *onboardingv1.UpdateProjectMemberRoleRequest, opts ...grpc.CallOption) (*onboardingv1.UpdateProjectMemberRoleResponse, error)
	ListProjectMembers(ctx context.Context, in *onboardingv1.ListProjectMembersRequest, opts ...grpc.CallOption) (*onboardingv1.ListProjectMembersResponse, error)
	GetProject(ctx context.Context, in *onboardingv1.GetProjectRequest, opts ...grpc.CallOption) (*onboardingv1.GetProjectResponse, error)
}
