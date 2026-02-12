package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

var migrationsFS fs.FS

func SetMigrations(fsys fs.FS) {
	migrationsFS = fsys
}

var rootCmd = &cobra.Command{
	Use:   "workbuddy",
	Short: "A note-taking application with easy search and creation",
	Long:  ``,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return runMigrations()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func runMigrations() error {
	dbPath := getDBPath()
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "sqlite", driver)
	if err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}

	if err := ensureNoteColumns(db); err != nil {
		return fmt.Errorf("failed to ensure note columns: %w", err)
	}

	return nil
}

func ensureNoteColumns(db *sql.DB) error {
	ctx := context.Background()
	rows, err := db.QueryContext(ctx, "PRAGMA table_info(note);")
	if err != nil {
		return err
	}
	defer rows.Close()

	columns := map[string]bool{}
	for rows.Next() {
		var cid int
		var name string
		var colType string
		var notNull int
		var defaultVal sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &colType, &notNull, &defaultVal, &pk); err != nil {
			return err
		}
		columns[name] = true
	}
	if err := rows.Err(); err != nil {
		return err
	}

	if !columns["completed_at"] {
		if _, err := db.ExecContext(ctx, "ALTER TABLE note ADD COLUMN completed_at DATETIME;"); err != nil {
			return err
		}
	}
	if !columns["is_task"] {
		if _, err := db.ExecContext(ctx, "ALTER TABLE note ADD COLUMN is_task INTEGER NOT NULL DEFAULT 0;"); err != nil {
			return err
		}
	}

	return nil
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
