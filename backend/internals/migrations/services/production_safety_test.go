package services

import (
	"context"
	"os"
	"testing"
)

func TestValidateProductionAccess(t *testing.T) {
	t.Parallel()

	t.Run("GivenNonProduction_WhenValidated_ThenSucceeds", func(t *testing.T) {
		// Given
		_ = os.Unsetenv("APP_ENV")
		_ = os.Unsetenv("ENVIRONMENT")

		// When
		err := ValidateProductionAccess(context.Background(), "dev", false)

		// Then
		if err != nil {
			t.Fatalf("expected success, got error: %v", err)
		}
	})

	t.Run("GivenProductionWithoutApproval_WhenValidated_ThenFails", func(t *testing.T) {
		// Given
		if err := os.Setenv("APP_ENV", "prd"); err != nil {
			t.Fatalf("set APP_ENV: %v", err)
		}
		t.Cleanup(func() { _ = os.Unsetenv("APP_ENV") })

		// When
		err := ValidateProductionAccess(context.Background(), "prd", false)

		// Then
		if err == nil {
			t.Fatal("expected production approval error, got nil")
		}
	})
}
