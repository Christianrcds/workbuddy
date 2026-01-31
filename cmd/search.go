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
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search notes by tag",
	Long: `Search for all notes that have a specific tag.

  Example:
    workbuddy search --tag golang
    workbuddy search -t learning`,
	Run: func(cmd *cobra.Command, args []string) {
		tag, _ := cmd.Flags().GetString("tag")
		fmt.Printf("Searching for tag: %s\n", tag)

		db, err := sql.Open("sqlite", "internal/database/workbuddy.db")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		ctx := context.Background()
		service := note.NewService(db)
		notes, err := service.GetNotesByTag(ctx, tag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error searching notes: %v\n", err)
			os.Exit(1)
		}
		if len(notes) == 0 {
			fmt.Println("No notes found.")
			return
		}
		fmt.Printf("Found %d note(s):\n\n", len(notes))

		for _, note := range notes {
			fmt.Printf("Note: %s\n", note.Content)
			fmt.Printf("Created: %s\n", note.CreatedAt.In(time.Local).Format(time.RFC1123))
			fmt.Println("---")
		}
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringP("tag", "t", "standup", "One tag to search for")
}
