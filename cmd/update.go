package cmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"

	"workbuddy/internal/note"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update note content inline or in your $EDITOR",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil || id <= 0 {
			fmt.Fprintf(os.Stderr, "Invalid note id: %s\n", args[0])
			os.Exit(1)
		}

		content, _ := cmd.Flags().GetString("content")
		addTags, _ := cmd.Flags().GetStringSlice("add-tag")
		removeTags, _ := cmd.Flags().GetStringSlice("remove-tag")
		setTags, _ := cmd.Flags().GetStringSlice("set-tags")
		contentChanged := cmd.Flags().Changed("content")
		setTagsChanged := cmd.Flags().Changed("set-tags")

		db, err := openDatabase()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		ctx := context.Background()
		service := note.NewService(db)

		if !contentChanged {
			existingNote, err := service.GetNoteByID(ctx, id)
			if err != nil {
				switch {
				case errors.Is(err, sql.ErrNoRows):
					fmt.Fprintf(os.Stderr, "Note not found: %d\n", id)
				default:
					fmt.Fprintf(os.Stderr, "Error loading note: %v\n", err)
				}
				os.Exit(1)
			}

			content, err = openEditorForExistingContent(existingNote.Content)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error editing note: %v\n", err)
				os.Exit(1)
			}
		}

		params := note.UpdateParams{
			AddTags:    addTags,
			RemoveTags: removeTags,
		}
		params.Content = &content
		if setTagsChanged {
			params.SetTags = setTags
		}

		updatedNote, err := service.UpdateNote(ctx, id, params)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				fmt.Fprintf(os.Stderr, "Note not found: %d\n", id)
			default:
				fmt.Fprintf(os.Stderr, "Error updating note: %v\n", err)
			}
			os.Exit(1)
		}

		label := "note"
		if updatedNote.IsTask == 1 {
			label = "task"
		}
		fmt.Printf("Updated %s %d.\n", label, updatedNote.ID)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().StringP("content", "c", "", "Updated note content (opens $EDITOR when omitted)")
	updateCmd.Flags().StringSlice("add-tag", nil, "Add one or more tags")
	updateCmd.Flags().StringSlice("remove-tag", nil, "Remove one or more tags")
	updateCmd.Flags().StringSlice("set-tags", nil, "Replace all tags with the provided set")
}
