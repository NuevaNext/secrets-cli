## 2025-05-14 - Interactive Confirmation for Destructive CLI Actions
**Learning:** Destructive operations (like deleting vaults or keys) should always offer an interactive confirmation prompt when run in a terminal. This prevents accidental data loss while maintaining workflow efficiency for power users via a `--force` flag.
**Action:** Use the `Confirm` helper in `internal/cmd/root.go` for all destructive actions. Ensure the prompt clearly identifies the entity being deleted using quoted formatting (`%q`).
