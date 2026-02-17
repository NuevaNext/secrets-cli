## 2026-02-01 - Argument Injection in CLI Wrappers
**Vulnerability:** User-controlled positional arguments (like emails or secret names) starting with hyphens could be interpreted as flags by underlying CLI tools (`gpg`, `pass`), leading to argument injection.
**Learning:** Wrapping external CLI tools requires careful handling of positional arguments. Even if `exec.Command` prevents shell injection, it doesn't prevent the target binary from misinterpreting arguments as flags.
**Prevention:** Always use the `--` separator to explicitly mark the end of flags and the beginning of positional arguments when calling external binaries with user-provided data.
## 2026-01-31 - Path Traversal and Argument Injection in Secrets CLI
**Vulnerability:** Path traversal in `key add` and `vault create` commands allowed arbitrary file writes via manipulated email or vault names. Additionally, missing `--` separators in GPG/Pass wrappers allowed argument injection.
**Learning:** CLI tools that construct file paths from user arguments are susceptible to traversal if not explicitly sanitized. Positional arguments starting with hyphens can also be misinterpreted as flags by underlying tools.
**Prevention:** Use a centralized validation function to reject dangerous characters (`..`, `/`, `\`) and leading hyphens in names. Always use the `--` delimiter to separate options from positional arguments when calling external binaries.

## 2026-02-17 - Comprehensive Input Validation for Secrets and Vaults
**Vulnerability:** Many CLI entry points (get, set, delete, rename, copy, info, etc.) lacked input validation for vault and secret names, potentially allowing path traversal and argument injection.
**Learning:** Security validation must be applied at every entry point where user input is used to construct file paths or command arguments. Even if some commands are validated, others might be missed.
**Prevention:** Implement distinct validation functions for different input types (e.g., `validateName` for flat names and `validateSecretName` for hierarchical paths) and apply them consistently to all command arguments.
