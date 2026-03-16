package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"workbuddy/internal/note"

	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check <id>",
	Short: "Mark a note as completed",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil || id <= 0 {
			fmt.Fprintf(os.Stderr, "Invalid note id: %s\n", args[0])
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

		_, err = service.CompleteTask(ctx, id)
		if err != nil {
			if errors.Is(err, note.ErrTaskNotCompletable()) {
				fmt.Fprintf(os.Stderr, "Task not found, already completed, or not a task: %d\n", id)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Error marking note as completed: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Marked task %d as completed.\n", id)
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
