package deployer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"path/filepath"

	"github.com/compose-network/localnet-control-plane/internal/l2/docker"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/ethdb/leveldb"
	"github.com/ethereum/go-ethereum/ethdb/pebble"
)

// ExportGenesis exports genesis.json files for both rollups and returns their computed hashes.
func ExportGenesis(ctx context.Context, stateDir, networksDir string, rollupAChainID, rollupBChainID int, walletAddress, sequencerAddress string, genesisBalanceWei string, coordinatorPrivateKey string) (map[int]string, error) {
	slog.Info("exporting genesis files")

	absStateDir, err := filepath.Abs(stateDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute state path: %w", err)
	}

	rollupADir := filepath.Join(networksDir, "rollup-a")
	rollupBDir := filepath.Join(networksDir, "rollup-b")

	for _, dir := range []string{rollupADir, rollupBDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create network directory: %w", err)
		}
	}

	hashA, err := exportGenesisForChain(ctx, absStateDir, rollupADir, rollupAChainID, walletAddress, sequencerAddress, genesisBalanceWei, coordinatorPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to export genesis for rollup-a: %w", err)
	}

	hashB, err := exportGenesisForChain(ctx, absStateDir, rollupBDir, rollupBChainID, walletAddress, sequencerAddress, genesisBalanceWei, coordinatorPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to export genesis for rollup-b: %w", err)
	}

	genesisHashes := map[int]string{
		rollupAChainID: hashA,
		rollupBChainID: hashB,
	}

	slog.Info("genesis files exported successfully")

	return genesisHashes, nil
}

func exportGenesisForChain(ctx context.Context, stateDir, networkDir string, chainID int, walletAddress, sequencerAddress, genesisBalanceWei string, coordinatorPrivateKey string) (string, error) {
	genesisPath := filepath.Join(networkDir, "genesis.json")

	user := fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid())

	// Run op-deployer inspect genesis
	dockerClient, err := docker.New()
	if err != nil {
		return "", fmt.Errorf("failed to create docker client: %w", err)
	}
	defer dockerClient.Close()

	output, err := dockerClient.Run(ctx, docker.RunOptions{
		Image: deployerImage,
		Cmd: []string{
			"inspect",
			"genesis",
			fmt.Sprintf("%d", chainID),
		},
		Env: []string{
			"HOME=/work",
		},
		Volumes: map[string]string{
			stateDir: "/work",
		},
		WorkDir:    "/work",
		User:       user,
		AutoRemove: true,
		CaptureOut: true,
	})

	if err != nil {
		return "", fmt.Errorf("failed to run op-deployer inspect genesis: %w", err)
	}

	slog.Debug("op-deployer inspect genesis output", "chainID", chainID, "outputLength", len(output), "firstBytes", fmt.Sprintf("%q", output[:min(100, len(output))]))

	var genesis map[string]interface{}
	if err := json.Unmarshal([]byte(output), &genesis); err != nil {
		return "", fmt.Errorf("failed to parse genesis JSON (output length: %d, first 200 chars: %q): %w", len(output), output[:min(200, len(output))], err)
	}

	alloc, ok := genesis["alloc"].(map[string]interface{})
	if !ok {
		alloc = make(map[string]interface{})
		genesis["alloc"] = alloc
	}

	// Convert balance to hex format
	balanceWei := new(big.Int)
	var success bool
	balanceWei, success = balanceWei.SetString(genesisBalanceWei, 10)
	if !success {
		return "", fmt.Errorf("invalid genesis balance: %s", genesisBalanceWei)
	}
	balanceHex := fmt.Sprintf("0x%x", balanceWei)

	for _, addr := range []string{walletAddress, sequencerAddress} {
		if addr == "" {
			continue
		}
		cleanAddr := addr
		if len(cleanAddr) > 2 && cleanAddr[:2] == "0x" {
			cleanAddr = cleanAddr[2:]
		}
		cleanAddr = "0x" + cleanAddr

		accountData, ok := alloc[cleanAddr].(map[string]interface{})
		if !ok {
			accountData = make(map[string]interface{})
			alloc[cleanAddr] = accountData
		}
		accountData["balance"] = balanceHex
	}

	// Add accounts with balances to alloc, but preserve config from op-deployer
	// The config already has proper fork times and blob schedules configured

	// Add pragueTime and isthmusTime to config (required for op-geth to recognize Prague/Isthmus forks)
	// Note: isthmus_time is set by op-deployer in rollup.json, but op-geth also needs these in genesis config
	config, ok := genesis["config"].(map[string]interface{})
	if !ok {
		config = make(map[string]interface{})
		genesis["config"] = config
	}
	// Set both to 0 (active at genesis) to match isthmus_time in rollup.json
	// Op-geth validates that pragueTime must equal isthmusTime
	config["pragueTime"] = 0
	config["isthmusTime"] = 0

	genesisJSON, err := json.MarshalIndent(genesis, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal genesis JSON: %w", err)
	}

	if err := os.WriteFile(genesisPath, genesisJSON, 0644); err != nil {
		return "", fmt.Errorf("failed to write genesis file: %w", err)
	}

	slog.Info("genesis file written", "chainID", chainID, "path", genesisPath)

	// Compute genesis hash by initializing geth and extracting the hash from logs
	hash, err := computeGenesisHashViaGethInit(ctx, genesisPath, coordinatorPrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to compute genesis hash: %w", err)
	}

	slog.Info("genesis hash computed", "chainID", chainID, "hash", hash)
	return hash, nil
}

func computeGenesisHashViaGethInit(ctx context.Context, genesisPath string, coordinatorPrivateKey string) (string, error) {
	slog.Info("computing genesis hash from genesis file", "path", genesisPath)

	absGenesisPath, err := filepath.Abs(genesisPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	genesisDir := filepath.Dir(absGenesisPath)
	genesisFile := filepath.Base(absGenesisPath)

	tmpDataDir, err := os.MkdirTemp("", "geth-init-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp datadir: %w", err)
	}
	defer os.RemoveAll(tmpDataDir)

	dockerClient, err := docker.New()
	if err != nil {
		return "", fmt.Errorf("failed to create docker client: %w", err)
	}
	defer dockerClient.Close()

	slog.Info("running geth init", "image", "local/op-geth:dev", "genesisFile", genesisFile)
	_, err = dockerClient.Run(ctx, docker.RunOptions{
		Image: "local/op-geth:dev",
		Cmd: []string{
			"init",
			"--state.scheme=hash",
			"--datadir=/datadir",
			fmt.Sprintf("/genesis/%s", genesisFile),
		},
		Env: []string{
			fmt.Sprintf("GETH_COORDINATOR_KEY=%s", coordinatorPrivateKey),
		},
		Volumes: map[string]string{
			genesisDir: "/genesis",
			tmpDataDir: "/datadir",
		},
		AutoRemove: true,
		StreamLogs: false,
		CaptureOut: false,
	})

	if err != nil {
		slog.Error("geth init failed", "error", err)
		return "", fmt.Errorf("failed to run geth init: %w", err)
	}

	genesisBlockPath := filepath.Join(tmpDataDir, "geth", "chaindata")
	if _, err := os.Stat(genesisBlockPath); os.IsNotExist(err) {
		return "", fmt.Errorf("genesis database not created")
	}

	slog.Debug("opening database", "path", genesisBlockPath)

	// Try opening as pebble first (default for op-geth)
	var kvStore ethdb.KeyValueStore
	kvStore, err = pebble.New(genesisBlockPath, 16, 16, "", true)
	if err != nil {
		slog.Debug("pebble open failed, trying leveldb", "error", err)
		kvStore, err = leveldb.New(genesisBlockPath, 16, 16, "", true)
		if err != nil {
			return "", fmt.Errorf("failed to open database (tried pebble and leveldb): %w", err)
		}
	}
	defer kvStore.Close()

	slog.Debug("database opened successfully")

	db, err := rawdb.Open(kvStore, rawdb.OpenOptions{
		Ancient:  filepath.Join(tmpDataDir, "geth", "chaindata", "ancient"),
		ReadOnly: true,
	})
	if err != nil {
		kvStore.Close()
		return "", fmt.Errorf("failed to open database with freezer: %w", err)
	}
	defer db.Close()

	slog.Debug("reading genesis hash from database")

	// Read the genesis block hash (block number 0)
	genesisHash := rawdb.ReadCanonicalHash(db, 0)
	if genesisHash == (common.Hash{}) {
		return "", fmt.Errorf("genesis hash not found in database")
	}

	hash := genesisHash.Hex()
	slog.Info("genesis hash extracted from database", "hash", hash)
	return hash, nil
}
