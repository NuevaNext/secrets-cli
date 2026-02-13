## 2026-02-13 - Destructive CLI Actions and User Safety
**Learning:** Destructive actions (like deleting vaults or secrets) should provide an interactive confirmation prompt when run in a terminal. This prevents accidental data loss while allowing automation via a `--force` flag. Checking for a terminal using `os.ModeCharDevice` ensures the tool doesn't hang in non-interactive environments.
**Action:** Always implement a `confirm` helper that checks `isTerminal` before prompting, and provide a `--force` flag for all destructive operations.
