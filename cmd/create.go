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

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new note or task",
	Long:  `Create a new note or task, optionally with tags.`,
	Run: func(cmd *cobra.Command, args []string) {
		content, _ := cmd.Flags().GetString("content")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		isTask, _ := cmd.Flags().GetBool("task")
		if content == "" {
			var err error
			content, err = openEditorForContent()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error receiving content: %v\n", err)
				os.Exit(1)
			}
		}

		db, err := openDatabase()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		ctx := context.Background()
		service := note.NewService(db)
		var isTaskInt int64
		if isTask {
			isTaskInt = 1
		}

		note, err := service.CreateNote(ctx, content, tags, isTaskInt)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating note: %v\n", err)
			os.Exit(1)
		}

		label := "note"
		if isTask {
			label = "task"
		}
		fmt.Printf("Created %s with ID %d\n", label, note.ID)

	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringP("content", "c", "", "Note content (opens $EDITOR when omitted)")
	createCmd.Flags().BoolP("task", "k", false, "Create as a task (shows pending/checked in list)")
	createCmd.Flags().StringSliceP("tags", "t", []string{}, "Tags (comma-separated)")
}
