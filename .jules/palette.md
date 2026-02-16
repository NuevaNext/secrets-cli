## 2026-02-16 - Interactive CLI Confirmation Pattern
**Learning:** For destructive operations in CLI tools, requiring a `--force` flag is safe but can be frustrating for interactive users. A better UX pattern is to detect if the session is a terminal and provide an interactive confirmation prompt if the flag is missing.
**Action:** Use `isTerminal(os.Stdout)` and a `confirm()` helper in future CLI projects to provide a smoother interactive experience while maintaining script safety.
