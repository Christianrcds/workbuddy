# Workbuddy

A CLI note-taking application with tag-based organization, task tracking, and local SQLite storage.

## Installation

Requires Go 1.25 or higher.

```bash
git clone https://github.com/Christianrcds/workbuddy.git
cd workbuddy
go install
```

Make sure `$GOPATH/bin` (usually `~/go/bin`) is in your `$PATH`. On first run, the database is created automatically at `~/.config/workbuddy/workbuddy.db`.

## Usage

### Create a note

Without `-c`, workbuddy opens your `$EDITOR` — similar to `git commit` without `-m`:

```bash
workbuddy nt                              # opens editor
workbuddy nt -t work,urgent               # opens editor, tags applied on save
workbuddy nt -c "Note content"            # inline
workbuddy nt -c "Note content" -t work    # inline with tags
```

Tags accept comma-separated values or repeated flags: `-t tag1,tag2` or `-t tag1 -t tag2`.

### Create a task

```bash
workbuddy nt -c "Buy milk" --task
workbuddy nt -t work --task               # opens editor
```

### List & tags

```bash
workbuddy list    # all notes
workbuddy tags    # all tags
```

### Search

```bash
workbuddy search                          # all notes (up to 5)
workbuddy search -t work                  # filter by tag
workbuddy search -l 10                    # custom limit
workbuddy search --tasks                  # tasks only (-k)
workbuddy search --pending                # incomplete tasks (-p)
workbuddy search --completed              # completed tasks (-c)
workbuddy search -t work --pending -l 10  # flags can be combined
```

> `--completed` and `--pending` are mutually exclusive.

### Check & remove

```bash
workbuddy check <id>     # mark task as completed
workbuddy remove <id>    # delete note (asks for confirmation)
```

## Configuration

| Variable | Default | Description |
|---|---|---|
| `$EDITOR` | _(none — required for editor mode)_ | Editor opened by `workbuddy nt` |
| `$WORKBUDDY_DB` | `~/.config/workbuddy/workbuddy.db` | Custom database path |

## Technologies

**Go 1.25.6** · **SQLite** · **sqlc** · **golang-migrate** · **Cobra** · **lipgloss**

## Contributing

This is a learning project for understanding Go. Feel free to explore, fork, and experiment!
