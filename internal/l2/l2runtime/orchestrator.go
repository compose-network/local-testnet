package l2runtime

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/compose-network/local-testnet/configs"
	"github.com/compose-network/local-testnet/internal/l2/infra/docker"
	"github.com/compose-network/local-testnet/internal/l2/l1deployment"
	"github.com/compose-network/local-testnet/internal/l2/l2runtime/contracts"
	"github.com/compose-network/local-testnet/internal/l2/l2runtime/registry"
	"github.com/compose-network/local-testnet/internal/l2/l2runtime/services"
	"github.com/compose-network/local-testnet/internal/logger"
	"github.com/ethereum/go-ethereum/common"
)

// Orchestrator coordinates Phase 3: L2 runtime operations
//   - Builds Docker images via docker-compose
//   - Starts initial services (publisher, op-geth)
//   - Deploys L2 helper contracts
//   - Restarts services to pick up contract addresses
//   - Runs op-succinct setup calls and starts op-succinct services
type Orchestrator struct {
	rootDir     string
	localnetDir string
	networksDir string
	servicesDir string
	logger      *slog.Logger
}

// NewOrchestrator creates a new Phase 3 orchestrator
func NewOrchestrator(rootDir, localnetDir, networksDir, servicesDir string) *Orchestrator {
	return &Orchestrator{
		rootDir:     rootDir,
		localnetDir: localnetDir,
		networksDir: networksDir,
		servicesDir: servicesDir,
		logger:      logger.Named("l2_runtime_orchestrator"),
	}
}

// Execute runs Phase 3: Build images, start services, deploy contracts.
func (o *Orchestrator) Execute(ctx context.Context, cfg configs.L2, deploymentState l1deployment.DeploymentState) (map[configs.L2ChainName]map[contracts.ContractName]common.Address, error) {
	o.logger.Info("Phase 3: Starting L2 runtime operations")
	gameFactoryAddr := deploymentState.DisputeGameFactoryAddress

	publisherConfig := registry.NewConfigurator()
	if err := publisherConfig.SetupRegistry(o.localnetDir, cfg, gameFactoryAddr); err != nil {
		return nil, fmt.Errorf("failed to setup publisher registry: %w", err)
	}

	composePath, err := docker.EnsureComposeFile(o.localnetDir)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare docker-compose file: %w", err)
	}

	envBuilder := docker.NewEnvBuilder(o.rootDir, o.networksDir, o.servicesDir)
	envVars, err := envBuilder.BuildComposeEnv(cfg, gameFactoryAddr)
	if err != nil {
		return nil, err
	}

	enabledOpSuccinctChains := cfg.EnabledOpSuccinctChains()
	opSuccinctEnabled := isOpSuccinctEnabled(cfg, enabledOpSuccinctChains)
	opSuccinctPath := envVars["OP_SUCCINCT_PATH"]
	if opSuccinctEnabled {
		if opSuccinctPath == "" {
			return nil, fmt.Errorf("OP_SUCCINCT_PATH is empty")
		}
		if err := o.prepareOpSuccinctEnvFiles(cfg, envVars, deploymentState.DisputeGameFactoryProxyAddresses, opSuccinctPath); err != nil {
			return nil, fmt.Errorf("failed to prepare op-succinct env files: %w", err)
		}
	} else {
		o.logger.With("enabled_chains", enabledOpSuccinctChains).Info("op-succinct is disabled or repository is not configured; skipping op-succinct setup and services")
	}

	o.logger.With("env", envVars).Info("environment variables were constructed. Building compose services")
	if err := o.buildComposeServices(ctx, composePath, envVars, opSuccinctBuildServiceName(enabledOpSuccinctChains, opSuccinctEnabled)); err != nil {
		return nil, fmt.Errorf("failed to build compose services: %w", err)
	}

	o.logger.Info("docker-compose services built successfully")
	serviceManager := services.NewManager(o.rootDir, composePath)

	if cfg.Flashblocks.Enabled {
		o.logger.Info("flashblocks enabled, configuring services to use rollup-boost")
		flashblocksComposePath, err := docker.EnsureFlashblocksComposeFile(o.localnetDir)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare flashblocks compose file: %w", err)
		}
		serviceManager.WithFlashblocks(flashblocksComposePath)

		if cfg.Flashblocks.OpRbuilderImageTag != "" {
			envVars["OP_RBUILDER_IMAGE_TAG"] = cfg.Flashblocks.OpRbuilderImageTag
		}
		if cfg.Flashblocks.RollupBoostImageTag != "" {
			envVars["ROLLUP_BOOST_IMAGE_TAG"] = cfg.Flashblocks.RollupBoostImageTag
		}
	}

	if err := serviceManager.StartAll(ctx, envVars); err != nil {
		return nil, fmt.Errorf("failed to start base L2 services: %w", err)
	}

	// When flashblocks is enabled, use op-rbuilder RPC ports for contract deployment
	effectiveChainConfigs := cfg.ChainConfigs
	if cfg.Flashblocks.Enabled {
		effectiveChainConfigs = o.getFlashblocksChainConfigs(cfg)
		o.logger.Info("using flashblocks RPC ports for contract deployment",
			"rollup_a_port", effectiveChainConfigs[configs.L2ChainNameRollupA].RPCPort,
			"rollup_b_port", effectiveChainConfigs[configs.L2ChainNameRollupB].RPCPort)
	}

	contractDeployer := contracts.NewDeployer(o.networksDir)
	deployedContracts, err := contractDeployer.Deploy(ctx, effectiveChainConfigs, cfg.CoordinatorPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy contracts: %w", err)
	}

	o.logger.Info("restarting op-geth services to apply mailbox configuration")
	if err := o.restartOpGeth(ctx, composePath, envVars, deployedContracts); err != nil {
		return nil, fmt.Errorf("failed to restart op-geth services after contract deployment. Error: '%w'", err)
	}

	if opSuccinctEnabled {
		if err := o.setupOpSuccinct(ctx, cfg, opSuccinctPath, envVars); err != nil {
			return nil, fmt.Errorf("failed to run op-succinct setup calls: %w", err)
		}
		if err := o.finalizeOpSuccinctRuntimeEnvFiles(cfg, envVars); err != nil {
			return nil, fmt.Errorf("failed to finalize op-succinct runtime env files: %w", err)
		}

		if err := serviceManager.StartOpSuccinct(ctx, envVars, enabledOpSuccinctChains); err != nil {
			return nil, fmt.Errorf("failed to start op-succinct services: %w", err)
		}
	}

	o.logger.Info("Phase 3: L2 runtime operations completed successfully")

	return deployedContracts, nil
}

func (o *Orchestrator) restartOpGeth(ctx context.Context, composeFilePath string, env map[string]string, deployedContracts map[configs.L2ChainName]map[contracts.ContractName]common.Address) error {
	mailboxA := deployedContracts[configs.L2ChainNameRollupA][contracts.ContractNameMailbox]
	mailboxB := deployedContracts[configs.L2ChainNameRollupB][contracts.ContractNameMailbox]

	if mailboxA == (common.Address{}) || mailboxB == (common.Address{}) {
		return fmt.Errorf("mailbox addresses not found in deployed contracts")
	}

	env["MAILBOX_A"] = mailboxA.Hex()
	env["MAILBOX_B"] = mailboxB.Hex()

	o.logger.Info("restarting op-geth with mailbox addresses",
		"mailbox_a", mailboxA.Hex(),
		"mailbox_b", mailboxB.Hex())

	services := []string{"op-geth-a", "op-geth-b"}
	if err := docker.ComposeRestart(ctx, composeFilePath, env, services...); err != nil {
		return fmt.Errorf("failed to restart op-geth: %w", err)
	}

	o.logger.Info("op-geth services restarted successfully, waiting for them to be ready")

	return nil
}

// buildComposeServices builds services that are sourced locally.
func (o *Orchestrator) buildComposeServices(ctx context.Context, composeFilePath string, env map[string]string, opSuccinctBuildService string) error {
	services := []string{
		"publisher",
		"op-geth-a",
		"op-geth-b",
	}
	if opSuccinctBuildService != "" {
		services = append(services, opSuccinctBuildService)
	}

	if err := docker.ComposeBuild(ctx, composeFilePath, env, services...); err != nil {
		return fmt.Errorf("failed to build compose services: %w", err)
	}

	return nil
}

// getFlashblocksChainConfigs returns chain configs with op-rbuilder RPC ports.
func (o *Orchestrator) getFlashblocksChainConfigs(cfg configs.L2) map[configs.L2ChainName]configs.Chain {
	result := make(map[configs.L2ChainName]configs.Chain)

	for chainName, chainCfg := range cfg.ChainConfigs {
		modifiedCfg := chainCfg
		switch chainName {
		case configs.L2ChainNameRollupA:
			if cfg.Flashblocks.RollupARPCPort > 0 {
				modifiedCfg.RPCPort = cfg.Flashblocks.RollupARPCPort
			}
		case configs.L2ChainNameRollupB:
			if cfg.Flashblocks.RollupBRPCPort > 0 {
				modifiedCfg.RPCPort = cfg.Flashblocks.RollupBRPCPort
			}
		}
		result[chainName] = modifiedCfg
	}

	return result
}

func isOpSuccinctEnabled(cfg configs.L2, enabledChains []configs.L2ChainName) bool {
	if len(enabledChains) == 0 {
		return false
	}

	repo, exists := cfg.Repositories[configs.RepositoryNameOpSuccinct]
	if !exists {
		return false
	}

	return repo.LocalPath != "" || repo.URL != "" || repo.Branch != ""
}

func opSuccinctBuildServiceName(enabledChains []configs.L2ChainName, opSuccinctEnabled bool) string {
	if !opSuccinctEnabled {
		return ""
	}

	for _, chain := range enabledChains {
		switch chain {
		case configs.L2ChainNameRollupA:
			return "op-succinct-a"
		case configs.L2ChainNameRollupB:
			return "op-succinct-b"
		}
	}

	return ""
}
