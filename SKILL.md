---
name: secrets-cli Testing Requirements
description: Comprehensive testing requirements and guidelines for secrets-cli development. Every bug fix MUST include tests. All tests MUST run on PR checks.
---

# secrets-cli Testing Requirements

## Core Principle

**Every bug fix SHALL include additional tests (unit or e2e) that would have caught the bug.**

All tests MUST run automatically on PR checks to prevent regressions.

## Test Types

### 1. Unit Tests (Preferred when possible)
- **Location**: `*_test.go` files alongside source code
- **Run with**: `go test ./...`
- **Use for**: Testing individual functions, logic, data structures
- **Example**: Testing GPG key parsing, configuration validation, etc.

### 2. E2E Tests (Required for integration scenarios)
- **Location**: `tests/e2e-tests.sh`
- **Run with**: `./tests/run-tests.sh`
- **Use for**: Testing complete workflows, multi-component interactions
- **Example**: Vault creation, secret encryption, member management

## Testing Requirements for Bug Fixes

### Mandatory Test Checklist

When fixing a bug, you MUST:

1. ✅ **Understand the root cause** - Document why the bug occurred
2. ✅ **Write a test that fails** - Reproduce the bug in a test
3. ✅ **Implement the fix** - Make the test pass
4. ✅ **Verify the test catches the bug** - Temporarily revert fix, confirm test fails
5. ✅ **Document the test** - Add comments explaining what bug it prevents

### Test Requirements by Bug Type

#### Cryptographic Operations
When fixing bugs related to encryption, decryption, or key management:

**MUST verify actual cryptographic state**, not just command success:

```bash
# ❌ INSUFFICIENT - Only checks command success
assert_success "$SECRET_CLI vault add-member dev bob@example.com"

# ✅ REQUIRED - Verifies actual encryption
assert_success "$SECRET_CLI vault add-member dev bob@example.com"

# Verify .gpg file is actually encrypted for Bob's key
local bob_key_id=$(gpg --list-keys --with-colons bob@example.com | grep '^sub:' | grep ':e:' | head -1 | cut -d: -f5)
local recipients=$(gpg --list-packets secret.gpg 2>&1 | grep 'keyid')

if ! echo "$recipients" | grep -q "$bob_key_id"; then
    test_fail "Secret NOT encrypted for Bob's key"
fi
```

**MUST test with untrusted keys** when relevant:

```bash
# Simulate real-world scenario: import external untrusted key
GNUPGHOME="$temp_keyring" gpg --batch --gen-key <<EOF
...
EOF

# Export and import as untrusted
gpg --import external_key.asc

# Verify trust level is untrusted
trust_level=$(gpg --list-keys --with-colons user@example.com | grep '^pub:' | cut -d: -f2)
[[ "$trust_level" == "-" ]] || test_fail "Key should be untrusted"

# Now test the operation with untrusted key
```

#### File System Operations
When fixing bugs related to file creation, deletion, or modification:

**MUST verify actual file state**:

```bash
# ❌ INSUFFICIENT
assert_success "$SECRET_CLI vault create dev"

# ✅ REQUIRED
assert_success "$SECRET_CLI vault create dev"
assert_dir_exists ".secrets/vaults/dev"
assert_file_exists ".secrets/vaults/dev/vault.yaml"
assert_dir_exists ".secrets/vaults/dev/.password-store"
```

#### Multi-User/Permission Bugs
When fixing bugs related to access control or multi-user scenarios:

**MUST test with multiple users**:

```bash
# Test as Alice
assert_success "$SECRET_CLI --email alice@example.com set dev secret"

# Test as Bob (should fail if not a member)
assert_failure "$SECRET_CLI --email bob@example.com get dev secret"
assert_output_contains "Access denied"

# Add Bob, then verify access
assert_success "$SECRET_CLI --email alice@example.com vault add-member dev bob@example.com"
assert_success "$SECRET_CLI --email bob@example.com get dev secret"
```

## E2E Test Structure

### Adding a New E2E Test

1. **Define the test function** in `tests/e2e-tests.sh`:

```bash
test_your_feature_name() {
    cd "$TEST_DIR"
    
    # Setup: Create necessary state
    assert_success "$SECRET_CLI vault create test-vault"
    
    # Action: Perform the operation being tested
    assert_success "$SECRET_CLI some-command"
    
    # Verification: Check the results
    assert_file_exists "expected/file.txt"
    assert_output_contains "expected output"
    
    # Cleanup: Remove temporary resources (if needed)
    rm -rf temp-files
    
    return 0
}
```

2. **Register the test** in the `register_tests()` function:

```bash
register_test "your_feature_name" test_your_feature_name \
    "Brief description of what this test verifies"
```

3. **Update the test count** in `list_tests()`:

```bash
echo "Available tests (N total):"  # Increment N
```

### Test Naming Conventions

- **Function name**: `test_<feature>_<scenario>`
- **Registration name**: `<feature>_<scenario>` (no `test_` prefix)
- **Examples**:
  - `test_vault_add_member` → `vault_add_member`
  - `test_untrusted_key_reencryption` → `untrusted_key_reencryption`

## CI/CD Integration

### PR Workflow

All tests run automatically on PRs via `.github/workflows/pr.yml`:

```yaml
- name: Run E2E Tests
  run: ./tests/run-tests.sh --rebuild
```

### Test Execution

1. **Build**: `make build` creates the `secrets-cli` binary
2. **Docker**: Tests run in isolated Docker container (`tests/Dockerfile`)
3. **E2E**: All registered tests execute sequentially
4. **Report**: Results posted as PR comment

### Adding New Test Files

If you add a new test script (not in `e2e-tests.sh`):

1. **Copy to Docker image** in `tests/Dockerfile`:

```dockerfile
COPY tests/your-new-test.sh /workspace/tests/your-new-test.sh
RUN chmod +x /workspace/tests/your-new-test.sh
```

2. **Call from run-tests.sh** or integrate into `e2e-tests.sh`

## Common Testing Patterns

### Pattern 1: Verify Encryption Recipients

```bash
# Get expected key ID
local expected_key_id=$(gpg --list-keys --with-colons user@example.com | grep '^sub:' | grep ':e:' | head -1 | cut -d: -f5)

# Get actual recipients from .gpg file
local actual_recipients=$(gpg --list-packets file.gpg 2>&1 | grep 'keyid' | awk '{print $NF}')

# Verify
if ! echo "$actual_recipients" | grep -q "$expected_key_id"; then
    test_fail "File not encrypted for expected key"
fi
```

### Pattern 2: Test with Untrusted Key

```bash
# Create separate keyring
local temp_gnupghome=$(mktemp -d)

# Generate key in isolation
GNUPGHOME="$temp_gnupghome" gpg --batch --gen-key <<EOF
%no-protection
Key-Type: RSA
Key-Length: 2048
Name-Email: external@example.com
%commit
EOF

# Export and import as untrusted
GNUPGHOME="$temp_gnupghome" gpg --armor --export external@example.com > /tmp/key.asc
gpg --import /tmp/key.asc

# Verify untrusted
trust=$(gpg --list-keys --with-colons external@example.com | grep '^pub:' | cut -d: -f2)
[[ "$trust" == "-" ]] || test_fail "Key should be untrusted"

# Cleanup
rm -rf "$temp_gnupghome" /tmp/key.asc
```

### Pattern 3: Multi-User Access

```bash
# Setup: Create vault and secret as Alice
assert_success "$SECRET_CLI --email alice@example.com vault create shared"
assert_success "echo 'secret' | $SECRET_CLI --email alice@example.com set shared test"

# Test: Bob cannot access (not a member)
assert_failure "$SECRET_CLI --email bob@example.com get shared test"

# Action: Add Bob
assert_success "$SECRET_CLI --email alice@example.com vault add-member shared bob@example.com"

# Verify: Bob can now access
assert_success "$SECRET_CLI --email bob@example.com get shared test"
assert_output_equals "secret"
```

## Regression Test Requirements

### When to Add Regression Tests

Add a regression test when:

1. **Silent failure** - Operation appeared to succeed but didn't
2. **Edge case** - Bug only manifests in specific conditions
3. **Integration issue** - Bug involves multiple components
4. **Security issue** - Bug has security implications

### Regression Test Template

```bash
test_regression_<bug_description>() {
    cd "$TEST_DIR"
    
    # CONTEXT: Explain the bug this test prevents
    # Bug: <Brief description>
    # Root cause: <Why it happened>
    # Fixed in: <PR or commit reference>
    
    # SETUP: Create conditions that trigger the bug
    # ...
    
    # ACTION: Perform operation that would fail with bug
    # ...
    
    # VERIFICATION: Check that bug is fixed
    # This should verify the ACTUAL state, not just command success
    # ...
    
    return 0
}
```

### Example: Re-encryption Bug

```bash
test_untrusted_key_reencryption() {
    # CONTEXT: Regression test for silent re-encryption failure
    # Bug: Adding member with untrusted GPG key would report success
    #      but silently fail to re-encrypt secrets
    # Root cause: GPG refuses to encrypt to untrusted keys in --batch mode
    # Fixed in: PR #18
    
    # SETUP: Create untrusted key (simulates external import)
    local temp_keyring=$(mktemp -d)
    GNUPGHOME="$temp_keyring" gpg --batch --gen-key <<EOF
...
EOF
    
    # ACTION: Add untrusted key as vault member
    assert_success "$SECRET_CLI vault add-member test charlie@external.com"
    
    # VERIFICATION: Verify secret is ACTUALLY encrypted for Charlie
    local charlie_key_id=$(gpg --list-keys --with-colons charlie@external.com | grep '^sub:' | grep ':e:' | head -1 | cut -d: -f5)
    local recipients=$(gpg --list-packets secret.gpg 2>&1 | grep 'keyid')
    
    if ! echo "$recipients" | grep -q "$charlie_key_id"; then
        test_fail "Secret not encrypted for untrusted key (regression!)"
    fi
}
```

## Test Quality Standards

### Required Elements

Every test MUST have:

1. ✅ **Clear purpose** - Comment explaining what it tests
2. ✅ **Proper setup** - Create necessary preconditions
3. ✅ **Explicit verification** - Check actual state, not just success
4. ✅ **Cleanup** - Remove temporary resources
5. ✅ **Error handling** - Fail fast with clear messages

### Anti-Patterns to Avoid

❌ **Don't trust command success alone**:
```bash
# BAD
assert_success "$SECRET_CLI vault add-member dev bob@example.com"
# What if it succeeded but didn't actually add Bob?
```

❌ **Don't skip verification**:
```bash
# BAD
assert_success "$SECRET_CLI set dev secret"
# Did it actually create the file? Is it encrypted?
```

❌ **Don't test in isolation only**:
```bash
# BAD - Only tests with trusted keys
generate_gpg_key "bob@example.com"
assert_success "$SECRET_CLI vault add-member dev bob@example.com"
# Real users import external untrusted keys!
```

❌ **Don't ignore edge cases**:
```bash
# BAD - Only tests happy path
assert_success "$SECRET_CLI vault create dev"
# What about duplicate names? Invalid characters? Permissions?
```

## Documentation Policy

**IMPORTANT**: Do NOT add ephemeral, bug-specific documentation files to the repository.

### What NOT to Add

❌ **Avoid creating these types of files**:
- `BUGFIX-*.md` - Detailed analysis of specific bugs
- `TEST-COVERAGE-ANALYSIS.md` - Why tests didn't catch a specific bug  
- `SUMMARY.md` - Summaries of specific PRs or bug fixes
- Any other documentation that describes a specific bug or fix

### Why

- Bug-specific documentation becomes stale and clutters the repository
- The information belongs in:
  - **PR descriptions** - Context for reviewers
  - **Commit messages** - Permanent git history
  - **Code comments** - Inline explanations
  - **This SKILL.md** - General patterns and requirements

### What TO Document

✅ **DO add**:
- **Code comments** - Explain WHY the code does something, especially for bug fixes
- **Test comments** - Explain what bug a regression test prevents
- **SKILL.md updates** - Add general patterns learned from the bug

### Example

Instead of creating `BUGFIX-untrusted-keys.md`, add a comment in the code:

```go
// Set trust-model to always to handle untrusted GPG keys.
// Without this, GPG refuses to encrypt to untrusted keys in batch mode,
// causing silent re-encryption failures when adding new vault members.
existingOpts := os.Getenv("PASSWORD_STORE_GPG_OPTS")
gpgOpts := "--trust-model always"
```

And update this SKILL.md with the general pattern if it's broadly applicable.

## Quick Reference

### Running Tests Locally

```bash
# All e2e tests
./tests/run-tests.sh

# Specific test
./tests/run-tests.sh -t vault_create

# Verbose output
./tests/run-tests.sh -v

# Force rebuild
./tests/run-tests.sh --rebuild

# List all tests
./tests/run-tests.sh -l
```

### Test Assertions Available

```bash
assert_success "command"              # Command must succeed
assert_failure "command"              # Command must fail
assert_output_contains "text"         # Output must contain text
assert_output_equals "text"           # Output must equal text exactly
assert_output_not_contains "text"     # Output must NOT contain text
assert_file_exists "path"             # File must exist
assert_dir_exists "path"              # Directory must exist
test_fail "expected" "actual"         # Manual failure with message
test_log "message"                    # Log message (verbose mode)
```

## Summary

**Golden Rule**: If you fix a bug, write a test that would have caught it.

**Test Quality**: Verify actual state, not just command success.

**Coverage**: Test edge cases, especially untrusted keys and multi-user scenarios.

**CI/CD**: All tests run on PRs automatically. No exceptions.
