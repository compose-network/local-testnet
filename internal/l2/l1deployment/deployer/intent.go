package deployer

import (
	"bytes"
	_ "embed"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/compose-network/local-testnet/configs"
	"github.com/compose-network/local-testnet/internal/l2/infra/filesystem"
	"github.com/compose-network/local-testnet/internal/logger"
)

const intentFileName = "intent.toml"

//go:embed intent.tmpl
var intentTemplate string

// IntentWriter writes the intent.toml file
type IntentWriter struct {
	stateDir string
	writer   filesystem.Writer
	logger   *slog.Logger
}

// NewIntentWriter creates a new intent writer
func NewIntentWriter(stateDir string, writer filesystem.Writer) *IntentWriter {
	return &IntentWriter{
		stateDir: stateDir,
		writer:   writer,
		logger:   logger.Named("intent_writer"),
	}
}

// WriteIntent creates the intent.toml file for op-deployer
func (i *IntentWriter) WriteIntent(walletAddress, sequencerAddress string, l1ChainID int, l2Chains map[configs.L2ChainName]configs.ChainConfig) error {
	i.logger.
		With("file_name", intentFileName).
		Info("writing deployer intent file")

	intentPath := filepath.Join(i.stateDir, intentFileName)

	chains := make([]struct {
		ChainID string
	}, 0, len(l2Chains))
	for _, chainConfig := range l2Chains {
		chains = append(chains, struct {
			ChainID string
		}{
			ChainID: chainIDToHex(chainConfig.ID),
		})
	}

	data := struct {
		L1ChainID int
		Wallet    string
		Sequencer string
		Chains    []struct {
			ChainID string
		}
	}{
		L1ChainID: l1ChainID,
		Wallet:    strings.ToLower(walletAddress),
		Sequencer: strings.ToLower(sequencerAddress),
		Chains:    chains,
	}

	tmpl, err := template.New("intent").Parse(intentTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	if err := i.writer.WriteBytes(intentPath, buf.Bytes()); err != nil {
		return fmt.Errorf("failed to write intent file: %w", err)
	}

	slog.Info("intent file written successfully", "path", intentPath)
	return nil
}

// chainIDToHex converts a chain ID to 0x-prefixed 64-character hex string
func chainIDToHex(chainID int) string {
	return fmt.Sprintf("0x%064x", chainID)
}
