package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	secretsDir string
	userEmail  string
	gpgBinary  string
	verbose    bool

	// Version info
	versionInfo struct {
		Version string
		Commit  string
		Date    string
	}
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "secrets-cli",
	Short: "GPG-based secrets management for Git repositories",
	Long: `secrets-cli is a multi-user secrets management utility that uses GPG 
encryption and the pass password manager to securely store and share 
secrets within a Git repository.

Features:
  • Vault-based organization with access control
  • GPG encryption for all secrets
  • Multi-user support with key management
  • Export to env, dotenv, or JSON formats
  • Automatic re-encryption when members change

Quick Start:
  secrets-cli init --email you@example.com
  secrets-cli vault create dev
  secrets-cli set dev database/password "secret123"
  secrets-cli get dev database/password

For more information, visit: https://github.com/NuevaNext/secrets-cli`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// SetVersionInfo sets version information from build flags
func SetVersionInfo(version, commit, date string) {
	versionInfo.Version = version
	versionInfo.Commit = commit
	versionInfo.Date = date
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&secretsDir, "secrets-dir", ".secrets", "Path to secrets directory")
	rootCmd.PersistentFlags().StringVar(&userEmail, "email", "", "User email for GPG operations")
	rootCmd.PersistentFlags().StringVar(&gpgBinary, "gpg-binary", "gpg", "Path to GPG binary")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	// Version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("secrets-cli %s\n", versionInfo.Version)
			fmt.Printf("  commit: %s\n", versionInfo.Commit)
			fmt.Printf("  built:  %s\n", versionInfo.Date)
		},
	})
}

// GetSecretsDir returns the secrets directory path
func GetSecretsDir() string {
	if secretsDir == "" {
		if envDir := os.Getenv("SECRETS_DIR"); envDir != "" {
			return envDir
		}
		return ".secrets"
	}
	return secretsDir
}

// GetUserEmail returns the user email, auto-detecting if not explicitly set
func GetUserEmail() string {
	if userEmail != "" {
		return userEmail
	}
	if envEmail := os.Getenv("USER_EMAIL"); envEmail != "" {
		return envEmail
	}
	// Auto-detect email
	return detectUserEmail()
}

// detectUserEmail tries to detect email from git config or GPG keys
func detectUserEmail() string {
	// Try git config
	cmd := exec.Command("git", "config", "--get", "user.email")
	if output, err := cmd.Output(); err == nil {
		email := strings.TrimSpace(string(output))
		if email != "" {
			return email
		}
	}

	// Try GPG default key
	gpgBin := GetGPGBinary()
	cmd = exec.Command(gpgBin, "--list-secret-keys", "--keyid-format", "long")
	if output, err := cmd.Output(); err == nil {
		// Parse email from GPG output
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "uid") {
				// Extract email from uid line
				if start := strings.LastIndex(line, "<"); start != -1 {
					if end := strings.LastIndex(line, ">"); end > start {
						return line[start+1 : end]
					}
				}
			}
		}
	}

	return ""
}

// GetGPGBinary returns the GPG binary path
func GetGPGBinary() string {
	if gpgBinary == "" {
		if envGPG := os.Getenv("GPG_BINARY"); envGPG != "" {
			return envGPG
		}
		return "gpg"
	}
	return gpgBinary
}

// IsVerbose returns whether verbose mode is enabled
func IsVerbose() bool {
	return verbose
}

// validateName ensures a name is safe to use in file paths and command arguments
func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	// Prevent path traversal and argument injection
	if strings.Contains(name, "..") || strings.ContainsAny(name, "/\\") || strings.HasPrefix(name, "-") {
		return fmt.Errorf("invalid name: %s (contains illegal characters or path traversal)", name)
	}
	return nil
}
