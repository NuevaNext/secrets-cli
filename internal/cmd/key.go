package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/NuevaNext/secrets-cli/internal/config"
	"github.com/NuevaNext/secrets-cli/internal/gpg"
	"github.com/spf13/cobra"
)

var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "Manage GPG public keys for team members",
	Long: `Manage GPG public keys stored in the secrets repository.

Keys are stored in .secrets/keys/ and must be added before a team member 
can be added to a vault.

Examples:
  secrets-cli key list
  secrets-cli key add alice@example.com
  secrets-cli key add bob@example.com --key-file ./bob.asc
  secrets-cli key import`,
}

var keyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all stored public keys",
	Long: `List all GPG public keys stored in the secrets repository.

These keys can be used to add members to vaults.`,
	RunE: runKeyList,
}

var keyAddCmd = &cobra.Command{
	Use:   "add <email>",
	Short: "Add a team member's public key",
	Long: `Add a team member's public key to the secrets repository.

If the key exists in your GPG keyring, it will be exported automatically.
Otherwise, use --key-file to specify an ASCII-armored key file.

Examples:
  secrets-cli key add alice@example.com                # Export from GPG keyring
  secrets-cli key add bob@example.com --key-file ./bob.asc  # From file`,
	Args: cobra.ExactArgs(1),
	RunE: runKeyAdd,
}

var keyRemoveCmd = &cobra.Command{
	Use:   "remove <email>",
	Short: "Remove a public key from the store",
	Long: `Remove a team member's public key from the secrets repository.

Note: This does not revoke access to vaults. Use 'vault remove-member' first.`,
	Args: cobra.ExactArgs(1),
	RunE: runKeyRemove,
}

var keyImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import all stored keys to your GPG keyring",
	Long: `Import all stored public keys into your local GPG keyring.

This is typically run after cloning a repository with secrets, or is 
called automatically by 'secrets-cli setup'.`,
	RunE: runKeyImport,
}

var keyFile string

func init() {
	rootCmd.AddCommand(keyCmd)
	keyCmd.AddCommand(keyListCmd)
	keyCmd.AddCommand(keyAddCmd)
	keyCmd.AddCommand(keyRemoveCmd)
	keyCmd.AddCommand(keyImportCmd)

	keyAddCmd.Flags().StringVar(&keyFile, "key-file", "", "Path to key file (optional)")
}

func runKeyList(cmd *cobra.Command, args []string) error {
	secretsDir := GetSecretsDir()

	if _, err := os.Stat(secretsDir); os.IsNotExist(err) {
		return fmt.Errorf("✗ Secrets directory not found: %s. Run 'secrets-cli init' first", secretsDir)
	}

	keysDir := config.GetKeysDir(secretsDir)
	entries, err := os.ReadDir(keysDir)
	if err != nil {
		return fmt.Errorf("failed to read keys directory: %w", err)
	}

	fmt.Println("Stored public keys:")
	count := 0
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".asc" {
			email := entry.Name()[:len(entry.Name())-4] // Remove .asc
			fmt.Printf("  %s\n", email)
			count++
		}
	}

	if count == 0 {
		fmt.Println("  (none)")
	}

	return nil
}

func runKeyAdd(cmd *cobra.Command, args []string) error {
	secretsDir := GetSecretsDir()
	email := args[0]

	if err := validateName(email); err != nil {
		return err
	}

	if _, err := os.Stat(secretsDir); os.IsNotExist(err) {
		return fmt.Errorf("✗ Secrets directory not found: %s. Run 'secrets-cli init' first", secretsDir)
	}

	keysDir := config.GetKeysDir(secretsDir)
	keyPath := filepath.Join(keysDir, email+".asc")

	// Check if already exists
	if _, err := os.Stat(keyPath); !os.IsNotExist(err) {
		return fmt.Errorf("key already exists for %s", email)
	}

	g := gpg.New(GetGPGBinary())

	if keyFile != "" {
		// Copy from specified file
		data, err := os.ReadFile(keyFile)
		if err != nil {
			return fmt.Errorf("failed to read key file: %w", err)
		}
		if err := os.WriteFile(keyPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write key: %w", err)
		}
	} else {
		// Export from GPG keyring
		if !g.KeyExists(email) {
			return fmt.Errorf("no GPG key found for %s. Use --key-file to specify a key file", email)
		}
		if err := g.ExportPublicKeyToFile(email, keyPath); err != nil {
			return fmt.Errorf("failed to export key: %w", err)
		}
	}

	fmt.Printf("✓ Added key for %s\n", email)
	return nil
}

func runKeyRemove(cmd *cobra.Command, args []string) error {
	secretsDir := GetSecretsDir()
	email := args[0]

	if _, err := os.Stat(secretsDir); os.IsNotExist(err) {
		return fmt.Errorf("✗ Secrets directory not found: %s. Run 'secrets-cli init' first", secretsDir)
	}

	keysDir := config.GetKeysDir(secretsDir)
	keyPath := filepath.Join(keysDir, email+".asc")

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return fmt.Errorf("no key found for %s", email)
	}

	if err := os.Remove(keyPath); err != nil {
		return fmt.Errorf("failed to remove key: %w", err)
	}

	fmt.Printf("✓ Removed key for %s\n", email)
	return nil
}

func runKeyImport(cmd *cobra.Command, args []string) error {
	secretsDir := GetSecretsDir()

	if _, err := os.Stat(secretsDir); os.IsNotExist(err) {
		return fmt.Errorf("✗ Secrets directory not found: %s. Run 'secrets-cli init' first", secretsDir)
	}

	keysDir := config.GetKeysDir(secretsDir)
	g := gpg.New(GetGPGBinary())

	imported, err := g.ImportKeyFromDir(keysDir)
	if err != nil {
		return fmt.Errorf("failed to import keys: %w", err)
	}

	fmt.Printf("✓ Imported %d key(s) to GPG keyring\n", imported)
	return nil
}
