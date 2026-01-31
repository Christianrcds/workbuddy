/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "workbuddy",
	Short: "A note-taking application with easy search and creation",
	Long:  ``,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return runMigrations()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// runMigrations applies all pending database migrations
func runMigrations() error {
	// Get database path from environment or use default
	dbPath := os.Getenv("WORKBUDDY_DB")
	if dbPath == "" {
		dbPath = "internal/database/workbuddy.db"
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

	// Get migrations path from environment or use default
	migrationsPath := os.Getenv("WORKBUDDY_MIGRATIONS")
	if migrationsPath == "" {
		migrationsPath = "file://migrations"
	}

	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "sqlite", driver)
	if err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}
	defer m.Close()

	// Run up migrations (ignore ErrNoChange which means migrations are already applied)
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.workbuddy.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
