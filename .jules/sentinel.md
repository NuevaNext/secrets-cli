## 2026-02-15 - Inconsistent Input Validation across CLI Commands
**Vulnerability:** While `validateName` was implemented, it was inconsistently applied across commands. `init`, `setup`, `key remove`, `vault info/delete/add-member/remove-member`, and all `secrets` commands lacked validation, exposing them to path traversal and argument injection.
**Learning:** Security controls must be applied systemically to all entry points. Hierarchical inputs (like secret paths) require specialized validation (`validateSecretName`) that balances organizational needs with security (allowing `/` but preventing `//`, `..`, etc.).
**Prevention:** Use a centralized validation strategy and ensure every CLI command that accepts user-provided names or paths invokes the appropriate validator immediately after argument parsing.

## 2026-02-01 - Argument Injection in CLI Wrappers
**Vulnerability:** User-controlled positional arguments (like emails or secret names) starting with hyphens could be interpreted as flags by underlying CLI tools (`gpg`, `pass`), leading to argument injection.
**Learning:** Wrapping external CLI tools requires careful handling of positional arguments. Even if `exec.Command` prevents shell injection, it doesn't prevent the target binary from misinterpreting arguments as flags.
**Prevention:** Always use the `--` separator to explicitly mark the end of flags and the beginning of positional arguments when calling external binaries with user-provided data.
## 2026-01-31 - Path Traversal and Argument Injection in Secrets CLI
**Vulnerability:** Path traversal in `key add` and `vault create` commands allowed arbitrary file writes via manipulated email or vault names. Additionally, missing `--` separators in GPG/Pass wrappers allowed argument injection.
**Learning:** CLI tools that construct file paths from user arguments are susceptible to traversal if not explicitly sanitized. Positional arguments starting with hyphens can also be misinterpreted as flags by underlying tools.
**Prevention:** Use a centralized validation function to reject dangerous characters (`..`, `/`, `\`) and leading hyphens in names. Always use the `--` delimiter to separate options from positional arguments when calling external binaries.
