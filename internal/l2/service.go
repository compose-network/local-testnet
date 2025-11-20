package l2

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"

	"github.com/compose-network/local-testnet/configs"
	"github.com/compose-network/local-testnet/internal/l2/blockscout"
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
	blockscoutService interface {
		Run(context.Context, []blockscout.ChainConfig) error
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
		blockscoutService     blockscoutService
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
	blockscoutService blockscoutService,
	outputGenerator outputGenerator) *Service {
	return &Service{
		rootDir:               rootDir,
		cloner:                cloner,
		l1Orchestrator:        l1Orchestrator,
		l2ConfigOrchestrator:  l2ConfigOrchestrator,
		l2RuntimeOrchestrator: l2RuntimeOrchestrator,
		blockscoutService:     blockscoutService,
		outputGenerator:       outputGenerator,
		logger:                logger.Named("l2_service"),
	}
}

func (s *Service) Deploy(ctx context.Context, cfg configs.L2) error {
	s.logger.Info("starting L2 deployment process")

	if err := s.cloneRepositories(ctx, cfg); err != nil {
		return fmt.Errorf("failed to clone repositories: %w", err)
	}

	s.logger.Info("running phase 1 - L1 deployments")
	deploymentState, err := s.l1Orchestrator.Execute(ctx, cfg)
	if err != nil {
		return fmt.Errorf("phase 1 failed: %w", err)
	}

	s.logger.Info("running phase 2 - L2 config generation", "deployment_state", deploymentState)
	err = s.l2ConfigOrchestrator.Execute(ctx, cfg, deploymentState)
	if err != nil {
		return fmt.Errorf("phase 2 failed: %w", err)
	}

	s.logger.Info("running phase 3 - L2 launch")
	deployedContracts, err := s.l2RuntimeOrchestrator.Execute(ctx, cfg, deploymentState.DisputeGameFactoryAddress)
	if err != nil {
		return fmt.Errorf("phase 3 failed: %w", err)
	}

	s.logger.Info("restarting op-geth services to apply mailbox configuration")
	if err := s.restartOpGeth(ctx); err != nil {
		const msg = "failed to restart op-geth services"
		s.logger.Error(msg, "error", err)
		return fmt.Errorf("%s: %w", msg, err)
	}

	s.logger.Info("L2 deployment completed successfully. Generating output file")

	var chainConfigs []blockscout.ChainConfig
	for chainName, config := range cfg.ChainConfigs {
		var hostName string
		switch chainName {
		case configs.L2ChainNameRollupA:
			hostName = "op-geth-a"
		case configs.L2ChainNameRollupB:
			hostName = "op-geth-b"
		default:
			return fmt.Errorf("unknown chain name: %s", chainName)
		}

		chainConfigs = append(chainConfigs, blockscout.ChainConfig{
			ID:         config.ID,
			ELHostName: hostName,
			RPCPort:    config.RPCPort,
			WSPort:     8546,
		})
	}
	if err := s.blockscoutService.Run(ctx, chainConfigs); err != nil {
		return fmt.Errorf("failed to start Blockscout service: %w", err)
	}

	s.logger.Info("L2 deployment completed successfully. Generating output file")

	if err := s.outputGenerator.Generate(ctx, deployedContracts); err != nil {
		return fmt.Errorf("failed to generate output file: %w", err)
	}

	s.logger.Info("output file generated successfully")

	return nil
}

// restartOpGeth restarts op-geth services to pick up new mailbox configuration
func (s *Service) restartOpGeth(ctx context.Context) error {
	localnetDir := filepath.Join(s.rootDir, localnetDirName)
	composeFile := filepath.Join(localnetDir, "docker-compose.yml")

	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", composeFile, "restart", "op-geth-a", "op-geth-b")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("docker compose restart failed: %w, output: %s", err, string(output))
	}

	s.logger.Info("op-geth services restarted successfully, waiting for them to be ready")

	return nil
}

// cloneRepositories clones all required git repositories
func (s *Service) cloneRepositories(ctx context.Context, cfg configs.L2) error {
	s.logger.Info("cloning required repositories")

	repos := make([]git.Repository, 0, len(cfg.Repositories))

	for name, repo := range cfg.Repositories {
		if repo.URL != "" {
			repos = append(repos, git.Repository{
				Name: string(name),
				URL:  repo.URL,
				Ref:  repo.Branch,
			})
			continue
		}

		if repo.LocalPath != "" {
			absPath, err := filepath.Abs(repo.LocalPath)
			if err != nil {
				return fmt.Errorf("failed to resolve absolute path for local repository %s: %w", name, err)
			}
			s.logger.With("name", name, "local_path", repo.LocalPath, "resolved_path", absPath).Info("using local repository path; skipping clone")
			continue
		}

		return fmt.Errorf("repository %s has neither URL nor local-path set", name)
	}

	l2Dir := filepath.Join(s.rootDir, localnetDirName, servicesDirName)
	if err := s.cloner.CloneAll(ctx, l2Dir, repos); err != nil {
		return fmt.Errorf("failed to clone repositories: %w", err)
	}

	s.logger.Info("repositories cloned successfully")

	return nil
}
