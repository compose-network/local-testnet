package runtime

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/compose-network/localnet-control-plane/internal/logger"
)

const fileName = "runtime.env"

// Generator generates runtime.env for L2 chains
type Generator struct {
	logger *slog.Logger
}

func NewGenerator() *Generator {
	return &Generator{
		logger: logger.Named("runtime.env_generator"),
	}
}

func (g *Generator) Generate(gameFactoryProxyAddr string, path string) error {
	if gameFactoryProxyAddr == "" {
		return fmt.Errorf("could not generate %s, gameFactoryProxyAddr cannot be empty", fileName)
	}
	runtimeFilePath := filepath.Join(path, fileName)

	var lines []string
	lines = append(lines, fmt.Sprintf("DISPUTE_GAME_FACTORY_ADDRESS=%s", gameFactoryProxyAddr))
	lines = append(lines, fmt.Sprintf("OP_PROPOSER_GAME_FACTORY_ADDRESS=%s", gameFactoryProxyAddr))

	content := strings.Join(lines, "\n") + "\n"

	if err := os.WriteFile(runtimeFilePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", fileName, err)
	}

	g.logger.
		With("path", runtimeFilePath).
		Info("file was successfully written")

	return nil
}
