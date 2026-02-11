## 2026-02-01 - Argument Injection in CLI Wrappers
**Vulnerability:** User-controlled positional arguments (like emails or secret names) starting with hyphens could be interpreted as flags by underlying CLI tools (`gpg`, `pass`), leading to argument injection.
**Learning:** Wrapping external CLI tools requires careful handling of positional arguments. Even if `exec.Command` prevents shell injection, it doesn't prevent the target binary from misinterpreting arguments as flags.
**Prevention:** Always use the `--` separator to explicitly mark the end of flags and the beginning of positional arguments when calling external binaries with user-provided data.
## 2026-01-31 - Path Traversal and Argument Injection in Secrets CLI
**Vulnerability:** Path traversal in `key add` and `vault create` commands allowed arbitrary file writes via manipulated email or vault names. Additionally, missing `--` separators in GPG/Pass wrappers allowed argument injection.
**Learning:** CLI tools that construct file paths from user arguments are susceptible to traversal if not explicitly sanitized. Positional arguments starting with hyphens can also be misinterpreted as flags by underlying tools.
**Prevention:** Use a centralized validation function to reject dangerous characters (`..`, `/`, `\`) and leading hyphens in names. Always use the `--` delimiter to separate options from positional arguments when calling external binaries.
## 2026-02-11 - Pervasive Lack of Input Validation in CLI Commands
**Vulnerability:** Most CLI commands (vault, secrets, export) accepted user-provided names (vaults, emails, secrets) without any validation, allowing path traversal and argument injection.
**Learning:** While some specific commands (like `key add`) had validation, the rest of the application was inconsistently secured. A secure architecture requires centralized and consistently applied input validation for all user-controlled data that enters file paths or external command arguments.
**Prevention:** Implement centralized validation helpers (like `validateName` and `validateSecretName`) and apply them at the earliest possible entry point for every CLI command that accepts user input.
