package cmd

import (
	"fmt"
	"os"

	"github.com/NuevaNext/secrets-cli/internal/config"
	"github.com/NuevaNext/secrets-cli/internal/gpg"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Configure access after cloning a secrets repository",
	Long: `Set up your local environment after cloning a repository containing secrets.

This command:
  1. Verifies your GPG key exists in the stored keys
  2. Imports all stored public keys to your GPG keyring
  3. Lists vaults and shows your access status

Example:
  git clone git@github.com:org/project.git
  cd project
  secrets-cli setup --email you@example.com`,
	RunE: runSetup,
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func runSetup(cmd *cobra.Command, args []string) error {
	secretsDir := GetSecretsDir()
	email := GetUserEmail()

	// Check if secrets directory exists
	if _, err := os.Stat(secretsDir); os.IsNotExist(err) {
		return fmt.Errorf("secrets directory not found: %s. Run 'secrets-cli init' first", secretsDir)
	}

	// Require email
	if email == "" {
		return fmt.Errorf("email is required. Use --email flag or set USER_EMAIL environment variable")
	}

	// Validate email to prevent path traversal and argument injection
	if err := validateName(email); err != nil {
		return err
	}

	// Load config
	cfg, err := config.LoadConfig(secretsDir)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Printf("Setting up secrets for: %s\n", email)
	fmt.Printf("Store owner: %s\n", cfg.Owner)
	fmt.Println()

	// Check if user's key exists in store
	keysDir := config.GetKeysDir(secretsDir)
	keyFile := fmt.Sprintf("%s/%s.asc", keysDir, email)

	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		return fmt.Errorf("your key (%s) is not in the store. Ask an admin to add it", email)
	}

	fmt.Printf("✓ Found your key: %s\n", keyFile)

	// Import all keys
	g := gpg.New(GetGPGBinary())
	imported, err := g.ImportKeyFromDir(keysDir)
	if err != nil {
		return fmt.Errorf("failed to import keys: %w", err)
	}

	fmt.Printf("✓ Imported %d key(s) to your GPG keyring\n", imported)

	// List vaults and check access
	vaults, err := config.ListVaults(secretsDir)
	if err != nil {
		return fmt.Errorf("failed to list vaults: %w", err)
	}

	if len(vaults) > 0 {
		fmt.Println()
		fmt.Println("Available vaults:")
		for _, vault := range vaults {
			vaultDir := config.GetVaultDir(secretsDir, vault)
			vaultCfg, err := config.LoadVaultConfig(vaultDir)
			if err != nil {
				continue
			}

			hasAccess := false
			for _, member := range vaultCfg.Members {
				if member == email {
					hasAccess = true
					break
				}
			}

			if hasAccess {
				fmt.Printf("  ✓ %s (access granted)\n", vault)
			} else {
				fmt.Printf("  ✗ %s (no access)\n", vault)
			}
		}
	}

	fmt.Println()
	fmt.Println("Setup complete!")

	return nil
}
