## 2026-02-20 - Interactive confirmation for destructive CLI actions
**Learning:** Destructive actions in CLIs should not just fail when a force flag is missing, but should offer an interactive prompt in TTY environments to improve user flow while maintaining safety.
**Action:** Use a `Confirm` helper that checks for TTY (os.Stdin.Stat().Mode() & os.ModeCharDevice) to provide interactivity without breaking CI/CD pipelines.
