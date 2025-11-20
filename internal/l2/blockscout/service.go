package blockscout

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"path/filepath"

	"github.com/compose-network/local-testnet/internal/l2/infra/docker"
	"github.com/compose-network/local-testnet/internal/logger"
)

const (
	blockscoutVersion         = "9.0.2"
	blockscoutFrontendVersion = "v2.3.5"
)

type (
	ChainConfig struct {
		ID         int
		ELHostName string
		RPCPort    int
		WSPort     int
	}
	Service struct {
		localnetDir string
		networksDir string
		logger      *slog.Logger
	}
)

func New(localnetDir, networksDir string) *Service {
	return &Service{
		localnetDir: localnetDir,
		networksDir: networksDir,
		logger:      logger.Named("blockscout"),
	}
}

func (s *Service) Run(ctx context.Context, chainConfigs []ChainConfig) error {
	s.logger.Info("starting Blockscout service")

	if len(chainConfigs) != 2 {
		return fmt.Errorf("expected exactly 2 chain configs, got %d", len(chainConfigs))
	}

	if err := GenerateNginxConfigs(s.networksDir); err != nil {
		return fmt.Errorf("failed to generate nginx configs: %w", err)
	}

	composePath, err := EnsureComposeFile(s.localnetDir)
	if err != nil {
		return fmt.Errorf("failed to prepare blockscout compose file: %w", err)
	}

	envVars := s.buildAllEnvVars(chainConfigs)
	s.logger.With("env", envVars).Info("environment variables built. Starting services")

	if err := docker.ComposeUp(ctx, composePath, envVars); err != nil {
		return fmt.Errorf("failed to start blockscout services: %w", err)
	}

	s.logger.Info("blockscout services started successfully")
	return nil
}

func (s *Service) buildAllEnvVars(chainConfigs []ChainConfig) map[string]string {
	envVars := make(map[string]string)

	envVars["BLOCKSCOUT_VERSION"] = blockscoutVersion
	envVars["BLOCKSCOUT_FRONTEND_VERSION"] = blockscoutFrontendVersion
	envVars["BLOCKSCOUT_A_PUBLIC_PORT"] = "19000"
	envVars["BLOCKSCOUT_B_PUBLIC_PORT"] = "29000"

	rollupANginxConf := filepath.Join(s.networksDir, "rollup-a", "blockscout-nginx.conf")
	rollupBNginxConf := filepath.Join(s.networksDir, "rollup-b", "blockscout-nginx.conf")
	envVars["ROLLUP_A_NGINX_CONF"] = rollupANginxConf
	envVars["ROLLUP_B_NGINX_CONF"] = rollupBNginxConf

	prefixes := []string{"ROLLUP_A_", "ROLLUP_B_"}

	for i, config := range chainConfigs {
		rollupVars := s.buildRollupEnvVars(config)
		mergeWithPrefix(envVars, rollupVars, prefixes[i])
	}

	return envVars
}

func (s *Service) buildRollupEnvVars(config ChainConfig) map[string]string {
	envVars := make(map[string]string)

	backend := s.buildEnvVars(config.ELHostName, config.ID, config.RPCPort, config.WSPort)
	maps.Copy(envVars, backend)

	frontend := s.buildFrontendEnvVars(config.ID)
	maps.Copy(envVars, frontend)

	return envVars
}

func mergeWithPrefix(dst, src map[string]string, prefix string) {
	for k, v := range src {
		dst[prefix+k] = v
	}
}

func (s *Service) buildEnvVars(host string, chainID, rpcPort, wsPort int) map[string]string {
	httpURL := fmt.Sprintf("http://%s:%d", host, rpcPort)

	return map[string]string{
		"CHAIN_ID":                   fmt.Sprintf("%d", chainID),
		"ETHEREUM_JSONRPC_HTTP_URL":  httpURL,
		"ETHEREUM_JSONRPC_TRACE_URL": httpURL,
		"ETHEREUM_JSONRPC_WS_URL":    fmt.Sprintf("ws://%s:%d", host, wsPort),
	}
}

func (s *Service) buildFrontendEnvVars(chainID int) map[string]string {
	return map[string]string{
		"NETWORK_ID": fmt.Sprintf("%d", chainID),
	}
}
