package l2

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/compose-network/local-testnet/configs"
	"github.com/compose-network/local-testnet/internal/l2/infra/git"
	"github.com/compose-network/local-testnet/internal/l2/l1deployment"
	"github.com/compose-network/local-testnet/internal/l2/l2config"
	"github.com/compose-network/local-testnet/internal/l2/l2runtime"
	"github.com/compose-network/local-testnet/internal/l2/output"
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

		stateDir := filepath.Join(rootDir, ".localnet", "state")
		networksDir := filepath.Join(rootDir, ".localnet", "networks")
		compiledContractsDir := filepath.Join(rootDir, "internal", "l2", "l2runtime", "contracts", "compiled")

		l1Orchestrator := l1deployment.NewOrchestrator(rootDir, stateDir)
		l2ConfigOrchestrator := l2config.NewOrchestrator(rootDir, stateDir, networksDir)
		runtimeOrchestrator := l2runtime.NewOrchestrator(rootDir, networksDir)

		service := NewService(rootDir, git.NewCloner(), l1Orchestrator, l2ConfigOrchestrator, runtimeOrchestrator, output.NewGenerator(compiledContractsDir))
		if err := service.Deploy(cmd.Context(), configs.Values.L2); err != nil {
			return fmt.Errorf("l2 deployment failed: %w", err)
		}

		slog.Info("l2 deployment completed successfully")

		return nil
	},
}
