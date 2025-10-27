package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/compose-network/local-testnet/internal/l2/infra/docker"
	"github.com/compose-network/local-testnet/internal/logger"
)

// Manager manages L2 service lifecycle via docker-compose
type Manager struct {
	rootDir string
	logger  *slog.Logger
}

// NewManager creates a new service manager
func NewManager(rootDir string) *Manager {
	return &Manager{
		rootDir: rootDir,
		logger:  logger.Named("service_manager"),
	}
}

// StartAll starts all L2 services
func (m *Manager) StartAll(ctx context.Context, env map[string]string) error {
	services := []string{
		"publisher",
		"op-geth-a",
		"op-geth-b",
		"op-node-a",
		"op-node-b",
		"op-batcher-a",
		"op-batcher-b",
		"op-proposer-a",
		"op-proposer-b",
	}

	m.logger.With("services", services).Info("starting all L2 services")

	if err := docker.ComposeUp(ctx, env, services...); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	m.logger.Info("all L2 services started successfully")
	return nil
}
