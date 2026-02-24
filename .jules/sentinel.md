## 2026-02-01 - Argument Injection in CLI Wrappers
**Vulnerability:** User-controlled positional arguments (like emails or secret names) starting with hyphens could be interpreted as flags by underlying CLI tools (`gpg`, `pass`), leading to argument injection.
**Learning:** Wrapping external CLI tools requires careful handling of positional arguments. Even if `exec.Command` prevents shell injection, it doesn't prevent the target binary from misinterpreting arguments as flags.
**Prevention:** Always use the `--` separator to explicitly mark the end of flags and the beginning of positional arguments when calling external binaries with user-provided data.
## 2026-01-31 - Path Traversal and Argument Injection in Secrets CLI
**Vulnerability:** Path traversal in `key add` and `vault create` commands allowed arbitrary file writes via manipulated email or vault names. Additionally, missing `--` separators in GPG/Pass wrappers allowed argument injection.
**Learning:** CLI tools that construct file paths from user arguments are susceptible to traversal if not explicitly sanitized. Positional arguments starting with hyphens can also be misinterpreted as flags by underlying tools.
**Prevention:** Use a centralized validation function to reject dangerous characters (`..`, `/`, `\`) and leading hyphens in names. Always use the `--` delimiter to separate options from positional arguments when calling external binaries.

## 2026-02-22 - Argument Injection and Path Traversal in Secret Commands
**Vulnerability:** Command handlers for secret operations (`get`, `set`) and vault member management (`add-member`) were missing input validation, potentially allowing path traversal. Additionally, some internal GPG/Pass calls lacked the `--` separator, risking argument injection.
**Learning:** Even with existing validation functions like `validateName`, it is easy to miss applying them to new command handlers or specific arguments like secret paths that require different rules (e.g., allowing slashes).
**Prevention:** Apply `validateName` or `validateSecretName` to all user-controlled positional arguments. Ensure all CLI wrappers use the `--` separator before positional arguments to prevent them from being interpreted as flags.

## 2026-02-24 - Comprehensive Input Validation Across All CLI Commands
**Vulnerability:** Multiple command handlers (`list`, `delete`, `rename`, `copy`, `vault info`, `vault delete`, `vault remove-member`, `key remove`, `export`, `sync`, `init`, `setup`) were missing input validation for vault names, secret names, or member emails, risking path traversal.
**Learning:** Security validation must be applied comprehensively to all entry points, not just the most obvious ones (like `get`/`set`). Even helper commands like `init` or `setup` can be vulnerable if they use user-provided data to construct paths.
**Prevention:** Perform a comprehensive audit of all command handlers to ensure `validateName` or `validateSecretName` is applied to every user-controlled argument used in file system operations.
