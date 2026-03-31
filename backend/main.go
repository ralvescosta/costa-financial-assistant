package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/ralvescosta/costa-financial-assistant/backend/cmd/bff"
	"github.com/ralvescosta/costa-financial-assistant/backend/cmd/bills"
	"github.com/ralvescosta/costa-financial-assistant/backend/cmd/files"
	"github.com/ralvescosta/costa-financial-assistant/backend/cmd/identity"
	"github.com/ralvescosta/costa-financial-assistant/backend/cmd/migrations"
	"github.com/ralvescosta/costa-financial-assistant/backend/cmd/onboarding"
	"github.com/ralvescosta/costa-financial-assistant/backend/cmd/payments"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	root := &cobra.Command{
		Use:   "financial-assistant",
		Short: "Costa Financial Assistant — multi-service monorepo entry point",
	}

	root.AddCommand(
		bff.NewCommand(),
		bills.NewCommand(),
		files.NewCommand(),
		identity.NewCommand(),
		migrations.NewCommand(),
		onboarding.NewCommand(),
		payments.NewCommand(),
	)

	if err := root.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
