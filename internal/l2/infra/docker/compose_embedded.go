package docker

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed docker-compose.yml docker-compose.flashblocks.yml
var embeddedComposeFS embed.FS

const (
	composeFileName            = "docker-compose.yml"
	composeFlashblocksFileName = "docker-compose.flashblocks.yml"
)

// EnsureComposeFile ensures the docker-compose.yml file exists in the specified directory
// and returns its path. It always writes the embedded content to ensure the file is up-to-date.
// This allows the compose file to be used from anywhere (including when running
// the binary from a different directory).
func EnsureComposeFile(localnetDir string) (string, error) {
	composePath := filepath.Join(localnetDir, composeFileName)

	content, err := getDockerComposeContent()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(composePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create %s directory: %w", localnetDir, err)
	}

	if err := os.WriteFile(composePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write %s: %w", composeFileName, err)
	}

	return composePath, nil
}

// getDockerComposeContent returns the embedded docker-compose.yml content.
func getDockerComposeContent() (string, error) {
	content, err := embeddedComposeFS.ReadFile(composeFileName)
	if err != nil {
		return "", fmt.Errorf("failed to read embedded %s: %w", composeFileName, err)
	}
	return string(content), nil
}

// EnsureFlashblocksComposeFile ensures the docker-compose.flashblocks.yml file exists
// in the specified directory and returns its path.
func EnsureFlashblocksComposeFile(localnetDir string) (string, error) {
	composePath := filepath.Join(localnetDir, composeFlashblocksFileName)

	content, err := getFlashblocksComposeContent()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(composePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create %s directory: %w", localnetDir, err)
	}

	if err := os.WriteFile(composePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write %s: %w", composeFlashblocksFileName, err)
	}

	return composePath, nil
}

// getFlashblocksComposeContent returns the embedded docker-compose.flashblocks.yml content.
func getFlashblocksComposeContent() (string, error) {
	content, err := embeddedComposeFS.ReadFile(composeFlashblocksFileName)
	if err != nil {
		return "", fmt.Errorf("failed to read embedded %s: %w", composeFlashblocksFileName, err)
	}
	return string(content), nil
}
