package git

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
)

// Repository represents a git repository to clone
type Repository struct {
	Name string
	URL  string
	Ref  string // branch, tag, or commit
}

// Cloner handles git repository operations
type Cloner struct{}

// NewCloner creates a new git cloner
func NewCloner() *Cloner {
	return &Cloner{}
}

// CloneAll clones multiple repositories in parallel
func (c *Cloner) CloneAll(ctx context.Context, destDir string, repos []Repository) error {
	slog.Info("cloning repositories", "count", len(repos), "destination", destDir)

	for _, repo := range repos {
		if err := c.Clone(ctx, destDir, repo); err != nil {
			return fmt.Errorf("failed to clone %s: %w", repo.Name, err)
		}
	}

	slog.Info("all repositories cloned successfully")
	return nil
}

// Clone clones a single repository
func (c *Cloner) Clone(ctx context.Context, destDir string, repo Repository) error {
	repoPath := filepath.Join(destDir, repo.Name)

	// Check if already cloned
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); err == nil {
		slog.Info("repository already cloned, skipping", "name", repo.Name, "path", repoPath)
		return nil
	}

	slog.Info("cloning repository", "name", repo.Name, "url", repo.URL, "ref", repo.Ref)

	// Ensure parent directory exists
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Clone the repository
	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", "--branch", repo.Ref, repo.URL, repoPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	slog.Info("repository cloned successfully", "name", repo.Name)
	return nil
}
