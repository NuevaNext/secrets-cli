## 2026-02-12 - Interactive Deletion Safeguards
**Learning:** Destructive actions in CLI tools should always prompt for confirmation in interactive sessions, while maintaining non-interactive support via a force flag. This prevents accidental data loss without breaking automation.
**Action:** Always implement a `confirm` helper and `isTerminal` check for any deletion or irreversible operation.
