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

// tagsCmd represents the tags command
var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "List all tags",
	Long:  `List all tags.`,
	Run: func(cmd *cobra.Command, args []string) {
		content, _ := cmd.Flags().GetString("name")
		pattern := "%" + content + "%"
		db, err := openDatabase()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		ctx := context.Background()

		repo := note.NewRepository(db)

		tags, err := repo.ListTags(ctx, pattern)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing tags: %v\n", err)
			os.Exit(1)
		}

		if len(tags) == 0 {
			fmt.Println("No tags found.")
			return
		}

		for _, tag := range tags {
			fmt.Printf("Tag: %s\n", tag.Name)
		}
	},
}

// Add functionality to search for tags by name (fuzzy search)
func init() {
	rootCmd.AddCommand(tagsCmd)
	tagsCmd.Flags().StringP("name", "n", "", "Tag name to search for (fuzzy search)")
}
