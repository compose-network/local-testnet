package l2config

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/compose-network/localnet-control-plane/configs"
	"github.com/compose-network/localnet-control-plane/internal/l2/domain"
	"github.com/compose-network/localnet-control-plane/internal/l2/infra/docker"
	"github.com/compose-network/localnet-control-plane/internal/l2/infra/filesystem/json"
	"github.com/compose-network/localnet-control-plane/internal/l2/l1deployment/deployer"
	"github.com/compose-network/localnet-control-plane/internal/l2/l2config/addresses"
	"github.com/compose-network/localnet-control-plane/internal/l2/l2config/contracts"
	"github.com/compose-network/localnet-control-plane/internal/l2/l2config/crypto"
	"github.com/compose-network/localnet-control-plane/internal/l2/l2config/genesis"
	"github.com/compose-network/localnet-control-plane/internal/l2/l2config/rollup"
	"github.com/compose-network/localnet-control-plane/internal/l2/l2config/runtime"
	"github.com/compose-network/localnet-control-plane/internal/l2/l2config/secrets"
	"github.com/compose-network/localnet-control-plane/internal/logger"
)

// Orchestrator coordinates Phase 2: L2 configuration generation
//   - Generates genesis.json for each L2 chain
//   - Generates rollup.json for each L2 chain
//   - Generates JWT secrets and passwords
//   - Extracts L1 contract addresses from state.json
//   - Writes contracts.json for each chain
//   - Builds runtime environment variables for docker-compose
type Orchestrator struct {
	rootDir     string
	stateDir    string
	networksDir string
	logger      *slog.Logger
}

// NewOrchestrator creates a new Phase 2 orchestrator
func NewOrchestrator(rootDir, stateDir, networksDir string) *Orchestrator {
	return &Orchestrator{
		rootDir:     rootDir,
		stateDir:    stateDir,
		networksDir: networksDir,
		logger:      logger.Named("l2_config_orchestrator"),
	}
}

// Execute runs Phase 2: Generate all L2 configuration files
func (o *Orchestrator) Execute(ctx context.Context, cfg configs.L2, stateDeployment *domain.DeploymentState) (string, error) {
	o.logger.Info("Phase 2: Starting L2 configuration generation")

	dockerClient, err := docker.New()
	if err != nil {
		return "", fmt.Errorf("failed to create docker client: %w", err)
	}
	defer dockerClient.Close()

	var (
		writer = json.NewWriter()

		opDeployer       = deployer.NewDeployer(o.rootDir, o.stateDir, cfg.OPDeployerVersion, dockerClient)
		genesisGen       = genesis.NewGenerator(opDeployer, dockerClient, writer)
		rollupGen        = rollup.NewGenerator(json.NewReader(), opDeployer, writer)
		secretsGen       = secrets.NewGenerator(writer)
		contractsGen     = contracts.NewGenerator(writer)
		runtimeGen       = runtime.NewGenerator()
	)
	)

	var disputeGameFactoryImplAddress string

	for chainName, chainConfig := range cfg.ChainConfigs {
		configPath := filepath.Join(o.networksDir, string(chainName))

		logger := o.logger.With("chain_name", chainName).With("chain_id", chainConfig.ID)
		logger.Info("generating l2 chain configuration")

		var chainDeployment *domain.OpChainDeployment
		for _, deployment := range stateDeployment.OpChainDeployments {
			if deployment.ID == fmt.Sprintf("0x%064x", chainConfig.ID) {
				chainDeployment = &deployment
				break
			}
		}
		if chainDeployment == nil {
			return "", fmt.Errorf("chain deployment not found for chain ID %d", chainConfig.ID)
		}

		disputeGameFactoryImplAddress = chainDeployment.DisputeGameFactoryProxyAddress

		sequencerAddress, err := crypto.AddressFromPrivateKey(cfg.CoordinatorPrivateKey)
		if err != nil {
			return "", fmt.Errorf("failed to derive sequencer address from coordinator PK for chain %d: %w", chainConfig.ID, err)
		}

		logger.Info("generating genesis file")
		genesisHash, err := genesisGen.Generate(
			ctx,
			chainConfig.ID,
			configPath,
			cfg.Wallet.Address,
			sequencerAddress,
			cfg.GenesisBalanceWei,
			cfg.CoordinatorPrivateKey,
		)
		if err != nil {
			return "", fmt.Errorf("failed to generate genesis for chain %d: %w", chainConfig.ID, err)
		}

		err = rollupGen.Generate(ctx, chainConfig.ID, configPath, genesisHash, chainDeployment.StartBlock.Hash, chainDeployment.StartBlock.Number)
		if err != nil {
			return "", fmt.Errorf("failed to generate rollup for chain %d: %w", chainConfig.ID, err)
		}

		err = secretsGen.GenerateJWT(configPath)
		if err != nil {
			return "", fmt.Errorf("failed to generate JWT for chain %d: %w", chainConfig.ID, err)
		}

		if err := secretsGen.GeneratePassword(configPath); err != nil {
			return "", fmt.Errorf("failed to generate password for chain %d: %w", chainConfig.ID, err)
		}

		if err := contractsGen.GeneratePlaceholders(configPath, chainConfig.ID); err != nil {
			return "", fmt.Errorf("failed to generate contract placeholders for chain %d: %w", chainConfig.ID, err)
		}

		if err := runtimeGen.Generate(stateDeployment.ImplementationsDeployment.DisputeGameFactoryImplAddress, configPath); err != nil {
			return "", fmt.Errorf("failed to generate runtime file, %w", err)
		}

	}

	o.logger.Info("Phase 2: L2 configuration generation completed successfully")

	return disputeGameFactoryImplAddress, nil
}
