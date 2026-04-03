package services

import (
	"os"
	"strings"
	"testing"
)

func TestIdentityOnboardingPaymentsBoundaryPointerPolicy(t *testing.T) {
	t.Parallel()

	t.Run("GivenIdentityServiceWhenAuditedThenStructBoundariesUsePointers", func(t *testing.T) {
		// Given
		identityPath := "../../identity/services/token_service.go"

		// Arrange
		content, err := os.ReadFile(identityPath)
		if err != nil {
			t.Fatalf("failed to read identity token service: %v", err)
		}
		text := string(content)

		// When
		hasPointerClaims := strings.Contains(text, "ValidateToken(ctx context.Context, token string) (bool, *identityv1.JwtClaims, error)")
		hasPointerJWKS := strings.Contains(text, "GetJwksMetadata(ctx context.Context) (*identityv1.JwksMetadata, error)")

		// Then
		if !hasPointerClaims || !hasPointerJWKS {
			t.Fatalf("expected pointer-based identity service boundaries")
		}
	})

	t.Run("GivenOnboardingServiceWhenAuditedThenStructBoundariesUsePointers", func(t *testing.T) {
		// Given
		onboardingPath := "../../onboarding/services/project_members_service.go"

		// Arrange
		content, err := os.ReadFile(onboardingPath)
		if err != nil {
			t.Fatalf("failed to read onboarding service: %v", err)
		}
		text := string(content)

		// When
		hasPointerProjectReturn := strings.Contains(text, "CreateProject(ctx context.Context, ownerID, name string, projectType onboardingv1.ProjectType) (*onboardingv1.Project, error)")
		hasPointerMemberReturn := strings.Contains(text, "InviteProjectMember(ctx context.Context, projectID, inviteeEmail string, role onboardingv1.ProjectMemberRole, invitedBy string) (*onboardingv1.ProjectMember, error)")

		// Then
		if !hasPointerProjectReturn || !hasPointerMemberReturn {
			t.Fatalf("expected pointer-based onboarding service boundaries")
		}
	})

	t.Run("GivenPaymentsCycleServiceWhenAuditedThenStructBoundariesUsePointers", func(t *testing.T) {
		// Given
		paymentsPath := "payment_cycle_service.go"

		// Arrange
		content, err := os.ReadFile(paymentsPath)
		if err != nil {
			t.Fatalf("failed to read payments cycle service: %v", err)
		}
		text := string(content)

		// When
		hasPointerReturn := strings.Contains(text, "GetCyclePreference(ctx context.Context, projectID string) (*interfaces.CyclePreference, error)")

		// Then
		if !hasPointerReturn {
			t.Fatalf("expected pointer-based payment cycle service boundaries")
		}
	})
}
