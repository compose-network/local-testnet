package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ComposeBuild builds docker compose services.
func ComposeBuild(ctx context.Context, composeFilePath string, env map[string]string, services ...string) error {
	args := append([]string{"build", "--parallel"}, services...)
	return composeRun(ctx, composeFilePath, env, args...)
}

// ComposeUp starts docker compose services in detached mode.
func ComposeUp(ctx context.Context, composeFilePath string, env map[string]string, services ...string) error {
	args := append([]string{"up", "-d"}, services...)
	return composeRun(ctx, composeFilePath, env, args...)
}

// ComposeRestart restarts docker compose services.
func ComposeRestart(ctx context.Context, composeFilePath string, env map[string]string, services ...string) error {
	args := append([]string{"up", "-d", "--force-recreate"}, services...)
	return composeRun(ctx, composeFilePath, env, args...)
}

// ComposeDown stops docker compose services.
func ComposeDown(ctx context.Context, composeFilePath string, env map[string]string, removeVolumes bool) error {
	args := []string{"down"}
	if removeVolumes {
		args = append(args, "-v")
	}
	return composeRun(ctx, composeFilePath, env, args...)
}

// ComposeRun executes a docker compose command with environment variables.
func composeRun(ctx context.Context, composeFilePath string, env map[string]string, args ...string) error {
	fullArgs := append([]string{"compose", "-f", composeFilePath}, args...)
	cmd := exec.CommandContext(ctx, "docker", fullArgs...)
	cmd.Dir = filepath.Dir(composeFilePath)

	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker compose %s failed: %w", strings.Join(args, " "), err)
	}

	return nil
}
