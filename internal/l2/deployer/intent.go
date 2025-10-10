package deployer

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const intentTemplate = `configType = "custom"
l1ChainID = {{.L1ChainID}}
fundDevAccounts = false
l1ContractsLocator = "tag://op-contracts/v3.0.0"
l2ContractsLocator = "tag://op-contracts/v3.0.0"

[superchainRoles]
  proxyAdminOwner = "{{.Wallet}}"
  protocolVersionsOwner = "{{.Wallet}}"
  guardian = "{{.Wallet}}"

[[chains]]
  id = "{{.RollupA.ChainID}}"
  baseFeeVaultRecipient = "{{.Wallet}}"
  l1FeeVaultRecipient = "{{.Wallet}}"
  sequencerFeeVaultRecipient = "{{.Sequencer}}"
  eip1559DenominatorCanyon = 250
  eip1559Denominator = 50
  eip1559Elasticity = 6
  gasLimit = 60000000
  operatorFeeScalar = 0
  operatorFeeConstant = 0
  minBaseFee = 0
  [chains.roles]
    l1ProxyAdminOwner = "{{.Wallet}}"
    l2ProxyAdminOwner = "{{.Wallet}}"
    systemConfigOwner = "{{.Wallet}}"
    unsafeBlockSigner = "{{.Wallet}}"
    batcher = "{{.Wallet}}"
    proposer = "{{.Wallet}}"
    challenger = "{{.Wallet}}"

[[chains]]
  id = "{{.RollupB.ChainID}}"
  baseFeeVaultRecipient = "{{.Wallet}}"
  l1FeeVaultRecipient = "{{.Wallet}}"
  sequencerFeeVaultRecipient = "{{.Sequencer}}"
  eip1559DenominatorCanyon = 250
  eip1559Denominator = 50
  eip1559Elasticity = 6
  gasLimit = 60000000
  operatorFeeScalar = 0
  operatorFeeConstant = 0
  minBaseFee = 0
  [chains.roles]
    l1ProxyAdminOwner = "{{.Wallet}}"
    l2ProxyAdminOwner = "{{.Wallet}}"
    systemConfigOwner = "{{.Wallet}}"
    unsafeBlockSigner = "{{.Wallet}}"
    batcher = "{{.Wallet}}"
    proposer = "{{.Wallet}}"
    challenger = "{{.Wallet}}"
`

type intentData struct {
	L1ChainID int
	Wallet    string
	Sequencer string
	RollupA   rollupConfig
	RollupB   rollupConfig
}

type rollupConfig struct {
	ChainID string
}

// WriteIntent creates the intent.toml file for op-deployer.
func WriteIntent(stateDir, walletAddress, sequencerAddress string, l1ChainID, rollupAChainID, rollupBChainID int) error {
	slog.Info("writing op-deployer intent file", "stateDir", stateDir)

	intentPath := filepath.Join(stateDir, "intent.toml")

	data := intentData{
		L1ChainID: l1ChainID,
		Wallet:    strings.ToLower(walletAddress),
		Sequencer: strings.ToLower(sequencerAddress),
		RollupA: rollupConfig{
			ChainID: chainIDToHex(rollupAChainID),
		},
		RollupB: rollupConfig{
			ChainID: chainIDToHex(rollupBChainID),
		},
	}

	tmpl, err := template.New("intent").Parse(intentTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	if err := os.WriteFile(intentPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write intent file: %w", err)
	}

	slog.Info("intent file written successfully", "path", intentPath)
	return nil
}

// chainIDToHex converts a chain ID to 0x-prefixed 64-character hex string.
func chainIDToHex(chainID int) string {
	return fmt.Sprintf("0x%064x", chainID)
}
