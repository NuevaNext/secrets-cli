## 2026-02-11 - Prompt for stdin reading in CLI
**Learning:** CLI tools that read from stdin without a prompt can appear to "hang" when run interactively, causing user confusion.
**Action:** Always check if stdin is a terminal using `isTerminal` (via `os.ModeCharDevice`) and provide a descriptive prompt (e.g., "Enter value: ") when reading interactively, while remaining silent for piped/automated input.
