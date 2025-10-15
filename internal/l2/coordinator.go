package l2

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/compose-network/localnet-control-plane/configs"
	"github.com/compose-network/localnet-control-plane/internal/l2/infra/git"
	"github.com/compose-network/localnet-control-plane/internal/l2/l1deployment"
	"github.com/compose-network/localnet-control-plane/internal/l2/l2config"
	"github.com/compose-network/localnet-control-plane/internal/l2/l2runtime"
	"github.com/compose-network/localnet-control-plane/internal/logger"
)

// Coordinator orchestrates the entire L2 deployment process
type Coordinator struct {
	rootDir     string
	stateDir    string // For L1 deployment state
	networksDir string // For L2 chain configs
	logger      *slog.Logger
}

// NewCoordinator creates a new coordinator
func NewCoordinator(rootDir string) *Coordinator {
	return &Coordinator{
		rootDir:     rootDir,
		stateDir:    filepath.Join(rootDir, "internal", "l2", "state"),
		networksDir: filepath.Join(rootDir, "internal", "l2", "networks"),
		logger:      logger.Named("l2_coordinator"),
	}
}

func (c *Coordinator) Deploy(ctx context.Context, cfg configs.L2) error {
	c.logger.Info("starting L2 deployment process")

	if err := c.cloneRepositories(ctx, cfg); err != nil {
		return fmt.Errorf("failed to clone repositories: %w", err)
	}

	l1Orchestrator := l1deployment.NewOrchestrator(c.rootDir, c.stateDir)
	deployment, err := l1Orchestrator.Execute(ctx, cfg)
	if err != nil {
		return fmt.Errorf("phase 1 failed: %w", err)
	}

	l2ConfigOrchestrator := l2config.NewOrchestrator(c.rootDir, c.stateDir, c.networksDir)
	l2Config, err := l2ConfigOrchestrator.Execute(ctx, cfg, deployment)
	if err != nil {
		return fmt.Errorf("phase 2 failed: %w", err)
	}

	l2Orchestrator := l2runtime.NewOrchestrator(c.rootDir, c.networksDir)
	if err := l2Orchestrator.Execute(ctx, cfg, l2Config); err != nil {
		return fmt.Errorf("phase 3 failed: %w", err)
	}

	c.logger.Info("L2 deployment completed successfully")

	return nil
}

// cloneRepositories clones all required git repositories
func (c *Coordinator) cloneRepositories(ctx context.Context, cfg configs.L2) error {
	c.logger.Info("cloning required repositories")

	cloner := git.NewCloner()
	repos := make([]git.Repository, 0, len(cfg.Repositories))

	for name, repo := range cfg.Repositories {
		repos = append(repos, git.Repository{
			Name: string(name),
			URL:  repo.URL,
			Ref:  repo.Branch,
		})
	}

	l2Dir := filepath.Join(c.rootDir, "internal", "l2", "services")
	if err := cloner.CloneAll(ctx, l2Dir, repos); err != nil {
		return fmt.Errorf("failed to clone repositories: %w", err)
	}

	c.logger.Info("repositories cloned successfully")

	return nil
}
