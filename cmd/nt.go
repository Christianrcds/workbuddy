/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"workbuddy/internal/note"

	"github.com/spf13/cobra"
)

// ntCmd represents the nt command
// nt stands for note and tag management
var ntCmd = &cobra.Command{
	Use:   "nt",
	Short: "Creates Note with Tags",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("nt called")

		content, _ := cmd.Flags().GetString("content")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		db, err := sql.Open("sqlite", "internal/database/workbuddy.db")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		ctx := context.Background()

		trx, err := db.BeginTx(ctx, nil)
		if err != nil {
			os.Exit(1)
		}
		defer trx.Rollback()

		repo := note.NewRepositoryWithTx(trx)

		createdNote, err := repo.CreateNote(ctx, content)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating note: %v\n", err)
			os.Exit(1)
		}

		for _, tagName := range tags {
			tagName = strings.TrimSpace(tagName)

			tag, err := repo.CreateTag(ctx, tagName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating tag: %v\n", err)
				os.Exit(1)
			}

			err = repo.AddTagToNote(ctx, note.AddTagToNoteParams{
				NoteID: createdNote.ID,
				TagID:  tag.ID,
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error adding tag to note: %v\n", err)
				os.Exit(1)
			}
		}

		trx.Commit()

	},
}

func init() {
	rootCmd.AddCommand(ntCmd)

	ntCmd.Flags().StringP("content", "c", "", "Note content (required)")
	ntCmd.Flags().StringSliceP("tags", "t", []string{}, "Tags (comma-separated)")

	ntCmd.MarkFlagRequired("content")
}
