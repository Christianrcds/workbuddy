# Workbuddy

A CLI note-taking application with tag-based organization and full-text search.

## Features

- Create and manage notes from the command line
- Organize notes with tags
- Full-text search across all notes
- SQLite database for fast, local storage

## Installation

### Prerequisites

- Go 1.25 or higher

### Setup

1. **Clone the repository:**

   ```bash
   git clone https://github.com/Christianrcds/workbuddy.git
   cd workbuddy
   ```

2. **Choose your installation method:**

#### Option 1: Install with `go install` (Recommended)

This allows you to run `workbuddy` from anywhere on your system:

```bash
go install
```

**Important:** Make sure `$GOPATH/bin` (usually `~/go/bin`) is in your `$PATH`.

To check:

```bash
echo $PATH | grep -q "$HOME/go/bin" && echo "✅ Already in PATH" || echo "❌ Not in PATH"
```

If not in PATH, add this to your `~/.zshrc` or `~/.bashrc`:

```bash
export PATH="$HOME/go/bin:$PATH"
```

Then reload your shell:

```bash
source ~/.zshrc  # or source ~/.bashrc
```

Now you can run workbuddy from any directory:

```bash
cd ~
workbuddy note list
```

#### Option 2: System-wide Installation

For a traditional system-wide installation:

```bash
# Build the binary
go build -o workbuddy

# Install to /usr/local/bin (requires sudo)
sudo mv workbuddy /usr/local/bin/

# Verify installation
which workbuddy
```

#### Option 3: Run from Project Directory (Development)

If you just want to test without installing:

```bash
go build
./workbuddy note list

# Or run directly without building
go run main.go note list
```

### First Run

On first run, workbuddy will automatically:

- Create the database directory at `~/.config/workbuddy/`
- Create the database file `workbuddy.db`
- Run all migrations to set up the schema

No manual setup required! 🎉

## Usage

### Create a note

```bash
workbuddy nt -t tag_name -c "Note content"
```

### Create a task (todo item)

```bash
workbuddy create -c "Buy milk" --task
```

With tags:

```bash
workbuddy nt -c "Send report" -t work,urgent --task
```

### Create a note with multiple tags

```bash
workbuddy nt -t tag1 -t tag2 -c "Note content"
```

or

```bash
workbuddy nt -t tag1,tag2 -c "Note content"
```

### List all notes

```bash
workbuddy list
```

Tasks show status indicators `[ ]` or `[x]` only when created with `--task`.

### List all tags

```bash
workbuddy tags
```

### Search notes

Search notes by tag (returns up to 5 results by default):

```bash
workbuddy search -t "work"
```

To return more results, use the `-l` flag:

```bash
workbuddy search -t "work" -l 10
```

Search all notes without filtering by tag:

```bash
workbuddy search
```

#### Filter by task status

Show only tasks (notes created with `--task`):

```bash
workbuddy search --tasks
workbuddy search -k
```

Show only **pending** (incomplete) tasks:

```bash
workbuddy search --pending
workbuddy search -p
```

Show only **completed** tasks:

```bash
workbuddy search --completed
workbuddy search -c
```

> **Note:** `--completed` and `--pending` are mutually exclusive — using both at the same time returns an error.

#### Combining flags

All flags can be combined. For example, show pending tasks tagged `work`, returning up to 10:

```bash
workbuddy search -t "work" --pending -l 10
```

Show all completed tasks across all tags:

```bash
workbuddy search --completed
```

### Mark task as completed

```bash
workbuddy check <id>
```

### Remove a note or task

```bash
workbuddy remove <id>
```

You will be asked for confirmation before deletion.

## Configuration

### Database Location

By default, the database is stored at:

- **Linux/macOS:** `~/.config/workbuddy/workbuddy.db`
- **Windows:** `%APPDATA%\workbuddy\workbuddy.db`

To use a custom location, set the `WORKBUDDY_DB` environment variable:

```bash
# Temporary (one-time use)
WORKBUDDY_DB="/path/to/my/notes.db" workbuddy note list

# Permanent (add to ~/.zshrc or ~/.bashrc)
export WORKBUDDY_DB="$HOME/Dropbox/notes/workbuddy.db"
```

## Technologies

- **Go 1.25.6** - Programming language
- **SQLite** - Embedded database (pure Go driver)
- **sqlc** - Type-safe SQL code generation
- **golang-migrate** - Database migrations
- **Cobra** - CLI framework
- **lipgloss** - Terminal styling

## Contributing

This is a learning project for understanding Go programming. Feel free to explore, fork, and experiment!
