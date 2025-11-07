package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	var (
		title    string
		project  string
		phase    string
		priority int
		tags     []string
		comments []string
		dueDate  string
	)

	cmd := &cobra.Command{
		Use:   "update [task-id]",
		Short: "Update a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]

			// Check if at least one flag was provided
			if !cmd.Flags().Changed("title") &&
				!cmd.Flags().Changed("project") &&
				!cmd.Flags().Changed("phase") &&
				!cmd.Flags().Changed("priority") &&
				!cmd.Flags().Changed("tags") &&
				!cmd.Flags().Changed("comments") &&
				!cmd.Flags().Changed("due-date") {
				return fmt.Errorf("at least one field must be specified for update")
			}

			// Parse due date if provided
			var parsedDueDate *time.Time
			if dueDate != "" {
				// Try parsing different formats
				formats := []string{
					"2006-01-02 15:04:05",
					"2006-01-02 15:04",
					"2006-01-02",
					"06-01-02 15:04:05",
					"06-01-02 15:04",
					"06-01-02",
				}
				var parsed time.Time
				var err error
				for _, format := range formats {
					parsed, err = time.Parse(format, dueDate)
					if err == nil {
						parsedDueDate = &parsed
						break
					}
				}
				if err != nil {
					return fmt.Errorf("invalid due date format: %s", dueDate)
				}
			}

			err := App.TaskService.UpdateTask(
				context.Background(),
				taskID,
				title,
				project,
				phase,
				priority,
				tags,
				comments,
				parsedDueDate,
			)
			if err != nil {
				return fmt.Errorf("failed to update task: %w", err)
			}

			pterm.Success.Printf("Task %s updated successfully\n", taskID)
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "New task title")
	cmd.Flags().StringVar(&project, "project", "", "New project name")
	cmd.Flags().StringVar(&phase, "phase", "", "New phase/column name")
	cmd.Flags().IntVarP(&priority, "priority", "p", 0, "New task priority (1-5)")
	cmd.Flags().StringSliceVarP(&tags, "tags", "t", nil, "New task tags (space-separated)")
	cmd.Flags().StringSliceVarP(&comments, "comments", "c", nil, "New task comments")
	cmd.Flags().StringVar(&dueDate, "due-date", "", "Due date (format: YYYY-MM-DD or YYYY-MM-DD HH:MM:SS)")

	return cmd
}
