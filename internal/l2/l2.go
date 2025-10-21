package l2

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/compose-network/local-testnet/configs"
	"github.com/spf13/cobra"
)

func init() {
	CMD.AddCommand(compileCmd)
}

var CMD = &cobra.Command{
	Use:   "l2",
	Short: "Commands for running L2 network",
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.Info("starting l2 command. Validating config", slog.Any("config", configs.Values.L2))

		if err := configs.Values.L2.Validate(); err != nil {
			return err
		}

		slog.Info("config validation successful. Starting l2 deployment...")

		rootDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}

		service := NewService(rootDir)
		if err := service.Deploy(cmd.Context(), configs.Values.L2); err != nil {
			return fmt.Errorf("l2 deployment failed: %w", err)
		}

		slog.Info("l2 deployment completed successfully")

		return nil
	},
}
