# Workbuddy TODO

Product-focused backlog for the `productize-workbuddy` branch.

## Recently Completed

### 1. Unify Note Creation Flow

Status: Done

- Standardized on `workbuddy create` as the single creation command.
- Removed the duplicate `nt` command path.
- Routed note and task creation through the same service flow.
- Kept editor mode, tags, and task creation under one command surface.
- Updated README examples to reflect the new command.

### 2. Add Content Search

Status: Done

- Added content search to `workbuddy search [query]`.
- Backed content search with SQLite FTS.
- Preserved tag, task, pending, and completed filters.
- Added tests for content matches and combined filter behavior.

### 3. Add Update Command

Status: Done

- Added `workbuddy update <id> -c "..."` for inline note edits.
- Added editor-based updating to `workbuddy update <id>` when `-c` is omitted.
- Added tag updates to `workbuddy update` with `--add-tag`, `--remove-tag`, and `--set-tags`.
- Reused the existing editor workflow with seeded note content.
- Added repository and service support for note lookup and content updates.
- Removed the duplicate `edit` command to keep one canonical update path.

### 5. Improve Tag Loading Performance

Status: Done

- Replaced the per-note tag lookup pattern with a batched query.
- Kept the CLI output unchanged while improving the data-loading path.
- Added tests for notes with no tags, one tag, and multiple tags.

### 6. Expand Integration Tests

Status: Done

- Added end-to-end tests for create, search, list, check, remove, and update flows.
- Tested migrations against a fresh database and an already-initialized database.
- Covered invalid inputs and user-facing error cases.
- Added regression tests for duplicate tags, empty notes, and task-only operations.
- Kept the tests easy to run locally with `go test ./...`.

### 11. Add Recurring Tasks With Simple Cadence Rules

Status: Done

- Added recurring task series backed by schema migrations instead of treating recurrence as a per-row flag.
- Added `daily`, `weekly`, and `monthly` recurrence with validation for due dates, weekdays, and day-of-month rules.
- Extended `create`, `update`, `check`, `remove`, `list`, and `search` so recurring tasks behave consistently in the CLI.
- Preserved completed history when removing a recurring series, while preventing duplicate pending occurrences.
- Added service and CLI integration tests for recurrence validation, series updates, completion rollover, and monthly clamping.

## Near-Term Priorities

### 7. Improve Validation and User Feedback

Goal: make invalid operations fail clearly and predictably across commands.

- Audit command error messages for consistency around IDs, empty content, conflicting flags, and missing records.
- Push remaining validation rules into the service layer where behavior should be shared.
- Add CLI-focused tests for the most important failure cases.
- Update README examples if any command behavior or wording changes.

### 8. Improve Search and List Ergonomics

Goal: make common retrieval workflows faster and easier from the terminal.

- Add sorting options for `workbuddy list` and `workbuddy search`.
- Support richer filter combinations, starting with multiple tags.
- Keep output behavior predictable across styled terminal output and future scripting modes.
- Add tests for combined filters and sort behavior.

### 9. Add Due Dates and Overdue Task Views

Goal: make task tracking more useful for real daily planning.

- Finish the broader due-date product after the recurrence groundwork already landed.
- Add overdue and upcoming task views without weakening the current command model.
- Cover parsing, validation, and overdue filtering in tests.

### 10. Add Tag Management Commands

Goal: manage tag cleanup without direct database edits.

- Add commands to rename and delete tags safely.
- Keep note-tag relationships consistent when tags are renamed or removed.
- Decide and document the command semantics before implementation.
- Add integration tests for tag rename/delete workflows.

## Brainstorm

Ideas worth exploring after the near-term priorities are underway.

- Add due dates and overdue task views.
- Add priorities for tasks and sorting by urgency.
- Support note archiving instead of only deletion.
- Add tag management commands such as rename and delete.
- Add bulk operations such as complete/delete/tag multiple notes at once.
- Add richer search filters such as date ranges and multiple tags.
- Add saved searches or named views.
- Add JSON export and import for backups and migration.
- Add Markdown-aware note rendering for richer terminal reading.
- Add shell completion support for common commands and flags.
- Add a config file for defaults such as database path, output mode, and editor.
- Add optional plain-text output for scripting in addition to styled terminal output.
- Add `workbuddy today` or `workbuddy agenda` views for daily task review.
- Add sorting options for list and search results.
- Add pinning or starring for important notes.
- Add note/task IDs in a more user-friendly format if numeric IDs become cumbersome.
- Add note templates for recurring workflows such as meeting notes or journaling.
- Add note duplication for quick reuse.
- Add a timeline/history command to review recent note and task activity.
- Add attachments or file references stored alongside notes.
- Add clipboard integration for quick note capture.
- Add reminder hooks for due tasks via local notifications.
- Add database backup and restore commands.
- Add database integrity/doctor checks for troubleshooting.
- Add an interactive TUI mode later, after the core CLI is stable.
- Add syncing or remote backup support later, after the local experience is stable.
