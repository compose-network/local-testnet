package bridge

import (
	"github.com/compose-network/localnet-control-plane/configs"
	"github.com/spf13/cobra"
)

var CMD = &cobra.Command{
	Use:   "bridge",
	Short: "Cross-chain token bridge commands",
}

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send tokens from rollup A to rollup B",
	RunE:  runSend,
}

var (
	flagAmount    string
	flagSessionID string
	flagWait      int
)

func init() {
	sendCmd.Flags().StringVar(&flagAmount, "amount", "100", "Amount to bridge in wei")
	sendCmd.Flags().StringVar(&flagSessionID, "session", "", "Session ID (random if not specified)")
	sendCmd.Flags().IntVar(&flagWait, "wait", 30, "Seconds to wait for confirmation")

	CMD.AddCommand(sendCmd)
}

func runSend(cmd *cobra.Command, args []string) error {
	cfg := configs.Values.L2
	return ExecuteBridge(cmd.Context(), cfg, flagAmount, flagSessionID, flagWait)
}
