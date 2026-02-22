## 2026-02-22 - Reusable Interactive Confirmation Pattern
**Learning:** Destructive CLI actions should always be guarded by a confirmation prompt to prevent accidental data loss, but must respect a `--force` flag for scripting and non-interactive environments (CI). Checking for a TTY ensures the CLI doesn't hang when no terminal is present.
**Action:** Use the `Confirm(message string, force bool) bool` helper in `internal/cmd/root.go` for any new destructive commands.
