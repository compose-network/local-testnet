package l2

import (
	"fmt"
	"log/slog"

	"github.com/compose-network/localnet-control-plane/configs"
	"github.com/spf13/cobra"
)

var CMD = &cobra.Command{
	Use:   "l2",
	Short: "Commands for running L2 network",
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.Info("starting l2 command. Validating config", slog.Any("config", configs.Values.L2))

		if err := configs.Values.L2.Validate(); err != nil {
			return err
		}

		slog.Info("config validation successful. Starting l2 services...")

		if err := start(cmd.Context()); err != nil {
			return fmt.Errorf("error occurred starting l2: %w", err)
		}

		slog.Info("l2 services started successfully")

		return nil
	},
}
