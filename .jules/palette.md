# Palette's Journal - secrets-cli

## 2026-02-24 - Interactive Confirmations for Destructive Actions
**Learning:** Destructive actions should always offer an interactive confirmation if not explicitly forced via a flag. This prevents accidental data loss and improves the user's sense of safety. Using `%q` for entity names in prompts makes them stand out and clearly identifies what is being deleted.
**Action:** Implement a reusable `Confirm` helper and apply it to all destructive commands (`vault delete`, `delete`, `key remove`).
