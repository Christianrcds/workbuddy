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
		dueRaw, _ := cmd.Flags().GetString("due")
		recurRaw, _ := cmd.Flags().GetString("recur")
		weekdayRaw, _ := cmd.Flags().GetString("weekday")
		dayOfMonth, _ := cmd.Flags().GetInt("day-of-month")
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

		dueDate, err := parseDueDate(dueRaw)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating note: %v\n", err)
			os.Exit(1)
		}
		weekday, err := parseWeekday(weekdayRaw)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating note: %v\n", err)
			os.Exit(1)
		}

		var recurrenceRule note.RecurrenceRule
		if recurRaw != "" {
			recurrenceRule = note.RecurrenceRule(recurRaw)
		}
		var dayOfMonthPtr *int
		if cmd.Flags().Changed("day-of-month") {
			dayOfMonthPtr = &dayOfMonth
		}

		createdNote, err := service.CreateNoteWithParams(ctx, note.CreateParams{
			Content:              content,
			Tags:                 tags,
			IsTask:               isTask,
			DueDate:              dueDate,
			RecurrenceRule:       recurrenceRule,
			RecurrenceWeekday:    weekday,
			RecurrenceDayOfMonth: dayOfMonthPtr,
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating note: %v\n", err)
			os.Exit(1)
		}

		label := "note"
		if isTask {
			label = "task"
		}
		fmt.Printf("Created %s with ID %d\n", label, createdNote.ID)

	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringP("content", "c", "", "Note content (opens $EDITOR when omitted)")
	createCmd.Flags().BoolP("task", "k", false, "Create as a task (shows pending/checked in list)")
	createCmd.Flags().StringSliceP("tags", "t", []string{}, "Tags (comma-separated)")
	createCmd.Flags().String("due", "", "Due date in YYYY-MM-DD format")
	createCmd.Flags().String("recur", "", "Repeat cadence: daily, weekly, or monthly")
	createCmd.Flags().String("weekday", "", "Weekday for weekly recurrence: mon, tue, wed, thu, fri, sat, or sun")
	createCmd.Flags().Int("day-of-month", 0, "Day of month for monthly recurrence (1-31)")
}
