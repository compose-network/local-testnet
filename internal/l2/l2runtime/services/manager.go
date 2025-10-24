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

// StartInitial starts initial services (publisher, op-geth) before contract deployment
func (m *Manager) StartInitial(ctx context.Context, env map[string]string) error {
	initialServices := []string{
		"publisher",
		"op-geth-a",
		"op-geth-b",
	}

	m.logger.With("initial_services", initialServices).Info("starting initial services")

	if err := docker.ComposeUp(ctx, env, initialServices...); err != nil {
		return fmt.Errorf("failed to start initial services: %w", err)
	}

	m.logger.Info("initial services started successfully")
	return nil
}

// StartFinal starts final services (op-node, batcher, proposer)
func (m *Manager) StartFinal(ctx context.Context, env map[string]string) error {
	finalServices := []string{
		"op-node-a",
		"op-node-b",
		"op-batcher-a",
		"op-batcher-b",
		"op-proposer-a",
		"op-proposer-b",
	}

	m.logger.With("services", finalServices).Info("starting final services")

	if err := docker.ComposeUp(ctx, env, finalServices...); err != nil {
		return fmt.Errorf("failed to start final services: %w", err)
	}

	m.logger.Info("final services started successfully")

	return nil
}
