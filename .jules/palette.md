## 2026-02-19 - Interactive Confirmation for Destructive CLI Actions
**Learning:** Destructive CLI actions like deleting vaults or secrets should ideally offer an interactive confirmation prompt when run in a terminal, rather than strictly requiring a --force flag. This improves usability for human users while maintaining safety for automated scripts.
**Action:** Use a `Confirm` helper that checks for a terminal (os.ModeCharDevice) to provide interactive prompts while falling back to non-interactive failure (requiring --force) in CI/CD environments.
