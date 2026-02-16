/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"workbuddy/internal/note"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all notes or tasks in a table format",
	Long: `
	The list command retrieves all notes and tasks from the database and displays them in a formatted table.`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := openDatabase()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		repo := note.NewRepository(db)
		ctx := context.Background()
		notesTypes := args[0]

		var notes []note.Note
		if len(notesTypes) == 0 {
			notes, err = repo.ListNotes(ctx)
		} else {
			notes, err = repo.ListTasks(ctx)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing notes: %v\n", err)
			os.Exit(1)
		}

		if len(notes) == 0 {
			fmt.Println("No notes found.")
			return
		}

		tagsByNoteID, err := buildTagsByNoteID(ctx, repo, notes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading tags: %v\n", err)
			os.Exit(1)
		}

		var title string
		if len(notesTypes) == 0 {
			title = fmt.Sprintf("Notes (%d)", len(notes))
		} else {
			title = fmt.Sprintf("Tasks (%d)", len(notes))
		}

		displayNotesAsTable(notes, tagsByNoteID, title)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
