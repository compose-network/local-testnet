package deployer

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/compose-network/local-testnet/configs"
	"github.com/compose-network/local-testnet/internal/l2/infra/docker"
	"github.com/compose-network/local-testnet/internal/logger"
)

const (
	dockerfileName = "op-deployer.Dockerfile"
	imageName      = "op-deployer"
	imageTag       = "local"
)

var imageWithTag = fmt.Sprintf("%s:%s", imageName, imageTag)

// Deployer wraps the op-deployer tool
type Deployer struct {
	rootDir  string
	stateDir string
	version  string
	docker   *docker.Client
	logger   *slog.Logger
}

// NewDeployer creates a new op-deployer wrapper
func NewDeployer(rootDir, stateDir, version string, dockerClient *docker.Client) *Deployer {
	return &Deployer{
		rootDir:  rootDir,
		stateDir: stateDir,
		version:  version,
		docker:   dockerClient,
		logger:   logger.Named("deployer"),
	}
}

// Init initializes the op-deployer state
func (o *Deployer) Init(ctx context.Context, l1ChainID int, l2Chains map[configs.L2ChainName]configs.ChainConfig) error {
	o.logger.
		With("state_dir", o.stateDir).
		Info("initializing deployer state. Ensuring image exists")

	if err := o.ensureImage(ctx); err != nil {
		return fmt.Errorf("failed to ensure op-deployer image: %w", err)
	}

	stateFile := filepath.Join(o.stateDir, stateFile)
	if _, err := os.Stat(stateFile); err == nil {
		o.logger.
			With("file_name", stateFile).
			Info("state already exists, skipping init")

		return nil
	}

	absStateDir, err := filepath.Abs(o.stateDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}
	var chainIDsStr []string
	for _, chainConfig := range l2Chains {
		chainIDsStr = append(chainIDsStr, fmt.Sprintf("%d", chainConfig.ID))
	}

	o.logger.Info("running docker image")
	_, err = o.docker.Run(ctx, docker.RunOptions{
		Image: imageWithTag,
		Cmd: []string{
			"init",
			"--intent-type", "custom",
			"--l1-chain-id", strconv.Itoa(l1ChainID),
			"--l2-chain-ids", strings.Join(chainIDsStr, ","),
		},
		Env: []string{
			"HOME=/work",
			"DEPLOYER_CACHE_DIR=/work/.cache",
		},
		Volumes: map[string]string{
			absStateDir: "/work",
		},
		WorkDir:    "/work",
		User:       fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()),
		AutoRemove: true,
		StreamLogs: true,
	})

	if err != nil {
		return fmt.Errorf("failed to run op-deployer init: %w", err)
	}

	o.logger.Info("deployer state initialized successfully")

	return nil
}

// EnsureImage ensures the op-deployer image exists
func (o *Deployer) ensureImage(ctx context.Context) error {
	logger := o.logger.With("image_name", imageName).With("image_tag", imageTag)

	exists, err := o.docker.ImageExists(ctx, imageWithTag)
	if err != nil {
		return fmt.Errorf("failed to check image: '%s' existence: %w", imageWithTag, err)
	}

	if exists {
		logger.Info("image already exists")
		return nil
	}

	logger.With("version", o.version).Info("building image")
	absRootDir, err := filepath.Abs(o.rootDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute root path: %w", err)
	}

	buildArgs := map[string]*string{
		"OP_DEPLOYER_VERSION": &o.version,
	}

	dockerfilePath := filepath.Join("internal", "l2", "l1deployment", "deployer", dockerfileName)
	if err := o.docker.BuildImage(ctx, dockerfilePath, absRootDir, imageWithTag, buildArgs); err != nil {
		return fmt.Errorf("failed to build image: '%s', %w", imageWithTag, err)
	}

	return nil
}

// Apply runs op-deployer apply to deploy L1 contracts
func (o *Deployer) Apply(ctx context.Context, l1RpcURL, deployerPrivateKey, deploymentTarget string) error {
	o.logger.
		With("deployment_target", deploymentTarget).
		Info("running deployer apply")

	absStateDir, err := filepath.Abs(o.stateDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	_, err = o.docker.Run(ctx, docker.RunOptions{
		Image: imageWithTag,
		Cmd: []string{
			"apply",
			"--deployment-target", deploymentTarget,
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
		User:       fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()),
		AutoRemove: true,
		StreamLogs: true,
	})

	if err != nil {
		return fmt.Errorf("failed to run op-deployer apply: %w", err)
	}

	o.logger.Info("deployer apply completed successfully")

	return nil
}

// InspectGenesis exports genesis JSON for a chain
func (o *Deployer) InspectGenesis(ctx context.Context, chainID int) (string, error) {
	absStateDir, err := filepath.Abs(o.stateDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	output, err := o.docker.Run(ctx, docker.RunOptions{
		Image: imageWithTag,
		Cmd: []string{
			"inspect",
			"genesis",
			fmt.Sprintf("%d", chainID),
		},
		Env: []string{
			"HOME=/work",
		},
		Volumes: map[string]string{
			absStateDir: "/work",
		},
		WorkDir:    "/work",
		User:       fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()),
		AutoRemove: true,
		CaptureOut: true,
	})

	if err != nil {
		return "", fmt.Errorf("failed to run op-deployer inspect genesis: %w", err)
	}

	return output, nil
}

// InspectRollup exports rollup config for a chain
func (o *Deployer) InspectRollup(ctx context.Context, chainID int, outputPath string) error {
	absStateDir, err := filepath.Abs(o.stateDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	absOutputDir := filepath.Dir(outputPath)
	outputFile := filepath.Base(outputPath)

	_, err = o.docker.Run(ctx, docker.RunOptions{
		Image: imageWithTag,
		Cmd: []string{
			"inspect",
			"rollup",
			"--outfile", fmt.Sprintf("/output/%s", outputFile),
			fmt.Sprintf("%d", chainID),
		},
		Env: []string{
			"HOME=/work",
		},
		Volumes: map[string]string{
			absStateDir:  "/work",
			absOutputDir: "/output",
		},
		WorkDir:    "/work",
		User:       fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()),
		AutoRemove: true,
	})

	if err != nil {
		return fmt.Errorf("failed to run op-deployer inspect rollup: %w", err)
	}

	return nil
}
