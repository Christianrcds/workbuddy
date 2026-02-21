package note

import (
	"database/sql"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	migratesqlite "github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "modernc.org/sqlite"
)

// newTestDB creates a fresh in-memory SQLite database with all migrations applied.
// The database is automatically closed when the test ends via t.Cleanup.
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}

	driver, err := migratesqlite.WithInstance(db, &migratesqlite.Config{})
	if err != nil {
		t.Fatalf("failed to create migration driver: %v", err)
	}

	// os.DirFS("../..") points to the project root from internal/note/
	// "migrations" is the subdirectory within that root containing the SQL files
	sourceDriver, err := iofs.New(os.DirFS("../.."), "migrations")
	if err != nil {
		t.Fatalf("failed to create migration source: %v", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "sqlite", driver)
	if err != nil {
		t.Fatalf("failed to initialize migrator: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("migration failed: %v", err)
	}

	t.Cleanup(func() {
		m.Close()
		db.Close()
	})

	return db
}
