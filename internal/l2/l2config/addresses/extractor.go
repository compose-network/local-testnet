package addresses

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/compose-network/localnet-control-plane/internal/l2/domain"
	"github.com/compose-network/localnet-control-plane/internal/l2/infra/filesystem"
	"github.com/compose-network/localnet-control-plane/internal/logger"
)

const fileName = "addresses.json"

type (
	// Extractor extracts L1 contract addresses from deployment state
	Extractor struct {
		writer filesystem.Writer
		logger *slog.Logger
	}
)

// NewExtractor creates a new address extractor
func NewExtractor(writer filesystem.Writer) *Extractor {
	return &Extractor{
		writer: writer,
		logger: logger.Named("addresses_extractor"),
	}
}

// Extract extracts addresses for a specific chain from deployment
func (e *Extractor) ExtractDisputeGameFactoryAddr(state *domain.DeploymentState, chainID int, path string) (string, error) {
	logger := e.logger.With("chain_id", chainID)
	logger.Info("extracting addresses for chain")

	var (
		chainDeployment domain.OpChainDeployment
		found           bool
	)
	for _, chain := range state.OpChainDeployments {
		if chain.ID == fmt.Sprintf("0x%064x", chainID) {
			chainDeployment = chain
			found = true
		}
	}
	if !found {
		return "", fmt.Errorf("failed to get chain deployment")
	}

	type addr struct {
		OptimismPortal     string `json:"OPTIMISM_PORTAL,omitempty"`
		L1StandardBridge   string `json:"L1_STANDARD_BRIDGE,omitempty"`
		SystemConfig       string `json:"SYSTEM_CONFIG,omitempty"`
		DisputeGameFactory string `json:"DISPUTE_GAME_FACTORY,omitempty"`
	}

	addresses := addr{
		OptimismPortal:     chainDeployment.OptimismPortalProxyAddress,
		L1StandardBridge:   chainDeployment.L1StandardBridgeProxyAddress,
		SystemConfig:       chainDeployment.SystemConfigProxyAddress,
		DisputeGameFactory: chainDeployment.DisputeGameFactoryProxyAddress,
	}

	filePath := filepath.Join(path, fileName)

	logger.
		With("file_path", filePath).
		Info("addresses extracted successfully. Writing file")
	if err := e.writer.WriteJSON(filePath, addresses); err != nil {
		return "", fmt.Errorf("failed to write '%s' for chain %d: %w", fileName, chainID, err)
	}

	return chainDeployment.DisputeGameFactoryProxyAddress, nil
}
