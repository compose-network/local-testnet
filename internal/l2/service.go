package l2

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/compose-network/local-testnet/configs"
	"github.com/compose-network/local-testnet/internal/l2/infra/git"
	"github.com/compose-network/local-testnet/internal/l2/l1deployment"
	"github.com/compose-network/local-testnet/internal/l2/l2runtime/contracts"
	"github.com/compose-network/local-testnet/internal/logger"
	"github.com/ethereum/go-ethereum/common"
)

// Service orchestrates the entire L2 deployment process
type (
	cloner interface {
		CloneAll(ctx context.Context, baseDir string, repos []git.Repository) error
	}
	l1Orchestrator interface {
		Execute(ctx context.Context, cfg configs.L2) (l1deployment.DeploymentState, error)
	}
	l2ConfigOrchestrator interface {
		Execute(ctx context.Context, cfg configs.L2, state l1deployment.DeploymentState) error
	}
	l2RuntimeOrchestrator interface {
		Execute(ctx context.Context, cfg configs.L2, disputeGameFactory common.Address) (map[configs.L2ChainName]map[contracts.ContractName]common.Address, error)
	}
	outputGenerator interface {
		Generate(context.Context, map[configs.L2ChainName]map[contracts.ContractName]common.Address) error
	}

	Service struct {
		rootDir               string
		cloner                cloner
		l1Orchestrator        l1Orchestrator
		l2ConfigOrchestrator  l2ConfigOrchestrator
		l2RuntimeOrchestrator l2RuntimeOrchestrator
		outputGenerator       outputGenerator
		logger                *slog.Logger
	}
)

// NewService creates a new l2 service
func NewService(
	rootDir string,
	cloner cloner,
	l1Orchestrator l1Orchestrator,
	l2ConfigOrchestrator l2ConfigOrchestrator,
	l2RuntimeOrchestrator l2RuntimeOrchestrator,
	outputGenerator outputGenerator) *Service {
	return &Service{
		rootDir:               rootDir,
		cloner:                cloner,
		l1Orchestrator:        l1Orchestrator,
		l2ConfigOrchestrator:  l2ConfigOrchestrator,
		l2RuntimeOrchestrator: l2RuntimeOrchestrator,
		outputGenerator:       outputGenerator,
		logger:                logger.Named("l2_service"),
	}
}

func (c *Service) Deploy(ctx context.Context, cfg configs.L2) error {
	c.logger.Info("starting L2 deployment process")

	if err := c.cloneRepositories(ctx, cfg); err != nil {
		return fmt.Errorf("failed to clone repositories: %w", err)
	}

	c.logger.Info("running phase 1 - L1 deployments")
	deploymentState, err := c.l1Orchestrator.Execute(ctx, cfg)
	if err != nil {
		return fmt.Errorf("phase 1 failed: %w", err)
	}

	c.logger.Info("running phase 2 - L2 config generation")
	err = c.l2ConfigOrchestrator.Execute(ctx, cfg, deploymentState)
	if err != nil {
		return fmt.Errorf("phase 2 failed: %w", err)
	}

	c.logger.Info("running phase 3 - L2 launch")
	deployedContracts, err := c.l2RuntimeOrchestrator.Execute(ctx, cfg, deploymentState.DisputeGameFactoryAddress)
	if err != nil {
		return fmt.Errorf("phase 3 failed: %w", err)
	}

	c.logger.Info("L2 deployment completed successfully. Generating output file")

	if err := c.outputGenerator.Generate(ctx, deployedContracts); err != nil {
		return fmt.Errorf("failed to generate output file: %w", err)
	}

	c.logger.Info("output file generated successfully")

	return nil
}

// cloneRepositories clones all required git repositories
func (c *Service) cloneRepositories(ctx context.Context, cfg configs.L2) error {
	c.logger.Info("cloning required repositories")

	repos := make([]git.Repository, 0, len(cfg.Repositories))

	for name, repo := range cfg.Repositories {
		repos = append(repos, git.Repository{
			Name: string(name),
			URL:  repo.URL,
			Ref:  repo.Branch,
		})
	}

	l2Dir := filepath.Join(c.rootDir, "internal", "l2", "services")
	if err := c.cloner.CloneAll(ctx, l2Dir, repos); err != nil {
		return fmt.Errorf("failed to clone repositories: %w", err)
	}

	c.logger.Info("repositories cloned successfully")

	return nil
}
