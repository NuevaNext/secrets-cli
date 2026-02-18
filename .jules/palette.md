## 2026-02-18 - Interactive Confirmation for Destructive Actions
**Learning:** Hard-failing on missing safety flags (like --force) in interactive CLI sessions is a frustrating UX. Users expect the tool to ask them if they are sure before failing.
**Action:** Use a terminal-aware `Confirm` helper for all destructive actions to provide a "smooth" experience for interactive users while maintaining safety for automated scripts.
