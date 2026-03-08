# Workbuddy

Workbuddy is a local-first CLI for notes and tasks. It stores everything in SQLite, supports tag-based organization, and is designed to stay fast and scriptable from the terminal.

This branch, `productize-workbuddy`, shifts the project from a learning exercise toward a practical tool with a clearer product direction.

## Current Features

- Create notes inline or in your editor
- Update notes inline or in your editor
- Create tasks and mark them as completed
- Organize notes with tags
- Search by tag, task status, and completion state
- Store data locally in SQLite with automatic schema migrations

## Installation

Requires Go 1.25 or newer.

```bash
git clone https://github.com/Christianrcds/workbuddy.git
cd workbuddy
go install
```

Make sure your Go bin directory is in `PATH`. On first run, Workbuddy creates its database automatically at `~/.config/workbuddy/workbuddy.db`.

## Quick Start

Create a note:

```bash
workbuddy create -c "Write weekly plan"
workbuddy create -c "Call Alice" -t personal,follow-up
```

Create a task:

```bash
workbuddy create -c "Ship release notes" --task
```

Create a note in your editor:

```bash
workbuddy create
workbuddy create -t work,urgent
```

List notes and tags:

```bash
workbuddy list
workbuddy tags
```

Search:

```bash
workbuddy search
workbuddy search "release notes"
workbuddy search integrations
workbuddy search -t work
workbuddy search integrations -t work
workbuddy search --tasks
workbuddy search --pending
workbuddy search --completed
workbuddy search -t work --pending -l 10
```

Complete or remove items:

```bash
workbuddy check 12
workbuddy update 12 -c "Updated content"
workbuddy update 12
workbuddy update 12 --add-tag urgent --remove-tag someday
workbuddy update 12 --set-tags work,backend
workbuddy remove 12
```

## Command Notes

- `workbuddy create` is the primary creation command and supports tags plus editor-based input.
- `workbuddy update` updates content inline, and opens the current note in your editor when `-c` is omitted.
- `workbuddy update` also supports tag changes with `--add-tag`, `--remove-tag`, and `--set-tags`.
- If `-c` is omitted, `update` opens the editor even when tag flags are also present.
- Tags accept comma-separated values or repeated flags: `-t work,urgent` or `-t work -t urgent`.
- `workbuddy search [query]` accepts an optional content query plus the existing filters.
- `--completed` and `--pending` are mutually exclusive on `search`.
- `check` only completes tasks.
- `remove` currently deletes notes by ID after confirmation.

## Configuration

| Variable | Default | Description |
|---|---|---|
| `$EDITOR` | Required for editor mode | Editor opened by `workbuddy create` when `-c` is omitted |
| `$WORKBUDDY_DB` | `~/.config/workbuddy/workbuddy.db` | Custom database path |

## Architecture

Workbuddy keeps a small, explicit structure:

- `cmd/` contains the Cobra CLI commands
- `internal/note/` contains the service layer, repository abstraction, sqlc models, and generated queries
- `migrations/` contains schema changes applied automatically at startup

Core stack:

- Go 1.25.6
- SQLite via `modernc.org/sqlite`
- sqlc for typed SQL access
- golang-migrate for schema migrations
- Cobra for the CLI
- lipgloss for terminal output

## Product Direction

The next phase of Workbuddy should focus on usefulness, reliability, and a cleaner CLI experience.

Near-term priorities:

- Add content search using the existing FTS groundwork
- Improve validation and user-facing error messages
- Expand integration tests around commands, migrations, and database behavior
- Improve list and search performance by reducing repeated tag lookups

## Contributing

Contributions should prioritize stability, usability, and predictable CLI behavior. Small, focused changes are preferred, especially when they improve command consistency, error handling, tests, or search.
