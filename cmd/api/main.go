package main

import (
	"database/sql"
	"log"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "modernc.org/sqlite"
)

func main() {
	// Setup logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	// Open database
	db, err := sql.Open("sqlite", "internal/database/workbuddy.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Enable foreign keys for SQLite
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		log.Fatal("Failed to enable foreign keys:", err)
	}

	// Run migrations
	if err := runMigrations(db, logger); err != nil {
		log.Fatal("Migration failed:", err)
	}

	logger.Info("Database setup completed successfully!")
	logger.Info("Application ready to start...")
}

func runMigrations(db *sql.DB, logger *slog.Logger) error {
	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"sqlite", driver)
	if err != nil {
		return err
	}
	defer m.Close()

	logger.Info("Running database migrations...")
	
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	version, dirty, err := m.Version()
	if err != nil {
		return err
	}
	
	logger.Info("Migration completed", "version", version, "dirty", dirty)
	return nil
}