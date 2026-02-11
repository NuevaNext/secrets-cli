---
name: secrets-cli Development Guidelines
description: Testing requirements, documentation policy, and essential references for secrets-cli development
---

# secrets-cli Development Guidelines

## 🚨 CRITICAL: Read This First

**Before submitting any PR or making changes, you MUST read:**

### 📋 [PR-PROCESS.md](PR-PROCESS.md) - Complete Pull Request Process

This document is **MANDATORY** reading. It covers:
- ✅ Prerequisites before PR submission (Step 0)
- 🔄 Complete review and iteration process (Steps 1-8)
- 📝 How to track and address review comments
- ✅ Final verification before merge

**Key principle**: The PR process has a loop (Steps 2 → 3 → 3.5 → back to 2) that you'll iterate through multiple times. Understanding this flow is essential.

---

## Testing Requirements

### Core Principle

**Every bug fix SHALL include additional tests (unit or e2e) that would have caught the bug.**

All tests MUST run automatically on PR checks to prevent regressions.

### Test Types

#### 1. Unit Tests (Preferred when possible)
- **Location**: `*_test.go` files alongside source code
- **Run with**: `go test ./...`
- **Use for**: Testing individual functions, logic, data structures
- **Example**: Testing GPG key parsing, configuration validation, etc.

#### 2. E2E Tests (Required for integration scenarios)
- **Location**: `tests/e2e-tests.sh`
- **Run with**: `./tests/run-tests.sh`
- **Use for**: Testing complete workflows, multi-component interactions
- **Example**: Vault creation, secret encryption, member management

### Testing Philosophy

#### Prefer Unit Tests Over E2E Tests

**Always try to reproduce and fix bugs with unit tests first**, before relying on e2e tests:

✅ **DO**:
```go
// Test the specific function in isolation
func TestVerifyEncryption(t *testing.T) {
    // Create minimal test case
    // Debug the actual behavior
    // Fix the root cause
}
```

❌ **DON'T**:
```bash
# Only rely on e2e tests that test everything at once
# Harder to debug, slower feedback loop
```

**Why?**
- Unit tests are **fast** (milliseconds vs seconds)
- Unit tests **isolate** the problem to a specific function
- Unit tests give **precise** error messages
- E2E tests test too many things at once, making debugging hard

#### Focus on Correctness, Not Just Passing Tests

❌ **DON'T** make changes just to make tests pass  
✅ **DO** understand the root cause and fix it correctly

**Example**: If a test expects key ID matching to work, don't hack the matching logic. Instead, understand WHY the matching fails (e.g., GPG version differences) and implement a robust solution (e.g., count-based verification).

### Test Quality Standards

When writing tests:

✅ **DO**:
- Test actual behavior and state changes
- Use descriptive test names
- Include edge cases
- Test error conditions
- Add comments explaining what bug the test prevents

❌ **DON'T**:
- Test only success paths
- Only verify command exit codes (verify actual state!)
- Skip edge cases like untrusted keys, empty inputs, etc.

---

## Documentation Policy

### What NOT to Document

❌ **NEVER create these files**:
- `BUGFIX-*.md` - Bug-specific documentation
- `TEST-COVERAGE-ANALYSIS.md` - Test analysis docs
- `SUMMARY.md` - Temporary status/summary files
- Any other documentation that describes a specific bug or fix

### Why

- Bug-specific documentation becomes stale and clutters the repository
- The information belongs in:
  - **PR descriptions** - Context for reviewers
  - **Commit messages** - Permanent git history
  - **Code comments** - Inline explanations
  - **This SKILL.md** - General patterns and requirements
  - **PR-PROCESS.md** - Process guidelines

### What TO Document

✅ **DO add**:
- **Code comments** - Explain WHY the code does something, especially for bug fixes
- **Test comments** - Explain what bug a regression test prevents
- **SKILL.md updates** - Add general patterns learned from the bug
- **PR-PROCESS.md updates** - Add process improvements (if applicable)

### Example

Instead of creating `BUGFIX-untrusted-keys.md`, add a comment in the code:

```go
// Use --trust-model always to allow encryption to untrusted keys.
// Without this, GPG silently fails when keys have trust level "-" (unknown),
// causing re-encryption to appear successful but not actually update the files.
cmd.Env = append(os.Environ(), "PASSWORD_STORE_GPG_OPTS=--trust-model always")
```

---

## E2E Test Helpers

### Available Helper Functions

```bash
# Assertions
assert_success                        # Command must succeed (exit 0)
assert_failure                        # Command must fail (exit != 0)
assert_output_contains "text"         # Output must contain text
assert_output_not_contains "text"     # Output must NOT contain text
assert_file_exists "path"             # File must exist
assert_dir_exists "path"              # Directory must exist
test_fail "expected" "actual"         # Manual failure with message
test_log "message"                    # Log message (verbose mode)
```

---

## Summary

### Golden Rules

1. **Read PR-PROCESS.md** - Follow the complete process, no shortcuts
2. **Unit tests first** - Debug bugs with focused unit tests
3. **Focus on correctness** - Don't just make tests pass
4. **Every bug needs a test** - That would have caught it
5. **No ephemeral docs** - Use code comments and PR descriptions
6. **Reply to all comments** - See PR-PROCESS.md for details

### Quick Links

- 📋 **[PR-PROCESS.md](PR-PROCESS.md)** - Complete PR submission and review process
- 📝 **[.github/pull_request_template.md](.github/pull_request_template.md)** - PR template with checklists

---

**Remember**: Understanding and following the PR process in PR-PROCESS.md is not optional. It's a requirement for all contributions.
