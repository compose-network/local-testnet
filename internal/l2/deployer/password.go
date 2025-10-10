package deployer

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// GeneratePasswordFiles generates empty password files for geth account import.
func GeneratePasswordFiles(networksDir string) error {
	slog.Info("generating password files")

	rollupADir := filepath.Join(networksDir, "rollup-a")
	rollupBDir := filepath.Join(networksDir, "rollup-b")

	for _, dir := range []string{rollupADir, rollupBDir} {
		passwordPath := filepath.Join(dir, "password.txt")
		if err := os.WriteFile(passwordPath, []byte("\n"), 0600); err != nil {
			return fmt.Errorf("failed to write password file: %w", err)
		}
		slog.Info("password file written", "path", passwordPath)
	}

	slog.Info("password files generated successfully")

	return nil
}
