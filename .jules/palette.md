## 2026-02-23 - Interactive Confirmation for Destructive Actions
**Learning:** For a security-focused CLI tool like secrets-cli, providing immediate feedback and confirmation for destructive actions (vault delete, secret delete, key remove) significantly improves user safety and confidence. Balancing this with a --force flag ensures that automated scripts remain functional.
**Action:** Always implement a Confirm helper that checks for TTY status before prompting, and provide a --force override for consistency.
