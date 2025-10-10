package l2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/compose-network/localnet-control-plane/configs"
	"github.com/compose-network/localnet-control-plane/internal/l2/contracts"
	"github.com/compose-network/localnet-control-plane/internal/l2/crypto"
	"github.com/compose-network/localnet-control-plane/internal/l2/deployer"
	"github.com/compose-network/localnet-control-plane/internal/l2/docker"
	"github.com/compose-network/localnet-control-plane/internal/l2/repository"
)

const (
	servicesDir = "./internal/l2/services"
	stateDir    = "./internal/l2/state"
	networksDir = "./internal/l2/networks"
	rootDir     = "."
)

func start(ctx context.Context) error {
	if err := cloneRepositories(ctx); err != nil {
		return err
	}

	if err := deployer.EnsureImage(ctx, rootDir); err != nil {
		return err
	}

	cfg := configs.Values.L2
	if err := deployer.InitState(ctx, stateDir, cfg.L1ChainID, cfg.ChainIDs.RollupA, cfg.ChainIDs.RollupB); err != nil {
		return err
	}

	coordinatorAddress, err := crypto.AddressFromPrivateKey(cfg.CoordinatorPrivateKey)
	if err != nil {
		return err
	}

	if err := deployer.WriteIntent(stateDir, cfg.Wallet.Address, coordinatorAddress, cfg.L1ChainID, cfg.ChainIDs.RollupA, cfg.ChainIDs.RollupB); err != nil {
		return err
	}

	if err := deployer.ApplyDeployment(ctx, stateDir, cfg.L1ElURL, cfg.Wallet.PrivateKey, cfg.DeploymentTarget); err != nil {
		return err
	}

	env := buildDockerComposeEnv(cfg)

	slog.Info("building op-geth docker image (required for genesis hash computation)")
	if err := docker.ComposeBuild(ctx, env, "op-geth-a"); err != nil {
		return err
	}

	genesisHashes, err := deployer.ExportGenesis(ctx, stateDir, networksDir, cfg.ChainIDs.RollupA, cfg.ChainIDs.RollupB, cfg.Wallet.Address, coordinatorAddress, cfg.GenesisBalanceWei, cfg.CoordinatorPrivateKey)
	if err != nil {
		return err
	}

	if err := deployer.ExportRollupConfigs(ctx, stateDir, networksDir, cfg.ChainIDs.RollupA, cfg.ChainIDs.RollupB, genesisHashes); err != nil {
		return err
	}

	if err := deployer.GenerateJWTSecrets(networksDir); err != nil {
		return err
	}

	if err := deployer.GeneratePasswordFiles(networksDir); err != nil {
		return err
	}

	if err := deployer.ExportAddresses(stateDir, networksDir, cfg.ChainIDs.RollupA, cfg.ChainIDs.RollupB); err != nil {
		return err
	}

	slog.Info("creating placeholder contracts.json files")
	if err := createPlaceholderContractsJSON(networksDir, cfg); err != nil {
		return err
	}

	slog.Info("building remaining docker images")
	// Only build one service per unique image to avoid conflicts when building in parallel
	// op-geth-a already built earlier, op-geth-b uses same image
	// op-node-a/b, op-batcher-a/b, op-proposer-a/b each share images
	allServices := []string{
		"publisher",
		"op-node-a",
		"op-batcher-a",
		"op-proposer-a",
	}
	if err := docker.ComposeBuild(ctx, env, allServices...); err != nil {
		return err
	}

	slog.Info("starting initial docker compose services (excluding op-node)")

	// Start publisher and op-geth first, op-node will be started after contracts are deployed
	initialServices := []string{
		"publisher",
		"op-geth-a",
		"op-geth-b",
	}
	if err := docker.ComposeUp(ctx, env, initialServices...); err != nil {
		return err
	}

	contractsDir := filepath.Join(rootDir, "internal", "l2", "contracts")

	deployCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := contracts.DeployContracts(deployCtx, contractsDir, networksDir, cfg); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			slog.Warn("contract deployment timed out (publisher may not be producing blocks yet); continuing with placeholder addresses")
		} else {
			slog.Warn("contract deployment failed; continuing with placeholder addresses", "error", err)
		}
	}

	slog.Info("restarting publisher and op-geth to apply helper configuration")
	restartServices := []string{
		"publisher",
		"op-geth-a",
		"op-geth-b",
	}
	if err := docker.ComposeRestart(ctx, env, restartServices...); err != nil {
		return err
	}

	slog.Info("starting op-node, batcher, and proposer services")
	finalServices := []string{
		"op-node-a",
		"op-node-b",
		"op-batcher-a",
		"op-batcher-b",
		"op-proposer-a",
		"op-proposer-b",
	}
	if err := docker.ComposeUp(ctx, env, finalServices...); err != nil {
		return err
	}

	slog.Info("L2 rollup deployment completed successfully")

	return nil
}

func cloneRepositories(ctx context.Context) error {
	repos := configs.Values.L2.Repositories

	return repository.Clone(
		ctx,
		repository.Repository{
			Name: "op-geth",
			URL:  repos["op-geth"].URL,
			Ref:  repos["op-geth"].Branch,
			Dest: filepath.Join(servicesDir, "op-geth"),
		},
		repository.Repository{
			Name: "optimism",
			URL:  repos["optimism"].URL,
			Ref:  repos["optimism"].Branch,
			Dest: filepath.Join(servicesDir, "optimism"),
		},
		repository.Repository{
			Name: "publisher",
			URL:  repos["publisher"].URL,
			Ref:  repos["publisher"].Branch,
			Dest: filepath.Join(servicesDir, "publisher"),
		},
	)
}

func createPlaceholderContractsJSON(networksDir string, cfg configs.L2) error {
	placeholder := map[string]any{
		"chainInfo": map[string]any{
			"chainId": 0,
		},
		"addresses": map[string]string{
			"Mailbox":     "0x0000000000000000000000000000000000000000",
			"PingPong":    "0x0000000000000000000000000000000000000000",
			"Bridge":      "0x0000000000000000000000000000000000000000",
			"MyToken":     "0x0000000000000000000000000000000000000000",
			"Coordinator": "0x0000000000000000000000000000000000000000",
		},
	}

	data, err := json.MarshalIndent(placeholder, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal placeholder contracts.json: %w", err)
	}

	rollupAPath := filepath.Join(networksDir, "rollup-a", "contracts.json")
	if err := os.WriteFile(rollupAPath, append(data, '\n'), 0644); err != nil {
		return fmt.Errorf("failed to write placeholder contracts.json for Rollup A: %w", err)
	}

	rollupBPath := filepath.Join(networksDir, "rollup-b", "contracts.json")
	if err := os.WriteFile(rollupBPath, append(data, '\n'), 0644); err != nil {
		return fmt.Errorf("failed to write placeholder contracts.json for Rollup B: %w", err)
	}

	return nil
}

func buildDockerComposeEnv(cfg configs.L2) map[string]string {
	absRootDir, err := filepath.Abs(rootDir)
	if err != nil {
		absRootDir = rootDir
	}

	return map[string]string{
		"L1_EL_URL":                  cfg.L1ElURL,
		"L1_CL_URL":                  cfg.L1ClURL,
		"L1_CHAIN_ID":                fmt.Sprintf("%d", cfg.L1ChainID),
		"WALLET_PRIVATE_KEY":         strings.TrimPrefix(cfg.Wallet.PrivateKey, "0x"),
		"WALLET_ADDRESS":             cfg.Wallet.Address,
		"SEQUENCER_PRIVATE_KEY":      strings.TrimPrefix(cfg.CoordinatorPrivateKey, "0x"),
		"COORDINATOR_PRIVATE_KEY":    strings.TrimPrefix(cfg.CoordinatorPrivateKey, "0x"),
		"ROLLUP_A_CHAIN_ID":          fmt.Sprintf("%d", cfg.ChainIDs.RollupA),
		"ROLLUP_B_CHAIN_ID":          fmt.Sprintf("%d", cfg.ChainIDs.RollupB),
		"SP_L1_SUPERBLOCK_CONTRACT":  "0x0000000000000000000000000000000000000001",
		"SP_L1_DISPUTE_GAME_FACTORY": "0x0000000000000000000000000000000000000001",
		"OP_GETH_PATH":               filepath.Join(absRootDir, "internal", "l2", "services", "op-geth"),
		"OPTIMISM_PATH":              filepath.Join(absRootDir, "internal", "l2", "services", "optimism"),
		"PUBLISHER_PATH":             filepath.Join(absRootDir, "internal", "l2", "services", "publisher"),
	}
}
