## 2024-05-23 - Interactive Confirmations for Destructive Actions
**Learning:** Destructive actions like deleting vaults or secrets should ideally offer an interactive confirmation prompt when run in a terminal, rather than just failing and demanding a `--force` flag. This provides a smoother user experience for manual operations while maintaining safety.
**Action:** Implement a shared `Confirm` helper that handles both interactive terminals (prompting) and non-interactive environments (requiring `--force`).
