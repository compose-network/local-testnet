package deployer

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/compose-network/local-testnet/internal/l2/domain"
	"github.com/compose-network/local-testnet/internal/l2/infra/filesystem"
	"github.com/compose-network/local-testnet/internal/logger"
)

const stateFile = "state.json"

// StateManager manages the deployment state (state.json)
type StateManager struct {
	stateDir string
	reader   filesystem.Reader
	logger   *slog.Logger
}

// NewStateManager creates a new state manager
func NewStateManager(stateDir string, reader filesystem.Reader) *StateManager {
	return &StateManager{
		stateDir: stateDir,
		reader:   reader,
		logger:   logger.Named("state_manager"),
	}
}

// EnsureStateDir ensures the state directory and cache exist
func (s *StateManager) EnsureStateDir() error {
	if err := os.MkdirAll(s.stateDir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	cacheDir := filepath.Join(s.stateDir, ".cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	return nil
}

// Load reads the deployment state from state.json
func (s *StateManager) Load() (*domain.DeploymentState, error) {
	statePath := filepath.Join(s.stateDir, stateFile)

	var state domain.DeploymentState
	if err := s.reader.ReadJSON(statePath, &state); err != nil {
		return nil, fmt.Errorf("failed to read '%s': %w", stateFile, err)
	}

	return &state, nil
}
