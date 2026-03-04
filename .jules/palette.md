## 2026-03-04 - Terminal Detection for Interactive Prompts
**Learning:** For CLI tools, always check if stdin is a terminal using `os.Stdin.Stat()` before prompting for interactive input. This prevents the tool from hanging in CI/CD environments or when piped, while still providing a better experience for human users.
**Action:** Use the `Confirm` helper pattern which checks `os.ModeCharDevice` and returns a safe default (false) for non-terminal environments.
