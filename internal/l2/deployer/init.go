package deployer

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	"github.com/compose-network/localnet-control-plane/internal/l2/docker"
)

// InitState initializes the op-deployer state directory.
func InitState(ctx context.Context, stateDir string, l1ChainID, rollupAChainID, rollupBChainID int) error {
	slog.Info("initializing op-deployer state", "stateDir", stateDir)

	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	cacheDir := filepath.Join(stateDir, ".cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	stateFile := filepath.Join(stateDir, "state.json")
	if _, err := os.Stat(stateFile); err == nil {
		slog.Info("state.json already exists, skipping init")
		return nil
	}

	absStateDir, err := filepath.Abs(stateDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Get current user UID:GID for docker
	user := fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid())

	dockerClient, err := docker.New()
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer dockerClient.Close()

	l2ChainIDs := fmt.Sprintf("%d,%d", rollupAChainID, rollupBChainID)

	_, err = dockerClient.Run(ctx, docker.RunOptions{
		Image: deployerImage,
		Cmd: []string{
			"init",
			"--intent-type", "custom",
			"--l1-chain-id", strconv.Itoa(l1ChainID),
			"--l2-chain-ids", l2ChainIDs,
		},
		Env: []string{
			"HOME=/work",
			"DEPLOYER_CACHE_DIR=/work/.cache",
		},
		Volumes: map[string]string{
			absStateDir: "/work",
		},
		WorkDir:    "/work",
		User:       user,
		AutoRemove: true,
		StreamLogs: true,
	})

	if err != nil {
		return fmt.Errorf("failed to run op-deployer init: %w", err)
	}

	slog.Info("op-deployer state initialized successfully")

	return nil
}
