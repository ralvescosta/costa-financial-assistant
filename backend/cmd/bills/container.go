package bills

import (
	"context"
	"fmt"

	pkgotel "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/otel"
)

// run wires the dependency container and starts the bills gRPC service.
func run(_ context.Context) error {
	if err := pkgotel.RegisterServiceMetrics("bills"); err != nil {
		return fmt.Errorf("bills: register metrics: %w", err)
	}
	return nil
}
