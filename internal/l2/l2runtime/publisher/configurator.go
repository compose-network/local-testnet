package publisher

import (
	"embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"text/template"

	"github.com/compose-network/local-testnet/internal/logger"
)

//go:embed *.tmpl
var templatesFS embed.FS

// Configurator handles Publisher-specific configuration setup
// This includes creating the custom registry structure that prevents
// Publisher from loading embedded chain definitions
type Configurator struct {
	logger *slog.Logger
}

func NewConfigurator() *Configurator {
	return &Configurator{
		logger: logger.Named("publisher_configurator"),
	}
}

// SetupRegistry creates the custom registry directory structure for Publisher
// This prevents Publisher from loading embedded chain definitions from the registry library
// Publisher will only use chains specified in REGISTRY_STATIC_CHAIN_IDS environment variable
func (c *Configurator) SetupRegistry(localnetDir, networkName string) error {
	const composeFileName = "compose.toml"
	// Create registry directory structure: registry/networks/<network-name>/
	registryNetworkDir := filepath.Join(localnetDir, "registry", "networks", networkName)
	if err := os.MkdirAll(registryNetworkDir, 0755); err != nil {
		return fmt.Errorf("failed to create registry network directory: %w", err)
	}

	c.logger.Info("created registry network directory", "path", registryNetworkDir)

	tmplContent, err := templatesFS.ReadFile("compose.toml.tmpl")
	if err != nil {
		return fmt.Errorf("failed to read %s template: %w", composeFileName, err)
	}

	tmpl, err := template.New(composeFileName).Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse %s template: %w", composeFileName, err)
	}

	composeTomlPath := filepath.Join(registryNetworkDir, composeFileName)
	file, err := os.Create(composeTomlPath)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", composeFileName, err)
	}
	defer file.Close()

	// Execute template (currently no data needed, but structure is in place for future)
	if err := tmpl.Execute(file, nil); err != nil {
		return fmt.Errorf("failed to execute %s template: %w", composeFileName, err)
	}

	c.logger.Info("created minimal registry configuration", "path", composeTomlPath)

	return nil
}
