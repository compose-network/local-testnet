package l2runtime

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/compose-network/local-testnet/configs"
	"github.com/compose-network/local-testnet/internal/l2/infra/docker"
	"github.com/compose-network/local-testnet/internal/l2/l2runtime/contracts"
	"github.com/compose-network/local-testnet/internal/l2/l2runtime/services"
	"github.com/compose-network/local-testnet/internal/logger"
	"github.com/ethereum/go-ethereum/common"
)

// Orchestrator coordinates Phase 3: L2 runtime operations
//   - Builds Docker images via docker-compose
//   - Starts initial services (publisher, op-geth)
//   - Deploys L2 helper contracts
//   - Restarts services to pick up contract addresses
//   - Starts final services (op-node, batcher, proposer)
type Orchestrator struct {
	rootDir     string
	networksDir string
	logger      *slog.Logger
}

// NewOrchestrator creates a new Phase 3 orchestrator
func NewOrchestrator(rootDir, networksDir string) *Orchestrator {
	return &Orchestrator{
		rootDir:     rootDir,
		networksDir: networksDir,
		logger:      logger.Named("l2_runtime_orchestrator"),
	}
}

// Execute runs Phase 3: Build images, start services, deploy contracts
func (o *Orchestrator) Execute(ctx context.Context, cfg configs.L2, gameFactoryAddr common.Address) error {
	o.logger.Info("Phase 3: Starting L2 runtime operations")

	env := o.buildDockerComposeEnv(cfg, gameFactoryAddr)

	o.logger.With("env", env).Info("environment variables were constructed. Building compose services")
	if err := o.buildComposeServices(ctx, env); err != nil {
		return fmt.Errorf("failed to build compose services: %w", err)
	}

	o.logger.Info("docker-compose services built successfully")
	serviceManager := services.NewManager(o.rootDir)
	if err := serviceManager.StartInitial(ctx, env); err != nil {
		return fmt.Errorf("failed to start initial services: %w", err)
	}

	contractDeployer := contracts.NewDeployer(o.rootDir, o.networksDir)
	if err := contractDeployer.Deploy(ctx, cfg.ChainConfigs, cfg.CoordinatorPrivateKey); err != nil {
		return fmt.Errorf("failed to deploy contracts: %w", err)
	}

	if err := serviceManager.RestartInitial(ctx, env); err != nil {
		return fmt.Errorf("failed to restart initial services: %w", err)
	}

	if err := serviceManager.StartFinal(ctx, env); err != nil {
		return fmt.Errorf("failed to start final services: %w", err)
	}

	o.logger.Info("Phase 3: L2 runtime operations completed successfully")

	return nil
}

// buildDockerComposeEnv creates environment variables for docker-compose
func (o *Orchestrator) buildDockerComposeEnv(cfg configs.L2, gameFactoryAddr common.Address) map[string]string {
	env := make(map[string]string)

	env["ROOT_DIR"] = o.rootDir
	env["WALLET_PRIVATE_KEY"] = cfg.Wallet.PrivateKey
	env["WALLET_ADDRESS"] = cfg.Wallet.Address
	env["L1_EL_URL"] = cfg.L1ElURL
	env["L1_CL_URL"] = cfg.L1ClURL
	env["L1_CHAIN_ID"] = fmt.Sprintf("%d", cfg.L1ChainID)
	env["COORDINATOR_PRIVATE_KEY"] = cfg.CoordinatorPrivateKey
	env["SEQUENCER_PRIVATE_KEY"] = cfg.CoordinatorPrivateKey
	env["SP_L1_SUPERBLOCK_CONTRACT"] = ""

	env["PUBLISHER_PATH"] = filepath.Join(o.rootDir, "internal", "l2", "services", string(configs.RepositoryNamePublisher))
	env["OP_GETH_PATH"] = filepath.Join(o.rootDir, "internal", "l2", "services", string(configs.RepositoryNameOpGeth))
	env["OPTIMISM_PATH"] = filepath.Join(o.rootDir, "internal", "l2", "services", string(configs.RepositoryNameOptimism))

	env["ROLLUP_A_CHAIN_ID"] = fmt.Sprintf("%d", cfg.ChainConfigs[configs.L2ChainNameRollupA].ID)
	env["ROLLUP_A_RPC_PORT"] = fmt.Sprintf("%d", cfg.ChainConfigs[configs.L2ChainNameRollupA].RPCPort)

	env["ROLLUP_B_CHAIN_ID"] = fmt.Sprintf("%d", cfg.ChainConfigs[configs.L2ChainNameRollupB].ID)
	env["ROLLUP_B_RPC_PORT"] = fmt.Sprintf("%d", cfg.ChainConfigs[configs.L2ChainNameRollupB].RPCPort)

	env["SP_L1_DISPUTE_GAME_FACTORY"] = gameFactoryAddr.Hex()

	return env
}

// buildComposeServices builds services using docker-compose
func (o *Orchestrator) buildComposeServices(ctx context.Context, env map[string]string) error {
	services := []string{
		"publisher",
		"op-geth-a",
		"op-node-a",
		"op-batcher-a",
		"op-proposer-a",
	}

	if err := docker.ComposeBuild(ctx, env, services...); err != nil {
		return fmt.Errorf("failed to build compose services: %w", err)
	}

	return nil
}
