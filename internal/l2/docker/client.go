package docker

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/client"
	"github.com/moby/go-archive"
)

type Client struct {
	cli *client.Client
}

// New creates a new Docker client.
func New() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &Client{cli: cli}, nil
}

// Close closes the Docker client connection.
func (c *Client) Close() error {
	return c.cli.Close()
}

// ImageExists checks if a Docker image exists locally.
func (c *Client) ImageExists(ctx context.Context, imageName string) (bool, error) {
	_, err := c.cli.ImageInspect(ctx, imageName)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// BuildImage builds a Docker image from a Dockerfile.
func (c *Client) BuildImage(ctx context.Context, dockerfilePath, contextPath, tag string) error {
	slog.Info("building docker image", "tag", tag, "dockerfile", dockerfilePath)

	buildContext, err := archive.TarWithOptions(contextPath, &archive.TarOptions{})
	if err != nil {
		return fmt.Errorf("failed to create build context: %w", err)
	}
	defer buildContext.Close()

	buildOptions := build.ImageBuildOptions{
		Tags:       []string{tag},
		Dockerfile: dockerfilePath,
		Remove:     true,
	}

	resp, err := c.cli.ImageBuild(ctx, buildContext, buildOptions)
	if err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		slog.Debug(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading build output: %w", err)
	}

	slog.Info("docker image built successfully", "tag", tag)
	return nil
}

// ComposeRun executes a docker compose command with environment variables.
func ComposeRun(ctx context.Context, env map[string]string, args ...string) error {
	fullArgs := append([]string{"compose", "-f", "internal/l2/deploy/docker-compose.yml"}, args...)
	cmd := exec.CommandContext(ctx, "docker", fullArgs...)

	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	slog.Debug("running docker compose command", "args", strings.Join(args, " "))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker compose %s failed: %w", strings.Join(args, " "), err)
	}

	return nil
}

// ComposeBuild builds docker compose services.
func ComposeBuild(ctx context.Context, env map[string]string, services ...string) error {
	args := append([]string{"build", "--parallel"}, services...)
	return ComposeRun(ctx, env, args...)
}

// ComposeUp starts docker compose services in detached mode.
func ComposeUp(ctx context.Context, env map[string]string, services ...string) error {
	args := append([]string{"up", "-d"}, services...)
	return ComposeRun(ctx, env, args...)
}

// ComposeRestart restarts docker compose services.
func ComposeRestart(ctx context.Context, env map[string]string, services ...string) error {
	args := append([]string{"restart"}, services...)
	return ComposeRun(ctx, env, args...)
}

// ComposeDown stops docker compose services.
func ComposeDown(ctx context.Context, env map[string]string, removeVolumes bool) error {
	args := []string{"down"}
	if removeVolumes {
		args = append(args, "-v")
	}
	return ComposeRun(ctx, env, args...)
}
