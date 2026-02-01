package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/NuevaNext/secrets-cli/internal/config"
	"github.com/NuevaNext/secrets-cli/internal/gpg"
	"github.com/NuevaNext/secrets-cli/internal/pass"
	"github.com/spf13/cobra"
)

var vaultCmd = &cobra.Command{
	Use:   "vault",
	Short: "Manage vaults for organizing secrets",
	Long: `Manage vaults for organizing secrets by environment or access level.

Vaults are isolated containers for secrets with independent access control.
Each vault has its own set of members and encrypted password store.

Examples:
  secrets-cli vault list
  secrets-cli vault create dev --description "Development secrets"
  secrets-cli vault info dev
  secrets-cli vault add-member dev alice@example.com
  secrets-cli vault delete dev --force`,
}

var vaultListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all vaults",
	Long: `List all vaults and show your access status (✓/✗).

If --email is set, access status is shown for each vault.`,
	RunE: runVaultList,
}

var vaultCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new vault",
	Long: `Create a new vault with you as the first member.

The vault name should be short and descriptive (e.g., dev, staging, production).
You will be automatically added as the first member.

Examples:
  secrets-cli vault create dev
  secrets-cli vault create production --description "Production credentials"`,
	Args: cobra.ExactArgs(1),
	RunE: runVaultCreate,
}

var vaultInfoCmd = &cobra.Command{
	Use:   "info <vault>",
	Short: "Show vault details and members",
	Long: `Display detailed information about a vault including:
  - Description and creation date
  - Number of secrets
  - List of members with access`,
	Args: cobra.ExactArgs(1),
	RunE: runVaultInfo,
}

var vaultDeleteCmd = &cobra.Command{
	Use:   "delete <vault>",
	Short: "Delete a vault and all its secrets",
	Long: `Permanently delete a vault and all secrets it contains.

This action cannot be undone. Use --force to confirm.`,
	Args: cobra.ExactArgs(1),
	RunE: runVaultDelete,
}

var vaultAddMemberCmd = &cobra.Command{
	Use:   "add-member <vault> <email>",
	Short: "Grant vault access to a team member",
	Long: `Add a member to a vault, granting them read/write access.

The member's GPG key must first be added with 'secrets-cli key add'.
All secrets will be re-encrypted to include the new member.`,
	Args: cobra.ExactArgs(2),
	RunE: runVaultAddMember,
}

var vaultRemoveMemberCmd = &cobra.Command{
	Use:   "remove-member <vault> <email>",
	Short: "Revoke vault access from a team member",
	Long: `Remove a member from a vault, revoking their access.

All secrets will be re-encrypted to exclude the removed member.
Note: The removed member may still have copies of secrets they previously viewed.`,
	Args: cobra.ExactArgs(2),
	RunE: runVaultRemoveMember,
}

var (
	vaultDescription string
	forceDelete      bool
)

func init() {
	rootCmd.AddCommand(vaultCmd)
	vaultCmd.AddCommand(vaultListCmd)
	vaultCmd.AddCommand(vaultCreateCmd)
	vaultCmd.AddCommand(vaultInfoCmd)
	vaultCmd.AddCommand(vaultDeleteCmd)
	vaultCmd.AddCommand(vaultAddMemberCmd)
	vaultCmd.AddCommand(vaultRemoveMemberCmd)

	vaultCreateCmd.Flags().StringVarP(&vaultDescription, "description", "d", "", "Vault description")
	vaultDeleteCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "Force delete without confirmation")
}

func runVaultList(cmd *cobra.Command, args []string) error {
	secretsDir := GetSecretsDir()
	email := GetUserEmail()

	if _, err := os.Stat(secretsDir); os.IsNotExist(err) {
		return fmt.Errorf("✗ Secrets directory not found: %s. Run 'secrets-cli init' first", secretsDir)
	}

	vaults, err := config.ListVaults(secretsDir)
	if err != nil {
		return err
	}

	if len(vaults) == 0 {
		fmt.Println("No vaults found. Create one with: secrets-cli vault create <name>")
		return nil
	}

	fmt.Println("Vaults:")
	for _, vault := range vaults {
		vaultDir := config.GetVaultDir(secretsDir, vault)
		vaultCfg, err := config.LoadVaultConfig(vaultDir)
		if err != nil {
			fmt.Printf("  %s (error loading config)\n", vault)
			continue
		}

		hasAccess := false
		if email != "" {
			for _, member := range vaultCfg.Members {
				if member == email {
					hasAccess = true
					break
				}
			}
		}

		status := ""
		if email != "" {
			if hasAccess {
				status = " ✓"
			} else {
				status = " ✗"
			}
		}

		desc := ""
		if vaultCfg.Description != "" {
			desc = fmt.Sprintf(" - %s", vaultCfg.Description)
		}

		fmt.Printf("  %s%s%s\n", vault, status, desc)
	}

	return nil
}

func runVaultCreate(cmd *cobra.Command, args []string) error {
	secretsDir := GetSecretsDir()
	email := GetUserEmail()
	vaultName := args[0]

	if err := validateName(vaultName); err != nil {
		return err
	}

	if _, err := os.Stat(secretsDir); os.IsNotExist(err) {
		return fmt.Errorf("✗ Secrets directory not found: %s. Run 'secrets-cli init' first", secretsDir)
	}

	if email == "" {
		return fmt.Errorf("email is required. Use --email flag or set USER_EMAIL environment variable")
	}

	// Check vault doesn't exist
	vaultDir := config.GetVaultDir(secretsDir, vaultName)
	if _, err := os.Stat(vaultDir); !os.IsNotExist(err) {
		return fmt.Errorf("vault already exists: %s", vaultName)
	}

	// Check GPG key exists
	g := gpg.New(GetGPGBinary())
	if !g.KeyExists(email) {
		return fmt.Errorf("no GPG key found for %s", email)
	}

	// Create vault directory
	if err := os.MkdirAll(vaultDir, 0755); err != nil {
		return fmt.Errorf("failed to create vault directory: %w", err)
	}

	// Create vault config
	now := time.Now().UTC().Format(time.RFC3339)
	vaultCfg := &config.VaultConfig{
		Name:        vaultName,
		Description: vaultDescription,
		Members:     []string{email},
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := config.SaveVaultConfig(vaultDir, vaultCfg); err != nil {
		os.RemoveAll(vaultDir)
		return fmt.Errorf("failed to create vault config: %w", err)
	}

	// Initialize password store
	storeDir := filepath.Join(vaultDir, ".password-store")
	if err := os.MkdirAll(storeDir, 0700); err != nil {
		os.RemoveAll(vaultDir)
		return fmt.Errorf("failed to create password store: %w", err)
	}

	p := pass.New(storeDir)
	if err := p.Init([]string{email}); err != nil {
		os.RemoveAll(vaultDir)
		return fmt.Errorf("failed to initialize password store: %w", err)
	}

	fmt.Printf("✓ Created vault: %s\n", vaultName)
	if vaultDescription != "" {
		fmt.Printf("  Description: %s\n", vaultDescription)
	}
	fmt.Printf("  Owner: %s\n", email)

	return nil
}

func runVaultInfo(cmd *cobra.Command, args []string) error {
	secretsDir := GetSecretsDir()
	vaultName := args[0]

	vaultDir := config.GetVaultDir(secretsDir, vaultName)
	if _, err := os.Stat(vaultDir); os.IsNotExist(err) {
		return fmt.Errorf("vault not found: %s", vaultName)
	}

	vaultCfg, err := config.LoadVaultConfig(vaultDir)
	if err != nil {
		return fmt.Errorf("failed to load vault config: %w", err)
	}

	// Count secrets
	storeDir := filepath.Join(vaultDir, ".password-store")
	p := pass.New(storeDir)
	secrets, _ := p.List()

	fmt.Printf("Vault: %s\n", vaultCfg.Name)
	if vaultCfg.Description != "" {
		fmt.Printf("Description: %s\n", vaultCfg.Description)
	}
	fmt.Printf("Created: %s\n", vaultCfg.CreatedAt)
	if vaultCfg.UpdatedAt != "" && vaultCfg.UpdatedAt != vaultCfg.CreatedAt {
		fmt.Printf("Updated: %s\n", vaultCfg.UpdatedAt)
	}
	fmt.Printf("Secrets: %d\n", len(secrets))
	fmt.Println()
	fmt.Println("Members:")
	for _, member := range vaultCfg.Members {
		fmt.Printf("  - %s\n", member)
	}

	return nil
}

func runVaultDelete(cmd *cobra.Command, args []string) error {
	secretsDir := GetSecretsDir()
	vaultName := args[0]

	vaultDir := config.GetVaultDir(secretsDir, vaultName)
	if _, err := os.Stat(vaultDir); os.IsNotExist(err) {
		return fmt.Errorf("vault not found: %s", vaultName)
	}

	if !forceDelete {
		return fmt.Errorf("use --force to confirm deletion of vault: %s", vaultName)
	}

	if err := os.RemoveAll(vaultDir); err != nil {
		return fmt.Errorf("failed to delete vault: %w", err)
	}

	fmt.Printf("✓ Deleted vault: %s\n", vaultName)
	return nil
}

func runVaultAddMember(cmd *cobra.Command, args []string) error {
	secretsDir := GetSecretsDir()
	email := GetUserEmail()
	vaultName := args[0]
	memberEmail := args[1]

	vaultDir := config.GetVaultDir(secretsDir, vaultName)
	if _, err := os.Stat(vaultDir); os.IsNotExist(err) {
		return fmt.Errorf("vault not found: %s", vaultName)
	}

	// Load vault config
	vaultCfg, err := config.LoadVaultConfig(vaultDir)
	if err != nil {
		return fmt.Errorf("failed to load vault config: %w", err)
	}

	// Check caller has access (is a member)
	if email != "" {
		hasAccess := false
		for _, member := range vaultCfg.Members {
			if member == email {
				hasAccess = true
				break
			}
		}
		if !hasAccess {
			return fmt.Errorf("access denied: you are not a member of vault %s", vaultName)
		}
	}

	// Check member's key exists
	keysDir := config.GetKeysDir(secretsDir)
	keyFile := filepath.Join(keysDir, memberEmail+".asc")
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		return fmt.Errorf("key not found for %s. Add it with: secrets-cli key add %s", memberEmail, memberEmail)
	}

	// Check not already a member
	for _, m := range vaultCfg.Members {
		if m == memberEmail {
			return fmt.Errorf("%s is already a member of %s", memberEmail, vaultName)
		}
	}

	// Import the member's key to GPG
	g := gpg.New(GetGPGBinary())
	if err := g.ImportKey(keyFile); err != nil {
		return fmt.Errorf("failed to import key: %w", err)
	}

	// Add member
	vaultCfg.Members = append(vaultCfg.Members, memberEmail)
	vaultCfg.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	if err := config.SaveVaultConfig(vaultDir, vaultCfg); err != nil {
		return fmt.Errorf("failed to save vault config: %w", err)
	}

	// Re-encrypt secrets with new member
	storeDir := filepath.Join(vaultDir, ".password-store")
	p := pass.New(storeDir)
	if err := p.ReInit(vaultCfg.Members); err != nil {
		return fmt.Errorf("failed to re-encrypt secrets: %w", err)
	}

	fmt.Printf("✓ Added %s to vault %s\n", memberEmail, vaultName)
	fmt.Printf("✓ Re-encrypted %d secret(s)\n", countSecrets(storeDir))

	return nil
}

func runVaultRemoveMember(cmd *cobra.Command, args []string) error {
	secretsDir := GetSecretsDir()
	email := GetUserEmail()
	vaultName := args[0]
	memberEmail := args[1]

	vaultDir := config.GetVaultDir(secretsDir, vaultName)
	if _, err := os.Stat(vaultDir); os.IsNotExist(err) {
		return fmt.Errorf("vault not found: %s", vaultName)
	}

	// Load vault config
	vaultCfg, err := config.LoadVaultConfig(vaultDir)
	if err != nil {
		return fmt.Errorf("failed to load vault config: %w", err)
	}

	// Check caller has access
	if email != "" {
		hasAccess := false
		for _, member := range vaultCfg.Members {
			if member == email {
				hasAccess = true
				break
			}
		}
		if !hasAccess {
			return fmt.Errorf("access denied: you are not a member of vault %s", vaultName)
		}
	}

	// Check is a member
	memberIndex := -1
	for i, m := range vaultCfg.Members {
		if m == memberEmail {
			memberIndex = i
			break
		}
	}
	if memberIndex == -1 {
		return fmt.Errorf("%s is not a member of %s", memberEmail, vaultName)
	}

	// Cannot remove last member
	if len(vaultCfg.Members) == 1 {
		return fmt.Errorf("cannot remove the last member from a vault")
	}

	// Remove member
	vaultCfg.Members = append(vaultCfg.Members[:memberIndex], vaultCfg.Members[memberIndex+1:]...)
	vaultCfg.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	if err := config.SaveVaultConfig(vaultDir, vaultCfg); err != nil {
		return fmt.Errorf("failed to save vault config: %w", err)
	}

	// Re-encrypt secrets without removed member
	storeDir := filepath.Join(vaultDir, ".password-store")
	p := pass.New(storeDir)
	if err := p.ReInit(vaultCfg.Members); err != nil {
		return fmt.Errorf("failed to re-encrypt secrets: %w", err)
	}

	fmt.Printf("✓ Removed %s from vault %s\n", memberEmail, vaultName)
	fmt.Printf("✓ Re-encrypted %d secret(s)\n", countSecrets(storeDir))

	return nil
}

func countSecrets(storeDir string) int {
	p := pass.New(storeDir)
	secrets, _ := p.List()
	return len(secrets)
}

// hasVaultAccess checks if an email has access to a vault
func hasVaultAccess(secretsDir, vaultName, email string) bool {
	if email == "" {
		return false
	}
	vaultDir := config.GetVaultDir(secretsDir, vaultName)
	vaultCfg, err := config.LoadVaultConfig(vaultDir)
	if err != nil {
		return false
	}
	for _, member := range vaultCfg.Members {
		if strings.EqualFold(member, email) {
			return true
		}
	}
	return false
}
