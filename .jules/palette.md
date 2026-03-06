## 2026-03-06 - Interactive Confirmation for Destructive CLI Actions
**Learning:** For destructive operations in a CLI, failing with an error message requiring a `--force` flag can be frustrating for human users. Providing an interactive confirmation prompt (`[y/N]`) instead improves usability without sacrificing safety.
**Action:** Implement a shared `Confirm` helper that detects terminal environments and provides an interactive prompt for destructive actions, while still supporting a bypass flag for automated workflows.
