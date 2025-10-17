package debug

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/compose-network/localnet-control-plane/configs"
	"github.com/compose-network/localnet-control-plane/internal/l2/debug/balance"
	"github.com/compose-network/localnet-control-plane/internal/l2/debug/logs"
	"github.com/compose-network/localnet-control-plane/internal/l2/debug/mailbox"
	"github.com/compose-network/localnet-control-plane/internal/l2/debug/publisher"
	fsjson "github.com/compose-network/localnet-control-plane/internal/l2/infra/filesystem/json"
	"github.com/compose-network/localnet-control-plane/internal/logger"
	"github.com/ethereum/go-ethereum/common"
)

// Service orchestrates debug-bridge operations
type Service struct {
	rootDir        string
	networksDir    string
	mailboxScanner *mailbox.Scanner
	balanceChecker *balance.Checker
	statsFetcher   *publisher.StatsFetcher
	logsCollector  *logs.Collector
	logger         *slog.Logger
}

// NewService creates a new debug service
func NewService(rootDir string) *Service {
	return &Service{
		rootDir:        rootDir,
		networksDir:    filepath.Join(rootDir, "internal", "l2", "networks"),
		mailboxScanner: mailbox.NewScanner(),
		balanceChecker: balance.NewChecker(),
		statsFetcher:   publisher.NewStatsFetcher(),
		logsCollector:  logs.NewCollector(),
		logger:         logger.Named("debug_service"),
	}
}

// RollupConfig holds rollup-specific configuration
type RollupConfig struct {
	Name          string
	RPCURL        string
	Container     string
	ChainID       int
	MailboxAddr   common.Address
	BridgeAddr    common.Address
	TokenAddr     common.Address
	WalletAddress common.Address
}

// RunDebug executes debug mode - full diagnostics
func (s *Service) RunDebug(ctx context.Context, cfg configs.L2, blocks int, sessionFilter *uint64, logsSince string) error {
	s.logger.Info("running debug-bridge in debug mode")

	rollups, err := s.loadRollupConfigs(cfg)
	if err != nil {
		return fmt.Errorf("failed to load rollup configs: %w", err)
	}

	publisherStatsURL := cfg.DebugBridge.PublisherStatsURL

	s.logger.Info("collecting mailbox activity", "blocks", blocks)

	// Collect mailbox activity from both rollups
	for _, rollup := range rollups {
		calls, err := s.mailboxScanner.ScanBlocks(ctx, rollup.RPCURL, rollup.MailboxAddr, blocks, sessionFilter)
		if err != nil {
			s.logger.With("rollup", rollup.Name).With("err", err).Warn("failed to collect mailbox activity")
			continue
		}

		if len(calls) == 0 {
			s.logger.With("rollup", rollup.Name).With("blocks", blocks).Info("no mailbox activity detected")
			continue
		}

		s.logger.With("rollup", rollup.Name).Info("mailbox activity found")
		for _, call := range calls {
			s.logger.Info(mailbox.FormatCall(call))
		}
	}

	// Fetch publisher stats
	stats, err := s.statsFetcher.FetchStats(ctx, publisherStatsURL)
	if err != nil {
		s.logger.With("err", err).Warn("shared publisher stats unavailable")
	} else {
		s.logger.Info("shared publisher stats")
		var prettyJSON map[string]interface{}
		_ = json.Unmarshal([]byte(stats.Raw), &prettyJSON)
		formatted, _ := json.MarshalIndent(prettyJSON, "", "  ")
		// Limit to 2000 chars like Python version
		output := string(formatted)
		if len(output) > 2000 {
			output = output[:2000]
		}
		s.logger.Info(output)
	}

	// Check balances
	s.logger.Info("recent balances")
	for _, rollup := range rollups {
		ethBal, err := s.balanceChecker.GetETHBalance(ctx, rollup.RPCURL, rollup.WalletAddress)
		s.logger.Info(balance.FormatETHBalance(rollup.Name, rollup.WalletAddress.Hex(), ethBal, err))
	}

	s.logger.Info("token balances")
	for _, rollup := range rollups {
		if rollup.TokenAddr == (common.Address{}) {
			s.logger.With("rollup", rollup.Name).Info("token address unavailable")
			continue
		}
		tokenBal, err := s.balanceChecker.GetTokenBalance(ctx, rollup.RPCURL, rollup.TokenAddr, rollup.WalletAddress)
		s.logger.Info(balance.FormatTokenBalance(rollup.Name, rollup.TokenAddr.Hex(), rollup.WalletAddress.Hex(), tokenBal, err))
	}

	// Collect logs
	s.logger.Info("log snippets")
	for _, rollup := range rollups {
		logLines, err := s.logsCollector.CollectLogs(ctx, rollup.Container, logsSince)
		if err != nil {
			s.logger.With("container", rollup.Container).With("err", err).Warn("failed to collect logs")
			continue
		}
		s.logger.Info(logs.FormatLogs(rollup.Container, logLines))
	}

	return nil
}

// RunCheck executes check mode - quick health check
func (s *Service) RunCheck(ctx context.Context, cfg configs.L2) error {
	s.logger.Info("running debug-bridge in check mode")

	rollups, err := s.loadRollupConfigs(cfg)
	if err != nil {
		return fmt.Errorf("failed to load rollup configs: %w", err)
	}

	publisherStatsURL := cfg.DebugBridge.PublisherStatsURL

	s.logger.With("address", cfg.Wallet.Address).Info("wallet address")

	s.logger.Info("balances")
	for _, rollup := range rollups {
		ethBal, err := s.balanceChecker.GetETHBalance(ctx, rollup.RPCURL, rollup.WalletAddress)
		s.logger.Info(balance.FormatETHBalance(rollup.Name, rollup.WalletAddress.Hex(), ethBal, err))
	}

	s.logger.Info("token balances")
	for _, rollup := range rollups {
		if rollup.TokenAddr == (common.Address{}) {
			s.logger.With("rollup", rollup.Name).Info("token address unavailable")
			continue
		}
		tokenBal, err := s.balanceChecker.GetTokenBalance(ctx, rollup.RPCURL, rollup.TokenAddr, rollup.WalletAddress)
		s.logger.Info(balance.FormatTokenBalance(rollup.Name, rollup.TokenAddr.Hex(), rollup.WalletAddress.Hex(), tokenBal, err))
	}

	stats, err := s.statsFetcher.FetchStats(ctx, publisherStatsURL)
	if err != nil {
		s.logger.With("err", err).Warn("shared publisher stats unavailable")
	} else {
		s.logger.Info(publisher.FormatStats(publisherStatsURL, stats, nil))
	}

	return nil
}

// loadRollupConfigs loads rollup configurations from contracts.json files
func (s *Service) loadRollupConfigs(cfg configs.L2) ([]RollupConfig, error) {
	rollups := []RollupConfig{}
	walletAddr := common.HexToAddress(cfg.Wallet.Address)

	for chainName, chainConfig := range cfg.ChainConfigs {
		contractsPath := filepath.Join(s.networksDir, string(chainName), "contracts.json")

		var data map[string]any
		reader := fsjson.NewReader()
		err := reader.ReadJSON(contractsPath, &data)
		if err != nil {
			return nil, fmt.Errorf("failed to read contracts.json for %s: %w", chainName, err)
		}

		addresses, ok := data["addresses"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid contracts.json format for %s", chainName)
		}

		var mailboxAddr, bridgeAddr, tokenAddr common.Address

		if addr, ok := addresses["Mailbox"].(string); ok {
			mailboxAddr = common.HexToAddress(addr)
		}
		if addr, ok := addresses["Bridge"].(string); ok {
			bridgeAddr = common.HexToAddress(addr)
		}
		if addr, ok := addresses["MyToken"].(string); ok {
			tokenAddr = common.HexToAddress(addr)
		}

		rpcURL := fmt.Sprintf("http://localhost:%d", chainConfig.RPCPort)
		containerName := fmt.Sprintf("op-geth-%s", string(chainName)[len(string(chainName))-1:])

		rollups = append(rollups, RollupConfig{
			Name:          string(chainName),
			RPCURL:        rpcURL,
			Container:     containerName,
			ChainID:       chainConfig.ID,
			MailboxAddr:   mailboxAddr,
			BridgeAddr:    bridgeAddr,
			TokenAddr:     tokenAddr,
			WalletAddress: walletAddr,
		})
	}

	return rollups, nil
}
