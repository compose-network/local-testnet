package repository

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

type Repository struct {
	Name, URL, Ref, Dest string
}

// Clone clones one or more repositories in parallel.
// Each repository is cloned in a separate goroutine.
// Returns on the first error encountered or if context is cancelled.
func Clone(ctx context.Context, repos ...Repository) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(repos))

	for _, repo := range repos {
		wg.Add(1)
		go func(repo Repository) {
			defer wg.Done()

			slog.Info("cloning repository", "name", repo.Name, "url", repo.URL, "ref", repo.Ref)
			if err := cloneRepo(ctx, repo.URL, repo.Ref, repo.Dest); err != nil {
				select {
				case errChan <- fmt.Errorf("failed to clone %s: %w", repo.Name, err):
				case <-ctx.Done():
					errChan <- ctx.Err()
				}
				return
			}

			slog.Info("repository cloned successfully", "name", repo.Name)
		}(repo)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	case <-done:
		return nil
	}
}

// cloneRepo clones a single Git repository to the destination path.
// If the repository already exists, it skips cloning.
// Attempts shallow clone first, falls back to full clone if needed.
func cloneRepo(ctx context.Context, repo, ref, dest string) error {
	if _, err := os.Stat(filepath.Join(dest, ".git")); err == nil {
		slog.Info("repository already exists, skipping clone", "dest", dest)
		return nil
	}

	if _, err := os.Stat(dest); err == nil {
		if err := os.RemoveAll(dest); err != nil {
			return fmt.Errorf("failed to remove existing directory: %w", err)
		}
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	cmd := exec.CommandContext(ctx, "git", "clone", "--branch", ref, "--single-branch", repo, dest)

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		cmd = exec.CommandContext(ctx, "git", "clone", repo, dest)
		if err := cmd.Run(); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("git clone failed: %w", err)
		}

		cmd = exec.CommandContext(ctx, "git", "-C", dest, "checkout", ref)
		if err := cmd.Run(); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("git checkout %s failed: %w", ref, err)
		}
	}

	return nil
}
