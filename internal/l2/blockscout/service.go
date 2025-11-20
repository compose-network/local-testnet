package blockscout

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/compose-network/local-testnet/internal/logger"
)

const (
	blockscoutVersion         = "9.0.2"
	blockscoutFrontendVersion = "v2.3.5"
)

type (
	ChainConfig struct {
		ID      int
		Host    string
		RPCPort int
		WSPort  int
	}
	Service struct {
		logger *slog.Logger
	}
)

func New() *Service {
	return &Service{
		logger: logger.Named("blockscout"),
	}
}

func (s *Service) Run(ctx context.Context, chainConfigs []ChainConfig) error {
	s.logger.Info("starting Blockscout service. Building environment variables")

	for _, config := range chainConfigs {
		envVars := s.buildEnvVars(config.Host, config.ID, config.RPCPort, config.WSPort)
		s.logger.Info("environment variables built", slog.Any("envVars", envVars))
	}

	return nil
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
