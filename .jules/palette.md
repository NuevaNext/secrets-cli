## 2026-03-02 - Interactive Confirmation for CLI Safety
**Learning:** For destructive CLI actions, interactive confirmation prompts [y/N] enhance UX by preventing accidental data loss, but must detect non-terminal environments (e.g., CI/CD) to avoid hanging processes.
**Action:** Use `os.Stdin.Stat()` to check for `os.ModeCharDevice` before prompting for user input, and always provide a `--force` flag to bypass the prompt for automation.
