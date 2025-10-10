package deployer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/compose-network/localnet-control-plane/internal/l2/docker"
)

// ExportRollupConfigs exports rollup.json files for both rollups with the correct genesis hashes.
func ExportRollupConfigs(ctx context.Context, stateDir, networksDir string, rollupAChainID, rollupBChainID int, genesisHashes map[int]string) error {
	slog.Info("exporting rollup configs")

	absStateDir, err := filepath.Abs(stateDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute state path: %w", err)
	}

	rollupADir := filepath.Join(networksDir, "rollup-a")
	rollupBDir := filepath.Join(networksDir, "rollup-b")

	if err := exportRollupConfigForChain(ctx, absStateDir, rollupADir, rollupAChainID, genesisHashes[rollupAChainID]); err != nil {
		return fmt.Errorf("failed to export rollup config for rollup-a: %w", err)
	}

	if err := exportRollupConfigForChain(ctx, absStateDir, rollupBDir, rollupBChainID, genesisHashes[rollupBChainID]); err != nil {
		return fmt.Errorf("failed to export rollup config for rollup-b: %w", err)
	}

	slog.Info("rollup configs exported successfully")
	return nil
}

func exportRollupConfigForChain(ctx context.Context, stateDir, networkDir string, chainID int, genesisHash string) error {
	rollupPath := filepath.Join(networkDir, "rollup.json")

	absNetworkDir, err := filepath.Abs(networkDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute network path: %w", err)
	}

	user := fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid())

	dockerClient, err := docker.New()
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer dockerClient.Close()

	_, err = dockerClient.Run(ctx, docker.RunOptions{
		Image: deployerImage,
		Cmd: []string{
			"inspect",
			"rollup",
			"--outfile", "/network/rollup.json",
			fmt.Sprintf("%d", chainID),
		},
		Env: []string{
			"HOME=/work",
		},
		Volumes: map[string]string{
			stateDir:      "/work",
			absNetworkDir: "/network",
		},
		WorkDir:    "/work",
		User:       user,
		AutoRemove: true,
	})

	if err != nil {
		return fmt.Errorf("failed to run op-deployer inspect rollup: %w", err)
	}

	// Patch the rollup.json with the correct genesis hash
	data, err := os.ReadFile(rollupPath)
	if err != nil {
		return fmt.Errorf("failed to read rollup.json: %w", err)
	}

	var rollupConfig map[string]interface{}
	if err := json.Unmarshal(data, &rollupConfig); err != nil {
		return fmt.Errorf("failed to parse rollup.json: %w", err)
	}

	// Set isthmus_time to 0 (matches Python implementation)
	rollupConfig["isthmus_time"] = 0

	// Patch genesis.l2.hash
	genesis, ok := rollupConfig["genesis"].(map[string]interface{})
	if !ok {
		genesis = make(map[string]interface{})
		rollupConfig["genesis"] = genesis
	}

	l2, ok := genesis["l2"].(map[string]interface{})
	if !ok {
		l2 = make(map[string]interface{})
		genesis["l2"] = l2
	}

	l2["hash"] = genesisHash

	patchedData, err := json.MarshalIndent(rollupConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal patched rollup.json: %w", err)
	}

	if err := os.WriteFile(rollupPath, append(patchedData, '\n'), 0644); err != nil {
		return fmt.Errorf("failed to write patched rollup.json: %w", err)
	}

	slog.Info("rollup config file written and patched", "chainID", chainID, "path", rollupPath, "genesisHash", genesisHash)
	return nil
}
