package deployer

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/compose-network/localnet-control-plane/internal/l2/docker"
)

const deployerImage = "op-deployer:local"

// EnsureImage ensures the op-deployer Docker image exists, building it if necessary.
func EnsureImage(ctx context.Context, rootDir string) error {
	dockerClient, err := docker.New()
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer dockerClient.Close()

	exists, err := dockerClient.ImageExists(ctx, deployerImage)
	if err != nil {
		return fmt.Errorf("failed to check if image exists: %w", err)
	}

	if exists {
		slog.Info("op-deployer image already exists", "tag", deployerImage)
		return nil
	}

	absRootDir, err := filepath.Abs(rootDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute root path: %w", err)
	}

	dockerfilePath := filepath.Join("internal", "l2", "deploy", "op-deployer.Dockerfile")

	if err := dockerClient.BuildImage(ctx, dockerfilePath, absRootDir, deployerImage); err != nil {
		return fmt.Errorf("failed to build op-deployer image: %w", err)
	}

	return nil
}
