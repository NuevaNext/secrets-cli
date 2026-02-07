package cmd

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindGitRoot traverses up from the current directory to find the git repository root.
// It returns the absolute path to the directory containing .git, or an error if not found.
func FindGitRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	for {
		gitDir := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitDir); err == nil {
			// .git can be a directory (normal repo) or a file (worktree/submodule)
			if info.IsDir() || info.Mode().IsRegular() {
				return dir, nil
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root without finding .git
			return "", ErrNotGitRepository
		}
		dir = parent
	}
}

// ErrNotGitRepository is returned when the current directory is not inside a git repository
var ErrNotGitRepository = fmt.Errorf("not inside a git repository")

// RequireGitRepository checks if we're inside a git repository and returns a user-friendly error if not.
// This is used by commands that require a git repository to function properly.
func RequireGitRepository() (string, error) {
	gitRoot, err := FindGitRoot()
	if err != nil {
		return "", fmt.Errorf("secrets-cli requires a Git repository to ensure correct preservation of secrets across your project.\n\nPlease run 'git init' first, or navigate to a directory within an existing Git repository")
	}
	return gitRoot, nil
}
