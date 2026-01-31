/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"
	"workbuddy/internal/note"

	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		db, err := sql.Open("sqlite", "internal/database/workbuddy.db")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		repo := note.NewRepository(db)

		ctx := context.Background()

		notes, err := repo.ListNotes(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing notes: %v\n", err)
			os.Exit(1)
		}

		if len(notes) == 0 {
			fmt.Println("No notes found.")
			return
		}
		fmt.Printf("Found %d note(s):\n\n", len(notes))

		for _, note := range notes {
			fmt.Printf("Name: %s\n", note.Name)
			fmt.Printf("Content: %s\n", note.Content)
			fmt.Printf("Created: %s\n", note.CreatedAt.In(time.Local).Format(time.RFC1123))
			fmt.Println("---")
		}

	},
}

func init() {
	rootCmd.AddCommand(listCmd)

}
