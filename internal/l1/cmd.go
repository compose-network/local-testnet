package l1

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

var CMD = &cobra.Command{
	Use:   "l1",
	Short: "Commands for running L1 network",
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.Info("starting l1")
		err := start(cmd.Context())
		if err != nil {
			return fmt.Errorf("error occurred starting l1: %w", err)
		}

		slog.Info("l1 started.")

		return nil
	},
}
