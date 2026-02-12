## 2026-02-01 - Argument Injection in CLI Wrappers
**Vulnerability:** User-controlled positional arguments (like emails or secret names) starting with hyphens could be interpreted as flags by underlying CLI tools (`gpg`, `pass`), leading to argument injection.
**Learning:** Wrapping external CLI tools requires careful handling of positional arguments. Even if `exec.Command` prevents shell injection, it doesn't prevent the target binary from misinterpreting arguments as flags.
**Prevention:** Always use the `--` separator to explicitly mark the end of flags and the beginning of positional arguments when calling external binaries with user-provided data.
## 2026-01-31 - Path Traversal and Argument Injection in Secrets CLI
**Vulnerability:** Path traversal in `key add` and `vault create` commands allowed arbitrary file writes via manipulated email or vault names. Additionally, missing `--` separators in GPG/Pass wrappers allowed argument injection.
**Learning:** CLI tools that construct file paths from user arguments are susceptible to traversal if not explicitly sanitized. Positional arguments starting with hyphens can also be misinterpreted as flags by underlying tools.
**Prevention:** Use a centralized validation function to reject dangerous characters (`..`, `/`, `\`) and leading hyphens in names. Always use the `--` delimiter to separate options from positional arguments when calling external binaries.

## 2026-02-12 - Path Traversal in Secret Names
**Vulnerability:** Secret names were used in file paths without validation, allowing path traversal (e.g., `../../something`). While vault names were validated, the secret names themselves were not.
**Learning:** Hierarchical secret names (e.g., `database/password`) require a specialized validation function (`validateSecretName`) that allows slashes but strictly forbids traversal patterns like `..`, double slashes `//`, and backslashes.
**Prevention:** Implement and consistently use `validateSecretName` for all secret-related commands and `validateName` for all vault and email-related commands at the CLI entry point.
