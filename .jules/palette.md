## 2026-03-09 - Interactive Confirmation Pattern for CLI
**Learning:** Destructive CLI actions should have a safety net. While a --force flag is standard for automation, interactive users benefit greatly from a confirmation prompt. Detecting the terminal environment via os.Stdin.Stat() allows for a seamless transition between interactive use and automated scripts.
**Action:** Use the shared Confirm helper in internal/cmd/root.go for all destructive operations (delete, remove, etc.) to ensure a consistent and safe user experience.
