package blockscout

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed docker-compose.blockscout.yml networks
var embeddedFS embed.FS

const (
	composeFileName = "docker-compose.blockscout.yml"
)

func EnsureComposeFile(localnetDir string) (string, error) {
	composePath := filepath.Join(localnetDir, composeFileName)

	content, err := embeddedFS.ReadFile(composeFileName)
	if err != nil {
		return "", fmt.Errorf("failed to read embedded %s: %w", composeFileName, err)
	}

	if err := os.MkdirAll(filepath.Dir(composePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create %s directory: %w", localnetDir, err)
	}

	if err := os.WriteFile(composePath, content, 0644); err != nil {
		return "", fmt.Errorf("failed to write %s: %w", composeFileName, err)
	}

	if err := copyNetworksDir(localnetDir); err != nil {
		return "", fmt.Errorf("failed to copy networks directory: %w", err)
	}

	return composePath, nil
}

func copyNetworksDir(localnetDir string) error {
	networksDir := filepath.Join(localnetDir, "networks")
	if err := os.MkdirAll(networksDir, 0755); err != nil {
		return fmt.Errorf("failed to create networks directory: %w", err)
	}

	entries, err := embeddedFS.ReadDir("networks")
	if err != nil {
		return fmt.Errorf("failed to read embedded networks directory: %w", err)
	}

	for _, entry := range entries {
		if err := copyRollupDir(localnetDir, entry.Name()); err != nil {
			return err
		}
	}

	return nil
}

func copyRollupDir(localnetDir, rollupName string) error {
	rollupDir := filepath.Join(localnetDir, "networks", rollupName)
	if err := os.MkdirAll(rollupDir, 0755); err != nil {
		return fmt.Errorf("failed to create %s directory: %w", rollupDir, err)
	}

	entries, err := embeddedFS.ReadDir(filepath.Join("networks", rollupName))
	if err != nil {
		return fmt.Errorf("failed to read embedded %s directory: %w", rollupName, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		srcPath := filepath.Join("networks", rollupName, entry.Name())
		content, err := embeddedFS.ReadFile(srcPath)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", srcPath, err)
		}

		dstPath := filepath.Join(rollupDir, entry.Name())
		if err := os.WriteFile(dstPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", dstPath, err)
		}
	}

	return nil
}
