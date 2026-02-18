package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/NuevaNext/secrets-cli/internal/config"
	"github.com/NuevaNext/secrets-cli/internal/pass"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list <vault>",
	Short: "List all secrets in a vault",
	Long: `List all secrets stored in a vault.

Use --format names to get just secret names (useful for scripting).

Examples:
  secrets-cli list dev
  secrets-cli list production --format names`,
	Args: cobra.ExactArgs(1),
	RunE: runList,
}

var getCmd = &cobra.Command{
	Use:   "get <vault> <secret>",
	Short: "Retrieve and display a secret value",
	Long: `Retrieve and display the decrypted value of a secret.

The secret name can use slashes for organization (e.g., database/password).

Examples:
  secrets-cli get dev database/password
  secrets-cli get production api/key`,
	Args: cobra.ExactArgs(2),
	RunE: runGet,
}

var setCmd = &cobra.Command{
	Use:   "set <vault> <secret> [value]",
	Short: "Set a secret value",
	Long: `Set a secret value. If no value is provided, reads from stdin.

Examples:
  secrets-cli set development database/password "my-password"
  echo "my-password" | secrets-cli set development database/password`,
	Args: cobra.RangeArgs(2, 3),
	RunE: runSet,
}

var deleteCmd = &cobra.Command{
	Use:     "delete <vault> <secret>",
	Aliases: []string{"rm"},
	Short:   "Permanently delete a secret",
	Long: `Permanently delete a secret from a vault.

This action cannot be undone. Use --force to confirm.

Example:
  secrets-cli delete dev temp/test-secret --force`,
	Args: cobra.ExactArgs(2),
	RunE: runDelete,
}

var renameCmd = &cobra.Command{
	Use:     "rename <vault> <old-name> <new-name>",
	Aliases: []string{"mv"},
	Short:   "Rename or move a secret within a vault",
	Long: `Rename a secret or move it to a different path within the same vault.

Example:
  secrets-cli rename dev old/path new/path`,
	Args: cobra.ExactArgs(3),
	RunE: runRename,
}

var copyCmd = &cobra.Command{
	Use:     "copy <src-vault> <secret> <dst-vault>",
	Aliases: []string{"cp"},
	Short:   "Copy a secret to another vault",
	Long: `Copy a secret from one vault to another.

You must have access to both vaults. Use --new-name to rename during copy.

Examples:
  secrets-cli copy dev database/password staging
  secrets-cli copy dev api/key production --new-name api/dev_key_backup`,
	Args: cobra.ExactArgs(3),
	RunE: runCopy,
}

var (
	listFormat    string
	forceSecret   bool
	newSecretName string
)

func init() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(renameCmd)
	rootCmd.AddCommand(copyCmd)

	listCmd.Flags().StringVar(&listFormat, "format", "table", "Output format: table, names")
	deleteCmd.Flags().BoolVarP(&forceSecret, "force", "f", false, "Force delete without confirmation")
	copyCmd.Flags().StringVar(&newSecretName, "new-name", "", "New name for the copied secret")
}

func runList(cmd *cobra.Command, args []string) error {
	secretsDir := GetSecretsDir()
	email := GetUserEmail()
	vaultName := args[0]

	if _, err := os.Stat(secretsDir); os.IsNotExist(err) {
		return fmt.Errorf("✗ Secrets directory not found: %s. Run 'secrets-cli init' first", secretsDir)
	}

	// Check vault exists
	vaultDir := config.GetVaultDir(secretsDir, vaultName)
	if _, err := os.Stat(vaultDir); os.IsNotExist(err) {
		return fmt.Errorf("vault not found: %s", vaultName)
	}

	// Check access
	if !hasVaultAccess(secretsDir, vaultName, email) && email != "" {
		return fmt.Errorf("Access denied: you are not a member of vault %s", vaultName)
	}

	// List secrets
	storeDir := filepath.Join(vaultDir, ".password-store")
	p := pass.New(storeDir)
	secrets, err := p.List()
	if err != nil {
		return fmt.Errorf("failed to list secrets: %w", err)
	}

	if len(secrets) == 0 {
		fmt.Printf("No secrets in vault: %s\n", vaultName)
		return nil
	}

	switch listFormat {
	case "names":
		for _, secret := range secrets {
			fmt.Println(secret)
		}
	default: // table
		fmt.Printf("Secrets in vault '%s':\n", vaultName)
		for _, secret := range secrets {
			fmt.Printf("  %s\n", secret)
		}
	}

	return nil
}

func runGet(cmd *cobra.Command, args []string) error {
	secretsDir := GetSecretsDir()
	email := GetUserEmail()
	vaultName := args[0]
	secretName := args[1]

	if _, err := os.Stat(secretsDir); os.IsNotExist(err) {
		return fmt.Errorf("✗ Secrets directory not found: %s. Run 'secrets-cli init' first", secretsDir)
	}

	// Check vault exists
	vaultDir := config.GetVaultDir(secretsDir, vaultName)
	if _, err := os.Stat(vaultDir); os.IsNotExist(err) {
		return fmt.Errorf("vault not found: %s", vaultName)
	}

	// Check access
	if !hasVaultAccess(secretsDir, vaultName, email) && email != "" {
		return fmt.Errorf("Access denied: you are not a member of vault %s", vaultName)
	}

	// Get secret
	storeDir := filepath.Join(vaultDir, ".password-store")
	p := pass.New(storeDir)

	if !p.Exists(secretName) {
		return fmt.Errorf("secret not found: %s/%s", vaultName, secretName)
	}

	value, err := p.Show(secretName)
	if err != nil {
		return fmt.Errorf("failed to get secret: %w", err)
	}

	fmt.Println(value)
	return nil
}

func runSet(cmd *cobra.Command, args []string) error {
	secretsDir := GetSecretsDir()
	email := GetUserEmail()
	vaultName := args[0]
	secretName := args[1]

	if _, err := os.Stat(secretsDir); os.IsNotExist(err) {
		return fmt.Errorf("✗ Secrets directory not found: %s. Run 'secrets-cli init' first", secretsDir)
	}

	// Check vault exists
	vaultDir := config.GetVaultDir(secretsDir, vaultName)
	if _, err := os.Stat(vaultDir); os.IsNotExist(err) {
		return fmt.Errorf("vault not found: %s", vaultName)
	}

	// Check access
	if !hasVaultAccess(secretsDir, vaultName, email) && email != "" {
		return fmt.Errorf("Access denied: you are not a member of vault %s", vaultName)
	}

	// Get value
	var value string
	if len(args) > 2 {
		value = args[2]
	} else {
		// Read from stdin
		reader := bufio.NewReader(os.Stdin)
		data, err := reader.ReadString('\n')
		if err != nil && err.Error() != "EOF" {
			// Try reading without newline
			data, err = reader.ReadString('\000')
			if err != nil && err.Error() != "EOF" {
				return fmt.Errorf("failed to read from stdin: %w", err)
			}
		}
		value = strings.TrimSuffix(data, "\n")
	}

	if value == "" {
		return fmt.Errorf("empty secret value not allowed")
	}

	// Set secret
	storeDir := filepath.Join(vaultDir, ".password-store")
	p := pass.New(storeDir)

	if err := p.Insert(secretName, value); err != nil {
		return fmt.Errorf("failed to set secret: %w", err)
	}

	fmt.Printf("✓ Set secret: %s/%s\n", vaultName, secretName)
	return nil
}

func runDelete(cmd *cobra.Command, args []string) error {
	secretsDir := GetSecretsDir()
	email := GetUserEmail()
	vaultName := args[0]
	secretName := args[1]

	if _, err := os.Stat(secretsDir); os.IsNotExist(err) {
		return fmt.Errorf("✗ Secrets directory not found: %s. Run 'secrets-cli init' first", secretsDir)
	}

	// Check vault exists
	vaultDir := config.GetVaultDir(secretsDir, vaultName)
	if _, err := os.Stat(vaultDir); os.IsNotExist(err) {
		return fmt.Errorf("vault not found: %s", vaultName)
	}

	// Check access
	if !hasVaultAccess(secretsDir, vaultName, email) && email != "" {
		return fmt.Errorf("Access denied: you are not a member of vault %s", vaultName)
	}

	if !Confirm(fmt.Sprintf("Are you sure you want to delete secret %s/%s?", vaultName, secretName), forceSecret) {
		return fmt.Errorf("deletion of secret %s/%s cancelled", vaultName, secretName)
	}

	// Delete secret
	storeDir := filepath.Join(vaultDir, ".password-store")
	p := pass.New(storeDir)

	if err := p.Remove(secretName); err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	fmt.Printf("✓ Deleted secret: %s/%s\n", vaultName, secretName)
	return nil
}

func runRename(cmd *cobra.Command, args []string) error {
	secretsDir := GetSecretsDir()
	email := GetUserEmail()
	vaultName := args[0]
	oldName := args[1]
	newName := args[2]

	if _, err := os.Stat(secretsDir); os.IsNotExist(err) {
		return fmt.Errorf("✗ Secrets directory not found: %s. Run 'secrets-cli init' first", secretsDir)
	}

	// Check vault exists
	vaultDir := config.GetVaultDir(secretsDir, vaultName)
	if _, err := os.Stat(vaultDir); os.IsNotExist(err) {
		return fmt.Errorf("vault not found: %s", vaultName)
	}

	// Check access
	if !hasVaultAccess(secretsDir, vaultName, email) && email != "" {
		return fmt.Errorf("Access denied: you are not a member of vault %s", vaultName)
	}

	// Rename secret
	storeDir := filepath.Join(vaultDir, ".password-store")
	p := pass.New(storeDir)

	if !p.Exists(oldName) {
		return fmt.Errorf("secret not found: %s/%s", vaultName, oldName)
	}

	if err := p.Move(oldName, newName); err != nil {
		return fmt.Errorf("failed to rename secret: %w", err)
	}

	fmt.Printf("✓ Renamed secret: %s/%s -> %s/%s\n", vaultName, oldName, vaultName, newName)
	return nil
}

func runCopy(cmd *cobra.Command, args []string) error {
	secretsDir := GetSecretsDir()
	email := GetUserEmail()
	srcVault := args[0]
	secretName := args[1]
	dstVault := args[2]

	if _, err := os.Stat(secretsDir); os.IsNotExist(err) {
		return fmt.Errorf("✗ Secrets directory not found: %s. Run 'secrets-cli init' first", secretsDir)
	}

	// Check source vault exists and access
	srcVaultDir := config.GetVaultDir(secretsDir, srcVault)
	if _, err := os.Stat(srcVaultDir); os.IsNotExist(err) {
		return fmt.Errorf("source vault not found: %s", srcVault)
	}

	if !hasVaultAccess(secretsDir, srcVault, email) && email != "" {
		return fmt.Errorf("Access denied: you are not a member of vault %s", srcVault)
	}

	// Check destination vault exists and access
	dstVaultDir := config.GetVaultDir(secretsDir, dstVault)
	if _, err := os.Stat(dstVaultDir); os.IsNotExist(err) {
		return fmt.Errorf("destination vault not found: %s", dstVault)
	}

	if !hasVaultAccess(secretsDir, dstVault, email) && email != "" {
		return fmt.Errorf("Access denied: you are not a member of vault %s", dstVault)
	}

	// Get source secret
	srcStoreDir := filepath.Join(srcVaultDir, ".password-store")
	srcPass := pass.New(srcStoreDir)

	if !srcPass.Exists(secretName) {
		return fmt.Errorf("secret not found: %s/%s", srcVault, secretName)
	}

	value, err := srcPass.Show(secretName)
	if err != nil {
		return fmt.Errorf("failed to read source secret: %w", err)
	}

	// Set in destination
	dstStoreDir := filepath.Join(dstVaultDir, ".password-store")
	dstPass := pass.New(dstStoreDir)

	dstSecretName := secretName
	if newSecretName != "" {
		dstSecretName = newSecretName
	}

	if err := dstPass.Insert(dstSecretName, value); err != nil {
		return fmt.Errorf("failed to copy secret to destination: %w", err)
	}

	fmt.Printf("✓ Copied secret: %s/%s -> %s/%s\n", srcVault, secretName, dstVault, dstSecretName)
	return nil
}
