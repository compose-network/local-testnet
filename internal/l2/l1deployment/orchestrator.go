package l1deployment

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/compose-network/local-testnet/configs"
	"github.com/compose-network/local-testnet/internal/l2/domain"
	"github.com/compose-network/local-testnet/internal/l2/infra/docker"
	"github.com/compose-network/local-testnet/internal/l2/infra/filesystem/json"
	"github.com/compose-network/local-testnet/internal/l2/l1deployment/deployer"
	"github.com/compose-network/local-testnet/internal/l2/l1deployment/dispute"
	"github.com/compose-network/local-testnet/internal/l2/l2config/crypto"
	"github.com/compose-network/local-testnet/internal/logger"
)

/*
Orchestrator coordinates Phase 1: L1 deployment
  - Initializes op-deployer state and writes intent.toml
  - Deploys OP Stack L1 contracts to the L1 chain
  - Outputs state.json with contract addresses
*/
type Orchestrator struct {
	rootDir  string
	stateDir string
	logger   *slog.Logger
}

// NewOrchestrator creates a new Phase 1 orchestrator
func NewOrchestrator(rootDir, stateDir string) *Orchestrator {
	return &Orchestrator{
		rootDir:  rootDir,
		stateDir: stateDir,
		logger:   logger.Named("l1_orchestrator"),
	}
}

// Execute runs Phase 1: Deploy L1 contracts and dispute contracts
// Returns deployment state and DisputeGameFactory proxy address
func (o *Orchestrator) Execute(ctx context.Context, cfg configs.L2) (*domain.DeploymentState, string, error) {
	o.logger.Info("Phase 1: Starting L1 deployment")

	stateManager := deployer.NewStateManager(o.stateDir, json.NewReader())

	o.logger.Info("ensuring state directory created")
	if err := stateManager.EnsureStateDir(); err != nil {
		return nil, "", fmt.Errorf("failed to ensure state directory: %w", err)
	}

	o.logger.Info("instantiating Docker client")
	dockerClient, err := docker.New()
	if err != nil {
		return nil, "", fmt.Errorf("failed to create docker client: %w", err)
	}
	defer dockerClient.Close()

	o.logger.Info("instantiating Deployer")
	opDeployer := deployer.NewDeployer(o.rootDir, o.stateDir, cfg.OPDeployerVersion, dockerClient)

	o.logger.Info("initializing Deployer")
	if err := opDeployer.Init(ctx, cfg.L1ChainID, cfg.ChainConfigs); err != nil {
		return nil, "", fmt.Errorf("failed to initialize op-deployer state: %w", err)
	}

	o.logger.Info("converting coordinator PK to address")
	coordinatorAddress, err := crypto.AddressFromPrivateKey(cfg.CoordinatorPrivateKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to derive coordinator address: %w", err)
	}

	o.logger.Info("generating intent file")
	intentWriter := deployer.NewIntentWriter(o.stateDir, json.NewWriter())
	if err := intentWriter.WriteIntent(
		cfg.Wallet.Address,
		coordinatorAddress,
		cfg.L1ChainID,
		cfg.ChainConfigs,
	); err != nil {
		return nil, "", fmt.Errorf("failed to write intent: %w", err)
	}

	if err := opDeployer.Apply(ctx, cfg.L1ElURL, cfg.Wallet.PrivateKey, cfg.DeploymentTarget); err != nil {
		return nil, "", fmt.Errorf("failed to deploy L1 contracts: %w", err)
	}

	deployment, err := stateManager.Load()
	if err != nil {
		return nil, "", fmt.Errorf("failed to load deployment state: %w", err)
	}

	o.logger.Info("deploying dispute contracts")
	disputeService := dispute.NewService(o.rootDir, cfg)
	gameFactoryAddr, err := disputeService.Deploy(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to deploy dispute contracts: %w", err)
	}

	o.logger.With("game_factory_address", gameFactoryAddr).Info("Phase 1: L1 deployment completed successfully")

	return deployment, gameFactoryAddr, nil
}
