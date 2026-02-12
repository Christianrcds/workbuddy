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

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new note",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		content, _ := cmd.Flags().GetString("content")
		isTask, _ := cmd.Flags().GetBool("task")

		db, err := openDatabase()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()
		repo := note.NewRepository(db)
		ctx := context.Background()

		note, err := repo.CreateNote(ctx, note.CreateNoteParams{Content: content, IsTask: isTask})

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

	createCmd.Flags().StringP("content", "c", "", "Note content (required)")
	createCmd.Flags().BoolP("task", "k", false, "Create as a task (shows pending/checked in list)")
	createCmd.Flags().StringSliceP("tags", "t", []string{}, "Tags (comma-separated)")

	createCmd.MarkFlagRequired("content")

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
