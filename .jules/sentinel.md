## 2026-03-06 - Path Traversal in Vault and Secret Commands
**Vulnerability:** Command handlers for vault and secret operations (e.g., `vault delete`, `list`, `get`, `set`, `delete`, `rename`, `copy`, `export`, `sync`) were missing input validation for vault names and secret paths, allowing path traversal attacks.
**Learning:** Even if some commands (like `vault create`) have validation, other commands that use the same input to construct file paths might be overlooked, creating inconsistent security coverage.
**Prevention:** Systematically apply `validateName` (for vaults and emails) and `validateSecretName` (for secret paths) to all user-controlled positional arguments and flags that are used in filesystem operations.
