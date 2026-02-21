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
		limit, _ := cmd.Flags().GetInt("limit")
		completed, _ := cmd.Flags().GetBool("completed")
		is_task, _ := cmd.Flags().GetBool("tasks")
		pending, _ := cmd.Flags().GetBool("pending")

		if completed && pending {
			fmt.Fprintln(os.Stderr, "Error: --completed and --pending are mutually exclusive")
			os.Exit(1)
		}
		db, err := openDatabase()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		ctx := context.Background()
		service := note.NewService(db)

		var completedFilter *bool
		if completed {
			t := true
			completedFilter = &t
		} else if pending {
			f := false
			completedFilter = &f
		}
		notes, err := service.SearchNotes(ctx, note.SearchParams{
			Tag:       tag,
			Limit:     int64(limit),
			TasksOnly: is_task,
			Completed: completedFilter,
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error searching notes: %v\n", err)
			os.Exit(1)
		}

		if len(notes) == 0 {
			fmt.Println("No notes found.")
			return
		}

		repo := note.NewRepository(db)
		tagsByNoteID, err := buildTagsByNoteID(ctx, repo, notes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading tags: %v\n", err)
			os.Exit(1)
		}

		displayNotes(notes, tagsByNoteID, fmt.Sprintf("🔍 Search Results: '%s' (%d found)", tag, len(notes)))
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringP("tag", "t", "", "Tag to search for")
	searchCmd.Flags().IntP("limit", "l", 5, "Maximum number of notes to return")
	searchCmd.Flags().BoolP("completed", "c", false, "Search completed notes")
	searchCmd.Flags().BoolP("tasks", "k", false, "Search tasks")
	searchCmd.Flags().BoolP("pending", "p", false, "Search pending notes")
}
