## 2026-02-01 - Argument Injection in CLI Wrappers
**Vulnerability:** User-controlled positional arguments (like emails or secret names) starting with hyphens could be interpreted as flags by underlying CLI tools (`gpg`, `pass`), leading to argument injection.
**Learning:** Wrapping external CLI tools requires careful handling of positional arguments. Even if `exec.Command` prevents shell injection, it doesn't prevent the target binary from misinterpreting arguments as flags.
**Prevention:** Always use the `--` separator to explicitly mark the end of flags and the beginning of positional arguments when calling external binaries with user-provided data.
## 2026-01-31 - Path Traversal and Argument Injection in Secrets CLI
**Vulnerability:** Path traversal in `key add` and `vault create` commands allowed arbitrary file writes via manipulated email or vault names. Additionally, missing `--` separators in GPG/Pass wrappers allowed argument injection.
**Learning:** CLI tools that construct file paths from user arguments are susceptible to traversal if not explicitly sanitized. Positional arguments starting with hyphens can also be misinterpreted as flags by underlying tools.
**Prevention:** Use a centralized validation function to reject dangerous characters (`..`, `/`, `\`) and leading hyphens in names. Always use the `--` delimiter to separate options from positional arguments when calling external binaries.
## 2026-02-16 - Path Traversal in Secret Names
**Vulnerability:** While vault names were validated, secret names were not, allowing potential path traversal (e.g., `../../file`) even if underlying tools had some protection.
**Learning:** Defense in depth requires validation at the application level, even if delegated tools (like `pass`) have their own protections. Secret names need a specific validator that allows slashes for organization but rejects traversal patterns.
**Prevention:** Implement `validateSecretName` to allow single slashes but reject `..`, leading/trailing slashes, and double slashes. Apply this validation to all commands handling secret names.
