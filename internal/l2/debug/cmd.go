package debug

import (
	"fmt"
	"os"

	"github.com/compose-network/localnet-control-plane/configs"
	"github.com/spf13/cobra"
)

var CMD = &cobra.Command{
	Use:   "debug-bridge",
	Short: "Diagnose cross-rollup bridge activity and system health",
	Long: `Debug-bridge provides diagnostics for the Compose rollup bridge system.

Modes:
  - debug: Full diagnostics including mailbox activity, publisher stats, balances, and logs
  - check: Quick health check with balances and publisher status

Examples:
  localnet l2 debug-bridge --mode=debug --blocks=20
  localnet l2 debug-bridge --mode=check
  localnet l2 debug-bridge --mode=debug --session=12345 --blocks=50
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate configuration
		if err := validateConfig(configs.Values.L2); err != nil {
			return fmt.Errorf("configuration validation failed: %w", err)
		}

		mode, _ := cmd.Flags().GetString("mode")
		blocks, _ := cmd.Flags().GetInt("blocks")
		session, _ := cmd.Flags().GetInt("session")
		since, _ := cmd.Flags().GetString("since")

		rootDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}

		service := NewService(rootDir)

		var sessionFilter *uint64
		if session > 0 {
			sessionVal := uint64(session)
			sessionFilter = &sessionVal
		}

		ctx := cmd.Context()

		switch mode {
		case "debug":
			return service.RunDebug(ctx, configs.Values.L2, blocks, sessionFilter, since)
		case "check":
			return service.RunCheck(ctx, configs.Values.L2)
		default:
			return fmt.Errorf("invalid mode: %s (must be 'debug' or 'check')", mode)
		}
	},
}

func validateConfig(cfg configs.L2) error {
	if cfg.DebugBridge.PublisherStatsURL == "" {
		return fmt.Errorf("l2.debug-bridge.publisher-stats-url is required")
	}
	if cfg.DebugBridge.DefaultBlocks <= 0 {
		return fmt.Errorf("l2.debug-bridge.default-blocks must be greater than 0")
	}
	if cfg.DebugBridge.DefaultLogWindow == "" {
		return fmt.Errorf("l2.debug-bridge.default-log-window is required")
	}
	return nil
}

func init() {
	CMD.Flags().String("mode", "debug", "Operation mode: 'debug' or 'check'")
	CMD.Flags().Int("blocks", 12, "Number of recent blocks to scan for mailbox activity")
	CMD.Flags().Int("session", 0, "Filter mailbox activity by session ID (0 = no filter)")
	CMD.Flags().String("since", "120s", "Docker logs time window (e.g., '5m', '120s')")
}
