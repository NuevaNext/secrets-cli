## 2026-02-17 - Interactive Confirmation for Destructive Actions
**Learning:** Hard-failing a command when a required confirmation flag is missing (like `--force`) creates friction for interactive users. Providing a prompt instead allows the user to stay in their flow while still maintaining safety for scripts.
**Action:** Use a `Confirm` helper that checks if stdin is a terminal before prompting, ensuring non-interactive environments still require the explicit flag.
