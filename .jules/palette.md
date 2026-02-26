## 2026-02-26 - Interactive Confirmation for Destructive Actions
**Learning:** For CLI tools, UX is improved by providing interactive confirmation prompts for destructive operations in a TTY, while maintaining strict "fail-fast" behavior in non-interactive environments (CI) unless explicitly overridden by a force flag. Using quoted string formatting (`%q`) for identifiers in prompts helps distinguish them from the surrounding message.
**Action:** Always implement a `Confirm` helper that detects TTY status and use it in all delete/remove command handlers with `%q` for entity names.

## 2026-02-26 - Flexible E2E Test Environments
**Learning:** Hardcoded paths in E2E scripts hinder local testing on developer machines. Providing environment variable overrides with sensible defaults (e.g., `WORKSPACE="${WORKSPACE:-/workspace}"`) allows for greater flexibility and easier debugging without breaking the primary containerized test environment.
**Action:** Use the `${VAR:-default}` shell parameter expansion for critical paths in test scripts to support environment overrides.
