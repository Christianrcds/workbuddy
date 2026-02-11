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
workbuddy note list
```

### List all tags

```bash
workbuddy tag list
```

### Filter notes by tag

```bash
workbuddy search -t "work"
```

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

## Development

### Database Migrations

Migrations are embedded in the binary and run automatically on startup. The migration files are located in the `migrations/` directory.

To create a new migration:

```bash
migrate create -ext sql -dir migrations -seq description_of_change
```

**Note:** Migrations are bundled into the binary using Go's `embed` package, so they don't need to be distributed separately.

### Project Structure

```
cmd/           → CLI commands (Cobra)
internal/note/ → Business logic and data access
  ├─ service.go    → Business logic layer
  ├─ repository.go → Data access layer
  └─ queries.sql   → SQL queries (for sqlc)
migrations/    → Database schema versions (embedded in binary)
main.go        → Application entry point with embedded migrations
```

## Technologies

- **Go 1.25.6** - Programming language
- **SQLite** - Embedded database (pure Go driver)
- **sqlc** - Type-safe SQL code generation
- **golang-migrate** - Database migrations
- **Cobra** - CLI framework
- **lipgloss** - Terminal styling

## Uninstalling

### If installed with `go install`:

```bash
rm $(go env GOPATH)/bin/workbuddy
```

### If installed to `/usr/local/bin`:

```bash
sudo rm /usr/local/bin/workbuddy
```

### Remove data (optional):

```bash
rm -rf ~/.config/workbuddy
```

## License

See LICENSE file for details.

## Contributing

This is a learning project for understanding Go programming. Feel free to explore, fork, and experiment!
