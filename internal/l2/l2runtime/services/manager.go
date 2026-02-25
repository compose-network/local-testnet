package services

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"time"

	"github.com/compose-network/local-testnet/configs"
	"github.com/compose-network/local-testnet/internal/l2/infra/docker"
	"github.com/compose-network/local-testnet/internal/logger"
)

// Manager manages L2 service lifecycle via docker-compose
type Manager struct {
	rootDir                    string
	composeFilePath            string
	flashblocksComposeFilePath string
	flashblocksEnabled         bool
	logger                     *slog.Logger
}

// NewManager creates a new service manager
func NewManager(rootDir, composeFilePath string) *Manager {
	return &Manager{
		rootDir:         rootDir,
		composeFilePath: composeFilePath,
		logger:          logger.Named("service_manager"),
	}
}

// WithFlashblocks enables flashblocks support with the specified compose file
func (m *Manager) WithFlashblocks(flashblocksComposeFilePath string) *Manager {
	m.flashblocksComposeFilePath = flashblocksComposeFilePath
	m.flashblocksEnabled = true
	return m
}

// StartAll starts the core L2 services (excluding op-succinct).
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

	if m.flashblocksEnabled && m.flashblocksComposeFilePath != "" {
		services = append(services,
			"op-rbuilder-a",
			"op-rbuilder-b",
			"rollup-boost-a",
			"rollup-boost-b",
		)

		m.logger.With("services", services).Info("starting all L2 services with flashblocks")

		composeFiles := []string{m.composeFilePath, m.flashblocksComposeFilePath}
		if err := docker.ComposeUpMultiFile(ctx, composeFiles, env, services...); err != nil {
			return fmt.Errorf("failed to start services with flashblocks: %w", err)
		}
	} else {
		m.logger.With("services", services).Info("starting all L2 services")

		if err := docker.ComposeUp(ctx, m.composeFilePath, env, services...); err != nil {
			return fmt.Errorf("failed to start services: %w", err)
		}
	}

	m.logger.Info("all L2 services started successfully")
	return nil
}

// StartOpSuccinct starts selected op-succinct services after the rest of L2 is running.
func (m *Manager) StartOpSuccinct(ctx context.Context, env map[string]string, enabledChains []configs.L2ChainName) error {
	if len(enabledChains) == 0 {
		m.logger.Info("op-succinct has no enabled chains; skipping service startup")
		return nil
	}

	dbServices := make([]string, 0, len(enabledChains))
	proposerServices := make([]string, 0, len(enabledChains))

	for _, chain := range enabledChains {
		switch chain {
		case configs.L2ChainNameRollupA:
			dbServices = append(dbServices, "op-succinct-db-a")
			proposerServices = append(proposerServices, "op-succinct-a")
		case configs.L2ChainNameRollupB:
			dbServices = append(dbServices, "op-succinct-db-b")
			proposerServices = append(proposerServices, "op-succinct-b")
		default:
			m.logger.With("chain", chain).Warn("unknown op-succinct chain; skipping")
		}
	}

	if len(proposerServices) == 0 {
		m.logger.Info("op-succinct has no valid enabled chains; skipping service startup")
		return nil
	}

	m.logger.With("services", dbServices).Info("starting op-succinct database services")
	if err := docker.ComposeUp(ctx, m.composeFilePath, env, dbServices...); err != nil {
		return fmt.Errorf("failed to start op-succinct database services: %w", err)
	}

	if err := m.clearOpSuccinctLocks(ctx, dbServices); err != nil {
		return err
	}

	m.logger.With("services", proposerServices).Info("starting op-succinct proposer services")
	if err := docker.ComposeUp(ctx, m.composeFilePath, env, proposerServices...); err != nil {
		return fmt.Errorf("failed to start op-succinct proposer services: %w", err)
	}

	m.logger.Info("op-succinct services started successfully")
	return nil
}

func (m *Manager) clearOpSuccinctLocks(ctx context.Context, dbContainers []string) error {
	for _, container := range dbContainers {
		if err := waitForPostgresReady(ctx, container); err != nil {
			return fmt.Errorf("op-succinct database %s is not ready: %w", container, err)
		}
		if err := clearChainLocks(ctx, container); err != nil {
			return fmt.Errorf("failed to clear op-succinct chain locks in %s: %w", container, err)
		}
	}

	m.logger.Info("cleared stale op-succinct database chain locks")
	return nil
}

func waitForPostgresReady(ctx context.Context, container string) error {
	deadline := time.Now().Add(60 * time.Second)
	var lastErr error

	for {
		cmd := exec.CommandContext(ctx, "docker", "exec", container, "pg_isready", "-U", "op-succinct", "-d", "op-succinct")
		if err := cmd.Run(); err == nil {
			return nil
		} else {
			lastErr = err
		}

		if time.Now().After(deadline) {
			if lastErr == nil {
				lastErr = fmt.Errorf("timed out waiting for postgres readiness")
			}
			return lastErr
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}

func clearChainLocks(ctx context.Context, container string) error {
	const cleanupSQL = "DO $$ BEGIN IF to_regclass('public.chain_locks') IS NOT NULL THEN DELETE FROM chain_locks; END IF; END $$;"

	cmd := exec.CommandContext(
		ctx,
		"docker", "exec",
		container,
		"psql", "-U", "op-succinct", "-d", "op-succinct",
		"-v", "ON_ERROR_STOP=1",
		"-c", cleanupSQL,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("psql cleanup failed: %w: %s", err, strings.TrimSpace(string(output)))
	}

	return nil
}

// StartFlashblocks starts flashblocks services (op-rbuilder and rollup-boost)
func (m *Manager) StartFlashblocks(ctx context.Context, env map[string]string) error {
	if !m.flashblocksEnabled || m.flashblocksComposeFilePath == "" {
		return fmt.Errorf("flashblocks not enabled or compose file not set")
	}

	services := []string{
		"op-rbuilder-a",
		"op-rbuilder-b",
		"rollup-boost-a",
		"rollup-boost-b",
	}

	m.logger.With("services", services).Info("starting flashblocks services")

	if err := docker.ComposeUpMultiFile(ctx, []string{m.composeFilePath, m.flashblocksComposeFilePath}, env, services...); err != nil {
		return fmt.Errorf("failed to start flashblocks services: %w", err)
	}

	m.logger.Info("flashblocks services started successfully")
	return nil
}
