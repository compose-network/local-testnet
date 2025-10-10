package deployer

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/compose-network/localnet-control-plane/internal/l2/docker"
)

// ApplyDeployment runs op-deployer apply to deploy L1 contracts.
func ApplyDeployment(ctx context.Context, stateDir, l1RpcURL, deployerPrivateKey, deploymentTarget string) error {
	slog.Info("running op-deployer apply", "deploymentTarget", deploymentTarget)

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

	_, err = dockerClient.Run(ctx, docker.RunOptions{
		Image: deployerImage,
		Cmd: []string{
			"apply",
			fmt.Sprintf("--deployment-target=%s", deploymentTarget),
		},
		Env: []string{
			"HOME=/work",
			"DEPLOYER_CACHE_DIR=/work/.cache",
			fmt.Sprintf("L1_RPC_URL=%s", l1RpcURL),
			fmt.Sprintf("DEPLOYER_PRIVATE_KEY=%s", deployerPrivateKey),
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
		return fmt.Errorf("failed to run op-deployer apply: %w", err)
	}

	slog.Info("op-deployer apply completed successfully")

	return nil
}
