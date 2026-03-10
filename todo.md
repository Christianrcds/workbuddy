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

## Near-Term Priorities

### 5. Improve Tag Loading Performance

Goal: avoid repeated database queries when listing or searching notes.

- Replace the current per-note tag lookup pattern with a batched query.
- Measure the current flow and compare it with the batched version.
- Keep the CLI output unchanged while improving the data-loading path.
- Add tests for notes with no tags, one tag, and multiple tags.

### 6. Expand Integration Tests

Goal: verify real workflows against SQLite and migrations.

- Add end-to-end tests for create, search, list, check, remove, and update flows.
- Test migrations against a fresh database and an already-initialized database.
- Cover invalid inputs and user-facing error cases.
- Add regression tests for duplicate tags, empty notes, and task-only operations.
- Keep tests easy to run locally with `go test ./...`.

## Brainstorm

Ideas worth exploring after the near-term priorities are underway.

- Add recurring tasks with simple cadence rules.
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
