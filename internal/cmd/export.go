package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/NuevaNext/secrets-cli/internal/config"
	"github.com/NuevaNext/secrets-cli/internal/pass"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export <vault>",
	Short: "Export secrets as environment variables",
	Long: `Export secrets from a vault in various formats.

Formats:
  env    - Shell export format: export VAR=value
  dotenv - Dotenv format: VAR=value
  json   - JSON object: {"key": "value"}`,
	Args: cobra.ExactArgs(1),
	RunE: runExport,
}

var syncCmd = &cobra.Command{
	Use:   "sync <vault>",
	Short: "Synchronize and verify vault integrity",
	Long: `Synchronize a vault by verifying integrity and re-encrypting if needed.

This ensures that all secrets are encrypted for all current members.`,
	Args: cobra.ExactArgs(1),
	RunE: runSync,
}

var (
	exportFormat string
	exportPrefix string
)

func init() {
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(syncCmd)

	exportCmd.Flags().StringVar(&exportFormat, "format", "env", "Output format: env, dotenv, json")
	exportCmd.Flags().StringVar(&exportPrefix, "prefix", "", "Prefix for variable names")
}

func runExport(cmd *cobra.Command, args []string) error {
	secretsDir := GetSecretsDir()
	email := GetUserEmail()
	vaultName := args[0]

	// Validate vault name to prevent path traversal and argument injection
	if err := validateName(vaultName); err != nil {
		return err
	}

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

	// Get all secrets
	storeDir := filepath.Join(vaultDir, ".password-store")
	p := pass.New(storeDir)
	secrets, err := p.List()
	if err != nil {
		return fmt.Errorf("failed to list secrets: %w", err)
	}

	// Export based on format
	switch exportFormat {
	case "json":
		fmt.Println("{")
		for i, secret := range secrets {
			value, err := p.Show(secret)
			if err != nil {
				continue
			}
			// Escape JSON
			value = strings.ReplaceAll(value, "\\", "\\\\")
			value = strings.ReplaceAll(value, "\"", "\\\"")

			comma := ","
			if i == len(secrets)-1 {
				comma = ""
			}
			fmt.Printf("  \"%s%s\": \"%s\"%s\n", exportPrefix, secretToEnvName(secret), value, comma)
		}
		fmt.Println("}")

	case "dotenv":
		for _, secret := range secrets {
			value, err := p.Show(secret)
			if err != nil {
				continue
			}
			fmt.Printf("%s%s=%s\n", exportPrefix, secretToEnvName(secret), value)
		}

	default: // env
		for _, secret := range secrets {
			value, err := p.Show(secret)
			if err != nil {
				continue
			}
			fmt.Printf("export %s%s=%s\n", exportPrefix, secretToEnvName(secret), quoteForShell(value))
		}
	}

	return nil
}

func runSync(cmd *cobra.Command, args []string) error {
	secretsDir := GetSecretsDir()
	email := GetUserEmail()
	vaultName := args[0]

	// Validate vault name to prevent path traversal and argument injection
	if err := validateName(vaultName); err != nil {
		return err
	}

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

	// Load vault config
	vaultCfg, err := config.LoadVaultConfig(vaultDir)
	if err != nil {
		return fmt.Errorf("failed to load vault config: %w", err)
	}

	// Re-init password store with current members
	storeDir := filepath.Join(vaultDir, ".password-store")
	p := pass.New(storeDir)

	secrets, _ := p.List()
	fmt.Printf("Synchronizing vault: %s\n", vaultName)
	fmt.Printf("  Members: %d\n", len(vaultCfg.Members))
	fmt.Printf("  Secrets: %d\n", len(secrets))

	if err := p.ReInit(vaultCfg.Members); err != nil {
		return fmt.Errorf("failed to re-encrypt secrets: %w", err)
	}

	// Update timestamp
	vaultCfg.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := config.SaveVaultConfig(vaultDir, vaultCfg); err != nil {
		return fmt.Errorf("failed to save vault config: %w", err)
	}

	fmt.Printf("✓ Synchronized vault: %s\n", vaultName)
	return nil
}

// secretToEnvName converts a secret path to an environment variable name
// e.g., "database/password" -> "DATABASE_PASSWORD"
func secretToEnvName(secret string) string {
	name := strings.ToUpper(secret)
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, ".", "_")
	return name
}

// quoteForShell quotes a value for safe use in shell
func quoteForShell(value string) string {
	// If no special characters, return as-is
	if !strings.ContainsAny(value, " \t\n'\"$`\\!#&|;<>(){}[]") {
		return value
	}
	// Use single quotes and escape any single quotes in the value
	escaped := strings.ReplaceAll(value, "'", "'\"'\"'")
	return "'" + escaped + "'"
}
