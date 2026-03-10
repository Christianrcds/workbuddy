package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"workbuddy/internal/note"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search notes by content, tag, and task state",
	Long: `Search notes by optional content query, tag, and task state.

  Example:
    workbuddy search "release notes"
    workbuddy search integrations -t work
    workbuddy search --tag learning --pending`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := ""
		if len(args) == 1 {
			query = args[0]
		}
		tag, _ := cmd.Flags().GetString("tag")
		limit, _ := cmd.Flags().GetInt("limit")
		completed, _ := cmd.Flags().GetBool("completed")
		isTask, _ := cmd.Flags().GetBool("tasks")
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
			Query:     query,
			Tag:       tag,
			Limit:     int64(limit),
			TasksOnly: isTask,
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

		displayNotes(notes, tagsByNoteID, searchResultsTitle(query, tag, len(notes)))
	},
}

func searchResultsTitle(query, tag string, count int) string {
	filters := make([]string, 0, 2)
	if strings.TrimSpace(query) != "" {
		filters = append(filters, fmt.Sprintf("query=%q", query))
	}
	if tag != "" {
		filters = append(filters, fmt.Sprintf("tag=%q", tag))
	}
	if len(filters) == 0 {
		return fmt.Sprintf("🔍 Search Results (%d found)", count)
	}
	return fmt.Sprintf("🔍 Search Results: %s (%d found)", strings.Join(filters, ", "), count)
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringP("tag", "t", "", "Tag to filter by")
	searchCmd.Flags().IntP("limit", "l", 5, "Maximum number of notes to return")
	searchCmd.Flags().BoolP("completed", "c", false, "Search completed notes")
	searchCmd.Flags().BoolP("tasks", "k", false, "Search tasks")
	searchCmd.Flags().BoolP("pending", "p", false, "Search pending tasks")
}
