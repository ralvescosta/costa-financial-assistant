package onboarding

import (
	"context"
	"fmt"

	pkgotel "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/otel"
)

// run wires the dependency container and starts the onboarding gRPC service.
func run(_ context.Context) error {
	if err := pkgotel.RegisterServiceMetrics("onboarding"); err != nil {
		return fmt.Errorf("onboarding: register metrics: %w", err)
	}
	return nil
}
