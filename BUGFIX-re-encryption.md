# Bug Fix: Silent Re-encryption Failure

## Issue
When adding a new member to a vault, `secrets-cli` would report success ("✓ Re-encrypted N secret(s)") but the `.gpg` files were **not actually re-encrypted** to include the new member's key.

## Root Cause
The `pass` password manager uses GPG in `--batch` mode for re-encryption. When a GPG key has trust level `-` (unknown/untrusted), GPG **silently refuses** to encrypt to that key in batch mode:

```
gpg: 252D2A4F4CFF6725: There is no assurance this key belongs to the named user
gpg: [stdin]: encryption failed: Unusable public key
```

The `pass init` command would:
1. Update the `.gpg-id` file ✅
2. Attempt to re-encrypt all secrets
3. **Fail silently** when GPG refused to encrypt to untrusted keys ❌
4. Return success anyway ❌

## Impact
- New team members added to vaults could not decrypt secrets
- No error was shown to the user
- The vault config showed the member as added, creating a false sense of security

## Fix
Two changes were made to `/home/omar/src/github.com/nuevanext/secrets-cli/internal/pass/pass.go`:

### 1. Trust Model Override
Set `PASSWORD_STORE_GPG_OPTS=--trust-model always` when running `pass` commands:

```go
func (p *Pass) run(args ...string) (string, error) {
    cmd := exec.Command("pass", args...)
    cmd.Env = append(os.Environ(),
        "PASSWORD_STORE_DIR="+p.StoreDir,
        "PASSWORD_STORE_GPG_OPTS=--trust-model always",  // ← FIX
    )
    // ...
}
```

This tells GPG to encrypt to all specified recipients regardless of trust level.

### 2. Re-encryption Verification
Added `VerifyEncryption()` function that checks if secrets are actually encrypted for all expected recipients:

```go
func (p *Pass) ReInit(gpgIDs []string) error {
    // ... write .gpg-id and run pass init ...
    
    // Verify re-encryption succeeded
    secrets, err := p.List()
    if err != nil {
        return fmt.Errorf("failed to list secrets after re-init: %w", err)
    }
    
    if len(secrets) > 0 {
        if err := p.VerifyEncryption(secrets[0], gpgIDs); err != nil {
            return fmt.Errorf("re-encryption verification failed: %w", err)
        }
    }
    
    return nil
}
```

The verification:
- Uses `gpg --list-packets` to extract actual recipient key IDs from `.gpg` files
- Compares against expected key IDs from `gpg --list-keys`
- Returns an error if any expected recipient is missing

## Testing
Before fix:
```bash
$ gpg --list-packets .secrets/vaults/local/.password-store/DOCKERHUB_PASSWORD.gpg | grep keyid
:pubkey enc packet: version 3, algo 1, keyid 901B42D22BFC9E7C
# Only Omar's key ❌
```

After fix:
```bash
$ gpg --list-packets .secrets/vaults/local/.password-store/DOCKERHUB_PASSWORD.gpg | grep keyid
:pubkey enc packet: version 3, algo 1, keyid 901B42D22BFC9E7C
:pubkey enc packet: version 3, algo 1, keyid 252D2A4F4CFF6725
# Both Omar and Ryan's keys ✅
```

## Lessons Learned
1. **Never trust silent success** — Always verify critical operations
2. **GPG trust model matters** — In batch mode, GPG is strict about key trust
3. **Test with real scenarios** — The bug only appeared when adding keys that weren't in the trust database

## Related Files
- `/home/omar/src/github.com/nuevanext/secrets-cli/internal/pass/pass.go` — Core fix
- `/home/omar/src/github.com/nuevanext/secrets-cli/internal/cmd/vault.go` — Calls ReInit()
