package files

import (
	"context"
	"fmt"

	pkgotel "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/otel"
)

// run wires the dependency container and starts the files gRPC service.
func run(_ context.Context) error {
	if err := pkgotel.RegisterServiceMetrics("files"); err != nil {
		return fmt.Errorf("files: register metrics: %w", err)
	}
	return nil
}
