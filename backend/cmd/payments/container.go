package payments

import (
	"context"
	"fmt"

	pkgotel "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/otel"
)

// run wires the dependency container and starts the payments gRPC service.
func run(_ context.Context) error {
	if err := pkgotel.RegisterServiceMetrics("payments"); err != nil {
		return fmt.Errorf("payments: register metrics: %w", err)
	}
	return nil
}
