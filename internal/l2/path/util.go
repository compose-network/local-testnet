package path

import (
	"os"
	"path/filepath"
	"strings"
)

// GetHostPath converts a container path to the host path when running in Docker.
// This is needed because nested containers need to mount volumes using host paths.
//
// When HOST_PROJECT_PATH env var is set (running in Docker):
//   - Converts /workspace/... paths to $HOST_PROJECT_PATH/...
//   - Passes through other paths unchanged (like /tmp)
//
// When HOST_PROJECT_PATH is not set (running natively):
//   - Returns absolute path as-is
func GetHostPath(containerPath string) (string, error) {
	hostProjectPath := os.Getenv("HOST_PROJECT_PATH")
	if hostProjectPath == "" {
		// Not running in Docker, use absolute path
		return filepath.Abs(containerPath)
	}

	// Running in Docker: replace /workspace with the host's project path
	absPath, err := filepath.Abs(containerPath)
	if err != nil {
		return "", err
	}

	// Convert /workspace/... to $HOST_PROJECT_PATH/...
	if after, ok := strings.CutPrefix(absPath, "/workspace/"); ok {
		return filepath.Join(hostProjectPath, after), nil
	}
	if absPath == "/workspace" {
		return hostProjectPath, nil
	}

	// For paths outside /workspace (like /tmp), return as-is
	// These will fail when running in Docker, which is expected
	return absPath, nil
}
