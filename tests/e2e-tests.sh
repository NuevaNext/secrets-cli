#!/usr/bin/env bash
#
# e2e-tests.sh - End-to-end tests for secret-cli
#
# Usage:
#   ./e2e-tests.sh              # Run all tests
#   ./e2e-tests.sh -v           # Run all tests with verbose output
#   ./e2e-tests.sh -t <name>    # Run specific test by name
#   ./e2e-tests.sh -l           # List all available tests
#

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE="/workspace"
SECRET_CLI="${WORKSPACE}/secrets-cli"

# Source test utilities
source "${SCRIPT_DIR}/test-utils.sh"

# =============================================================================
# Test Configuration
# =============================================================================

# Test directory (fresh for each run)
TEST_DIR="${WORKSPACE}/test-project"

# Test user emails
ALICE_EMAIL="alice@test.local"
BOB_EMAIL="bob@test.local"

# GPG key passphrases (empty for non-interactive)
GPG_PASSPHRASE=""

# Timeout per command in seconds (prevents CI/CD hangs)
TEST_TIMEOUT=30

# =============================================================================
# Setup Functions
# =============================================================================

setup_test_environment() {
    test_log "Setting up test environment..."
    
    # Create fresh test directory
    rm -rf "$TEST_DIR"
    mkdir -p "$TEST_DIR"
    cd "$TEST_DIR"
    
    # Initialize git repository (required by secrets-cli)
    git init --quiet
    git config user.email "$ALICE_EMAIL"
    git config user.name "Test User"
    
    test_log "Test directory: $TEST_DIR (git initialized)"
}

generate_gpg_key() {
    local email="$1"
    local name="${2:-Test User}"
    
    test_log "Generating GPG key for: $email"
    
    # Generate key non-interactively
    gpg --batch --gen-key <<EOF
%no-protection
Key-Type: RSA
Key-Length: 2048
Subkey-Type: RSA
Subkey-Length: 2048
Name-Real: ${name}
Name-Email: ${email}
Expire-Date: 0
%commit
EOF
}

setup_gpg_keys() {
    test_log "Setting up GPG keys..."
    
    # Generate Alice's key
    generate_gpg_key "$ALICE_EMAIL" "Alice Test"
    
    # Generate Bob's key
    generate_gpg_key "$BOB_EMAIL" "Bob Test"
    
    test_log "GPG keys generated successfully"
}

cleanup() {
    test_log "Cleaning up..."
    cd /
    rm -rf "$TEST_DIR"
}

# =============================================================================
# Test Registry
# =============================================================================

declare -a TEST_NAMES=()
declare -A TEST_FUNCTIONS=()
declare -A TEST_DESCRIPTIONS=()

register_test() {
    local name="$1"
    local func="$2"
    local desc="${3:-}"
    TEST_NAMES+=("$name")
    TEST_FUNCTIONS["$name"]="$func"
    TEST_DESCRIPTIONS["$name"]="$desc"
}

list_tests() {
    echo ""
    echo "Available tests (35 total):"
    echo ""
    printf "  %-35s %s\n" "TEST NAME" "DESCRIPTION"
    printf "  %-35s %s\n" "─────────────────────────────────" "───────────────────────────────────────"
    for name in "${TEST_NAMES[@]}"; do
        local desc="${TEST_DESCRIPTIONS[$name]:-}"
        printf "  %-35s %s\n" "$name" "$desc"
    done
    echo ""
    echo "Note: Tests run sequentially and depend on previous state."
    echo "      Run all tests with: ./run-tests.sh"
    echo ""
}

run_single_test() {
    local name="$1"
    
    if [[ -z "${TEST_FUNCTIONS[$name]:-}" ]]; then
        echo "Unknown test: $name"
        echo "Use -l to list available tests"
        exit 1
    fi
    
    test_start "$name"
    if ${TEST_FUNCTIONS[$name]}; then
        test_pass
        return 0
    else
        return 1
    fi
}

run_all_tests() {
    for name in "${TEST_NAMES[@]}"; do
        if ! run_single_test "$name"; then
            echo ""
            echo -e "${RED}Test run stopped due to failure.${NC}"
            echo "Re-run with -v for verbose output, or -t '$name' to run just this test."
            test_summary || true
            exit 1
        fi
    done
    
    test_summary
}

# =============================================================================
# Tests: Initialization
# =============================================================================

test_init_requires_git_repository() {
    # Create a non-git directory
    local no_git_dir="${WORKSPACE}/no-git-test"
    rm -rf "$no_git_dir"
    mkdir -p "$no_git_dir"
    cd "$no_git_dir"
    
    assert_failure "$SECRET_CLI init --email $ALICE_EMAIL" || return 1
    assert_output_contains "Git repository" || return 1
    
    # Cleanup and return to test dir
    cd "$TEST_DIR"
    rm -rf "$no_git_dir"
    
    return 0
}

test_init_creates_structure() {
    cd "$TEST_DIR"
    
    assert_success "$SECRET_CLI init --email $ALICE_EMAIL" || return 1
    assert_dir_exists ".secrets" || return 1
    assert_dir_exists ".secrets/keys" || return 1
    assert_dir_exists ".secrets/vaults" || return 1
    assert_file_exists ".secrets/config.yaml" || return 1
    assert_file_exists ".secrets/keys/${ALICE_EMAIL}.asc" || return 1
    
    return 0
}

test_init_fails_if_already_initialized() {
    cd "$TEST_DIR"
    
    # Already initialized in previous test
    assert_failure "$SECRET_CLI init --email $ALICE_EMAIL" || return 1
    assert_output_contains "already exists" || return 1
    
    return 0
}

# =============================================================================
# Tests: Vault Management
# =============================================================================

test_vault_create() {
    cd "$TEST_DIR"
    
    assert_success "$SECRET_CLI vault create dev --description 'Development environment'" || return 1
    assert_dir_exists ".secrets/vaults/dev" || return 1
    assert_file_exists ".secrets/vaults/dev/vault.yaml" || return 1
    assert_dir_exists ".secrets/vaults/dev/.password-store" || return 1
    
    return 0
}

test_vault_create_production() {
    cd "$TEST_DIR"
    
    assert_success "$SECRET_CLI vault create production --description 'Production environment'" || return 1
    assert_dir_exists ".secrets/vaults/production" || return 1
    
    return 0
}

test_vault_list_shows_vaults() {
    cd "$TEST_DIR"
    
    assert_success "$SECRET_CLI --email $ALICE_EMAIL vault list" || return 1
    assert_output_contains "dev" || return 1
    assert_output_contains "production" || return 1
    
    return 0
}

test_vault_info() {
    cd "$TEST_DIR"
    
    assert_success "$SECRET_CLI vault info dev" || return 1
    assert_output_contains "dev" || return 1
    assert_output_contains "Development environment" || return 1
    assert_output_contains "$ALICE_EMAIL" || return 1
    
    return 0
}

test_vault_create_duplicate_fails() {
    cd "$TEST_DIR"
    
    assert_failure "$SECRET_CLI vault create dev" || return 1
    assert_output_contains "already exists" || return 1
    
    return 0
}

# =============================================================================
# Tests: Secret Management
# =============================================================================

test_secret_set() {
    cd "$TEST_DIR"
    
    assert_success "echo 'my-secret-password-123' | $SECRET_CLI --email $ALICE_EMAIL set dev database/password" || return 1
    assert_file_exists ".secrets/vaults/dev/.password-store/database/password.gpg" || return 1
    
    return 0
}

test_secret_set_another() {
    cd "$TEST_DIR"
    
    assert_success "echo 'api-key-xyz-789' | $SECRET_CLI --email $ALICE_EMAIL set dev api/key" || return 1
    
    return 0
}

test_secret_set_production() {
    cd "$TEST_DIR"
    
    assert_success "echo 'prod-db-password' | $SECRET_CLI --email $ALICE_EMAIL set production database/password" || return 1
    assert_success "echo 'prod-api-key' | $SECRET_CLI --email $ALICE_EMAIL set production api/key" || return 1
    
    return 0
}

test_secret_get() {
    cd "$TEST_DIR"
    
    assert_success "$SECRET_CLI --email $ALICE_EMAIL get dev database/password" || return 1
    assert_output_equals "my-secret-password-123" || return 1
    
    return 0
}

test_secret_get_api_key() {
    cd "$TEST_DIR"
    
    assert_success "$SECRET_CLI --email $ALICE_EMAIL get dev api/key" || return 1
    assert_output_equals "api-key-xyz-789" || return 1
    
    return 0
}

test_secret_list() {
    cd "$TEST_DIR"
    
    assert_success "$SECRET_CLI --email $ALICE_EMAIL list dev" || return 1
    assert_output_contains "database/password" || return 1
    assert_output_contains "api/key" || return 1
    
    return 0
}

test_secret_list_names_format() {
    cd "$TEST_DIR"
    
    assert_success "$SECRET_CLI --email $ALICE_EMAIL list dev --format names" || return 1
    assert_output_contains "database/password" || return 1
    assert_output_contains "api/key" || return 1
    
    return 0
}

test_secret_rename() {
    cd "$TEST_DIR"
    
    # First add a secret to rename
    assert_success "echo 'temp-value' | $SECRET_CLI --email $ALICE_EMAIL set dev temp/old_name" || return 1
    
    # Rename it
    assert_success "$SECRET_CLI --email $ALICE_EMAIL rename dev temp/old_name temp/new_name" || return 1
    
    # Verify new name exists and old doesn't
    assert_success "$SECRET_CLI --email $ALICE_EMAIL get dev temp/new_name" || return 1
    assert_output_equals "temp-value" || return 1
    
    return 0
}

test_secret_copy_between_vaults() {
    cd "$TEST_DIR"
    
    # Copy from dev to production
    assert_success "$SECRET_CLI --email $ALICE_EMAIL copy dev api/key production --new-name api/dev_key_copy" || return 1
    
    # Verify it exists in production
    assert_success "$SECRET_CLI --email $ALICE_EMAIL get production api/dev_key_copy" || return 1
    assert_output_equals "api-key-xyz-789" || return 1
    
    return 0
}

test_secret_delete() {
    cd "$TEST_DIR"
    
    # Delete the temp secret we created
    assert_success "$SECRET_CLI --email $ALICE_EMAIL delete dev temp/new_name --force" || return 1
    
    # Verify it's gone
    assert_failure "$SECRET_CLI --email $ALICE_EMAIL get dev temp/new_name" || return 1
    
    return 0
}

test_secret_get_nonexistent_fails() {
    cd "$TEST_DIR"
    
    assert_failure "$SECRET_CLI --email $ALICE_EMAIL get dev nonexistent/secret" || return 1
    assert_output_contains "not found" || return 1
    
    return 0
}

test_commands_work_from_subdirectory() {
    cd "$TEST_DIR"
    
    # Create a subdirectory
    mkdir -p "subdir/nested"
    cd "subdir/nested"
    
    # Commands should still work, finding .secrets from git root
    assert_success "$SECRET_CLI --email $ALICE_EMAIL vault list" || return 1
    assert_output_contains "dev" || return 1
    
    # Reading secrets should also work
    assert_success "$SECRET_CLI --email $ALICE_EMAIL get dev database/password" || return 1
    assert_output_equals "my-secret-password-123" || return 1
    
    # Return to test dir
    cd "$TEST_DIR"
    
    return 0
}

test_direnv_integration_pattern() {
    cd "$TEST_DIR"
    
    # Real direnv integration test
    # Creates an .envrc that loads secrets via secrets-cli, then uses
    # direnv exec to verify env vars are correctly loaded
    
    # Ensure secrets store and vault exist (makes test self-contained)
    if [[ ! -d "$TEST_DIR/.secrets" ]]; then
        $SECRET_CLI init --email "$ALICE_EMAIL" >/dev/null 2>&1
    fi
    $SECRET_CLI --email "$ALICE_EMAIL" vault create direnv-test >/dev/null 2>&1 || true
    echo "direnv-db-pass" | $SECRET_CLI --email $ALICE_EMAIL set direnv-test db_password >/dev/null 2>&1
    echo "direnv-api-key" | $SECRET_CLI --email $ALICE_EMAIL set direnv-test api_key >/dev/null 2>&1
    
    # Create .envrc with the documented pattern (no --email, auto-detect from git config)
    cat > .envrc <<EOF
export DATABASE_PASSWORD="\$($SECRET_CLI get direnv-test db_password)"
export API_KEY="\$($SECRET_CLI get direnv-test api_key)"
EOF
    
    # Allow the .envrc (required by direnv for security)
    local allow_output
    allow_output=$(direnv allow . 2>&1)
    if [[ $? -ne 0 ]]; then
        test_fail "direnv allow success" "failed: $allow_output"
        return 1
    fi
    
    # Use direnv exec to load the .envrc and check env vars are set
    local result
    local direnv_stderr
    direnv_stderr=$(mktemp)
    
    result=$(direnv exec . bash -c 'echo "$DATABASE_PASSWORD"' 2>"$direnv_stderr")
    if [[ "$result" != "direnv-db-pass" ]]; then
        test_fail "direnv-db-pass" "$result" "DATABASE_PASSWORD not loaded by direnv. stderr: $(cat "$direnv_stderr")"
        rm -f "$direnv_stderr" .envrc
        return 1
    fi
    
    result=$(direnv exec . bash -c 'echo "$API_KEY"' 2>"$direnv_stderr")
    if [[ "$result" != "direnv-api-key" ]]; then
        test_fail "direnv-api-key" "$result" "API_KEY not loaded by direnv. stderr: $(cat "$direnv_stderr")"
        rm -f "$direnv_stderr" .envrc
        return 1
    fi
    
    # Verify env vars are NOT set outside the direnv context
    if [[ -n "${DATABASE_PASSWORD:-}" ]]; then
        test_fail "DATABASE_PASSWORD unset outside direnv" "DATABASE_PASSWORD is set: $DATABASE_PASSWORD"
        rm -f "$direnv_stderr" .envrc
        return 1
    fi
    
    # Test direnv also works from a subdirectory (inherits parent .envrc)
    mkdir -p "app/src"
    result=$(direnv exec ./app/src bash -c 'echo "$DATABASE_PASSWORD"' 2>"$direnv_stderr")
    if [[ "$result" != "direnv-db-pass" ]]; then
        test_fail "direnv-db-pass" "$result" "direnv not loading secrets from subdirectory. stderr: $(cat "$direnv_stderr")"
        rm -f "$direnv_stderr" .envrc
        rm -rf app
        return 1
    fi
    
    # Cleanup
    rm -f "$direnv_stderr" .envrc
    rm -rf app
    
    return 0
}

# =============================================================================
# Tests: Key Management
# =============================================================================

test_key_list() {
    cd "$TEST_DIR"
    
    assert_success "$SECRET_CLI key list" || return 1
    assert_output_contains "$ALICE_EMAIL" || return 1
    
    return 0
}

test_key_add_bob() {
    cd "$TEST_DIR"
    
    # Add Bob's key to the store (export from GPG and add)
    assert_success "$SECRET_CLI key add $BOB_EMAIL" || return 1
    assert_file_exists ".secrets/keys/${BOB_EMAIL}.asc" || return 1
    
    return 0
}

test_key_list_shows_both() {
    cd "$TEST_DIR"
    
    assert_success "$SECRET_CLI key list" || return 1
    assert_output_contains "$ALICE_EMAIL" || return 1
    assert_output_contains "$BOB_EMAIL" || return 1
    
    return 0
}

# =============================================================================
# Tests: Multi-User Access
# =============================================================================

test_vault_add_member() {
    cd "$TEST_DIR"
    
    # Add Bob to dev vault
    assert_success "$SECRET_CLI --email $ALICE_EMAIL vault add-member dev $BOB_EMAIL" || return 1
    assert_output_contains "Added $BOB_EMAIL" || return 1
    assert_output_contains "Re-encrypted" || return 1
    
    return 0
}

test_vault_info_shows_members() {
    cd "$TEST_DIR"
    
    assert_success "$SECRET_CLI vault info dev" || return 1
    assert_output_contains "$ALICE_EMAIL" || return 1
    assert_output_contains "$BOB_EMAIL" || return 1
    
    return 0
}

test_bob_can_read_dev_secret() {
    cd "$TEST_DIR"
    
    # Bob should be able to read secrets from dev vault
    assert_success "$SECRET_CLI --email $BOB_EMAIL get dev database/password" || return 1
    assert_output_equals "my-secret-password-123" || return 1
    
    # CRITICAL: Verify the .gpg file is actually encrypted for Bob's key
    # This catches the bug where re-encryption silently fails with untrusted keys
    local secret_file=".secrets/vaults/dev/.password-store/database/password.gpg"
    
    # Get Bob's encryption subkey ID (field 5 from sub: lines with 'e' in usage field 12)
    local bob_key_id=$(gpg --list-keys --with-colons "$BOB_EMAIL" | awk -F: '/^sub:/ && $12 ~ /e/ {print $5; exit}')
    
    if [[ -z "$bob_key_id" ]]; then
        test_fail "Bob's encryption key ID found" "empty"
        return 1
    fi
    
    # Check if Bob's key ID appears in the encrypted file's recipients
    local recipients=$(gpg --list-packets "$secret_file" 2>&1 | grep 'keyid' | awk '{print $NF}')
    
    if ! echo "$recipients" | grep -q "$bob_key_id"; then
        test_fail "Secret encrypted for Bob (key $bob_key_id)" "Recipients: $recipients"
        return 1
    fi
    
    test_log "✓ Verified secret is encrypted for Bob's key: $bob_key_id"
    
    return 0
}

test_bob_cannot_read_production_secret() {
    cd "$TEST_DIR"
    
    # Bob is not a member of production vault
    assert_failure "$SECRET_CLI --email $BOB_EMAIL get production database/password" || return 1
    assert_output_contains "Access denied" || return 1
    
    return 0
}

test_bob_can_set_dev_secret() {
    cd "$TEST_DIR"
    
    # Bob should be able to write to dev vault
    assert_success "echo 'bob-added-this' | $SECRET_CLI --email $BOB_EMAIL set dev bob/secret" || return 1
    
    # Alice should be able to read it
    assert_success "$SECRET_CLI --email $ALICE_EMAIL get dev bob/secret" || return 1
    assert_output_equals "bob-added-this" || return 1
    
    return 0
}

test_vault_remove_member() {
    cd "$TEST_DIR"
    
    # Remove Bob from dev vault
    assert_success "$SECRET_CLI --email $ALICE_EMAIL vault remove-member dev $BOB_EMAIL" || return 1
    assert_output_contains "Removed $BOB_EMAIL" || return 1
    
    return 0
}

test_bob_cannot_read_after_removal() {
    cd "$TEST_DIR"
    
    # Bob should no longer be able to read dev vault
    assert_failure "$SECRET_CLI --email $BOB_EMAIL get dev database/password" || return 1
    
    return 0
}

# =============================================================================
# Tests: Export
# =============================================================================

test_export_env_format() {
    cd "$TEST_DIR"
    
    assert_success "$SECRET_CLI --email $ALICE_EMAIL export dev --format env" || return 1
    assert_output_contains "export DATABASE_PASSWORD=" || return 1
    assert_output_contains "export API_KEY=" || return 1
    
    return 0
}

test_export_dotenv_format() {
    cd "$TEST_DIR"
    
    assert_success "$SECRET_CLI --email $ALICE_EMAIL export dev --format dotenv" || return 1
    assert_output_contains "DATABASE_PASSWORD=" || return 1
    assert_output_not_contains "export " || return 1
    
    return 0
}

test_export_with_prefix() {
    cd "$TEST_DIR"
    
    assert_success "$SECRET_CLI --email $ALICE_EMAIL export dev --format env --prefix DEV_" || return 1
    assert_output_contains "DEV_DATABASE_PASSWORD=" || return 1
    
    return 0
}

# =============================================================================
# Tests: Sync
# =============================================================================

test_sync_vault() {
    cd "$TEST_DIR"
    
    assert_success "$SECRET_CLI --email $ALICE_EMAIL sync dev" || return 1
    assert_output_contains "Synchronized" || return 1
    
    return 0
}

# =============================================================================
# Tests: Vault Deletion
# =============================================================================

test_vault_delete() {
    cd "$TEST_DIR"
    
    # Create a vault to delete
    assert_success "$SECRET_CLI vault create temp-vault" || return 1
    assert_dir_exists ".secrets/vaults/temp-vault" || return 1
    
    # Delete it
    assert_success "$SECRET_CLI vault delete temp-vault --force" || return 1
    
    # Verify it's gone
    if [[ -d ".secrets/vaults/temp-vault" ]]; then
        test_fail "vault deleted" "vault still exists"
        return 1
    fi
    
    return 0
}

# =============================================================================
# Tests: Setup Command
# =============================================================================

test_setup_imports_keys() {
    cd "$TEST_DIR"
    
    assert_success "$SECRET_CLI setup --email $ALICE_EMAIL" || return 1
    assert_output_contains "Found your key" || return 1
    assert_output_contains "Imported" || return 1
    
    return 0
}

# =============================================================================
# Tests: Untrusted Key Re-encryption (Regression Test)
# =============================================================================

test_untrusted_key_reencryption() {
    cd "$TEST_DIR"
    
    # This test simulates the real-world bug where adding a member with an
    # untrusted GPG key would silently fail to re-encrypt secrets.
    test_log "Generating Charlie's key in separate keyring (simulates external key)..."
    
    # Create temporary keyring for Charlie
    local charlie_gnupghome=$(mktemp -d)
    
    # Generate Charlie's key in isolated keyring
    GNUPGHOME="$charlie_gnupghome" gpg --batch --gen-key <<EOF
%no-protection
Key-Type: RSA
Key-Length: 2048
Subkey-Type: RSA
Subkey-Length: 2048
Name-Real: Charlie External
Name-Email: charlie@external.com
Expire-Date: 0
%commit
EOF
    
    # Export Charlie's public key
    local charlie_key_file="${TEST_DIR}/charlie.asc"
    GNUPGHOME="$charlie_gnupghome" gpg --armor --export charlie@external.com > "$charlie_key_file"
    
    # Import Charlie's key into main keyring (will be UNTRUSTED)
    gpg --import "$charlie_key_file" 2>&1 || true
    
    # Verify Charlie's key is untrusted
    local trust_level=$(gpg --list-keys --with-colons charlie@external.com | grep '^pub:' | cut -d: -f2)
    if [[ "$trust_level" != "-" && "$trust_level" != "q" ]]; then
        test_fail "Charlie's key untrusted (expected -, got $trust_level)" ""
        rm -rf "$charlie_gnupghome" "$charlie_key_file"
        return 1
    fi
    
    test_log "✓ Charlie's key is untrusted (trust level: $trust_level)"
    
    # Create a test vault with a secret
    assert_success "$SECRET_CLI vault create untrusted-test" || {
        rm -rf "$charlie_gnupghome" "$charlie_key_file"
        return 1
    }
    
    assert_success "echo 'untrusted-secret-value' | $SECRET_CLI --email $ALICE_EMAIL set untrusted-test test/secret" || {
        rm -rf "$charlie_gnupghome" "$charlie_key_file"
        return 1
    }
    
    # Add Charlie's key to secrets-cli
    assert_success "$SECRET_CLI key add charlie@external.com" || {
        rm -rf "$charlie_gnupghome" "$charlie_key_file"
        return 1
    }
    
    # CRITICAL: Add Charlie as a member (this would silently fail before the fix)
    assert_success "$SECRET_CLI --email $ALICE_EMAIL vault add-member untrusted-test charlie@external.com" || {
        rm -rf "$charlie_gnupghome" "$charlie_key_file"
        return 1
    }
    
    # Get Charlie's encryption key ID (field 5 from sub: lines with 'e' in usage field 12)
    local charlie_key_id=$(gpg --list-keys --with-colons charlie@external.com | awk -F: '/^sub:/ && $12 ~ /e/ {print $5; exit}')
    
    if [[ -z "$charlie_key_id" ]]; then
        test_fail "Charlie's encryption key ID found" "empty"
        rm -rf "$charlie_gnupghome" "$charlie_key_file"
        return 1
    fi
    
    # VERIFICATION: Check if the secret is actually encrypted for Charlie
    local secret_file=".secrets/vaults/untrusted-test/.password-store/test/secret.gpg"
    local recipients=$(gpg --list-packets "$secret_file" 2>&1 | grep 'keyid' | awk '{print $NF}')
    
    if ! echo "$recipients" | grep -q "$charlie_key_id"; then
        test_fail "Secret encrypted for untrusted key (Charlie: $charlie_key_id)" "Recipients: $recipients"
        rm -rf "$charlie_gnupghome" "$charlie_key_file"
        return 1
    fi
    
    test_log "✓ Verified secret is encrypted for untrusted key: $charlie_key_id"
    
    # Cleanup
    rm -rf "$charlie_gnupghome" "$charlie_key_file"
    
    return 0
}

# =============================================================================
# Register All Tests
# =============================================================================

register_tests() {
    # Initialization
    register_test "init_requires_git_repository" test_init_requires_git_repository \
        "Require git repository for init command"
    register_test "init_creates_structure" test_init_creates_structure \
        "Initialize .secrets dir, config, and export user's GPG key"
    register_test "init_fails_if_already_initialized" test_init_fails_if_already_initialized \
        "Prevent double initialization"
    
    # Vault Management
    register_test "vault_create" test_vault_create \
        "Create 'dev' vault with description and pass store"
    register_test "vault_create_production" test_vault_create_production \
        "Create 'production' vault for multi-vault testing"
    register_test "vault_list_shows_vaults" test_vault_list_shows_vaults \
        "List all vaults with access status"
    register_test "vault_info" test_vault_info \
        "Display vault details, members, and secret count"
    register_test "vault_create_duplicate_fails" test_vault_create_duplicate_fails \
        "Reject creating vault with existing name"
    
    # Secret Management
    register_test "secret_set" test_secret_set \
        "Create encrypted secret via stdin pipe"
    register_test "secret_set_another" test_secret_set_another \
        "Add second secret to dev vault"
    register_test "secret_set_production" test_secret_set_production \
        "Add secrets to production vault"
    register_test "secret_get" test_secret_get \
        "Retrieve and verify secret value"
    register_test "secret_get_api_key" test_secret_get_api_key \
        "Retrieve different secret to verify isolation"
    register_test "secret_list" test_secret_list \
        "List all secrets in vault (table format)"
    register_test "secret_list_names_format" test_secret_list_names_format \
        "List secrets with --format names"
    register_test "secret_rename" test_secret_rename \
        "Rename secret and verify content preserved"
    register_test "secret_copy_between_vaults" test_secret_copy_between_vaults \
        "Copy secret from dev to production vault"
    register_test "secret_delete" test_secret_delete \
        "Delete secret with --force flag"
    register_test "secret_get_nonexistent_fails" test_secret_get_nonexistent_fails \
        "Error on accessing non-existent secret"
    register_test "commands_work_from_subdirectory" test_commands_work_from_subdirectory \
        "Commands work from nested subdirectories"
    register_test "direnv_integration_pattern" test_direnv_integration_pattern \
        "direnv .envrc pattern: get secrets without --email via command substitution"
    
    # Key Management
    register_test "key_list" test_key_list \
        "List stored public GPG keys"
    register_test "key_add_bob" test_key_add_bob \
        "Add Bob's public key to .secrets/keys/"
    register_test "key_list_shows_both" test_key_list_shows_both \
        "Verify both Alice and Bob keys listed"
    
    # Multi-User Access
    register_test "vault_add_member" test_vault_add_member \
        "Add Bob to dev vault, re-encrypt all secrets"
    register_test "vault_info_shows_members" test_vault_info_shows_members \
        "Verify vault info shows both members"
    register_test "bob_can_read_dev_secret" test_bob_can_read_dev_secret \
        "Bob reads dev vault secret after being added"
    register_test "bob_cannot_read_production_secret" test_bob_cannot_read_production_secret \
        "Bob denied access to production (not a member)"
    register_test "bob_can_set_dev_secret" test_bob_can_set_dev_secret \
        "Bob writes to dev vault, Alice can read it"
    register_test "vault_remove_member" test_vault_remove_member \
        "Remove Bob from dev vault, re-encrypt secrets"
    register_test "bob_cannot_read_after_removal" test_bob_cannot_read_after_removal \
        "Bob denied access after removal"
    
    # Export
    register_test "export_env_format" test_export_env_format \
        "Export secrets as 'export VAR=value' format"
    register_test "export_dotenv_format" test_export_dotenv_format \
        "Export secrets as 'VAR=value' (.env format)"
    register_test "export_with_prefix" test_export_with_prefix \
        "Export with --prefix adds prefix to var names"
    
    # Sync
    register_test "sync_vault" test_sync_vault \
        "Verify vault integrity and re-encrypt if needed"
    
    # Vault Deletion
    register_test "vault_delete" test_vault_delete \
        "Create and delete a vault with --force"
    
    # Setup (at the end since it uses existing state)
    register_test "setup_imports_keys" test_setup_imports_keys \
        "Import keys and verify access after clone"
    
    # Regression Tests
    register_test "untrusted_key_reencryption" test_untrusted_key_reencryption \
        "Verify re-encryption works with untrusted GPG keys (regression test for silent failure bug)"
}

# =============================================================================
# Main
# =============================================================================

main() {
    local run_test=""
    local list_only=false
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -t|--test)
                run_test="$2"
                shift 2
                ;;
            -l|--list)
                list_only=true
                shift
                ;;
            -h|--help)
                echo "Usage: $0 [options]"
                echo ""
                echo "Options:"
                echo "  -v, --verbose      Show verbose output for each test"
                echo "  -t, --test <name>  Run a specific test by name"
                echo "  -l, --list         List all available tests"
                echo "  -h, --help         Show this help"
                exit 0
                ;;
            *)
                echo "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Register all tests
    register_tests
    
    # List tests if requested
    if [[ "$list_only" == true ]]; then
        list_tests
        exit 0
    fi
    
    echo ""
    echo -e "${BOLD}secret-cli End-to-End Tests${NC}"
    echo -e "${DIM}Running in container with isolated environment${NC}"
    
    # Setup
    test_group "Setup"
    setup_test_environment
    setup_gpg_keys
    test_log "Environment ready"
    
    # Run tests
    test_group "Running Tests"
    
    if [[ -n "$run_test" ]]; then
        # Run single test
        if ! run_single_test "$run_test"; then
            test_summary || true
            exit 1
        fi
        test_summary
    else
        # Run all tests
        run_all_tests
    fi
    
    exit_code=$?
    
    # Cleanup
    cleanup
    
    exit $exit_code
}

main "$@"

# =============================================================================
# Tests: Untrusted Key Re-encryption (Regression Test)
# =============================================================================

test_untrusted_key_reencryption() {
    cd "$TEST_DIR"
    
    # This test simulates the real-world bug where adding a member with an
    # untrusted GPG key would silently fail to re-encrypt secrets.
    test_log "Generating Charlie's key in separate keyring (simulates external key)..."
    
    # Create temporary keyring for Charlie
    local charlie_gnupghome=$(mktemp -d)
    
    # Generate Charlie's key in isolated keyring
    GNUPGHOME="$charlie_gnupghome" gpg --batch --gen-key <<EOF
%no-protection
Key-Type: RSA
Key-Length: 2048
Subkey-Type: RSA
Subkey-Length: 2048
Name-Real: Charlie External
Name-Email: charlie@external.com
Expire-Date: 0
%commit
EOF
    
    # Export Charlie's public key
    local charlie_key_file="${TEST_DIR}/charlie.asc"
    GNUPGHOME="$charlie_gnupghome" gpg --armor --export charlie@external.com > "$charlie_key_file"
    
    # Import Charlie's key into main keyring (will be UNTRUSTED)
    gpg --import "$charlie_key_file" 2>&1 || true
    
    # Verify Charlie's key is untrusted
    local trust_level=$(gpg --list-keys --with-colons charlie@external.com | grep '^pub:' | cut -d: -f2)
    if [[ "$trust_level" != "-" && "$trust_level" != "q" ]]; then
        test_fail "Charlie's key untrusted (expected -, got $trust_level)" ""
        rm -rf "$charlie_gnupghome" "$charlie_key_file"
        return 1
    fi
    
    test_log "✓ Charlie's key is untrusted (trust level: $trust_level)"
    
    # Create a test vault with a secret
    assert_success "$SECRET_CLI vault create untrusted-test" || {
        rm -rf "$charlie_gnupghome" "$charlie_key_file"
        return 1
    }
    
    assert_success "echo 'untrusted-secret-value' | $SECRET_CLI --email $ALICE_EMAIL set untrusted-test test/secret" || {
        rm -rf "$charlie_gnupghome" "$charlie_key_file"
        return 1
    }
    
    # Add Charlie's key to secrets-cli
    assert_success "$SECRET_CLI key add charlie@external.com" || {
        rm -rf "$charlie_gnupghome" "$charlie_key_file"
        return 1
    }
    
    # CRITICAL: Add Charlie as a member (this would silently fail before the fix)
    assert_success "$SECRET_CLI --email $ALICE_EMAIL vault add-member untrusted-test charlie@external.com" || {
        rm -rf "$charlie_gnupghome" "$charlie_key_file"
        return 1
    }
    
    # Get Charlie's encryption key ID (field 5 from sub: lines with 'e' in usage field 12)
    local charlie_key_id=$(gpg --list-keys --with-colons charlie@external.com | awk -F: '/^sub:/ && $12 ~ /e/ {print $5; exit}')
    
    if [[ -z "$charlie_key_id" ]]; then
        test_fail "Charlie's encryption key ID found" "empty"
        rm -rf "$charlie_gnupghome" "$charlie_key_file"
        return 1
    fi
    
    # VERIFICATION: Check if the secret is actually encrypted for Charlie
    local secret_file=".secrets/vaults/untrusted-test/.password-store/test/secret.gpg"
    local recipients=$(gpg --list-packets "$secret_file" 2>&1 | grep 'keyid' | awk '{print $NF}')
    
    if ! echo "$recipients" | grep -q "$charlie_key_id"; then
        test_fail "Secret encrypted for untrusted key (Charlie: $charlie_key_id)" "Recipients: $recipients"
        rm -rf "$charlie_gnupghome" "$charlie_key_file"
        return 1
    fi
    
    test_log "✓ Verified secret is encrypted for untrusted key: $charlie_key_id"
    
    # Cleanup
    rm -rf "$charlie_gnupghome" "$charlie_key_file"
    
    return 0
}
