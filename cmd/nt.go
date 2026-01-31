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

// ntCmd represents the nt command
// nt stands for note and tag management
var ntCmd = &cobra.Command{
	Use:   "nt",
	Short: "Creates Note with Tags",
	Long:  `Create a new note with associated tags.`,
	Example: `
	workbuddy nt --content "Finish the report" --tags work,urgent
	workbuddy nt -c "Finish the report" -t "work, urgent"
	workbuddy nt -c "Read Go documentation" -t learning,go`,
	Run: func(cmd *cobra.Command, args []string) {
		content, _ := cmd.Flags().GetString("content")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		db, err := openDatabase()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		ctx := context.Background()
		service := note.NewService(db)

		note, err := service.CreateNoteWithTags(ctx, content, tags)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating note: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Created note with ID %d\n", note.ID)
	},
}

func init() {
	rootCmd.AddCommand(ntCmd)

	ntCmd.Flags().StringP("content", "c", "", "Note content (required)")
	ntCmd.Flags().StringSliceP("tags", "t", []string{}, "Tags (comma-separated)")

	ntCmd.MarkFlagRequired("content")
}
