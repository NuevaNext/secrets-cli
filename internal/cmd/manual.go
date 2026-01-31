package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var manualCmd = &cobra.Command{
	Use:   "manual",
	Short: "Display comprehensive documentation",
	Long:  `Display detailed man-like documentation for secrets-cli.`,
	Run:   runManual,
}

func init() {
	rootCmd.AddCommand(manualCmd)
}

func runManual(cmd *cobra.Command, args []string) {
	manual := `
NAME
    secrets-cli - Multi-user secrets management for Git repositories

SYNOPSIS
    secrets-cli <command> [options] [arguments]

DESCRIPTION
    secrets-cli is a command-line utility for managing encrypted secrets in
    Git repositories. It uses GPG encryption and the pass password manager
    to securely store and share secrets among team members.

    Secrets are organized into vaults, which are isolated containers with
    independent access control. Each vault can have multiple members, and
    all secrets are encrypted for all vault members.

COMMANDS
    init
        Initialize a new secrets store in the current directory. Creates
        the .secrets/ directory structure and exports your GPG public key.

        secrets-cli init --email you@example.com

    setup
        Configure access after cloning a repository with secrets. Imports
        all stored public keys and verifies your access to vaults.

        git clone git@github.com:org/project.git
        cd project
        secrets-cli setup --email you@example.com

    vault list
        List all vaults. Shows access status (✓/✗) for your email.

    vault create <name>
        Create a new vault. You are automatically added as the first member.

        secrets-cli vault create dev
        secrets-cli vault create production --description "Prod secrets"

    vault info <vault>
        Display vault details including description, member list, and
        number of secrets.

    vault delete <vault>
        Delete a vault and all its secrets. Requires --force flag.

        secrets-cli vault delete old-vault --force

    vault add-member <vault> <email>
        Grant a team member access to a vault. Their GPG key must first
        be added with 'key add'. All secrets are re-encrypted.

        secrets-cli vault add-member dev alice@example.com

    vault remove-member <vault> <email>
        Revoke a member's access. All secrets are re-encrypted to exclude
        the removed member.

        secrets-cli vault remove-member dev bob@example.com

    key list
        List all GPG public keys stored in the repository.

    key add <email>
        Add a team member's public key. If the key exists in your GPG
        keyring, it is exported automatically. Otherwise use --key-file.

        secrets-cli key add alice@example.com
        secrets-cli key add bob@example.com --key-file bob.asc

    key remove <email>
        Remove a public key from the store. Note: this does not revoke
        vault access. Use 'vault remove-member' first.

    key import
        Import all stored public keys into your local GPG keyring.

    list <vault>
        List all secrets in a vault.

        secrets-cli list dev
        secrets-cli list production --format names

    get <vault> <secret>
        Retrieve and display a secret value.

        secrets-cli get dev database/password
        secrets-cli get production api/stripe-key

    set <vault> <secret> [value]
        Store a secret. If value is omitted, reads from stdin.

        secrets-cli set dev database/password "my-secret"
        echo "secret123" | secrets-cli set dev api/key

    delete <vault> <secret>
        Delete a secret. Requires --force flag.

        secrets-cli delete dev old/secret --force

    rename <vault> <old> <new>
        Rename or move a secret within a vault.

        secrets-cli rename dev old/path new/path

    copy <src-vault> <secret> <dst-vault>
        Copy a secret to another vault. Use --new-name to rename.

        secrets-cli copy dev database/password staging
        secrets-cli copy dev api/key production --new-name api/dev-backup

    export <vault>
        Export all secrets from a vault in various formats.

        secrets-cli export dev                    # Shell format
        secrets-cli export dev --format dotenv    # .env format
        secrets-cli export dev --format json      # JSON format
        secrets-cli export dev --prefix APP_      # Add prefix

    sync <vault>
        Re-encrypt all secrets for current vault members. Use after
        membership changes or to verify vault integrity.

        secrets-cli sync production

    version
        Display version, commit hash, and build date.

GLOBAL OPTIONS
    --secrets-dir <path>
        Path to secrets directory. Default: .secrets
        Environment: SECRETS_DIR

    --email <email>
        Your email address for GPG operations.
        Environment: USER_EMAIL
        Auto-detected from: git config, GPG default key

    --gpg-binary <path>
        Path to GPG binary. Default: gpg
        Environment: GPG_BINARY

    -v, --verbose
        Enable verbose output.
        Environment: VERBOSE

DIRECTORY STRUCTURE
    .secrets/
    ├── config.yaml           # Store configuration
    ├── keys/                 # GPG public keys
    │   ├── alice@example.com.asc
    │   └── bob@example.com.asc
    └── vaults/
        ├── dev/
        │   ├── config.yaml   # Vault config (members, etc.)
        │   └── .password-store/  # Encrypted secrets
        └── production/
            ├── config.yaml
            └── .password-store/

EXAMPLES
    Initialize and create first vault:
        $ secrets-cli init --email admin@company.com
        $ secrets-cli vault create dev
        $ secrets-cli set dev database/password "super-secret"

    Add a team member:
        $ secrets-cli key add developer@company.com
        $ secrets-cli vault add-member dev developer@company.com
        $ git add .secrets && git commit -m "Add developer to dev vault"

    Team member setup after clone:
        $ git clone git@github.com:company/project.git
        $ cd project
        $ secrets-cli setup --email developer@company.com
        $ secrets-cli get dev database/password

    Export to .env file:
        $ secrets-cli export dev --format dotenv > .env

    Copy secrets between environments:
        $ secrets-cli copy dev database/password staging
        $ secrets-cli copy dev database/password production

SECURITY NOTES
    • Secrets are encrypted using GPG and can only be decrypted by vault
      members who possess the corresponding private keys.

    • When a member is removed, secrets are re-encrypted but the removed
      member may retain copies of previously accessed secrets.

    • The .secrets directory is designed to be committed to Git. Only
      encrypted data and public keys are stored.

    • Private GPG keys are never stored in the repository.

SEE ALSO
    gpg(1), pass(1)

    Project: https://github.com/NuevaNext/secrets-cli
`
	fmt.Print(manual)
}
