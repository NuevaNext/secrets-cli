## 2026-03-01 - Interactive Confirmation Pattern
**Learning:** Destructive CLI actions should support both interactive confirmation and a --force flag. Using `os.Stdin.Stat()` to detect `os.ModeCharDevice` allows the tool to provide a helpful prompt in terminals while safely failing in non-interactive environments (CI/CD), preventing hangs.
**Action:** Use the `Confirm(message string, force bool) bool` pattern for all destructive operations to improve micro-UX without breaking automation.
