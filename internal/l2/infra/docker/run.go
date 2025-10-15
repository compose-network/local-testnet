package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
)

type RunOptions struct {
	Image      string
	Cmd        []string
	Env        []string
	Volumes    map[string]string // host:container
	WorkDir    string
	User       string
	AutoRemove bool
	StreamLogs bool
	CaptureOut bool
}

// Run runs a Docker container and waits for it to complete.
func (c *Client) Run(ctx context.Context, opts RunOptions) (string, error) {
	config := &container.Config{
		Image:      opts.Image,
		Cmd:        opts.Cmd,
		Env:        opts.Env,
		WorkingDir: opts.WorkDir,
		User:       opts.User,
	}

	hostConfig := &container.HostConfig{
		AutoRemove: opts.AutoRemove,
	}

	if len(opts.Volumes) > 0 {
		binds := make([]string, 0, len(opts.Volumes))
		for host, containerPath := range opts.Volumes {
			binds = append(binds, fmt.Sprintf("%s:%s", host, containerPath))
		}
		hostConfig.Binds = binds
	}

	resp, err := c.cli.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	containerID := resp.ID

	defer func() {
		if err != nil && !opts.AutoRemove {
			_ = c.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
		}
	}()

	// Attach to container logs before starting (needed for AutoRemove containers)
	var stdout, stderr bytes.Buffer

	if opts.CaptureOut || opts.StreamLogs || opts.AutoRemove {
		attachResp, err := c.cli.ContainerAttach(ctx, containerID, container.AttachOptions{
			Stream: true,
			Stdout: true,
			Stderr: true,
		})
		if err != nil {
			return "", fmt.Errorf("failed to attach to container: %w", err)
		}
		defer attachResp.Close()

		// Start copying output in background
		go func() {
			if opts.StreamLogs {
				// Stream to console and capture
				outWriter := io.MultiWriter(os.Stdout, &stdout)
				errWriter := io.MultiWriter(os.Stderr, &stderr)
				_, _ = stdcopy.StdCopy(outWriter, errWriter, attachResp.Reader)
			} else {
				// Just capture
				_, _ = stdcopy.StdCopy(&stdout, &stderr, attachResp.Reader)
			}
		}()
	}

	if err := c.cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	statusCh, errCh := c.cli.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return "", fmt.Errorf("error waiting for container: %w", err)
		}
	case status := <-statusCh:
		if status.StatusCode != 0 {
			// Use already-captured output from attach
			errorOutput := stdout.String() + stderr.String()
			if errorOutput != "" {
				return "", fmt.Errorf("container exited with code %d: %s", status.StatusCode, errorOutput)
			}
			return "", fmt.Errorf("container exited with code %d", status.StatusCode)
		}
	}

	// Return captured output
	if opts.CaptureOut {
		return stdout.String(), nil
	}

	return "", nil
}
