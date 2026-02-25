package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/NuevaNext/secrets-cli/internal/config"
	"github.com/NuevaNext/secrets-cli/internal/gpg"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new secrets store in the current directory",
	Long: `Initialize a new secrets store in the current directory.

This command creates the .secrets directory structure containing:
  - config.yaml with store configuration
  - keys/ directory for GPG public keys
  - vaults/ directory for secret vaults

You must have a GPG key pair for your email address. If not, create one with:
  gpg --gen-key

Examples:
  secrets-cli init --email you@example.com
  secrets-cli init --email you@example.com --secrets-dir ./my-secrets`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	// Require git repository for proper secrets preservation
	gitRoot, err := RequireGitRepository()
	if err != nil {
		return err
	}
	if IsVerbose() {
		fmt.Printf("Git repository root: %s\n", gitRoot)
	}

	secretsDir := GetSecretsDir()
	email := GetUserEmail()

	// Check if already initialized
	if _, err := os.Stat(secretsDir); !os.IsNotExist(err) {
		return fmt.Errorf("secrets directory already exists: %s. Remove it first or use a different path", secretsDir)
	}

	// Require email
	if email == "" {
		return fmt.Errorf("email is required. Use --email flag or set USER_EMAIL environment variable")
	}

	// Validate email to prevent path traversal and argument injection
	if err := validateName(email); err != nil {
		return err
	}

	// Check GPG key exists
	g := gpg.New(GetGPGBinary())
	if !g.KeyExists(email) {
		return fmt.Errorf("no GPG key found for %s. Generate one with: gpg --gen-key", email)
	}

	// Create directory structure
	dirs := []string{
		secretsDir,
		filepath.Join(secretsDir, "keys"),
		filepath.Join(secretsDir, "vaults"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create config.yaml
	cfg := &config.Config{
		Version: "1",
		Owner:   email,
	}
	if err := config.SaveConfig(secretsDir, cfg); err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	// Export owner's public key
	keyPath := filepath.Join(secretsDir, "keys", email+".asc")
	if err := g.ExportPublicKeyToFile(email, keyPath); err != nil {
		return fmt.Errorf("failed to export public key: %w", err)
	}

	fmt.Printf("✓ Initialized secrets store in %s\n", secretsDir)
	fmt.Printf("✓ Exported your public key to %s\n", keyPath)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Create a vault:  secrets-cli vault create <name>")
	fmt.Println("  2. Add a secret:    secrets-cli set <vault> <secret>")
	fmt.Println("  3. Commit to git:   git add .secrets && git commit")

	return nil
}

// Helper to get current time in ISO format
func nowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}
