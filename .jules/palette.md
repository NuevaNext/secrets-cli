## 2026-02-25 - Interactive Confirmation Pattern
**Learning:** Destructive CLI actions (deleting secrets, vaults, keys) should always provide an interactive confirmation prompt when run in a TTY, while maintaining a --force flag for automation and non-interactive environments.
**Action:** Use the `Confirm(message string, force bool) bool` helper in `internal/cmd/root.go` for all destructive operations.
