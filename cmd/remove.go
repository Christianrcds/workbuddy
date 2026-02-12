package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"workbuddy/internal/note"

	"github.com/spf13/cobra"
)

// removeCmd deletes a task by id.
var removeCmd = &cobra.Command{
	Use:     "remove <id>",
	Aliases: []string{"rm"},
	Short:   "Remove a task",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil || id <= 0 {
			fmt.Fprintf(os.Stderr, "Invalid task id: %s\n", args[0])
			os.Exit(1)
		}

		db, err := openDatabase()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		ctx := context.Background()
		repo := note.NewRepository(db)

		confirmed, err := confirmRemoval(id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading confirmation: %v\n", err)
			os.Exit(1)
		}
		if !confirmed {
			fmt.Println("Canceled.")
			return
		}

		rows, err := repo.DeleteNoteByID(ctx, id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error removing task: %v\n", err)
			os.Exit(1)
		}
		if rows == 0 {
			fmt.Fprintf(os.Stderr, "Note not found: %d\n", id)
			os.Exit(1)
		}

		fmt.Printf("Removed note %d.\n", id)
	},
}

func confirmRemoval(id int64) (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Delete note %d? (y/N): ", id)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes", nil
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
