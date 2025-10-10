package observability

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

var CMD = &cobra.Command{
	Use:   "observability",
	Short: "Command for running observability services",
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.Info("Starting observability services")

		if err := start(cmd.Context()); err != nil {
			return fmt.Errorf("error occurred starting observability services: %w", err)
		}

		return nil
	},
}
