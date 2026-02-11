# Why E2E Tests Didn't Catch the Re-encryption Bug

## The Problem

The e2e tests passed even though the re-encryption bug existed. Here's why:

## Root Cause: Trusted vs Untrusted Keys

### What the Tests Did
```bash
# In setup_gpg_keys()
generate_gpg_key "$ALICE_EMAIL" "Alice Test"
generate_gpg_key "$BOB_EMAIL" "Bob Test"
```

Both Alice and Bob's keys were generated **in the same GPG keyring** during test setup. This means:
- Bob's key was **automatically trusted** (ultimate trust)
- GPG had no problem encrypting to Bob's key in `--batch` mode
- The re-encryption succeeded even without `--trust-model always`

### What Happens in Real World
```bash
# User imports an external key
gpg --import ryan@nuevanext.com.asc
# Trust level: - (unknown/untrusted)
```

When you import an external key:
- The key has trust level `-` (unknown) or `q` (undefined)
- GPG **refuses** to encrypt to untrusted keys in `--batch` mode
- The bug manifests: re-encryption silently fails

## What the Test Checked

### Original Test (Insufficient)
```bash
test_bob_can_read_dev_secret() {
    # Only checked if Bob could read the secret
    assert_success "$SECRET_CLI --email $BOB_EMAIL get dev database/password"
    assert_output_equals "my-secret-password-123"
}
```

**Problem**: This test only verified that:
1. The command succeeded
2. Bob could decrypt the secret

It **did not verify** that the `.gpg` file was actually re-encrypted for Bob's key.

### Why It Passed Despite the Bug

Bob could still read the secret because:
1. Bob's key was trusted (generated locally)
2. The re-encryption actually worked for trusted keys
3. The bug only manifests with **untrusted** keys

## The Fix

### 1. Enhanced E2E Test
```bash
test_bob_can_read_dev_secret() {
    # ... existing checks ...
    
    # CRITICAL: Verify the .gpg file is actually encrypted for Bob's key
    local secret_file=".secrets/vaults/dev/.password-store/database/password.gpg"
    local bob_key_id=$(gpg --list-keys --with-colons "$BOB_EMAIL" | grep '^sub:' | grep ':e:' | head -1 | cut -d: -f5)
    
    # Check if Bob's key ID appears in the encrypted file's recipients
    local recipients=$(gpg --list-packets "$secret_file" 2>&1 | grep 'keyid')
    
    if ! echo "$recipients" | grep -q "$bob_key_id"; then
        test_fail "Secret NOT encrypted for Bob's key"
        return 1
    fi
}
```

This now verifies the **actual encryption recipients** in the `.gpg` file.

### 2. Untrusted Key Test
Created `test-untrusted-key-reencryption.sh` that:
1. Generates a key in a **separate keyring** (simulates external key)
2. Imports it as **untrusted**
3. Adds the untrusted key as a vault member
4. **Verifies** the secret is actually encrypted for that key

This test would have **failed** before the fix and **passes** after.

## Lessons Learned

### Test What Actually Matters
- ✗ Testing if a command succeeds
- ✗ Testing if the user can read the secret
- ✅ Testing if the **underlying data structure** is correct

### Simulate Real-World Conditions
- ✗ Using only locally-generated trusted keys
- ✅ Simulating external/untrusted keys
- ✅ Testing edge cases that users will encounter

### Verify Side Effects
When a command claims to do something (e.g., "Re-encrypted 5 secret(s)"):
- ✗ Trust the success message
- ✅ Verify the actual files changed
- ✅ Check the encryption recipients match expectations

## Test Coverage Improvements

### Before
- ✓ Can add member to vault
- ✓ Member can read secrets
- ✗ Secrets are actually encrypted for member

### After
- ✓ Can add member to vault
- ✓ Member can read secrets
- ✓ Secrets are actually encrypted for member ← **NEW**
- ✓ Works with untrusted keys ← **NEW**

## Recommendation

For any cryptographic operation that claims to encrypt/re-encrypt:
1. **Verify the actual encryption** using `gpg --list-packets`
2. **Test with untrusted keys** to simulate real-world scenarios
3. **Don't rely on command success** — verify the underlying data
