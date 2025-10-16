package l2

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/compose-network/localnet-control-plane/configs"
	"github.com/compose-network/localnet-control-plane/internal/l2/infra/git"
	"github.com/compose-network/localnet-control-plane/internal/l2/l2runtime/contracts"
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

		coordinator := NewCoordinator(rootDir)
		if err := coordinator.Deploy(cmd.Context(), configs.Values.L2); err != nil {
			return fmt.Errorf("l2 deployment failed: %w", err)
		}

		slog.Info("l2 deployment completed successfully")

		return nil
	},
}

var compileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Compile L2 contracts from publisher repository",
	Long:  "Compiles Solidity contracts from the publisher repository and generates contracts.json with ABIs and bytecodes",
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.Info("running contract compilation command")
		ctx := cmd.Context()
		rootDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}

		slog.With("name", configs.RepositoryNamePublisher).Info("fetching repository settings from configuration before cloning")
		var (
			publisherRepo git.Repository
			found         bool
		)
		for name, repo := range configs.Values.L2.Repositories {
			if name == configs.RepositoryNamePublisher {
				publisherRepo = git.Repository{
					Name: string(name),
					URL:  repo.URL,
					Ref:  repo.Branch,
				}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("could not find: '%s' repository in the configuration", configs.RepositoryNamePublisher)
		}

		slog.Info("cloning repositories")
		cloner := git.NewCloner()
		if err := cloner.Clone(ctx, filepath.Join(rootDir, "internal", "l2", "services"), publisherRepo); err != nil {
			return fmt.Errorf("failed to clone repository: '%w'", err)
		}

		publisherContractsDir := filepath.Join(rootDir, "internal", "l2", "services", "publisher", "contracts")
		compiler := contracts.NewCompiler(
			publisherContractsDir,
			filepath.Join(publisherContractsDir, "src"),
			filepath.Join(rootDir, "internal", "l2", "l2runtime", "contracts", "compiled"),
		)

		slog.Info("starting contract compilation")
		if err := compiler.Compile(ctx); err != nil {
			return fmt.Errorf("contract compilation failed: %w", err)
		}

		slog.Info("contract compilation completed successfully")

		return nil
	},
}
