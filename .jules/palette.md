# Palette's UX Journal

## 2025-05-14 - Interactive Confirmation Pattern for Destructive Actions
**Learning:** Destructive CLI commands should favor interactive confirmation over immediate failure when run in a terminal. This provides a smoother experience than requiring the user to re-run the command with a `--force` flag.
**Action:** Use the `Confirm(message string, force bool) bool` helper in `internal/cmd/root.go` for all destructive actions. Ensure the helper handles non-interactive environments by failing safely unless forced.
