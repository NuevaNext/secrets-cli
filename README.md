# secrets-cli

A multi-user secrets management utility that uses GPG encryption and the `pass` password manager to securely store and share secrets within a Git repository.

## Features

- **Vault-based organization** — Group secrets by environment (dev, staging, production) or access level
- **GPG encryption** — All secrets encrypted using team members' GPG public keys
- **Multi-user access control** — Add or remove team members from individual vaults
- **Automatic re-encryption** — Secrets automatically re-encrypted when membership changes
- **Export formats** — Export secrets as shell variables, dotenv, or JSON
- **Git-friendly** — Designed to be committed alongside your code

## Requirements

- [GPG](https://gnupg.org/) (GnuPG 2.x recommended)
- [pass](https://www.passwordstore.org/) (the standard Unix password manager)
- Go 1.22+ (for building from source)

## Installation

### From GitHub Releases

Download the latest binary from the [Releases](https://github.com/NuevaNext/secrets-cli/releases) page:

```bash
# Linux (amd64)
curl -Lo secrets-cli https://github.com/NuevaNext/secrets-cli/releases/latest/download/secrets-cli-linux-amd64
chmod +x secrets-cli
sudo mv secrets-cli /usr/local/bin/

# macOS (Apple Silicon)
curl -Lo secrets-cli https://github.com/NuevaNext/secrets-cli/releases/latest/download/secrets-cli-darwin-arm64
chmod +x secrets-cli
sudo mv secrets-cli /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/NuevaNext/secrets-cli@latest
```

### From Source

```bash
git clone https://github.com/NuevaNext/secrets-cli.git
cd secrets-cli
make build
sudo mv secrets-cli /usr/local/bin/
```

## Quick Start

### 1. Initialize a secrets store

```bash
# Ensure you have a GPG key
gpg --list-secret-keys

# If not, generate one
gpg --gen-key

# Initialize the secrets store
secrets-cli init --email you@example.com
```

### 2. Create vaults and add secrets

```bash
# Create a vault
secrets-cli vault create dev --description "Development environment"

# Add secrets
secrets-cli set dev database/password "super-secret-123"
secrets-cli set dev api/key "sk-abc123xyz"

# View secrets
secrets-cli list dev
secrets-cli get dev database/password
```

### 3. Share with team members

```bash
# Add a team member's key
secrets-cli key add alice@example.com

# Grant vault access
secrets-cli vault add-member dev alice@example.com
```

### 4. Export for use

```bash
# Shell format
secrets-cli export dev --format env

# Dotenv format
secrets-cli export dev --format dotenv > .env

# JSON format
secrets-cli export dev --format json
```

## direnv Integration

[direnv](https://direnv.net/) can automatically load secrets as environment variables when you `cd` into your project, and unload them when you leave.

### Setup

1. [Install direnv](https://direnv.net/docs/installation.html) and add the shell hook to your profile
2. Create an `.envrc` file in your project root:

```bash
# .envrc
export DATABASE_PASSWORD="$(secrets-cli get dev database/password)"
export API_KEY="$(secrets-cli get dev api/key)"
```

3. Allow the file:

```bash
direnv allow
```

From now on, secrets are loaded automatically when you enter the project directory and unloaded when you leave. No `--email` flag is needed — secrets-cli auto-detects your identity from `git config user.email`.

> **Tip:** Add `.envrc` to `.gitignore` so each team member can choose which secrets to load.

## Commands

| Command | Description |
|---------|-------------|
| `init` | Initialize a new secrets store |
| `setup` | Configure access after cloning a repository |
| `vault list` | List all vaults |
| `vault create <name>` | Create a new vault |
| `vault info <vault>` | Show vault details |
| `vault delete <vault>` | Delete a vault |
| `vault add-member <vault> <email>` | Grant vault access |
| `vault remove-member <vault> <email>` | Revoke vault access |
| `key list` | List stored public keys |
| `key add <email>` | Add a team member's key |
| `key remove <email>` | Remove a key |
| `key import` | Import all keys to GPG |
| `list <vault>` | List secrets in a vault |
| `get <vault> <secret>` | Retrieve a secret |
| `set <vault> <secret> [value]` | Set a secret |
| `delete <vault> <secret>` | Delete a secret |
| `rename <vault> <old> <new>` | Rename a secret |
| `copy <src> <secret> <dst>` | Copy a secret to another vault |
| `export <vault>` | Export secrets |
| `sync <vault>` | Re-encrypt vault secrets |

Use `secrets-cli <command> --help` for detailed usage information.

## Team Workflow

### For Repository Owners

1. Initialize the store: `secrets-cli init --email owner@company.com`
2. Create vaults: `secrets-cli vault create production`
3. Add secrets: `secrets-cli set production database/url "postgresql://..."`
4. Add team members' keys: `secrets-cli key add developer@company.com`
5. Grant access: `secrets-cli vault add-member production developer@company.com`
6. Commit: `git add .secrets && git commit -m "Add production secrets"`

### For Team Members

1. Clone the repository: `git clone git@github.com:org/repo.git`
2. Set up access: `secrets-cli setup --email developer@company.com`
3. View secrets: `secrets-cli get production database/url`

## Configuration

### Global Flags

| Flag | Environment Variable | Description |
|------|---------------------|-------------|
| `--secrets-dir` | `SECRETS_DIR` | Path to secrets directory (default: `.secrets`) |
| `--email` | `USER_EMAIL` | Your email for GPG operations |
| `--gpg-binary` | `GPG_BINARY` | Path to GPG binary (default: `gpg`) |
| `--verbose`, `-v` | `VERBOSE` | Enable verbose output |

### Auto-detection

If `--email` is not provided, secrets-cli will attempt to detect your email from:
1. Git configuration (`git config user.email`)
2. GPG default secret key

---

## Developer Guide

### Building from Source

```bash
git clone https://github.com/NuevaNext/secrets-cli.git
cd secrets-cli

# Build
make build

# Build with specific version
make build VERSION=v1.0.0

# Build for all platforms
make build-all
```

### Running Tests

```bash
# Unit tests
make test

# End-to-end tests (requires Docker)
make test-e2e

# Verbose e2e tests
make test-e2e-verbose
```

### Project Structure

```
secrets-cli/
├── cmd/secrets-cli/          # CLI entrypoint
├── internal/
│   ├── cmd/                  # Cobra command implementations
│   ├── config/               # YAML configuration handling
│   ├── gpg/                  # GPG wrapper
│   └── pass/                 # pass wrapper
├── tests/
│   ├── Dockerfile            # E2E test environment
│   ├── e2e-tests.sh          # Test runner
│   └── test-utils.sh         # Test utilities
├── Makefile
├── go.mod
└── README.md
```

### Code Formatting

```bash
make fmt
make lint
```

## License

MIT License - see [LICENSE](LICENSE) for details.
